package checkers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestXMLDTDValid(t *testing.T) {
	dtd := "<!ELEMENT note (to)>\n<!ELEMENT to (#PCDATA)>\n<!ATTLIST note id CDATA #REQUIRED>\n"
	doc := `<note id="1"><to>Bob</to></note>`
	if errs := (XML{}).CheckSchema([]byte(doc), []byte(dtd), false); len(errs) != 0 {
		t.Fatalf("expected conforming document, got %v", errs)
	}
}

func TestXMLDTDUndeclaredElement(t *testing.T) {
	dtd := "<!ELEMENT note (to)>\n<!ELEMENT to (#PCDATA)>\n"
	doc := `<note><to>Bob</to><extra>x</extra></note>`
	if errs := (XML{}).CheckSchema([]byte(doc), []byte(dtd), false); len(errs) == 0 {
		t.Fatal("expected an error for the undeclared/unexpected element")
	}
}

func TestXMLDTDMissingRequiredAttr(t *testing.T) {
	dtd := "<!ELEMENT note (#PCDATA)>\n<!ATTLIST note id CDATA #REQUIRED>\n"
	doc := `<note>hi</note>`
	if errs := (XML{}).CheckSchema([]byte(doc), []byte(dtd), false); len(errs) == 0 {
		t.Fatal("expected a missing-required-attribute error")
	}
}

func TestXMLDTDEnumViolation(t *testing.T) {
	dtd := "<!ELEMENT note (#PCDATA)>\n<!ATTLIST note p (a|b) #IMPLIED>\n"
	doc := `<note p="c">hi</note>`
	if errs := (XML{}).CheckSchema([]byte(doc), []byte(dtd), false); len(errs) == 0 {
		t.Fatal("expected an enumeration violation")
	}
}

func TestXMLDTDMalformedDocument(t *testing.T) {
	dtd := "<!ELEMENT note (#PCDATA)>\n"
	if errs := (XML{}).CheckSchema([]byte("<note>oops</wrong>"), []byte(dtd), false); len(errs) == 0 {
		t.Fatal("expected a well-formedness error before DTD validation")
	}
}

// TestXMLDTDSamples is self-discovering over test-samples/dtd. Each
// <topic>.<variant>.xml document is validated against its sibling <topic>.dtd.
func TestXMLDTDSamples(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "test-samples", "dtd")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("dtd samples not available: %v", err)
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".xml" {
			continue
		}
		name := e.Name()
		dot := strings.IndexByte(name, '.')
		if dot < 0 {
			t.Errorf("%s: expected a <topic>.<variant>.xml name", name)
			continue
		}
		topic := name[:dot]

		doc, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("cannot read %s: %v", name, err)
		}
		schema, err := os.ReadFile(filepath.Join(dir, topic+".dtd"))
		if err != nil {
			t.Fatalf("cannot read DTD for %s: %v", name, err)
		}

		wantInvalid := strings.Contains(name, "_not_correct")
		t.Run(name, func(t *testing.T) {
			errs := XML{}.CheckSchema(doc, schema, true)
			switch {
			case wantInvalid && len(errs) == 0:
				t.Errorf("expected DTD violations, got none")
			case !wantInvalid && len(errs) != 0:
				t.Errorf("expected a conforming document, got %d error(s): %v", len(errs), errs)
			}
		})
	}
}
