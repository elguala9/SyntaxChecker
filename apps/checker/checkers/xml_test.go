package checkers

import "testing"

func TestXMLValid(t *testing.T) {
	errs := XML{}.Check([]byte(`<?xml version="1.0"?><root><a>1</a></root>`), false)
	if len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestXMLMismatchedTag(t *testing.T) {
	errs := XML{}.Check([]byte("<root>\n  <a>1</b>\n</root>"), false)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestXMLMultipleRoots(t *testing.T) {
	errs := XML{}.Check([]byte(`<a/><b/>`), false)
	if len(errs) != 1 {
		t.Fatalf("expected multi-root error, got %v", errs)
	}
}

func TestXMLEmpty(t *testing.T) {
	errs := XML{}.Check([]byte("   \n  "), false)
	if len(errs) != 1 {
		t.Fatalf("expected no-root error, got %v", errs)
	}
}
