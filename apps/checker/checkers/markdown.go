package checkers

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Markdown performs a parse-only check of CommonMark documents using goldmark.
// Markdown has no notion of "invalid" syntax — any byte sequence parses into
// some document — so this validator effectively confirms the file is decodable
// and only surfaces an error if the parser panics. It is included for
// completeness so .md files are recognized rather than rejected as an unknown
// type.
type Markdown struct{}

func (Markdown) Check(data []byte, strict bool) (errs []result.SyntaxError) {
	defer func() {
		if r := recover(); r != nil {
			errs = []result.SyntaxError{{Message: cleanMessage(fmt.Sprintf("markdown parse failed: %v", r))}}
		}
	}()

	md := goldmark.New()
	_ = md.Parser().Parse(text.NewReader(data))
	return nil
}
