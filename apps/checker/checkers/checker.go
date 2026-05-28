// Package checkers contains one syntax validator per supported file type.
package checkers

import (
	"strings"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Validator validates the syntax of a single file's contents. Each
// implementation is responsible for mapping its parser's native errors into the
// shared result.SyntaxError shape (including line/column normalization).
type Validator interface {
	// Check validates data and returns the found errors. The returned slice is
	// empty when the input is valid.
	Check(data []byte, strict bool) []result.SyntaxError
}

// maxMessageLen caps a parser message length (in runes).
const maxMessageLen = 200

// cleanMessage collapses whitespace into single spaces and truncates the
// result. SQL parsers (e.g. TiDB) embed the remaining source after the error
// point, which can be huge and span multiple lines.
func cleanMessage(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > maxMessageLen {
		return string(r[:maxMessageLen]) + "…"
	}
	return s
}
