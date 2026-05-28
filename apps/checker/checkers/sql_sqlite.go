package checkers

import (
	"errors"
	"strings"

	rqlite "github.com/rqlite/sql"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// SQLite validates SQLite SQL syntax using rqlite/sql, a pure-Go parser used in
// production by rqlite. It parses every statement in the input and reports the
// first syntax error, with its native line/column position.
type SQLite struct{}

func (SQLite) Check(data []byte, strict bool) []result.SyntaxError {
	p := rqlite.NewParser(strings.NewReader(string(data)))
	_, err := p.ParseStatements()
	if err == nil {
		return nil
	}
	se := result.SyntaxError{Message: cleanMessage(err.Error())}
	var pe *rqlite.Error
	if errors.As(err, &pe) {
		se.Line = pe.Pos.Line
		se.Column = pe.Pos.Column
	}
	return []result.SyntaxError{se}
}
