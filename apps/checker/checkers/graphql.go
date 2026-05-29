package checkers

import (
	"errors"
	"regexp"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// GraphQL validates GraphQL documents using gqlparser. A .graphql file may be
// either a schema (SDL) or an executable document (query/mutation/fragment); the
// two use different grammars. The document is considered valid when it parses as
// either kind, so neither grammar produces false positives against the other.
type GraphQL struct{}

// reGraphQLSchema matches a top-level SDL definition keyword. It only drives
// which grammar's error is reported when both grammars reject the input.
var reGraphQLSchema = regexp.MustCompile(`(?m)^\s*(type|schema|interface|input|enum|scalar|union|directive|extend)\b`)

func (GraphQL) Check(data []byte, strict bool) []result.SyntaxError {
	src := &ast.Source{Name: "input.graphql", Input: string(data)}

	if _, err := parser.ParseSchema(src); err == nil {
		return nil
	}
	if _, err := parser.ParseQuery(src); err == nil {
		return nil
	}

	// Both grammars rejected the input. Report the error from the grammar the
	// document most likely intended, for a relevant message.
	var err error
	if reGraphQLSchema.Match(data) {
		_, err = parser.ParseSchema(src)
	} else {
		_, err = parser.ParseQuery(src)
	}
	return graphqlErrors(err)
}

func graphqlErrors(err error) []result.SyntaxError {
	if err == nil {
		return nil
	}
	var ge *gqlerror.Error
	if errors.As(err, &ge) {
		se := result.SyntaxError{Message: cleanMessage(ge.Message)}
		if len(ge.Locations) > 0 {
			se.Line = ge.Locations[0].Line
			se.Column = ge.Locations[0].Column
		}
		return []result.SyntaxError{se}
	}
	return []result.SyntaxError{{Message: cleanMessage(err.Error())}}
}
