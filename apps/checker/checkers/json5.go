package checkers

import (
	"errors"

	"github.com/tidwall/jsonc"
	"github.com/titanous/json5"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// JSON5 validates JSON5 syntax (comments, trailing commas, single quotes,
// unquoted keys, hex/Infinity/NaN numbers) using titanous/json5. The decoder
// reports a byte offset only, which is normalized to line/column.
type JSON5 struct{}

func (JSON5) Check(data []byte, strict bool) []result.SyntaxError {
	var v any
	if err := json5.Unmarshal(data, &v); err != nil {
		var se *json5.SyntaxError
		if errors.As(err, &se) {
			line, col := result.OffsetToLineCol(data, int(se.Offset))
			return []result.SyntaxError{{Line: line, Column: col, Message: cleanMessage(se.Error())}}
		}
		return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
	}
	return nil
}

// JSONC validates JSONC (JSON with comments and trailing commas) by stripping
// the comments/commas to plain JSON and delegating to the standard JSON
// validator. tidwall/jsonc blanks removed bytes with whitespace, so byte
// offsets — and therefore the reported line/column — are preserved.
type JSONC struct{}

func (JSONC) Check(data []byte, strict bool) []result.SyntaxError {
	return JSON{}.Check(jsonc.ToJSON(data), strict)
}
