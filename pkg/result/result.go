// Package result defines the shared types produced by every checker.
package result

import "unicode/utf8"

// CheckResult is the outcome of validating a single file.
type CheckResult struct {
	File   string        `json:"file"`
	Type   string        `json:"type"`
	Valid  bool          `json:"valid"`
	Errors []SyntaxError `json:"errors,omitempty"`
}

// SyntaxError is a single problem found in a file. Line and Column are 1-based;
// they are omitted when the underlying parser cannot provide them.
type SyntaxError struct {
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
}

// OffsetToLineCol converts a 0-based byte offset into a 1-based line and column.
// Several parsers (encoding/json, go-pgquery) only report a byte offset; this
// helper normalizes them to the line/column expected by SyntaxError. The column
// counts runes, not bytes, so multibyte UTF-8 characters count as one column.
func OffsetToLineCol(data []byte, offset int) (line, col int) {
	if offset > len(data) {
		offset = len(data)
	}
	line, col = 1, 1
	for i := 0; i < offset; {
		if data[i] == '\n' {
			line++
			col = 1
			i++
			continue
		}
		_, size := utf8.DecodeRune(data[i:])
		if size < 1 {
			size = 1
		}
		col++
		i += size
	}
	return line, col
}
