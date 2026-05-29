package checkers

import (
	"errors"
	"go/parser"
	"go/scanner"
	"go/token"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Go validates the syntax of a Go source file using the standard library's
// go/parser. Only parsing is performed: imports are not resolved, types are not
// checked, and nothing is built. A valid file must begin with a package clause.
type Go struct{}

func (Go) Check(data []byte, strict bool) []result.SyntaxError {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "input.go", data,
		parser.AllErrors|parser.ParseComments|parser.SkipObjectResolution)
	if err == nil {
		return nil
	}

	// go/parser returns a scanner.ErrorList carrying every syntax error, each
	// with a 1-based line/column position.
	var list scanner.ErrorList
	if errors.As(err, &list) {
		out := make([]result.SyntaxError, 0, len(list))
		for _, e := range list {
			out = append(out, result.SyntaxError{
				Line:    e.Pos.Line,
				Column:  e.Pos.Column,
				Message: cleanMessage(e.Msg),
			})
		}
		return out
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
