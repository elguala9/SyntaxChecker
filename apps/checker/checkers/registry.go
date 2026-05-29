package checkers

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Registration describes one checker: its canonical type name, the aliases and
// file extensions that resolve to it, optional exact/prefix filename matches,
// and a factory that builds a fresh Validator.
//
// The registry is the single source of truth that the CLI (type resolution and
// extension auto-detection) and the tests all derive from. Adding a new
// language means adding one Registration to the table in init below plus its
// checker file — nothing else in this module needs to change.
type Registration struct {
	// Type is the canonical type name, e.g. "json" or "sql:mysql".
	Type string
	// Aliases are additional names accepted by Lookup, e.g. "golang" for "go".
	Aliases []string
	// Extensions are file extensions (with the leading dot, lowercase) that
	// auto-detect to this type, e.g. ".yml".
	Extensions []string
	// Filenames are exact base names (lowercase) that auto-detect to this type,
	// e.g. "dockerfile" or "build.bazel".
	Filenames []string
	// FilenamePrefixes are base-name prefixes (lowercase) that auto-detect to
	// this type, e.g. "dockerfile." matching "Dockerfile.prod".
	FilenamePrefixes []string
	// New builds a fresh Validator. It is nil only for a family entry whose sole
	// purpose is auto-detection (see Detect).
	New func() Validator
	// Detect resolves a content-dependent subtype for an extension that maps to
	// a family rather than to a single type (currently only ".sql"). When set,
	// New is nil; the resolved subtypes are registered separately with their own
	// New factory.
	Detect func(file string, data []byte) string
}

var (
	registrations []*Registration
	byType        = map[string]*Registration{} // canonical types and aliases
	byExt         = map[string]*Registration{}
	byFilename    = map[string]*Registration{}
)

// Register adds r to the registry. It panics on a duplicate type, alias,
// extension, or filename: those are programming errors detectable at startup.
func Register(r Registration) {
	rp := &r
	registrations = append(registrations, rp)
	addKey(byType, "type", strings.ToLower(rp.Type), rp)
	for _, a := range rp.Aliases {
		addKey(byType, "alias", strings.ToLower(a), rp)
	}
	for _, e := range rp.Extensions {
		addKey(byExt, "extension", strings.ToLower(e), rp)
	}
	for _, f := range rp.Filenames {
		addKey(byFilename, "filename", strings.ToLower(f), rp)
	}
}

func addKey(m map[string]*Registration, kind, key string, rp *Registration) {
	if key == "" {
		return
	}
	if _, dup := m[key]; dup {
		panic(fmt.Sprintf("checkers: duplicate %s %q in registry", kind, key))
	}
	m[key] = rp
}

// Lookup returns a fresh Validator for the given type name or alias
// (case-insensitive). ok is false for an unknown type.
func Lookup(typ string) (Validator, bool) {
	rp, ok := byType[strings.ToLower(typ)]
	if !ok || rp.New == nil {
		return nil, false
	}
	return rp.New(), true
}

// DetectByPath resolves a checker type from a file path, using the file
// contents only where a family extension (".sql") needs to sniff a dialect. It
// returns "" when the path matches no registered checker. Exact filenames and
// filename prefixes take precedence over extensions, since a file literally
// named "Dockerfile" has no extension.
func DetectByPath(file string, data []byte) string {
	base := strings.ToLower(filepath.Base(file))
	if rp, ok := byFilename[base]; ok {
		return rp.Type
	}
	for _, rp := range registrations {
		for _, p := range rp.FilenamePrefixes {
			if strings.HasPrefix(base, p) {
				return rp.Type
			}
		}
	}
	if rp, ok := byExt[strings.ToLower(filepath.Ext(file))]; ok {
		if rp.Detect != nil {
			return rp.Detect(file, data)
		}
		return rp.Type
	}
	return ""
}

// Types returns the sorted canonical type names that have a Validator (i.e.
// excluding family-only detection entries such as the ".sql" aggregator).
func Types() []string {
	out := make([]string, 0, len(registrations))
	for _, rp := range registrations {
		if rp.New != nil {
			out = append(out, rp.Type)
		}
	}
	sort.Strings(out)
	return out
}

