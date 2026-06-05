package checkers

import (
	"fmt"
	"sort"

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

	tree := parser.Tsql_file()

	// The bytebase T-SQL grammar makes the separating comma optional in table
	// element lists (column_def_table_constraints: element (','? element)* and
	// create_table: ... (','? table_indices)*), so a missing comma between two
	// column/constraint/index definitions parses cleanly even though SQL Server
	// rejects it. The error listener never fires for these, so we walk the tree
	// and flag any two adjacent elements with no comma token between them.
	errs := append(listener.errs, missingCommaErrors(tree)...)
	sort.SliceStable(errs, func(i, j int) bool {
		if errs[i].Line != errs[j].Line {
			return errs[i].Line < errs[j].Line
		}
		return errs[i].Column < errs[j].Column
	})
	return errs
}

// missingCommaErrors walks the parse tree and reports table element lists whose
// items are not comma-separated, recovering the strictness the lenient grammar
// drops. It covers column_def_table_constraints (used by CREATE TABLE, DECLARE
// @t TABLE, CREATE TYPE ... AS TABLE and table-valued parameters) and the
// table_indices sequence directly inside create_table.
func missingCommaErrors(tree antlr.Tree) []result.SyntaxError {
	var errs []result.SyntaxError

	var walk func(node antlr.Tree)
	walk = func(node antlr.Tree) {
		switch ctx := node.(type) {
		case tsql.IColumn_def_table_constraintsContext:
			errs = append(errs, checkSeparators(ctx, isConstraintElement)...)
		case tsql.ICreate_tableContext:
			errs = append(errs, checkSeparators(ctx, isCreateTableElement)...)
		}
		for i := 0; i < node.GetChildCount(); i++ {
			walk(node.GetChild(i))
		}
	}
	walk(tree)
	return errs
}

// checkSeparators scans the direct children of ctx in order and reports an error
// at the start of any element that immediately follows another element with no
// comma in between.
func checkSeparators(ctx antlr.Tree, isElement func(antlr.Tree) bool) []result.SyntaxError {
	var errs []result.SyntaxError
	prevWasElement := false
	for i := 0; i < ctx.GetChildCount(); i++ {
		child := ctx.GetChild(i)
		switch {
		case isComma(child):
			prevWasElement = false
		case isElement(child):
			if prevWasElement {
				if tok := startToken(child); tok != nil {
					errs = append(errs, result.SyntaxError{
						Line:    tok.GetLine(),
						Column:  tok.GetColumn() + 1,
						Message: fmt.Sprintf("missing ',' before '%s'", tok.GetText()),
					})
				}
			}
			prevWasElement = true
		}
	}
	return errs
}

func isComma(node antlr.Tree) bool {
	t, ok := node.(antlr.TerminalNode)
	if !ok {
		return false
	}
	return t.GetSymbol().GetTokenType() == tsql.TSqlParserCOMMA
}

func startToken(node antlr.Tree) antlr.Token {
	if prc, ok := node.(antlr.ParserRuleContext); ok {
		return prc.GetStart()
	}
	return nil
}

func isConstraintElement(node antlr.Tree) bool {
	_, ok := node.(tsql.IColumn_def_table_constraintContext)
	return ok
}

func isCreateTableElement(node antlr.Tree) bool {
	switch node.(type) {
	case tsql.IColumn_def_table_constraintsContext, tsql.ITable_indicesContext:
		return true
	}
	return false
}
