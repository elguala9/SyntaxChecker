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
