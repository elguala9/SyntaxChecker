package checkers

import (
	"errors"

	"github.com/itchyny/gojq"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// JQ validates the syntax of a jq program (a single .jq expression file) using
// gojq's parser. Only parsing is performed: the query is not compiled against an
// input nor are its referenced functions resolved.
type JQ struct{}

func (JQ) Check(data []byte, strict bool) []result.SyntaxError {
	_, err := gojq.Parse(string(data))
	if err == nil {
		return nil
	}

	// gojq reports a *ParseError with the byte offset reached when the error
	// occurred, which we normalize to a line/column.
	var pe *gojq.ParseError
	if errors.As(err, &pe) {
		line, col := result.OffsetToLineCol(data, pe.Offset)
		return []result.SyntaxError{{Line: line, Column: col, Message: cleanMessage(pe.Error())}}
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