func init() {
	Register(Registration{Type: "json", Extensions: []string{".json"}, New: func() Validator { return JSON{} }})
	Register(Registration{Type: "json5", Extensions: []string{".json5"}, New: func() Validator { return JSON5{} }})
	Register(Registration{Type: "jsonc", Extensions: []string{".jsonc"}, New: func() Validator { return JSONC{} }})
	Register(Registration{Type: "proto", Aliases: []string{"protobuf"}, Extensions: []string{".proto"}, New: func() Validator { return Protobuf{} }})
	Register(Registration{Type: "graphql", Aliases: []string{"gql"}, Extensions: []string{".graphql", ".gql"}, New: func() Validator { return GraphQL{} }})
	Register(Registration{Type: "xml", Extensions: []string{".xml", ".xhtml"}, New: func() Validator { return XML{} }})
	Register(Registration{Type: "html", Aliases: []string{"htm"}, Extensions: []string{".html", ".htm"}, New: func() Validator { return HTML{} }})
	Register(Registration{Type: "jq", Extensions: []string{".jq"}, New: func() Validator { return JQ{} }})
	Register(Registration{Type: "go", Aliases: []string{"golang"}, Extensions: []string{".go"}, New: func() Validator { return Go{} }})
	Register(Registration{Type: "ts", Aliases: []string{"typescript"}, Extensions: []string{".ts"}, New: func() Validator { return JS{Dialect: "ts"} }})
	Register(Registration{Type: "tsx", Extensions: []string{".tsx"}, New: func() Validator { return JS{Dialect: "tsx"} }})
	Register(Registration{Type: "js", Aliases: []string{"javascript"}, Extensions: []string{".js", ".mjs", ".cjs"}, New: func() Validator { return JS{Dialect: "js"} }})
	Register(Registration{Type: "jsx", Extensions: []string{".jsx"}, New: func() Validator { return JS{Dialect: "jsx"} }})
	Register(Registration{Type: "lua", Extensions: []string{".lua"}, New: func() Validator { return Lua{} }})
	Register(Registration{Type: "shell", Aliases: []string{"bash", "sh"}, Extensions: []string{".sh", ".bash"}, New: func() Validator { return Shell{} }})
	Register(Registration{Type: "starlark", Aliases: []string{"bzl"}, Extensions: []string{".star", ".bzl"}, Filenames: []string{"build.bazel", "workspace.bazel"}, New: func() Validator { return Starlark{} }})
	Register(Registration{Type: "yaml", Extensions: []string{".yml", ".yaml"}, New: func() Validator { return YAML{} }})
	Register(Registration{Type: "toml", Extensions: []string{".toml"}, New: func() Validator { return TOML{} }})
	Register(Registration{Type: "ini", Extensions: []string{".ini", ".cfg"}, New: func() Validator { return INI{} }})
	Register(Registration{Type: "csv", Extensions: []string{".csv"}, New: func() Validator { return CSV{Comma: ','} }})
	Register(Registration{Type: "tsv", Extensions: []string{".tsv"}, New: func() Validator { return CSV{Comma: '\t'} }})
	Register(Registration{Type: "hcl", Extensions: []string{".hcl", ".tf"}, New: func() Validator { return HCL{} }})
	Register(Registration{Type: "markdown", Extensions: []string{".md", ".markdown"}, New: func() Validator { return Markdown{} }})
	Register(Registration{Type: "env", Extensions: []string{".env"}, New: func() Validator { return Env{} }})
	Register(Registration{Type: "properties", Extensions: []string{".properties"}, New: func() Validator { return Properties{} }})
	Register(Registration{
		Type:             "dockerfile",
		Filenames:        []string{"dockerfile", "containerfile"},
		FilenamePrefixes: []string{"dockerfile."},
		Extensions:       []string{".dockerfile"},
		New:              func() Validator { return Dockerfile{} },
	})

	// SQL dialects. Each is a first-class type; the ".sql" extension maps to a
	// family entry that sniffs the dialect from the filename and contents.
	Register(Registration{Type: "sql:mysql", New: func() Validator { return MySQL{} }})
	Register(Registration{Type: "sql:postgres", Aliases: []string{"sql:postgresql", "sql:ansi"}, New: func() Validator { return Postgres{} }})
	Register(Registration{Type: "sql:sqlite", New: func() Validator { return SQLite{} }})
	Register(Registration{Type: "sql:mssql", Aliases: []string{"sql:tsql"}, New: func() Validator { return MSSQL{} }})
	Register(Registration{Type: "sql:oracle", Aliases: []string{"sql:plsql"}, New: func() Validator { return Oracle{} }})
	Register(Registration{Type: "sql", Extensions: []string{".sql"}, Detect: detectSQLDialect})
}

