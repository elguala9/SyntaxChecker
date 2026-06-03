package checkers

import (
	_ "embed"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// composeSchemaJSON is the official Compose Specification JSON Schema, vendored
// from github.com/compose-spec/compose-spec (schema/compose-spec.json, master).
// It is JSON Schema draft-07 and is validated by the same engine used for
// user-supplied schemas (see schema.go). Refresh it from upstream when the
// Compose Specification changes; it is the single source of truth for the
// structural checks below.
//
//go:embed compose-spec.json
var composeSchemaJSON []byte

// Compose validates Docker Compose / podman-compose files. A Compose file is a
// YAML document that must additionally conform to the Compose Specification, so
// Check runs two passes: YAML well-formedness first (reusing the YAML checker,
// which also reports duplicate keys and, in strict mode, custom tags) and then
// structural validation of the decoded document against the embedded Compose
// Specification schema.
//
// This is syntax + structural validation only: runtime semantics such as image
// resolution, build contexts, and "${VAR}" interpolation are intentionally not
// enforced (interpolation in particular would turn unset environment variables
// into spurious errors).
type Compose struct{}

// The embedded schema is compiled once and reused; compilation is deterministic
// and the result is safe for concurrent validation.
var (
	composeSchemaOnce sync.Once
	composeSchema     *jsonschema.Schema
	composeSchemaErr  error
)

func compiledComposeSchema() (*jsonschema.Schema, error) {
	composeSchemaOnce.Do(func() {
		composeSchema, composeSchemaErr = compileSchema(composeSchemaJSON)
	})
	return composeSchema, composeSchemaErr
}

func (Compose) Check(data []byte, strict bool) []result.SyntaxError {
	// Never schema-check a document that is not well-formed YAML. Delegating to
	// the YAML checker also reuses its duplicate-key and strict custom-tag logic.
	if errs := (YAML{}).Check(data, strict); len(errs) != 0 {
		return errs
	}

	var v any
	if err := yaml.Unmarshal(data, &v); err != nil {
		// Unreachable in practice: the YAML check above already passed.
		return []result.SyntaxError{yamlSyntaxError(err.Error())}
	}

	sch, err := compiledComposeSchema()
	if err != nil {
		return []result.SyntaxError{{Message: err.Error()}}
	}
	return validateAgainstSchema(sch, normalizeYAML(v))
}
