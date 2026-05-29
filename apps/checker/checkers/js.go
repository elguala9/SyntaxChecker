package checkers

import (
	"github.com/evanw/esbuild/pkg/api"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// JS validates the syntax of a file in the JavaScript/TypeScript family using
// esbuild's parser. Dialect selects how the source is parsed:
//
//	"ts"  (default) TypeScript
//	"tsx"           TypeScript + JSX
//	"js"            JavaScript
//	"jsx"           JavaScript + JSX
//
// Only parsing and esbuild's early syntactic checks run: modules are not
// resolved and no type checking is performed.
type JS struct {
	Dialect string
}

func (j JS) Check(data []byte, strict bool) []result.SyntaxError {
	var loader api.Loader
	switch j.Dialect {
	case "tsx":
		loader = api.LoaderTSX
	case "js":
		loader = api.LoaderJS
	case "jsx":
		loader = api.LoaderJSX
	default: // "ts" or empty
		loader = api.LoaderTS
	}

	res := api.Transform(string(data), api.TransformOptions{
		Loader:     loader,
		Sourcefile: "input",
	})
	if len(res.Errors) == 0 {
		return nil
	}

	out := make([]result.SyntaxError, 0, len(res.Errors))
	for _, m := range res.Errors {
		se := result.SyntaxError{Message: cleanMessage(m.Text)}
		if m.Location != nil {
			se.Line = m.Location.Line
			se.Column = m.Location.Column + 1 // esbuild columns are 0-based bytes
		}
		out = append(out, se)
	}
	return out
}
