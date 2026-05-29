package checkers

import "testing"

func TestGoValid(t *testing.T) {
	src := "package main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n"
	if errs := (Go{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestGoErrorPosition(t *testing.T) {
	src := "package main\n\nfunc main() {\n\tprintln(\"hi\"\n}\n"
	errs := Go{}.Check([]byte(src), false)
	if len(errs) == 0 {
		t.Fatal("expected an error")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestTypeScriptValid(t *testing.T) {
	src := "const n: number = 1;\nfunction f(x: string): string { return x; }\n"
	if errs := (JS{Dialect: "ts"}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestTypeScriptErrorPosition(t *testing.T) {
	errs := JS{Dialect: "ts"}.Check([]byte("function f(a: number {\n  return a;\n}\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected an error")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestTSXValid(t *testing.T) {
	src := "const el = <div className=\"x\">{label}</div>;\n"
	if errs := (JS{Dialect: "tsx"}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid tsx, got %v", errs)
	}
}

func TestJavaScriptValid(t *testing.T) {
	if errs := (JS{Dialect: "js"}).Check([]byte("const xs = [1, 2].map((x) => x * 2);\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestLuaValid(t *testing.T) {
	if errs := (Lua{}).Check([]byte("local x = 1\nprint(x + 2)\n"), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestLuaErrorPosition(t *testing.T) {
	// A mid-file token error carries a real position (unlike EOF errors).
	errs := Lua{}.Check([]byte("local x = = 1\nprint(x)\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected an error")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestLuaMissingEnd(t *testing.T) {
	// EOF errors have no usable position, but must still be reported.
	if errs := (Lua{}).Check([]byte("if x then\n  print(1)\n"), false); len(errs) == 0 {
		t.Fatal("expected an error for the missing end")
	}
}

func TestShellValid(t *testing.T) {
	src := "for i in 1 2 3; do\n  echo \"$i\"\ndone\n"
	if errs := (Shell{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestShellErrorPosition(t *testing.T) {
	errs := Shell{}.Check([]byte("if true; then\n  echo hi\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected an error for the unterminated if")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}

func TestStarlarkValid(t *testing.T) {
	src := "def f(x):\n  return x + 1\n\ny = f(2)\n"
	if errs := (Starlark{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestStarlarkErrorPosition(t *testing.T) {
	errs := Starlark{}.Check([]byte("xs = [1, 2,\ny = 3\n"), false)
	if len(errs) == 0 {
		t.Fatal("expected an error for the unclosed list")
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a line number, got %+v", errs[0])
	}
}
