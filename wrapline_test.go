package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runWrapline executes the wrapline program with given arguments and input
func runWrapline(t *testing.T, args []string, input string) (string, string, error) {
	cmd := exec.Command("./wrapline", args...)
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// TestBasicWrapping tests default double-quote wrapping
func TestBasicWrapping(t *testing.T) {
	input := "hello\nworld\n"
	expected := "\"hello\"\n\"world\"\n"

	stdout, stderr, err := runWrapline(t, []string{"-"}, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
	}

	if stdout != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, stdout)
	}
}

// TestCustomDelimiter tests custom delimiter option
func TestCustomDelimiter(t *testing.T) {
	tests := []struct {
		name      string
		delimiter string
		input     string
		expected  string
	}{
		{
			name:      "single quote",
			delimiter: "'",
			input:     "hello\n",
			expected:  "'hello'\n",
		},
		{
			name:      "pipe",
			delimiter: "|",
			input:     "test\n",
			expected:  "|test|\n",
		},
		{
			name:      "brackets",
			delimiter: "[]",
			input:     "data\n",
			expected:  "[]data[]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runWrapline(t, []string{"-d", tt.delimiter, "-"}, tt.input)

			if err != nil {
				t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
			}

			if stdout != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestHexDelimiter tests hexadecimal delimiter notation
func TestHexDelimiter(t *testing.T) {
	tests := []struct {
		name     string
		hexValue string
		input    string
		expected string
	}{
		{
			name:     "0x22 (double quote)",
			hexValue: "0x22",
			input:    "test\n",
			expected: "\"test\"\n",
		},
		{
			name:     "0x27 (single quote)",
			hexValue: "0x27",
			input:    "test\n",
			expected: "'test'\n",
		},
		{
			name:     "0x7C (pipe)",
			hexValue: "0x7C",
			input:    "test\n",
			expected: "|test|\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runWrapline(t, []string{"-d", tt.hexValue, "-"}, tt.input)

			if err != nil {
				t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
			}

			if stdout != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestStripWhitespace tests the -s flag
func TestStripWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "leading spaces",
			input:    "   hello\n",
			expected: "\"hello\"\n",
		},
		{
			name:     "trailing spaces",
			input:    "world   \n",
			expected: "\"world\"\n",
		},
		{
			name:     "both sides",
			input:    "  test  \n",
			expected: "\"test\"\n",
		},
		{
			name:     "tabs",
			input:    "\t\tdata\t\t\n",
			expected: "\"data\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runWrapline(t, []string{"-s", "-"}, tt.input)

			if err != nil {
				t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
			}

			if stdout != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestSkipEmptyLines tests the -e flag
func TestSkipEmptyLines(t *testing.T) {
	input := "hello\n\nworld\n\n"
	expected := "\"hello\"\n\"world\"\n"

	stdout, stderr, err := runWrapline(t, []string{"-e", "-"}, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
	}

	if stdout != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, stdout)
	}
}

// TestEmptyLastLine tests that empty last lines are always skipped
func TestEmptyLastLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line with trailing newline",
			input:    "hello\n",
			expected: "\"hello\"\n",
		},
		{
			name:     "multiple lines with empty last",
			input:    "hello\nworld\n\n",
			expected: "\"hello\"\n\"world\"\n",
		},
		{
			name:     "without -e flag, middle empty preserved",
			input:    "hello\n\nworld\n",
			expected: "\"hello\"\n\"\"\n\"world\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runWrapline(t, []string{"-"}, tt.input)

			if err != nil {
				t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
			}

			if stdout != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, stdout)
			}
		})
	}
}

// TestEscapeDelimiter tests the -escape flag
func TestEscapeDelimiter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single delimiter in content",
			input:    "She said \"hello\"\n",
			expected: "\"She said \\\"hello\\\"\"\n",
		},
		{
			name:     "multiple delimiters",
			input:    "\"quoted\" and \"more\"\n",
			expected: "\"\\\"quoted\\\" and \\\"more\\\"\"\n",
		},
		{
			name:     "no delimiters",
			input:    "plain text\n",
			expected: "\"plain text\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runWrapline(t, []string{"-escape", "-"}, tt.input)

			if err != nil {
				t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
			}

			if stdout != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, stdout)
			}
		})
	}
}

// TestNullTerminated tests the -0 flag
func TestNullTerminated(t *testing.T) {
	// Create input with null terminators
	input := "hello\x00world\x00test\x00"
	expected := "\"hello\"\n\"world\"\n\"test\"\n"

	stdout, stderr, err := runWrapline(t, []string{"-0", "-"}, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
	}

	if stdout != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, stdout)
	}
}

