package checkers

import (
	"github.com/magiconair/properties"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Properties validates Java .properties files using magiconair/properties.
// Expansion is left enabled so that circular or malformed ${...} references and
// invalid \uXXXX escape sequences are reported as errors.
type Properties struct{}

func (Properties) Check(data []byte, strict bool) []result.SyntaxError {
	_, err := properties.Load(data, properties.UTF8)
	if err == nil {
		return nil
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
