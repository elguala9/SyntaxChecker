package checkers

import "testing"

func TestComposeValid(t *testing.T) {
	src := `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
volumes:
  data:
`
	if errs := (Compose{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid compose file, got %v", errs)
	}
}

func TestComposeSchemaViolation(t *testing.T) {
	// "services" must be a mapping, not a list: a structural Compose violation.
	src := "services:\n  - image: nginx\n"
	if errs := (Compose{}).Check([]byte(src), false); len(errs) == 0 {
		t.Fatal("expected a schema violation for services-as-list")
	}
}

func TestComposeUnknownTopLevelKey(t *testing.T) {
	// The Compose schema is closed at the top level (additionalProperties: false).
	src := "services:\n  web:\n    image: nginx\nbogus: true\n"
	if errs := (Compose{}).Check([]byte(src), false); len(errs) == 0 {
		t.Fatal("expected a violation for an unknown top-level key")
	}
}

func TestComposeMalformedYAML(t *testing.T) {
	// A YAML well-formedness error must be reported before schema validation,
	// with the line number preserved.
	errs := Compose{}.Check([]byte("services:\n\tweb: {}\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected a YAML syntax error for tab indentation")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestComposeDetection(t *testing.T) {
	// Compose files resolve by exact filename or the docker-/podman-compose prefix.
	compose := []string{
		"compose.yaml", "compose.yml",
		"compose.override.yaml", "compose.override.yml",
		"docker-compose.yml", "docker-compose.yaml",
		"docker-compose.prod.yml", "podman-compose.yml",
	}
	for _, name := range compose {
		if got := DetectByPath(name, nil); got != "compose" {
			t.Errorf("DetectByPath(%q) = %q, want compose", name, got)
		}
	}

	// A plain YAML file must stay "yaml", and a Go file whose name starts with
	// "compose" must not be mistaken for a Compose file.
	notCompose := map[string]string{
		"config.yml":    "yaml",
		"values.yaml":   "yaml",
		"compose.go":    "go",
		"composer.json": "json",
	}
	for name, want := range notCompose {
		if got := DetectByPath(name, nil); got != want {
			t.Errorf("DetectByPath(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestComposeAliases(t *testing.T) {
	for _, alias := range []string{"compose", "docker-compose", "podman-compose"} {
		if _, ok := Lookup(alias); !ok {
			t.Errorf("Lookup(%q) failed; alias not registered", alias)
		}
	}
}
