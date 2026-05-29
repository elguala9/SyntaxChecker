package checkers

import (
	"bytes"
	"errors"

	luaparse "github.com/yuin/gopher-lua/parse"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Lua validates the syntax of a Lua 5.1 script using gopher-lua's parser. Only
// parsing is performed; the chunk is not compiled to bytecode nor executed.
type Lua struct{}

func (Lua) Check(data []byte, strict bool) []result.SyntaxError {
	_, err := luaparse.Parse(bytes.NewReader(data), "input.lua")
	if err == nil {
		return nil
	}

	var pe *luaparse.Error
	if errors.As(err, &pe) {
		msg := pe.Message
		if pe.Token != "" {
			msg += " near '" + pe.Token + "'"
		}
		se := result.SyntaxError{Message: cleanMessage(msg)}
		// gopher-lua uses a sentinel position (-1) for end-of-file errors;
		// only surface line/column when they are real (1-based).
		if pe.Pos.Line > 0 {
			se.Line = pe.Pos.Line
		}
		if pe.Pos.Column > 0 {
			se.Column = pe.Pos.Column
		}
		return []result.SyntaxError{se}
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
