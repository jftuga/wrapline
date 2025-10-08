# wrapline

A command-line tool that wraps each line of input with a specified delimiter.
Useful for formatting text data, preparing strings for code, or processing structured data.

## Features

- Wrap lines with any delimiter (default: double-quote)
- Support for hexadecimal delimiter notation
- Strip whitespace before wrapping
- Skip empty lines
- Escape delimiter characters within lines
- Read from files or STDIN
- Write to files or STDOUT
- Process null-terminated input (compatible with `find -print0`, `xargs -0`)
- Automatically skip empty last lines

## Installation

### Build from source

```bash
git clone https://github.com/jftuga/wrapline
cd wrapline
make
```

This creates the `wrapline` executable in the current directory.

## Usage

```
wrapline [options] <filename|-]
```

### Options

- `-d <delimiter>` - Delimiter to wrap lines with (default: `"`)
  - Supports literal strings: `-d "|"`
  - Supports hex notation: `-d 0x27` for single quote
- `-s` - Strip whitespace from lines before wrapping
- `-e` - Do not emit empty lines
- `-escape` - Escape delimiter characters within lines using backslash
- `-o <file>` - Write output to file instead of STDOUT
- `-0` - Read null-terminated records instead of newlines
- `-v` - Show version and exit

### Input

- Provide a filename to read from a file
- Use `-` to read from STDIN

## Examples

### Basic usage

Wrap lines with double quotes (default):

```bash
wrapline input.txt
```

**Input:**
```
hello
world
```

**Output:**
```
"hello"
"world"
```

### Custom delimiter

Wrap lines with single quotes:

```bash
wrapline -d "'" input.txt
```

**Output:**
```
'hello'
'world'
```

### Hexadecimal delimiters

Use hexadecimal notation for delimiters (useful for special characters):

```bash
# Pipe character (0x7C = ASCII 124)
wrapline -d 0x7C input.txt

# Tab character (0x09)
wrapline -d 0x09 input.txt
```

### Strip whitespace

Remove leading and trailing whitespace before wrapping:

```bash
wrapline -s input.txt
```

**Input:**
```
  hello
   world
```

**Output:**
```
"hello"
"world"
```

### Skip empty lines

Don't output empty lines:

```bash
wrapline -e input.txt
```

**Input:**
```
hello

world
```

**Output:**
```
"hello"
"world"
```

### Escape delimiters

Escape delimiter characters found within lines:

```bash
wrapline -escape input.txt
```

**Input:**
```
She said "hello" to me
```

**Output:**
```
"She said \"hello\" to me"
```

### Output to file

Write results to a file instead of STDOUT:

```bash
wrapline -d "|" input.txt -o output.txt
```

### Read from STDIN

Process piped input:

```bash
echo "hello world" | wrapline -
cat input.txt | wrapline -d "'" -s -
```

### Null-terminated input

Process null-terminated records (like `find -print0`):

```bash
find . -name "*.txt" -print0 | wrapline -0 -
```

This is useful when filenames may contain newlines or special characters.

### Combining options

Combine multiple options for complex processing:

```bash
# Strip whitespace, skip empty lines, escape quotes, write to file
wrapline -s -e -escape -o output.txt input.txt

# Process null-terminated, custom delimiter, output to file
find . -type f -print0 | wrapline -0 -d "'" -o filelist.txt -
```

## Common Use Cases

### Prepare strings for code

Convert a list of values into quoted strings for code:

```bash
wrapline cities.txt
```

**Input:**
```
New York
Los Angeles
Chicago
```

**Output:**
```
"New York"
"Los Angeles"
"Chicago"
```

### Create CSV-ready data

Wrap fields with quotes:

```bash
wrapline -escape data.txt
```

### Format file lists

Process files found by `find`:

```bash
find /var/log -name "*.log" -print0 | wrapline -0 -d "'" -
```

### Clean up text data

Strip whitespace and remove empty lines:

```bash
wrapline -s -e messy_data.txt -o clean_data.txt
```

## Notes

- Empty lines at the end of input are always skipped, regardless of flags
- When using `-e`, all empty lines are skipped
- The `-s` flag strips whitespace before checking if a line is empty
- Hexadecimal delimiter values must be prefixed with `0x`
- Without the `0x` prefix, numeric strings are treated as literal delimiters
