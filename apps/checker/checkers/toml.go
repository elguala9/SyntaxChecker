package checkers

import (
	"errors"

	"github.com/pelletier/go-toml/v2"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// TOML validates TOML syntax using pelletier/go-toml/v2. The decoder reports a
// precise row/column for the first error it encounters (duplicate keys,
// malformed values, bad table headers, etc.).
type TOML struct{}

func (TOML) Check(data []byte, strict bool) []result.SyntaxError {
	var v map[string]any
	err := toml.Unmarshal(data, &v)
	if err == nil {
		return nil
	}

	var de *toml.DecodeError
	if errors.As(err, &de) {
		row, col := de.Position()
		return []result.SyntaxError{{Line: row, Column: col, Message: cleanMessage(de.Error())}}
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
