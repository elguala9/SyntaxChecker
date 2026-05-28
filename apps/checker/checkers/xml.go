package checkers

import (
	"bytes"
	"encoding/xml"
	"io"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// XML validates XML well-formedness using the standard library. It does NOT
// validate against a DTD/XSD and deliberately does not implement
// SchemaValidator: there is no mature pure-Go XSD validator, and the only
// production-grade option (cgo + libxml2) would require a native library at
// build and run time, breaking the self-contained static binary and its
// Windows installer. Requesting a schema for XML therefore yields a clear
// "schema validation is not supported" error from the CLI.
type XML struct{}

func (XML) Check(data []byte, strict bool) []result.SyntaxError {
	dec := xml.NewDecoder(bytes.NewReader(data))
	depth, roots := 0, 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			// xml.SyntaxError exposes Line (1-based) but not column.
			if se, ok := err.(*xml.SyntaxError); ok {
				return []result.SyntaxError{{Line: se.Line, Message: se.Msg}}
			}
			return []result.SyntaxError{{Message: err.Error()}}
		}
		switch tok.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
			}
			depth++
		case xml.EndElement:
			depth--
		}
	}
	// encoding/xml is lenient about document structure: enforce a single root.
	switch {
	case roots == 0:
		return []result.SyntaxError{{Message: "no root element"}}
	case roots > 1:
		return []result.SyntaxError{{Message: "multiple root elements"}}
	}
	return nil
}
