package checkers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSchemaSamples is a self-discovering end-to-end test over the
// test-samples/schema directory. Each document file is paired with its sibling
// <topic>.schema.json (topic = the name up to the first dot) and validated
// through the format-appropriate SchemaValidator. A file whose name contains
// "_not_correct" must produce at least one violation; every other document
// must validate cleanly. Adding a sample file is enough to extend the test.
func TestSchemaSamples(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "test-samples", "schema")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("schema samples not available: %v", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".schema.json") {
			continue // a schema, not a document under test
		}

		var v SchemaValidator
		switch filepath.Ext(name) {
		case ".json":
			v = JSON{}
		case ".yaml", ".yml":
			v = YAML{}
		default:
			continue
		}

		dot := strings.IndexByte(name, '.')
		if dot < 0 {
			t.Errorf("%s: expected a <topic>.<variant> name", name)
			continue
		}
		topic := name[:dot]

		doc, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("cannot read %s: %v", name, err)
		}
		schema, err := os.ReadFile(filepath.Join(dir, topic+".schema.json"))
		if err != nil {
			t.Fatalf("cannot read schema for %s: %v", name, err)
		}

		wantInvalid := strings.Contains(name, "_not_correct")
		t.Run(name, func(t *testing.T) {
			errs := v.CheckSchema(doc, schema, true)
			switch {
			case wantInvalid && len(errs) == 0:
				t.Errorf("expected schema violations, got none")
			case !wantInvalid && len(errs) != 0:
				t.Errorf("expected a conforming document, got %d error(s): %v", len(errs), errs)
			}
		})
	}
}
