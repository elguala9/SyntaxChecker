package checkers

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/net/html"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// HTML validates HTML using golang.org/x/net/html. The HTML5 parsing algorithm
// is deliberately error-tolerant (it has well-defined recovery for almost any
// byte sequence), so lenient mode only reports genuine tokenizer errors and is
// effectively always valid — much like Markdown.
//
// With strict=true the tokenizer is replayed to enforce XHTML-style
// well-formedness: every non-void element must be explicitly closed and tags
// must nest correctly. This is what catches mismatched or unclosed tags, which
// a tolerant HTML5 parse silently repairs. Note that optional end tags allowed
// by the HTML spec (e.g. a bare <li> or <p>) are reported in strict mode.
type HTML struct{}

// htmlVoidElements never have a closing tag; they are not pushed on the nesting
// stack in strict mode.
var htmlVoidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

func (HTML) Check(data []byte, strict bool) []result.SyntaxError {
	if !strict {
		return htmlTokenizerErrors(data)
	}
	return htmlWellFormed(data)
}

// htmlTokenizerErrors reports only hard tokenizer failures (other than EOF),
// which the HTML5 tokenizer raises very rarely.
func htmlTokenizerErrors(data []byte) []result.SyntaxError {
	z := html.NewTokenizer(bytes.NewReader(data))
	for {
		if z.Next() == html.ErrorToken {
			if err := z.Err(); err != nil && err != io.EOF {
				return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
			}
			return nil
		}
	}
}

// openTag records an element awaiting its closing tag and where it opened.
type openTag struct {
	name   string
	offset int
}

// htmlWellFormed enforces balanced, correctly nested tags. It tracks a byte
// offset by summing the raw length of every token so positions can be reported.
func htmlWellFormed(data []byte) []result.SyntaxError {
	z := html.NewTokenizer(bytes.NewReader(data))
	var (
		errs   []result.SyntaxError
		stack  []openTag
		offset int
	)

	for {
		tt := z.Next()
		start := offset
		offset += len(z.Raw())

		switch tt {
		case html.ErrorToken:
			if err := z.Err(); err != nil && err != io.EOF {
				errs = append(errs, result.SyntaxError{Message: cleanMessage(err.Error())})
			}
			// Anything still open at EOF was never closed.
			for i := len(stack) - 1; i >= 0; i-- {
				line, col := result.OffsetToLineCol(data, stack[i].offset)
				errs = append(errs, result.SyntaxError{
					Line:    line,
					Column:  col,
					Message: fmt.Sprintf("unclosed element <%s>", stack[i].name),
				})
			}
			return errs

		case html.StartTagToken:
			name, _ := z.TagName()
			n := string(name)
			if !htmlVoidElements[n] {
				stack = append(stack, openTag{name: n, offset: start})
			}

		case html.EndTagToken:
			name, _ := z.TagName()
			n := string(name)
			if htmlVoidElements[n] {
				continue // a stray </br> and friends carry no nesting meaning
			}
			line, col := result.OffsetToLineCol(data, start)
			switch {
			case len(stack) == 0:
				errs = append(errs, result.SyntaxError{
					Line: line, Column: col,
					Message: fmt.Sprintf("unexpected closing tag </%s>", n),
				})
			case stack[len(stack)-1].name != n:
				errs = append(errs, result.SyntaxError{
					Line: line, Column: col,
					Message: fmt.Sprintf("mismatched closing tag </%s>, expected </%s>", n, stack[len(stack)-1].name),
				})
				stack = stack[:len(stack)-1]
			default:
				stack = stack[:len(stack)-1]
			}
		}
	}
}
