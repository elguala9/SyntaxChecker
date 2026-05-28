package checkers

import (
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// YAML validates YAML syntax using gopkg.in/yaml.v3. Decoding into a generic
// value also surfaces duplicate map keys, which yaml.v3 treats as an error.
type YAML struct{}

// yaml.v3 embeds the line number in the message, e.g. "yaml: line 4: ...".
var yamlLineRE = regexp.MustCompile(`(?:^|\s)line (\d+):`)

func (YAML) Check(data []byte, strict bool) []result.SyntaxError {
	var v any
	err := yaml.Unmarshal(data, &v)
	if err == nil {
		return nil
	}

	// A type error carries one or more messages; a plain error carries one.
	if te, ok := err.(*yaml.TypeError); ok {
		errs := make([]result.SyntaxError, 0, len(te.Errors))
		for _, msg := range te.Errors {
			errs = append(errs, yamlSyntaxError(msg))
		}
		return errs
	}
	return []result.SyntaxError{yamlSyntaxError(err.Error())}
}

func yamlSyntaxError(msg string) result.SyntaxError {
	line := 0
	if m := yamlLineRE.FindStringSubmatch(msg); m != nil {
		line, _ = strconv.Atoi(m[1])
	}
	return result.SyntaxError{Line: line, Message: msg}
}
