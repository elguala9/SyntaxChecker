package checkers

import "testing"

func TestJSONValid(t *testing.T) {
	errs := JSON{}.Check([]byte(`{"a":1,"b":[1,2,{"c":3}]}`), true)
	if len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestJSONMalformed(t *testing.T) {
	errs := JSON{}.Check([]byte("{\n  \"a\": 1\n  \"b\": 2\n}"), false)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestJSONDuplicateKeysStrict(t *testing.T) {
	// Line layout: 1 "{", 2 first "id", 3 dup "id", 4 "nested", 5 first "x", 6 dup "x".
	src := []byte("{\n  \"id\": 1,\n  \"id\": 2,\n  \"nested\": {\n    \"x\": 1,\n    \"x\": 2\n  }\n}")
	errs := JSON{}.Check(src, true)
	if len(errs) != 2 {
		t.Fatalf("expected 2 duplicate-key errors, got %d: %v", len(errs), errs)
	}
	if errs[0].Line != 3 {
		t.Errorf("duplicate \"id\" should be on line 3, got line %d", errs[0].Line)
	}
	if errs[1].Line != 6 {
		t.Errorf("duplicate \"x\" should be on line 6, got line %d", errs[1].Line)
	}
}

func TestJSONDuplicateKeysIgnoredWithoutStrict(t *testing.T) {
	src := []byte(`{"id":1,"id":2}`)
	errs := JSON{}.Check(src, false)
	if len(errs) != 0 {
		t.Fatalf("expected no errors without strict, got %v", errs)
	}
}

func TestJSONArrayElementsNotKeys(t *testing.T) {
	// Strings inside arrays must not be treated as object keys.
	src := []byte(`{"items":["a","a","a"]}`)
	errs := JSON{}.Check(src, true)
	if len(errs) != 0 {
		t.Fatalf("array string elements wrongly flagged: %v", errs)
	}
}

const jsonSchema = `{
  "type": "object",
  "required": ["name", "age"],
  "properties": {
    "name": {"type": "string"},
    "age": {"type": "integer", "minimum": 0}
  }
}`

func TestJSONSchemaValid(t *testing.T) {
	errs := JSON{}.CheckSchema([]byte(`{"name":"demo","age":30}`), []byte(jsonSchema), false)
	if len(errs) != 0 {
		t.Fatalf("expected conforming document, got %v", errs)
	}
}

func TestJSONSchemaViolations(t *testing.T) {
	// Wrong type for name, negative age: two violations.
	errs := JSON{}.CheckSchema([]byte(`{"name":123,"age":-5}`), []byte(jsonSchema), false)
	if len(errs) != 2 {
		t.Fatalf("expected 2 schema violations, got %d: %v", len(errs), errs)
	}
}

func TestJSONSchemaMissingRequired(t *testing.T) {
	errs := JSON{}.CheckSchema([]byte(`{"name":"demo"}`), []byte(jsonSchema), false)
	if len(errs) == 0 {
		t.Fatal("expected a violation for missing required property")
	}
}

func TestJSONSchemaMalformedDocument(t *testing.T) {
	// A malformed document is reported before schema validation runs.
	errs := JSON{}.CheckSchema([]byte(`{"name":`), []byte(jsonSchema), false)
	if len(errs) != 1 {
		t.Fatalf("expected 1 syntax error, got %d: %v", len(errs), errs)
	}
}

func TestJSONSchemaInvalidSchema(t *testing.T) {
	errs := JSON{}.CheckSchema([]byte(`{"name":"demo","age":30}`), []byte(`{not valid json`), false)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for invalid schema, got %d: %v", len(errs), errs)
	}
}
