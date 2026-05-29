package checkers

import (
	"github.com/joho/godotenv"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Env validates .env files using github.com/joho/godotenv. The parser reports
// the first malformed assignment (e.g. a line that is neither a comment nor a
// KEY=VALUE pair). godotenv does not expose a line number, so the message is
// surfaced as-is.
type Env struct{}

func (Env) Check(data []byte, strict bool) []result.SyntaxError {
	_, err := godotenv.UnmarshalBytes(data)
	if err == nil {
		return nil
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
