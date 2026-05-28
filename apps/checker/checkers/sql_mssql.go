package checkers

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/bytebase/parser/tsql"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// MSSQL validates SQL Server / T-SQL syntax using the bytebase ANTLR4 grammar.
// It reports every syntax error the parser recovers from, with line/column.
type MSSQL struct{}

func (MSSQL) Check(data []byte, strict bool) []result.SyntaxError {
	listener := newANTLRErrorListener()

	lexer := tsql.NewTSqlLexer(antlr.NewInputStream(string(data)))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := tsql.NewTSqlParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)

	parser.Tsql_file()
	return listener.errs
}
