package checkers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Focused unit tests -----------------------------------------------------

func TestTOMLValid(t *testing.T) {
	if errs := (TOML{}).Check([]byte("a = 1\n[s]\nb = \"x\"\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestTOMLErrorPosition(t *testing.T) {
	errs := TOML{}.Check([]byte("title = \"unterminated\nport = 8080\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected an error")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestTOMLDuplicateKey(t *testing.T) {
	if errs := (TOML{}).Check([]byte("a = 1\na = 2\n"), false); len(errs) == 0 {
		t.Fatal("expected a duplicate-key error")
	}
}

func TestINIValid(t *testing.T) {
	if errs := (INI{}).Check([]byte("[s]\nk = v\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestINIMissingDelimiter(t *testing.T) {
	if errs := (INI{}).Check([]byte("[s]\nno delimiter here\n"), false); len(errs) == 0 {
		t.Fatal("expected an error for the line without a delimiter")
	}
}

func TestCSVValid(t *testing.T) {
	if errs := (CSV{Comma: ','}).Check([]byte("a,b\n1,2\n3,4\n"), true); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestCSVRaggedStrict(t *testing.T) {
	errs := CSV{Comma: ','}.Check([]byte("a,b,c\n1,2\n"), true)
	if len(errs) == 0 {
		t.Fatal("expected a field-count error in strict mode")
	}
	if errs[0].Line != 2 {
		t.Errorf("expected error on line 2, got %+v", errs[0])
	}
}

func TestCSVRaggedLenient(t *testing.T) {
	if errs := (CSV{Comma: ','}).Check([]byte("a,b,c\n1,2\n"), false); len(errs) != 0 {
		t.Fatalf("ragged rows must be tolerated when not strict, got %v", errs)
	}
}

func TestCSVUnterminatedQuoteLenient(t *testing.T) {
	errs := CSV{Comma: ','}.Check([]byte("id,name\n1,\"Alice\n2,Bob\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected an unterminated-quote error even when not strict")
	}
	if errs[0].Line != 2 {
		t.Errorf("expected error on line 2 where the quote opens, got %+v", errs[0])
	}
}

func TestCSVEscapedQuotesValid(t *testing.T) {
	if errs := (CSV{Comma: ','}).Check([]byte("a,b\n\"x \"\"y\"\" z\",2\n"), false); len(errs) != 0 {
		t.Fatalf("doubled quotes inside a quoted field are valid, got %v", errs)
	}
}

func TestTSVDelimiter(t *testing.T) {
	if errs := (CSV{Comma: '\t'}).Check([]byte("a\tb\n1\t2\n"), true); len(errs) != 0 {
		t.Fatalf("expected valid TSV, got %v", errs)
	}
}

func TestHCLValid(t *testing.T) {
	if errs := (HCL{}).Check([]byte("a = 1\nblock {\n  b = \"x\"\n}\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestHCLErrorPosition(t *testing.T) {
	errs := HCL{}.Check([]byte("block {\n  name =\n}\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected an error")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestMarkdownAlwaysValid(t *testing.T) {
	// Markdown has no invalid syntax; even pathological input must parse.
	if errs := (Markdown{}).Check([]byte("# h\n\n```\nunclosed fence"), false); len(errs) != 0 {
		t.Fatalf("markdown must always parse, got %v", errs)
	}
}

func TestEnvValid(t *testing.T) {
	if errs := (Env{}).Check([]byte("FOO=bar\n# c\nBAZ=qux\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestEnvBadLine(t *testing.T) {
	if errs := (Env{}).Check([]byte("FOO=bar\nNOT A PAIR\n"), false); len(errs) == 0 {
		t.Fatal("expected an error for the malformed line")
	}
}

func TestPropertiesValid(t *testing.T) {
	if errs := (Properties{}).Check([]byte("a=1\nb = ${a}\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestPropertiesCircular(t *testing.T) {
	if errs := (Properties{}).Check([]byte("a=${b}\nb=${a}\n"), false); len(errs) == 0 {
		t.Fatal("expected a circular-reference error")
	}
}

func TestProtobufValid(t *testing.T) {
	src := "syntax = \"proto3\";\nmessage M {\n  string id = 1;\n}\n"
	if errs := (Protobuf{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestProtobufErrorPosition(t *testing.T) {
	src := "syntax = \"proto3\";\nmessage M {\n  id = 1;\n}\n"
	errs := Protobuf{}.Check([]byte(src), false)
	if len(errs) == 0 {
		t.Fatal("expected an error")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestGraphQLSchemaValid(t *testing.T) {
	if errs := (GraphQL{}).Check([]byte("type Query {\n  hello: String\n}\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid schema, got %v", errs)
	}
}

func TestGraphQLQueryValid(t *testing.T) {
	if errs := (GraphQL{}).Check([]byte("{ user(id: 1) { name } }\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid query, got %v", errs)
	}
}

func TestGraphQLErrorPosition(t *testing.T) {
	errs := GraphQL{}.Check([]byte("query { user(id: 1) { name }\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected an error for the unbalanced braces")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

// --- Self-discovering end-to-end test over test-samples ---------------------

// validatorByExt returns the Validator for a sample file extension, and whether
// the extension is one of the simple-parser formats covered here.
func validatorByExt(ext string) (Validator, bool) {
	switch ext {
	case ".toml":
		return TOML{}, true
	case ".ini", ".cfg":
		return INI{}, true
	case ".csv":
		return CSV{Comma: ','}, true
	case ".tsv":
		return CSV{Comma: '\t'}, true
	case ".hcl", ".tf":
		return HCL{}, true
	case ".md", ".markdown":
		return Markdown{}, true
	case ".env":
		return Env{}, true
	case ".properties":
		return Properties{}, true
	case ".proto":
		return Protobuf{}, true
	case ".graphql", ".gql":
		return GraphQL{}, true
	default:
		return nil, false
	}
}

// TestParserSamples walks test-samples and validates every file whose extension
// maps to one of the simple parsers. A file whose name contains "_not_correct"
// must produce at least one error; every other file must validate cleanly.
// Strict mode is used so format-specific strict checks are exercised.
func TestParserSamples(t *testing.T) {
	root := filepath.Join("..", "..", "..", "test-samples")
	dirs := []string{"toml", "ini", "csv", "hcl", "markdown", "env", "properties", "proto", "graphql"}

	for _, d := range dirs {
		entries, err := os.ReadDir(filepath.Join(root, d))
		if err != nil {
			t.Skipf("samples for %s not available: %v", d, err)
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			v, ok := validatorByExt(strings.ToLower(filepath.Ext(name)))
			if !ok {
				continue
			}
			data, err := os.ReadFile(filepath.Join(root, d, name))
			if err != nil {
				t.Fatalf("cannot read %s: %v", name, err)
			}
			wantInvalid := strings.Contains(name, "_not_correct")
			t.Run(filepath.Join(d, name), func(t *testing.T) {
				errs := v.Check(data, true)
				switch {
				case wantInvalid && len(errs) == 0:
					t.Errorf("expected syntax errors, got none")
				case !wantInvalid && len(errs) != 0:
					t.Errorf("expected a valid file, got %d error(s): %v", len(errs), errs)
				}
			})
		}
	}
}
