package checkers

import (
	"bytes"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// schemaResourceURL is the synthetic URL under which the in-memory schema is
// registered with the compiler. Its value is irrelevant as long as it is used
// consistently for AddResource and Compile.
const schemaResourceURL = "mem://schema.json"

// compileSchema parses and compiles a JSON Schema document. A parse or compile
// failure is returned as an error (the schema itself is invalid), as opposed to
// a validation error against the document.
func compileSchema(schema []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema))
	if err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaResourceURL, doc); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}
	sch, err := c.Compile(schemaResourceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}
	return sch, nil
}

// validateAgainstSchema validates an already-parsed instance — the generic
// value model shared by JSON and YAML — against a compiled schema, flattening
// the validation error tree into one SyntaxError per leaf violation.
//
// JSON Schema validation runs on the decoded value, which has no source
// position, so Line/Column are left at 0; the instance location (a JSON
// pointer such as "/items/0/name") is embedded in the message instead.
func validateAgainstSchema(sch *jsonschema.Schema, inst any) []result.SyntaxError {
	err := sch.Validate(inst)
	if err == nil {
		return nil
	}
	ve, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
	}

	var errs []result.SyntaxError
	var walk func(u *jsonschema.OutputUnit)
	walk = func(u *jsonschema.OutputUnit) {
		if u.Error != nil {
			loc := u.InstanceLocation
			if loc == "" {
				loc = "/"
			}
			errs = append(errs, result.SyntaxError{
				Message: cleanMessage(fmt.Sprintf("at %s: %s", loc, u.Error.String())),
			})
		}
		for i := range u.Errors {
			walk(&u.Errors[i])
		}
	}
	walk(ve.BasicOutput())

	// A non-conforming document always yields at least one leaf error, but guard
	// against an unexpectedly empty tree so we never report "valid" on error.
	if len(errs) == 0 {
		errs = append(errs, result.SyntaxError{Message: cleanMessage(ve.Error())})
	}
	return errs
}
