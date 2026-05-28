package checkers

import (
	"regexp"
	"strconv"

	"github.com/parresia/syntaxchecker/pkg/result"

	"github.com/pingcap/tidb/pkg/parser"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

// MySQL validates MySQL/TiDB SQL syntax using the TiDB parser.
//
// The TiDB parser is NOT goroutine-safe, so a fresh parser is created per call.
// We register the lightweight test_driver (blank import above) rather than the
// full types parser_driver, which would pull in most of TiDB.
type MySQL struct{}

// TiDB embeds the position in the message, e.g. `line 1 column 13 near "..."`.
var tidbPosRE = regexp.MustCompile(`line (\d+) column (\d+)`)

func (MySQL) Check(data []byte, strict bool) []result.SyntaxError {
	p := parser.New()
	_, _, err := p.Parse(string(data), "", "")
	if err == nil {
		return nil
	}
	se := result.SyntaxError{Message: cleanMessage(err.Error())}
	if m := tidbPosRE.FindStringSubmatch(err.Error()); m != nil {
		se.Line, _ = strconv.Atoi(m[1])
		se.Column, _ = strconv.Atoi(m[2])
	}
	return []result.SyntaxError{se}
}
