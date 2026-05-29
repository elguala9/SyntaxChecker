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

func TestGoGenericsValid(t *testing.T) {
	src := "package p\n\nfunc Map[T, U any](xs []T, f func(T) U) []U {\n\tout := make([]U, len(xs))\n\tfor i, x := range xs {\n\t\tout[i] = f(x)\n\t}\n\treturn out\n}\n"
	if errs := (Go{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid generics, got %v", errs)
	}
}

func TestGoMissingPackage(t *testing.T) {
	// A Go file with no package clause must be rejected.
	if errs := (Go{}).Check([]byte("import \"fmt\"\n\nfunc main() {}\n"), false); len(errs) == 0 {
		t.Fatal("expected an error for the missing package clause")
	}
}

func TestTypeScriptValid(t *testing.T) {
	src := "const n: number = 1;\nfunction f(x: string): string { return x; }\n"
	if errs := (JS{Dialect: "ts"}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestTypeScriptGenericsValid(t *testing.T) {
	src := "class Box<T> { constructor(private v: T) {} get(): T { return this.v; } }\n"
	if errs := (JS{Dialect: "ts"}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid generics, got %v", errs)
	}
}

func TestTypeScriptMalformedEnum(t *testing.T) {
	if errs := (JS{Dialect: "ts"}).Check([]byte("enum E {\n  A\n  B\n}\n"), false); len(errs) == 0 {
		t.Fatal("expected an error for the missing comma between enum members")
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

func TestJavaScriptClassValid(t *testing.T) {
	src := "class C {\n  #x = 1;\n  async run() { return await Promise.resolve(this.#x); }\n}\n"
	if errs := (JS{Dialect: "js"}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid class, got %v", errs)
	}
}

func TestJavaScriptDuplicateParam(t *testing.T) {
	// Duplicate parameter names are a syntax error in strict/module code.
	if errs := (JS{Dialect: "js"}).Check([]byte("\"use strict\";\nfunction f(a, a) { return a; }\n"), false); len(errs) == 0 {
		t.Fatal("expected an error for the duplicate parameter")
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

func TestLuaTablesValid(t *testing.T) {
	src := "local t = setmetatable({ x = 1 }, { __index = function() return 0 end })\nprint(t.x, t.y)\n"
	if errs := (Lua{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid tables, got %v", errs)
	}
}

func TestShellValid(t *testing.T) {
	src := "for i in 1 2 3; do\n  echo \"$i\"\ndone\n"
	if errs := (Shell{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestShellCaseValid(t *testing.T) {
	src := "case \"$1\" in\n  a) echo one ;;\n  b|c) echo two ;;\n  *) echo other ;;\nesac\n"
	if errs := (Shell{}).Check([]byte(src), false); len(errs) != 0 {
		t.Fatalf("expected valid case, got %v", errs)
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
