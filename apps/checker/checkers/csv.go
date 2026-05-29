package checkers

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// CSV validates delimiter-separated files using the standard library. Comma
// selects the delimiter: ',' for CSV, '\t' for TSV (the zero value defaults to
// a comma). In strict mode every record must have the same number of fields as
// the first one and quotes must be well-formed; otherwise ragged rows and lazy
// quotes are tolerated.
type CSV struct {
	Comma rune
}

func (c CSV) Check(data []byte, strict bool) []result.SyntaxError {
	comma := c.Comma
	if comma == 0 {
		comma = ','
	}

	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = comma
	r.LazyQuotes = !strict
	if strict {
		r.FieldsPerRecord = 0 // 0 => infer from the first record, then enforce
	} else {
		r.FieldsPerRecord = -1 // allow ragged rows
	}

	var errs []result.SyntaxError
	for {
		_, err := r.Read()
		if err == io.EOF {
			break
		}
		if err == nil {
			continue
		}

		var pe *csv.ParseError
		if errors.As(err, &pe) {
			errs = append(errs, result.SyntaxError{
				Line:    pe.Line,
				Column:  pe.Column,
				Message: cleanMessage(pe.Err.Error()),
			})
			// A field-count mismatch is recoverable, so keep scanning to report
			// every ragged row; any other error aborts the stream.
			if errors.Is(pe.Err, csv.ErrFieldCount) {
				continue
			}
			break
		}
		errs = append(errs, result.SyntaxError{Message: cleanMessage(err.Error())})
		break
	}

	// LazyQuotes (non-strict) tolerates stray quotes inside fields, but a quoted
	// field that is never closed before the end of the file is always malformed.
	// The lenient reader silently swallows it, so detect it explicitly here.
	if !strict {
		if line, ok := unterminatedQuoteLine(data, byte(comma)); ok {
			errs = append(errs, result.SyntaxError{
				Line:    line,
				Message: "unterminated quoted field",
			})
		}
	}
	return errs
}

// unterminatedQuoteLine scans data and, if it ends while still inside a quoted
// field, returns the 1-based line where that quote was opened. Doubled quotes
// ("") inside a quoted field are treated as escaped and do not close it.
func unterminatedQuoteLine(data []byte, comma byte) (int, bool) {
	line, openLine := 1, 0
	inQuote, atFieldStart := false, true
	for i := 0; i < len(data); i++ {
		ch := data[i]
		if inQuote {
			switch ch {
			case '"':
				if i+1 < len(data) && data[i+1] == '"' {
					i++ // escaped quote, stay inside the field
					continue
				}
				inQuote = false
				atFieldStart = false
			case '\n':
				line++ // newlines are legal inside a quoted field
			}
			continue
		}
		switch ch {
		case '"':
			if atFieldStart {
				inQuote = true
				openLine = line
			}
			atFieldStart = false
		case comma:
			atFieldStart = true
		case '\n':
			line++
			atFieldStart = true
		case '\r':
			atFieldStart = true
		default:
			atFieldStart = false
		}
	}
	return openLine, inQuote
}