var (
	reMSSQLGoBatch = regexp.MustCompile(`(?im)^\s*GO\s*$`)
	reOracleMinus  = regexp.MustCompile(`(?i)\bMINUS\b`)
)

// detectSQLDialect picks an SQL dialect from filename hints first and then from
// content-level token sniffing. It falls back to sql:ansi when nothing
// distinctive is found.
func detectSQLDialect(file string, data []byte) string {
	base := strings.ToLower(filepath.Base(file))
	switch {
	case strings.Contains(base, "mysql"), strings.Contains(base, "mariadb"):
		return "sql:mysql"
	case strings.Contains(base, "postgres"), strings.Contains(base, "postgresql"),
		strings.Contains(base, "_pg."), strings.Contains(base, "_pg_"),
		strings.HasPrefix(base, "pg_"):
		return "sql:postgres"
	case strings.Contains(base, "mssql"), strings.Contains(base, "tsql"),
		strings.Contains(base, "sqlserver"), strings.Contains(base, "sql_server"):
		return "sql:mssql"
	case strings.Contains(base, "oracle"), strings.Contains(base, "plsql"):
		return "sql:oracle"
	case strings.Contains(base, "sqlite"):
		return "sql:sqlite"
	}

	text := string(data)
	upper := strings.ToUpper(text)
	scores := map[string]int{}
	addIf := func(d, token string, weight int) {
		if strings.Contains(upper, token) {
			scores[d] += weight
		}
	}

	// Postgres
	addIf("sql:postgres", "::", 1)
	addIf("sql:postgres", "RETURNING", 2)
	addIf("sql:postgres", "JSONB", 3)
	addIf("sql:postgres", " SERIAL", 2)
	addIf("sql:postgres", "BIGSERIAL", 3)
	addIf("sql:postgres", "ILIKE", 3)
	addIf("sql:postgres", "DISTINCT ON", 3)
	if strings.Contains(text, "$$") {
		scores["sql:postgres"] += 3
	}

	// MySQL
	if strings.Contains(text, "`") {
		scores["sql:mysql"] += 3
	}
	addIf("sql:mysql", "AUTO_INCREMENT", 3)
	addIf("sql:mysql", "ENGINE=", 3)
	addIf("sql:mysql", "ON DUPLICATE KEY", 3)
	addIf("sql:mysql", "STRAIGHT_JOIN", 3)
	addIf("sql:mysql", "UNSIGNED", 1)

	// MSSQL / T-SQL
	addIf("sql:mssql", "NVARCHAR", 2)
	addIf("sql:mssql", "@@IDENTITY", 3)
	addIf("sql:mssql", "@@ROWCOUNT", 3)
	addIf("sql:mssql", "OUTPUT INSERTED.", 3)
	addIf("sql:mssql", "GETDATE()", 2)
	addIf("sql:mssql", "SELECT TOP ", 2)
	if reMSSQLGoBatch.MatchString(text) {
		scores["sql:mssql"] += 3
	}

	// Oracle / PL/SQL
	addIf("sql:oracle", "CONNECT BY", 3)
	addIf("sql:oracle", "NVL(", 2)
	addIf("sql:oracle", "SYSDATE", 2)
	addIf("sql:oracle", "FROM DUAL", 3)
	addIf("sql:oracle", "VARCHAR2", 3)
	addIf("sql:oracle", "NUMBER(", 1)
	if reOracleMinus.MatchString(text) {
		scores["sql:oracle"] += 2
	}

	// SQLite
	addIf("sql:sqlite", "AUTOINCREMENT", 3)
	addIf("sql:sqlite", "PRAGMA ", 3)
	addIf("sql:sqlite", "WITHOUT ROWID", 3)

	best := "sql:ansi"
	bestScore := 0
	// Priority order for ties.
	for _, d := range []string{"sql:postgres", "sql:mysql", "sql:mssql", "sql:oracle", "sql:sqlite"} {
		if scores[d] > bestScore {
			best = d
			bestScore = scores[d]
		}
	}
	return best
}
