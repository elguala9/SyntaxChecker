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
	return errs
}
