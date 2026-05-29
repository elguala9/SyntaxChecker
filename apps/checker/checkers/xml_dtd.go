package checkers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// CheckSchema validates XML well-formedness and then validates the document
// against the given DTD (the schema argument is the DTD text). XSD and RelaxNG
// are deliberately not supported: no mature pure-Go implementation exists and
// the production-grade option (cgo + libxml2) would break the self-contained
// static binary (see the type comment on XML).
//
// The DTD validator covers the common, high-value validity constraints:
//   - every element used must be declared;
//   - child elements must be permitted by the parent's content model;
//   - EMPTY elements may not contain children or text; element-content (non
//     mixed) elements may not contain character data;
//   - #REQUIRED attributes must be present, #FIXED values must match, and
//     enumerated attribute values must be among the declared tokens;
//   - attributes must be declared for their element.
//
// It does not enforce content-model order or cardinality, expand entities, or
// resolve ID/IDREF references — those are reported as limitations rather than
// errors.
func (XML) CheckSchema(data, schema []byte, strict bool) []result.SyntaxError {
	if errs := (XML{}).Check(data, strict); len(errs) != 0 {
		return errs
	}
	d := parseDTD(schema)
	return d.validate(data)
}

// --- DTD model --------------------------------------------------------------

type dtdAttr struct {
	required bool
	fixed    bool
	fixedVal string
	enum     map[string]bool // non-nil when the attribute is an enumeration
}

type dtdElement struct {
	any      bool
	empty    bool
	mixed    bool            // content model includes #PCDATA
	children map[string]bool // element names allowed as children
	attrs    map[string]*dtdAttr
}

type dtd struct {
	elements map[string]*dtdElement
}

func (d *dtd) element(name string) *dtdElement {
	e := d.elements[name]
	if e == nil {
		e = &dtdElement{children: map[string]bool{}, attrs: map[string]*dtdAttr{}}
		d.elements[name] = e
	}
	return e
}

// --- DTD parsing ------------------------------------------------------------

var (
	reDTDComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	reDTDElement = regexp.MustCompile(`(?s)<!ELEMENT\s+([^\s>]+)\s+(.*?)\s*>`)
	reDTDAttlist = regexp.MustCompile(`(?s)<!ATTLIST\s+([^\s>]+)\s+(.*?)\s*>`)
	reDTDName    = regexp.MustCompile(`[A-Za-z_:][\w.:-]*`)
)

func parseDTD(schema []byte) *dtd {
	src := reDTDComment.ReplaceAllString(string(schema), "")
	d := &dtd{elements: map[string]*dtdElement{}}

	for _, m := range reDTDElement.FindAllStringSubmatch(src, -1) {
		el := d.element(m[1])
		spec := strings.TrimSpace(m[2])
		switch {
		case spec == "EMPTY":
			el.empty = true
		case spec == "ANY":
			el.any = true
		default:
			el.mixed = strings.Contains(spec, "#PCDATA")
			for _, name := range reDTDName.FindAllString(spec, -1) {
				if name != "PCDATA" { // the regex strips the leading '#'
					el.children[name] = true
				}
			}
		}
	}

	for _, m := range reDTDAttlist.FindAllStringSubmatch(src, -1) {
		el := d.element(m[1])
		parseAttlist(el, m[2])
	}
	return d
}

// parseAttlist parses an ATTLIST body ("name type default name type default …")
// into the element's attribute map.
func parseAttlist(el *dtdElement, body string) {
	toks := tokenizeAttlist(body)
	for i := 0; i+2 < len(toks)+1 && i+1 < len(toks); {
		name := toks[i]
		attType := toks[i+1]
		i += 2

		attr := &dtdAttr{}
		if strings.HasPrefix(attType, "(") {
			attr.enum = map[string]bool{}
			for _, v := range reDTDName.FindAllString(attType, -1) {
				attr.enum[v] = true
			}
		}

		if i >= len(toks) {
			el.attrs[name] = attr
			break
		}
		switch d := toks[i]; d {
		case "#REQUIRED":
			attr.required = true
			i++
		case "#IMPLIED":
			i++
		case "#FIXED":
			attr.fixed = true
			i++
			if i < len(toks) {
				attr.fixedVal = unquote(toks[i])
				i++
			}
		default:
			// A literal default value (quoted); nothing to enforce.
			i++
		}
		el.attrs[name] = attr
	}
}

