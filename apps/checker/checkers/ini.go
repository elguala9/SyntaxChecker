package checkers

import (
	"regexp"
	"strconv"

	"gopkg.in/ini.v1"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// INI validates INI / CFG syntax using gopkg.in/ini.v1. The loader reports the
// first malformed line (e.g. a key without a value delimiter).
type INI struct{}

// ini.v1 embeds the offending line number in its error messages, e.g.
// "error parsing line 3: ...".
var iniLineRE = regexp.MustCompile(`(?i)line (\d+)`)

func (INI) Check(data []byte, strict bool) []result.SyntaxError {
	_, err := ini.Load(data)
	if err == nil {
		return nil
	}
	return []result.SyntaxError{iniSyntaxError(err.Error())}
}

func iniSyntaxError(msg string) result.SyntaxError {
	line := 0
	if m := iniLineRE.FindStringSubmatch(msg); m != nil {
		line, _ = strconv.Atoi(m[1])
	}
	return result.SyntaxError{Line: line, Message: cleanMessage(msg)}
}
