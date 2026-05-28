package checkers

import (
	"github.com/antlr4-go/antlr/v4"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// antlrErrorListener collects the syntax errors emitted by an ANTLR4 lexer or
// parser into the shared SyntaxError shape. ANTLR reports a 1-based line and a
// 0-based charPositionInLine, so the column is normalized to 1-based.
type antlrErrorListener struct {
	*antlr.DefaultErrorListener
	errs []result.SyntaxError
}

func newANTLRErrorListener() *antlrErrorListener {
	return &antlrErrorListener{DefaultErrorListener: antlr.NewDefaultErrorListener()}
}

func (l *antlrErrorListener) SyntaxError(_ antlr.Recognizer, _ any, line, column int, msg string, _ antlr.RecognitionException) {
	l.errs = append(l.errs, result.SyntaxError{
		Line:    line,
		Column:  column + 1,
		Message: cleanMessage(msg),
	})
}
