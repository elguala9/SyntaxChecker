package checkers

import (
	"fmt"
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

// CheckSchema validates YAML well-formedness and then validates the document
// against the given JSON Schema. YAML maps onto the same data model as JSON, so
// the decoded value is normalized (see normalizeYAML) and run through the
// shared JSON Schema engine.
func (YAML) CheckSchema(data, schema []byte, strict bool) []result.SyntaxError {
	var v any
	if err := yaml.Unmarshal(data, &v); err != nil {
		if te, ok := err.(*yaml.TypeError); ok {
			errs := make([]result.SyntaxError, 0, len(te.Errors))
			for _, msg := range te.Errors {
				errs = append(errs, yamlSyntaxError(msg))
			}
			return errs
		}
		return []result.SyntaxError{yamlSyntaxError(err.Error())}
	}

	sch, err := compileSchema(schema)
	if err != nil {
		return []result.SyntaxError{{Message: err.Error()}}
	}
	return validateAgainstSchema(sch, normalizeYAML(v))
}

// normalizeYAML rewrites the value decoded by yaml.v3 into the model expected by
// the JSON Schema engine. yaml.v3 may decode a mapping with non-string keys as
// map[any]any; the engine requires map[string]any. Keys are stringified with
// the default %v formatting, matching how YAML renders scalar keys.
func normalizeYAML(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, e := range val {
			val[k] = normalizeYAML(e)
		}
		return val
	case map[any]any:
		m := make(map[string]any, len(val))
		for k, e := range val {
			m[fmt.Sprintf("%v", k)] = normalizeYAML(e)
		}
		return m
	case []any:
		for i, e := range val {
			val[i] = normalizeYAML(e)
		}
		return val
	default:
		return v
	}
}