// tokenizeAttlist splits an ATTLIST body into words, parenthesized groups, and
// quoted strings.
func tokenizeAttlist(s string) []string {
	var toks []string
	for i := 0; i < len(s); {
		switch c := s[i]; {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '"' || c == '\'':
			j := i + 1
			for j < len(s) && s[j] != c {
				j++
			}
			if j < len(s) {
				j++ // include closing quote
			}
			toks = append(toks, s[i:j])
			i = j
		case c == '(':
			depth, j := 0, i
			for j < len(s) {
				if s[j] == '(' {
					depth++
				} else if s[j] == ')' {
					depth--
					if depth == 0 {
						j++
						break
					}
				}
				j++
			}
			toks = append(toks, s[i:j])
			i = j
		default:
			j := i
			for j < len(s) && !strings.ContainsRune(" \t\n\r(\"'", rune(s[j])) {
				j++
			}
			toks = append(toks, s[i:j])
			i = j
		}
	}
	return toks
}

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}

// --- Validation -------------------------------------------------------------

// validate walks the document once, checking each element against the DTD. The
// document is assumed well-formed (CheckSchema verifies that first).
func (d *dtd) validate(data []byte) []result.SyntaxError {
	if len(d.elements) == 0 {
		return []result.SyntaxError{{Message: "DTD declares no elements"}}
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	var errs []result.SyntaxError
	var stack []*dtdElement

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break // well-formedness already validated
		}

		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local
			parent := top(stack)
			if parent != nil && !parent.any && !parent.children[name] {
				errs = append(errs, result.SyntaxError{
					Message: fmt.Sprintf("element <%s> is not allowed here", name),
				})
			}

			el := d.elements[name]
			if el == nil {
				errs = append(errs, result.SyntaxError{
					Message: fmt.Sprintf("undeclared element <%s>", name),
				})
			} else {
				errs = append(errs, validateAttrs(name, el, t.Attr)...)
			}
			stack = append(stack, el)

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			if cur := top(stack); cur != nil && strings.TrimSpace(string(t)) != "" {
				if cur.empty || (!cur.any && !cur.mixed) {
					errs = append(errs, result.SyntaxError{
						Message: "character data is not allowed in element-content",
					})
				}
			}
		}
	}
	return errs
}

func top(stack []*dtdElement) *dtdElement {
	if len(stack) == 0 {
		return nil
	}
	return stack[len(stack)-1]
}

func validateAttrs(name string, el *dtdElement, attrs []xml.Attr) []result.SyntaxError {
	var errs []result.SyntaxError
	seen := map[string]string{}
	for _, a := range attrs {
		if a.Name.Local == "xmlns" || a.Name.Space == "xmlns" {
			continue // namespace declarations are not subject to ATTLIST
		}
		seen[a.Name.Local] = a.Value
		def := el.attrs[a.Name.Local]
		if def == nil {
			errs = append(errs, result.SyntaxError{
				Message: fmt.Sprintf("undeclared attribute %q on <%s>", a.Name.Local, name),
			})
			continue
		}
		if def.fixed && a.Value != def.fixedVal {
			errs = append(errs, result.SyntaxError{
				Message: fmt.Sprintf("attribute %q on <%s> must equal fixed value %q", a.Name.Local, name, def.fixedVal),
			})
		}
		if def.enum != nil && !def.enum[a.Value] {
			errs = append(errs, result.SyntaxError{
				Message: fmt.Sprintf("attribute %q on <%s> has value %q outside its enumeration", a.Name.Local, name, a.Value),
			})
		}
	}
	for attrName, def := range el.attrs {
		if def.required {
			if _, ok := seen[attrName]; !ok {
				errs = append(errs, result.SyntaxError{
					Message: fmt.Sprintf("missing required attribute %q on <%s>", attrName, name),
				})
			}
		}
	}
	return errs
}
