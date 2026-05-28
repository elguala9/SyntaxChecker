package checkers

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/bytebase/parser/plsql"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Oracle validates Oracle PL/SQL syntax using the bytebase ANTLR4 grammar. It
// reports every syntax error the parser recovers from, with line/column.
type Oracle struct{}

func (Oracle) Check(data []byte, strict bool) []result.SyntaxError {
	listener := newANTLRErrorListener()

	lexer := plsql.NewPlSqlLexer(antlr.NewInputStream(string(data)))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := plsql.NewPlSqlParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)

	parser.Sql_script()
	return listener.errs
}
