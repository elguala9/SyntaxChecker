package checkers

import (
	"errors"

	"go.starlark.net/syntax"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Starlark validates the syntax of a Starlark file (Bazel BUILD/.bzl, .star)
// using go.starlark.net's parser. Only parsing is performed; the file is not
// executed.
type Starlark struct{}

func (Starlark) Check(data []byte, strict bool) []result.SyntaxError {
	_, err := syntax.Parse("input.star", data, 0)
	if err == nil {
		return nil
	}

	var se syntax.Error
	if errors.As(err, &se) {
		return []result.SyntaxError{{
			Line:    int(se.Pos.Line),
			Column:  int(se.Pos.Col),
			Message: cleanMessage(se.Msg),
		}}
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
