package checkers

import (
	"bytes"
	"errors"

	"mvdan.cc/sh/v3/syntax"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Shell validates the syntax of a POSIX/Bash shell script using mvdan.cc/sh.
// The parser runs in Bash language mode (a superset of POSIX sh). Only parsing
// is performed; the script is never executed.
type Shell struct{}

func (Shell) Check(data []byte, strict bool) []result.SyntaxError {
	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
	_, err := parser.Parse(bytes.NewReader(data), "input.sh")
	if err == nil {
		return nil
	}

	var pe syntax.ParseError
	if errors.As(err, &pe) {
		return []result.SyntaxError{{
			Line:    int(pe.Pos.Line()),
			Column:  int(pe.Pos.Col()),
			Message: cleanMessage(pe.Text),
		}}
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
