package checkers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// HCL validates HashiCorp Configuration Language (and Terraform .tf) syntax
// using the native hclsyntax parser. Only well-formedness is checked, not the
// semantic validity of blocks/attributes. Each error carries the precise
// start position of the offending construct.
type HCL struct{}

func (HCL) Check(data []byte, strict bool) []result.SyntaxError {
	_, diags := hclsyntax.ParseConfig(data, "input.hcl", hcl.Pos{Line: 1, Column: 1})
	if !diags.HasErrors() {
		return nil
	}

	var errs []result.SyntaxError
	for _, d := range diags {
		if d.Severity != hcl.DiagError {
			continue
		}
		msg := d.Summary
		if d.Detail != "" {
			msg += ": " + d.Detail
		}
		se := result.SyntaxError{Message: cleanMessage(msg)}
		if d.Subject != nil {
			se.Line = d.Subject.Start.Line
			se.Column = d.Subject.Start.Column
		}
		errs = append(errs, se)
	}
	return errs
}
