/*
wrapline.go
2025-10-08

A CLI tool that reads a file or STDIN line-by-line (or null-terminated) and wraps each line
with a specified delimiter. Supports whitespace stripping, empty line filtering, delimiter
escaping, and output redirection. Optimized for high performance with large files using
buffered I/O and memory-efficient streaming with lookahead.

DISCLAIMER
==========
Code generated with assistance from Claude (Anthropic AI) - October 2025
The code has been tested and validated, but users should review and test
it for their specific use cases before production use.
*/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

const pgmName string = "wrapline"
const pgmURL string = "https://github.com/jftuga/wrapline"
const pgmVersion string = "1.1.6"

// parseDelimiter converts a delimiter argument to a string.
// If the argument starts with "0x", it is interpreted as a hexadecimal
// value and converted to the corresponding character.
func parseDelimiter(arg string) (string, error) {
	if strings.HasPrefix(arg, "0x") {
		// Parse as hexadecimal
		value, err := strconv.ParseInt(arg[2:], 16, 32)
		if err != nil {
			return "", fmt.Errorf("invalid hex value '%s': %w", arg, err)
		}
		if value < 0 || value > 0x10FFFF {
			return "", fmt.Errorf("hex value '%s' out of valid Unicode range", arg)
		}
		return string(rune(value)), nil
	}
	// Use as literal string
	return arg, nil
}

// processLine builds a complete output line with delimiters and writes it in a single operation.
// Uses the provided buffer to avoid allocations. Optionally escapes delimiter characters within the line.
func processLine(writer *bufio.Writer, line []byte, delimiter string, escapeDelim bool, outputBuf *[]byte) error {
	// Reset the buffer for reuse
	*outputBuf = (*outputBuf)[:0]

	// Add opening delimiter
	*outputBuf = append(*outputBuf, delimiter...)

	// Add line content (escaped if needed)
	if escapeDelim && len(delimiter) > 0 {
		lineStr := string(line)
		escapedLine := strings.ReplaceAll(lineStr, delimiter, "\\"+delimiter)
		*outputBuf = append(*outputBuf, escapedLine...)
	} else {
		*outputBuf = append(*outputBuf, line...)
	}

	// Add closing delimiter and newline
	*outputBuf = append(*outputBuf, delimiter...)
	*outputBuf = append(*outputBuf, '\n')

	// Single write operation
	_, err := writer.Write(*outputBuf)
	return err
}

func main() {
	// Define command-line flags
	showVersion := flag.Bool("v", false, "show version and exit")
	delimiterArg := flag.String("d", "\"", "delimiter to wrap lines with (or hex value with 0x prefix)")
	stripWS := flag.Bool("s", false, "strip whitespace from lines before wrapping")
	skipEmpty := flag.Bool("e", false, "do not emit empty lines")
	escapeDelim := flag.Bool("escape", false, "escape delimiter characters within lines")
	outputFile := flag.String("o", "", "output file (default: STDOUT)")
	nullTerminated := flag.Bool("0", false, "read null-terminated records instead of newlines")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("%s v%s\n%s\n", pgmName, pgmVersion, pgmURL)
		os.Exit(0)
	}

	// Parse delimiter (handle hex notation)
	delimiter, err := parseDelimiter(*delimiterArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid delimiter: %v\n", err)
		os.Exit(1)
	}

	// Get filename from remaining arguments
	args := flag.Args()
	// Determine whether stdin is a terminal
	inputIsTerminal := term.IsTerminal(int(os.Stdin.Fd()))

	var filename string
	switch {
	case len(args) == 1:
		// User explicitly provided a filename or "-"
		filename = args[0]
	case len(args) == 0 && !inputIsTerminal:
		// No filename, but data is being piped in
		filename = "-"
	default:
		// Anything else is an error
		fmt.Fprintln(os.Stderr, "Error: exactly one filename (or '-' for STDIN) required")
		os.Exit(1)
	}

	// Open input source
	var input io.Reader
	if filename == "-" {
		input = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open file '%s': %v\n", filename, err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	}

	// Set up output destination
	var output io.Writer = os.Stdout
	if *outputFile != "" {
		outFile, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create output file '%s': %v\n", *outputFile, err)
			os.Exit(1)
		}
		defer outFile.Close()
		output = outFile
	}

	// Create buffered reader and writer for optimal I/O performance
	reader := bufio.NewReader(input)
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	// Create reusable output buffer to avoid allocations per line
	outputBuf := make([]byte, 0, 1024)

	// Determine delimiter byte for reading
	var delimByte byte = '\n'
	if *nullTerminated {
		delimByte = 0
	}

	// Read and process with one-line lookahead to detect last line
	var bufferedLine []byte
	var hasBufferedLine bool

	for {
		line, err := reader.ReadBytes(delimByte)
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error: failed to read input: %v\n", err)
			os.Exit(1)
		}

		// Remove the delimiter from the end
		if len(line) > 0 && line[len(line)-1] == delimByte {
			line = line[:len(line)-1]
		}

		// Check if we've reached EOF
		if err == io.EOF {
			// If we have a buffered line, it's the last line
			if hasBufferedLine {
				processedLine := bufferedLine
				if *stripWS {
					processedLine = bytes.TrimSpace(processedLine)
				}
				// Always skip empty last lines
				if len(processedLine) > 0 {
					if err := processLine(writer, processedLine, delimiter, *escapeDelim, &outputBuf); err != nil {
						fmt.Fprintf(os.Stderr, "Error: failed to write output: %v\n", err)
						os.Exit(1)
					}
				}
			}
			// If line has data (no trailing delimiter case), it's also the last line
			if len(line) > 0 {
				processedLine := line
				if *stripWS {
					processedLine = bytes.TrimSpace(processedLine)
				}
				// Always skip empty last lines
				if len(processedLine) > 0 {
					if err := processLine(writer, processedLine, delimiter, *escapeDelim, &outputBuf); err != nil {
						fmt.Fprintf(os.Stderr, "Error: failed to write output: %v\n", err)
						os.Exit(1)
					}
				}
			}
			break
		}

		// If we have a buffered line, process it now (we know it's not the last line)
		if hasBufferedLine {
			processedLine := bufferedLine

			// Strip whitespace if requested
			if *stripWS {
				processedLine = bytes.TrimSpace(processedLine)
			}

			// Output line unless it's empty and we're skipping empty lines
			if len(processedLine) > 0 || !*skipEmpty {
				if err := processLine(writer, processedLine, delimiter, *escapeDelim, &outputBuf); err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to write output: %v\n", err)
					os.Exit(1)
				}
			}
		}

		// Buffer current line for next iteration
		bufferedLine = line
		hasBufferedLine = true
	}
}
