package checkers

import (
	"bytes"
	"errors"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Dockerfile validates a Dockerfile using buildkit's frontend parser. It runs
// two passes: the structural parse (line continuations, heredocs, the escape
// directive) and a per-instruction parse that rejects unknown instructions and
// malformed arguments. Image/stage build semantics (e.g. requiring a leading
// FROM, resolving base images) are intentionally not enforced.
type Dockerfile struct{}

func (Dockerfile) Check(data []byte, strict bool) []result.SyntaxError {
	res, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return []result.SyntaxError{dockerfileError(err)}
	}

	var errs []result.SyntaxError
	for _, child := range res.AST.Children {
		if _, err := instructions.ParseInstruction(child); err != nil {
			errs = append(errs, dockerfileError(err))
		}
	}
	return errs
}

// dockerfileError maps a buildkit parse error to a SyntaxError, extracting the
// source line from the *parser.LocationError the parser attaches when known.
func dockerfileError(err error) result.SyntaxError {
	se := result.SyntaxError{Message: cleanMessage(err.Error())}
	var le *parser.LocationError
	if errors.As(err, &le) && len(le.Locations) > 0 && len(le.Locations[0]) > 0 {
		se.Line = le.Locations[0][0].Start.Line
	}
	return se
}
