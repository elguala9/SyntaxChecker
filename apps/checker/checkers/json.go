package checkers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// JSON validates JSON syntax using the standard library. With strict=true it
// also reports duplicate object keys, which encoding/json silently ignores.
type JSON struct{}

func (JSON) Check(data []byte, strict bool) []result.SyntaxError {
	// First pass: well-formedness. json.Unmarshal reports a byte offset only,
	// so we convert it to line/column.
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		var se *json.SyntaxError
		if errors.As(err, &se) {
			line, col := result.OffsetToLineCol(data, int(se.Offset))
			return []result.SyntaxError{{Line: line, Column: col, Message: se.Error()}}
		}
		return []result.SyntaxError{{Message: err.Error()}}
	}

	if !strict {
		return nil
	}
	return duplicateKeys(data)
}

// CheckSchema validates JSON well-formedness and then validates the document
// against the given JSON Schema. With strict=true it additionally reports
// duplicate object keys, as Check does.
func (JSON) CheckSchema(data, schema []byte, strict bool) []result.SyntaxError {
	var inst any
	if err := json.Unmarshal(data, &inst); err != nil {
		var se *json.SyntaxError
		if errors.As(err, &se) {
			line, col := result.OffsetToLineCol(data, int(se.Offset))
			return []result.SyntaxError{{Line: line, Column: col, Message: se.Error()}}
		}
		return []result.SyntaxError{{Message: err.Error()}}
	}

	sch, err := compileSchema(schema)
	if err != nil {
		return []result.SyntaxError{{Message: err.Error()}}
	}

	var errs []result.SyntaxError
	if strict {
		errs = append(errs, duplicateKeys(data)...)
	}
	errs = append(errs, validateAgainstSchema(sch, inst)...)
	return errs
}

// container tracks one open JSON object or array while streaming tokens.
type container struct {
	isObject bool
	keys     map[string]bool // populated only for objects
	wantKey  bool            // for objects: next string token is a key
}

// duplicateKeys streams the document with a token decoder, tracking the set of
// keys seen at each open object. A repeated key produces an error.
func duplicateKeys(data []byte) []result.SyntaxError {
	dec := json.NewDecoder(bytes.NewReader(data))
	var errs []result.SyntaxError
	var stack []*container

	top := func() *container {
		if len(stack) == 0 {
			return nil
		}
		return stack[len(stack)-1]
	}

	// valueConsumed advances the parent object's key/value alternation after a
	// complete value has been read.
	valueConsumed := func() {
		if c := top(); c != nil && c.isObject {
			c.wantKey = true
		}
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break // malformed input is already reported by the first pass
		}
		// InputOffset after Token() is the end of the token just read, so it
		// lands on the key's own line (reading it before Token() would point at
		// the end of the previous token, i.e. the previous line).
		off := dec.InputOffset()

		if delim, ok := tok.(json.Delim); ok {
			switch delim {
			case '{':
				stack = append(stack, &container{isObject: true, keys: map[string]bool{}, wantKey: true})
			case '[':
				stack = append(stack, &container{})
			case '}', ']':
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
				valueConsumed()
			}
			continue
		}

		c := top()
		if c != nil && c.isObject && c.wantKey {
			key, _ := tok.(string)
			if c.keys[key] {
				line, col := result.OffsetToLineCol(data, int(off))
				errs = append(errs, result.SyntaxError{
					Line:    line,
					Column:  col,
					Message: fmt.Sprintf("duplicate key %q", key),
				})
			}
			c.keys[key] = true
			c.wantKey = false
			continue
		}

		// A scalar value (in an object or array, or the top-level document).
		valueConsumed()
	}
	return errs
}
