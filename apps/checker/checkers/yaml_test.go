package checkers

import "testing"

func TestYAMLValid(t *testing.T) {
	errs := YAML{}.Check([]byte("name: demo\nlist:\n  - a\n  - b\n"), false)
	if len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestYAMLTabIndent(t *testing.T) {
	errs := YAML{}.Check([]byte("items:\n\t- a\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected error for tab indentation")
	}
}

func TestYAMLDuplicateKeys(t *testing.T) {
	errs := YAML{}.Check([]byte("name: demo\nname: other\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected duplicate-key error")
	}
}
