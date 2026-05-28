package checkers

import (
	pg_query "github.com/wasilibs/go-pgquery"
	pgparser "github.com/wasilibs/go-pgquery/parser"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Postgres validates PostgreSQL syntax using go-pgquery, the real PostgreSQL
// parser compiled to WASM (zero CGO). Also used as the backend for sql:ansi.
type Postgres struct{}

func (Postgres) Check(data []byte, strict bool) []result.SyntaxError {
	_, err := pg_query.Parse(string(data))
	if err == nil {
		return nil
	}
	se := result.SyntaxError{Message: cleanMessage(err.Error())}
	// pgparser.Error carries Cursorpos, a 1-based character position into the
	// query. Treating it as a byte offset is exact for ASCII SQL.
	if pe, ok := err.(*pgparser.Error); ok && pe.Cursorpos > 0 {
		se.Line, se.Column = result.OffsetToLineCol(data, pe.Cursorpos-1)
	}
	return []result.SyntaxError{se}
}
