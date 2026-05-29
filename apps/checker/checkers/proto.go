package checkers

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	protoparser "github.com/yoheimuta/go-protoparser/v4"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// Protobuf validates Protocol Buffers (.proto) syntax using go-protoparser in
// non-permissive mode. Only well-formedness is checked, not import resolution
// or type references.
//
// The parser's error string nests the offending token (with its position) and a
// Go source-location suffix; protoErrorPos/protoCleanMessage normalize it into a
// line/column and a readable message.
type Protobuf struct{}

var (
	// reProtoPos captures the line:column embedded in the parser's "Pos=..." token.
	reProtoPos = regexp.MustCompile(`Pos=[^:]*:(\d+):(\d+)`)
	// reProtoToken strips the "(Token=N, Pos=...)" debug noise from the message.
	reProtoToken = regexp.MustCompile(`\(Token=\d+,\s*Pos=[^)]*\)`)
	// reProtoDebug drops the " at <path>.go:NN:..." suffix the parser appends.
	reProtoDebug = regexp.MustCompile(`\s+at\s+\S+\.go:\d+.*$`)
)

func (Protobuf) Check(data []byte, strict bool) []result.SyntaxError {
	_, err := protoparser.Parse(
		bytes.NewReader(data),
		protoparser.WithFilename("input.proto"),
		protoparser.WithPermissive(false),
	)
	if err == nil {
		return nil
	}

	raw := err.Error()
	se := result.SyntaxError{Message: protoCleanMessage(raw)}
	if m := reProtoPos.FindStringSubmatch(raw); m != nil {
		se.Line, _ = strconv.Atoi(m[1])
		se.Column, _ = strconv.Atoi(m[2])
	}
	return []result.SyntaxError{se}
}

// protoCleanMessage turns the parser's nested error into a single readable
// clause, e.g. `found "=" but expected [fieldName]`.
func protoCleanMessage(s string) string {
	s = reProtoToken.ReplaceAllString(s, "")
	s = reProtoDebug.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `""`, `"`)
	return cleanMessage(s)
}
