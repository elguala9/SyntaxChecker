package checkers

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	rqlite "github.com/rqlite/sql"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// SQLite validates SQLite SQL syntax using rqlite/sql, a pure-Go parser used in
// production by rqlite. It parses every statement in the input and reports the
// first syntax error, with its native line/column position.
type SQLite struct{}

func (SQLite) Check(data []byte, strict bool) []result.SyntaxError {
	p := rqlite.NewParser(strings.NewReader(string(data)))
	stmts, err := p.ParseStatements()
	if err != nil {
		se := result.SyntaxError{Message: cleanMessage(err.Error())}
		var pe *rqlite.Error
		if errors.As(err, &pe) {
			se.Line = pe.Pos.Line
			se.Column = pe.Pos.Column
		}
		return []result.SyntaxError{se}
	}

	// rqlite/sql is more permissive than SQLite about the comma that separates
	// table elements: parseColumnDefinitions simply loops, so a column followed
	// by another column (or a table constraint) with no comma in between parses
	// cleanly even though SQLite rejects it ("near <x>: syntax error"). The
	// parser drops the comma tokens, so we re-scan the source and flag any two
	// adjacent CREATE TABLE elements without a separating comma.
	errs := sqliteMissingCommaErrors(data, stmts)
	if len(errs) == 0 {
		return nil
	}
	sort.SliceStable(errs, func(i, j int) bool {
		if errs[i].Line != errs[j].Line {
			return errs[i].Line < errs[j].Line
		}
		return errs[i].Column < errs[j].Column
	})
	return errs
}

// sqliteMissingCommaErrors finds every CREATE TABLE statement and reports table
// elements that are not comma-separated from the element before them.
func sqliteMissingCommaErrors(data []byte, stmts []rqlite.Statement) []result.SyntaxError {
	tables := collectCreateTables(stmts)
	if len(tables) == 0 {
		return nil
	}

	commas, litAt := scanCommasAndLits(data)

	var errs []result.SyntaxError
	for _, t := range tables {
		errs = append(errs, tableMissingCommaErrors(t, commas, litAt)...)
	}
	return errs
}

// commaTok records the source offset and parenthesis depth of a comma token.
type commaTok struct {
	offset int
	depth  int
}

// tokenInfo records the depth and literal text of a token at a given offset.
type tokenInfo struct {
	depth int
	lit   string
}

// scanCommasAndLits tokenizes the whole input once. It returns the comma tokens
// (with the paren depth they sit at) and, for every token, its depth and source
// text keyed by offset so element starts can be located precisely.
func scanCommasAndLits(data []byte) ([]commaTok, map[int]tokenInfo) {
	s := rqlite.NewScanner(strings.NewReader(string(data)))
	var commas []commaTok
	litAt := make(map[int]tokenInfo)

	depth := 0
	for {
		pos, tok, lit := s.Scan()
		if tok == rqlite.RP && depth > 0 {
			depth--
		}
		if lit == "" {
			lit = tok.String()
		}
		litAt[pos.Offset] = tokenInfo{depth: depth, lit: lit}
		switch tok {
		case rqlite.COMMA:
			commas = append(commas, commaTok{offset: pos.Offset, depth: depth})
		case rqlite.LP:
			depth++
		case rqlite.EOF:
			return commas, litAt
		}
	}
}

// tableMissingCommaErrors reports each element of a single CREATE TABLE that is
// not separated from the preceding element by a comma at the element's depth.
func tableMissingCommaErrors(t *rqlite.CreateTableStatement, commas []commaTok, litAt map[int]tokenInfo) []result.SyntaxError {
	type elem struct {
		pos rqlite.Pos
	}
	var elems []elem
	for _, c := range t.Columns {
		if c != nil && c.Name != nil {
			elems = append(elems, elem{pos: c.Name.NamePos})
		}
	}
	for _, c := range t.Constraints {
		if pos, ok := sqliteConstraintStart(c); ok {
			elems = append(elems, elem{pos: pos})
		}
	}
	sort.SliceStable(elems, func(i, j int) bool { return elems[i].pos.Offset < elems[j].pos.Offset })

	var errs []result.SyntaxError
	for i := 1; i < len(elems); i++ {
		prev, cur := elems[i-1], elems[i]
		depth := litAt[cur.pos.Offset].depth
		if hasCommaBetween(commas, prev.pos.Offset, cur.pos.Offset, depth) {
			continue
		}
		text := litAt[cur.pos.Offset].lit
		errs = append(errs, result.SyntaxError{
			Line:    cur.pos.Line,
			Column:  cur.pos.Column,
			Message: fmt.Sprintf("missing ',' before '%s'", text),
		})
	}
	return errs
}

// hasCommaBetween reports whether a comma at the given depth sits between two
// offsets (exclusive); deeper commas, e.g. inside UNIQUE (a, b), are ignored.
func hasCommaBetween(commas []commaTok, from, to, depth int) bool {
	for _, c := range commas {
		if c.offset > from && c.offset < to && c.depth == depth {
			return true
		}
	}
	return false
}

// sqliteConstraintStart returns the source position where a table constraint
// begins: the CONSTRAINT keyword when named, otherwise its leading keyword.
func sqliteConstraintStart(c rqlite.Constraint) (rqlite.Pos, bool) {
	switch t := c.(type) {
	case *rqlite.PrimaryKeyConstraint:
		if t.Constraint.Line > 0 {
			return t.Constraint, true
		}
		return t.Primary, true
	case *rqlite.NotNullConstraint:
		if t.Constraint.Line > 0 {
			return t.Constraint, true
		}
		return t.Not, true
	case *rqlite.UniqueConstraint:
		if t.Constraint.Line > 0 {
			return t.Constraint, true
		}
		return t.Unique, true
	case *rqlite.CheckConstraint:
		if t.Constraint.Line > 0 {
			return t.Constraint, true
		}
		return t.Check, true
	case *rqlite.ForeignKeyConstraint:
		if t.Constraint.Line > 0 {
			return t.Constraint, true
		}
		return t.Foreign, true
	}
	return rqlite.Pos{}, false
}

// collectCreateTables walks the parsed statements and returns every CREATE TABLE
// node, including those nested inside other statements (e.g. trigger bodies).
func collectCreateTables(stmts []rqlite.Statement) []*rqlite.CreateTableStatement {
	c := &createTableCollector{}
	for _, stmt := range stmts {
		if stmt == nil {
			continue
		}
		_, _ = rqlite.Walk(c, stmt)
	}
	return c.tables
}

type createTableCollector struct {
	tables []*rqlite.CreateTableStatement
}

func (c *createTableCollector) Visit(n rqlite.Node) (rqlite.Visitor, rqlite.Node, error) {
	if ct, ok := n.(*rqlite.CreateTableStatement); ok {
		c.tables = append(c.tables, ct)
	}
	return c, n, nil
}

func (c *createTableCollector) VisitEnd(n rqlite.Node) (rqlite.Node, error) {
	return n, nil
}
