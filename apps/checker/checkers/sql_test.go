package checkers

import "testing"

func TestSQLValid(t *testing.T) {
	cases := map[string]struct {
		v   Validator
		sql string
	}{
		"mysql":    {MySQL{}, "SELECT id, name FROM users WHERE age >= 18;"},
		"postgres": {Postgres{}, "SELECT id, name FROM users WHERE age >= 18;"},
		"sqlite":   {SQLite{}, "SELECT id, name FROM users WHERE age >= 18;"},
		"mssql":    {MSSQL{}, "SELECT TOP 5 id, name FROM dbo.users ORDER BY name;"},
		"oracle":   {Oracle{}, "select id, name from employees where dept_id = 10;"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if errs := c.v.Check([]byte(c.sql), false); len(errs) != 0 {
				t.Fatalf("expected valid, got %v", errs)
			}
		})
	}
}

func TestMSSQLMissingCommaInTableElements(t *testing.T) {
	cases := map[string]string{
		"constraint then column": "CREATE TABLE t (\n  a BIGINT NOT NULL,\n  CONSTRAINT uq UNIQUE (a)\n  b BIGINT NULL\n);",
		"column then column":     "CREATE TABLE t (\n  a BIGINT NOT NULL\n  b BIGINT NULL\n);",
		"declare table variable": "DECLARE @t TABLE (\n  a INT\n  b INT\n);",
	}
	for name, sql := range cases {
		t.Run(name, func(t *testing.T) {
			errs := MSSQL{}.Check([]byte(sql), false)
			if len(errs) == 0 {
				t.Fatalf("expected a missing-comma error, got none")
			}
			if errs[0].Line == 0 || errs[0].Column == 0 {
				t.Errorf("expected a line/column, got %+v", errs[0])
			}
		})
	}
}

func TestMSSQLCommaSeparatedTableStillValid(t *testing.T) {
	sql := "CREATE TABLE t (\n  a BIGINT NOT NULL,\n  b BIGINT NULL,\n  CONSTRAINT uq UNIQUE (a, b)\n);"
	if errs := (MSSQL{}).Check([]byte(sql), false); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestSQLiteMissingCommaInTableElements(t *testing.T) {
	cases := map[string]string{
		"column then column":     "CREATE TABLE t (\n  a INT NOT NULL\n  b INT\n);",
		"column then constraint": "CREATE TABLE t (\n  a INT\n  CONSTRAINT uq UNIQUE (a)\n);",
		"single line columns":    "CREATE TABLE t ( a INT NOT NULL  b INT );",
	}
	for name, sql := range cases {
		t.Run(name, func(t *testing.T) {
			errs := SQLite{}.Check([]byte(sql), false)
			if len(errs) == 0 {
				t.Fatalf("expected a missing-comma error, got none")
			}
			if errs[0].Line == 0 || errs[0].Column == 0 {
				t.Errorf("expected a line/column, got %+v", errs[0])
			}
		})
	}
}

func TestSQLiteCommaEdgeCasesStillValid(t *testing.T) {
	// SQLite type names may be several words (e.g. UNSIGNED BIG INT), so a column
	// with no explicit constraint legitimately spans multiple identifiers.
	cases := map[string]string{
		"multi-word type":    "CREATE TABLE t ( a INT b INT );",
		"comma separated":    "CREATE TABLE t (\n  a INT NOT NULL,\n  b INT,\n  CONSTRAINT uq UNIQUE (a, b)\n);",
		"nested paren comma": "CREATE TABLE t (\n  a INT,\n  b INT,\n  PRIMARY KEY (a, b)\n);",
	}
	for name, sql := range cases {
		t.Run(name, func(t *testing.T) {
			if errs := (SQLite{}).Check([]byte(sql), false); len(errs) != 0 {
				t.Fatalf("expected valid, got %v", errs)
			}
		})
	}
}

func TestSQLInvalidReportsPosition(t *testing.T) {
	cases := map[string]struct {
		v   Validator
		sql string
	}{
		"mysql":    {MySQL{}, "SELECT * FORM users;"},
		"postgres": {Postgres{}, "SELECT * FORM users;"},
		"sqlite":   {SQLite{}, "SELECT FROM WHERE;"},
		"mssql":    {MSSQL{}, "SELECT name FROM WHERE id = 1;"},
		"oracle":   {Oracle{}, "SELECT FROM employees WERE id = 1;"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			errs := c.v.Check([]byte(c.sql), false)
			if len(errs) == 0 {
				t.Fatalf("expected errors, got none")
			}
			if errs[0].Line == 0 {
				t.Errorf("expected a line number, got %+v", errs[0])
			}
			if errs[0].Column == 0 {
				t.Errorf("expected a column number, got %+v", errs[0])
			}
		})
	}
}
