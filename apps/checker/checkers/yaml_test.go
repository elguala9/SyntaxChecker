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

func TestYAMLCustomTagStrict(t *testing.T) {
	errs := YAML{}.Check([]byte("p: !Point [1, 2]\nname: x\n"), true)
	if len(errs) == 0 {
		t.Fatal("expected a custom-tag error in strict mode")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestYAMLCustomTagLenient(t *testing.T) {
	if errs := (YAML{}).Check([]byte("p: !Point [1, 2]\n"), false); len(errs) != 0 {
		t.Fatalf("lenient mode must accept custom tags, got %v", errs)
	}
}

func TestYAMLStandardTagsStrict(t *testing.T) {
	// Explicit standard tags and anchors/aliases stay valid in strict mode.
	src := "a: &x 1\nb: *x\nc: !!str 7\n"
	if errs := (YAML{}).Check([]byte(src), true); len(errs) != 0 {
		t.Fatalf("standard tags and anchors must be valid, got %v", errs)
	}
}

const yamlSchema = `{
  "type": "object",
  "required": ["name", "age"],
  "properties": {
    "name": {"type": "string"},
    "age": {"type": "integer", "minimum": 0}
  }
}`

func TestYAMLSchemaValid(t *testing.T) {
	errs := YAML{}.CheckSchema([]byte("name: demo\nage: 30\n"), []byte(yamlSchema), false)
	if len(errs) != 0 {
		t.Fatalf("expected conforming document, got %v", errs)
	}
}

func TestYAMLSchemaViolations(t *testing.T) {
	errs := YAML{}.CheckSchema([]byte("name: 123\nage: -5\n"), []byte(yamlSchema), false)
	if len(errs) != 2 {
		t.Fatalf("expected 2 schema violations, got %d: %v", len(errs), errs)
	}
}

func TestYAMLSchemaNestedMapNormalized(t *testing.T) {
	// A nested mapping must be normalized to map[string]any so the schema engine
	// can traverse it. The "details" object requires a "city" string.
	schema := `{
  "type": "object",
  "properties": {
    "details": {
      "type": "object",
      "required": ["city"],
      "properties": {"city": {"type": "string"}}
    }
  }
}`
	ok := YAML{}.CheckSchema([]byte("details:\n  city: Rome\n"), []byte(schema), false)
	if len(ok) != 0 {
		t.Fatalf("expected conforming nested document, got %v", ok)
	}
	bad := YAML{}.CheckSchema([]byte("details:\n  zip: 00100\n"), []byte(schema), false)
	if len(bad) == 0 {
		t.Fatal("expected a violation for missing nested required property")
	}
}

func TestYAMLSchemaMalformedDocument(t *testing.T) {
	errs := YAML{}.CheckSchema([]byte("items:\n\t- a\n"), []byte(yamlSchema), false)
	if len(errs) == 0 {
		t.Fatal("expected a syntax error for malformed YAML")
	}
}