// TestOutputFile tests the -o flag
func TestOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")

	input := "hello\nworld\n"
	expected := "\"hello\"\n\"world\"\n"

	_, stderr, err := runWrapline(t, []string{"-o", outputFile, "-"}, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, string(content))
	}
}

// TestFileInput tests reading from a file instead of STDIN
func TestFileInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")

	input := "hello\nworld\n"
	expected := "\"hello\"\n\"world\"\n"

	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	stdout, stderr, err := runWrapline(t, []string{inputFile}, "")

	if err != nil {
		t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
	}

	if stdout != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, stdout)
	}
}

// TestCombinations tests common flag combinations
func TestCombinations(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		input    string
		expected string
	}{
		{
			name:     "strip and skip empty",
			args:     []string{"-s", "-e", "-"},
			input:    "  hello  \n\n  world  \n",
			expected: "\"hello\"\n\"world\"\n",
		},
		{
			name:     "custom delimiter with escape",
			args:     []string{"-d", "'", "-escape", "-"},
			input:    "it's working\n",
			expected: "'it\\'s working'\n",
		},
		{
			name:     "strip, skip empty, and escape",
			args:     []string{"-s", "-e", "-escape", "-"},
			input:    "  \"test\"  \n\n  data  \n",
			expected: "\"\\\"test\\\"\"\n\"data\"\n",
		},
		{
			name:     "null-terminated with custom delimiter",
			args:     []string{"-0", "-d", "|", "-"},
			input:    "one\x00two\x00three\x00",
			expected: "|one|\n|two|\n|three|\n",
		},
		{
			name:     "all options combined",
			args:     []string{"-s", "-e", "-escape", "-d", "'", "-"},
			input:    "  it's  \n\n  'quoted'  \n",
			expected: "'it\\'s'\n'\\'quoted\\''\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runWrapline(t, tt.args, tt.input)

			if err != nil {
				t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
			}

			if stdout != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, stdout)
			}
		})
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			args:     []string{"-"},
			input:    "",
			expected: "",
		},
		{
			name:     "single empty line",
			args:     []string{"-"},
			input:    "\n",
			expected: "",
		},
		{
			name:     "only whitespace with strip",
			args:     []string{"-s", "-"},
			input:    "   \n",
			expected: "",
		},
		{
			name:     "very long line",
			args:     []string{"-"},
			input:    strings.Repeat("a", 10000) + "\n",
			expected: "\"" + strings.Repeat("a", 10000) + "\"\n",
		},
		{
			name:     "no trailing newline",
			args:     []string{"-"},
			input:    "hello",
			expected: "\"hello\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runWrapline(t, tt.args, tt.input)

			if err != nil {
				t.Fatalf("Expected no error, got: %v\nStderr: %s", err, stderr)
			}

			if stdout != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, stdout)
			}
		})
	}
}

// TestVersion tests the -v flag
func TestVersion(t *testing.T) {
	cmd := exec.Command("./wrapline", "-v")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, pgmName) {
		t.Errorf("Expected version output to contain program name, got: %q", output)
	}
	if !strings.Contains(output, pgmVersion) {
		t.Errorf("Expected version output to contain version, got: %q", output)
	}
	if !strings.Contains(output, pgmURL) {
		t.Errorf("Expected version output to contain URL, got: %q", output)
	}
}

// TestErrorCases tests error conditions
func TestErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		input       string
		expectError bool
	}{
		{
			name:        "no filename argument",
			args:        []string{},
			input:       "",
			expectError: true,
		},
		{
			name:        "too many arguments",
			args:        []string{"file1.txt", "file2.txt"},
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid hex delimiter",
			args:        []string{"-d", "0xZZ", "-"},
			input:       "test\n",
			expectError: true,
		},
		{
			name:        "nonexistent input file",
			args:        []string{"/nonexistent/file.txt"},
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, err := runWrapline(t, tt.args, tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got none. Stderr: %s", stderr)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v. Stderr: %s", err, stderr)
			}
		})
	}
}

// TestParseDelimiter tests the parseDelimiter function directly
func TestParseDelimiter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "literal string",
			input:       "test",
			expected:    "test",
			expectError: false,
		},
		{
			name:        "hex double quote",
			input:       "0x22",
			expected:    "\"",
			expectError: false,
		},
		{
			name:        "hex single quote",
			input:       "0x27",
			expected:    "'",
			expectError: false,
		},
		{
			name:        "invalid hex",
			input:       "0xGG",
			expected:    "",
			expectError: true,
		},
		{
			name:        "out of range",
			input:       "0x110000",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDelimiter(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
