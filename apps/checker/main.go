package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/parresia/syntaxchecker/apps/checker/checkers"
	"github.com/parresia/syntaxchecker/pkg/result"
)

// Set via -ldflags at build time.
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// Exit codes per the CLI contract.
const (
	exitOK       = 0 // valid
	exitInvalid  = 1 // syntax errors found
	exitInternal = 2 // file not found / internal error
)

var (
	reMSSQLGoBatch = regexp.MustCompile(`(?im)^\s*GO\s*$`)
	reOracleMinus  = regexp.MustCompile(`(?i)\bMINUS\b`)
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		file    string
		typ     string
		format  string
		schema  string
		strict  bool
		quiet   bool
		showVer bool
	)
	flag.StringVar(&file, "file", "", "path of the file to check (required)")
	flag.StringVar(&file, "f", "", "shorthand for --file")
	flag.StringVar(&typ, "type", "", "force type, e.g. sql:mysql, json (default: auto-detect from extension)")
	flag.StringVar(&typ, "t", "", "shorthand for --type")
	flag.StringVar(&format, "format", "text", "output format: text | json | json-pretty")
	flag.StringVar(&format, "o", "text", "shorthand for --format")
	flag.StringVar(&schema, "schema", "", "path of a JSON Schema to validate against (json, yaml)")
	flag.StringVar(&schema, "s", "", "shorthand for --schema")
	flag.BoolVar(&strict, "strict", false, "enable stricter checks where supported")
	flag.BoolVar(&quiet, "quiet", false, "no output, exit code only")
	flag.BoolVar(&showVer, "version", false, "print version and exit")
	flag.Parse()

	if showVer {
		fmt.Printf("syntax-checker %s (commit %s, built %s)\n", version, commit, buildDate)
		return exitOK
	}

	if file == "" {
		fmt.Fprintln(os.Stderr, "error: --file is required")
		return exitInternal
	}

	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitInternal
	}

	autoDetected := false
	if typ == "" {
		typ = detectType(file, data)
		if typ == "" {
			fmt.Fprintf(os.Stderr, "error: cannot auto-detect type for %q, use --type\n", file)
			return exitInternal
		}
		autoDetected = true
	}

	v, err := validatorFor(typ)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitInternal
	}

	if !quiet && strings.ToLower(typ) == "sql:ansi" {
		fmt.Fprintln(os.Stderr, "warning: validated using PostgreSQL grammar (closest to ANSI standard)")
	}

	var errs []result.SyntaxError
	if schema != "" {
		sv, ok := v.(checkers.SchemaValidator)
		if !ok {
			fmt.Fprintf(os.Stderr, "error: schema validation is not supported for type %q\n", typ)
			return exitInternal
		}
		schemaData, err := os.ReadFile(schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot read schema: %v\n", err)
			return exitInternal
		}
		errs = sv.CheckSchema(data, schemaData, strict)
	} else {
		errs = v.Check(data, strict)
	}
	res := result.CheckResult{
		File:         file,
		Type:         typ,
		AutoDetected: autoDetected,
		Valid:        len(errs) == 0,
		Errors:       errs,
	}

	if !quiet {
		printResult(res, format)
	}
	if res.Valid {
		return exitOK
	}
	return exitInvalid
}

// detectType maps a file extension to a checker type. For .sql it also
// inspects the filename and content to guess the dialect. Returns "" when
// the extension is unknown.
func detectType(file string, data []byte) string {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".json":
		return "json"
	case ".xml", ".xhtml":
		return "xml"
	case ".yml", ".yaml":
		return "yaml"
	case ".sql":
		return detectSQLDialect(file, data)
	default:
		return ""
	}
}

// detectSQLDialect picks an SQL dialect from filename hints first and then
// from content-level token sniffing. Falls back to sql:ansi when nothing
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

func validatorFor(typ string) (checkers.Validator, error) {
	switch strings.ToLower(typ) {
	case "json":
		return checkers.JSON{}, nil
	case "sql:mysql":
		return checkers.MySQL{}, nil
	case "sql:postgres", "sql:postgresql", "sql:ansi":
		return checkers.Postgres{}, nil
	case "sql:sqlite":
		return checkers.SQLite{}, nil
	case "sql:mssql", "sql:tsql":
		return checkers.MSSQL{}, nil
	case "sql:oracle", "sql:plsql":
		return checkers.Oracle{}, nil
	case "xml":
		return checkers.XML{}, nil
	case "yaml":
		return checkers.YAML{}, nil
	default:
		return nil, fmt.Errorf("unsupported type %q", typ)
	}
}

func printResult(res result.CheckResult, format string) {
	switch format {
	case "json":
		b, _ := json.Marshal(res)
		fmt.Println(string(b))
	case "json-pretty":
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
	default:
		printText(res)
	}
}

func printText(res result.CheckResult) {
	typeLabel := res.Type
	if res.AutoDetected {
		typeLabel += " (auto-detected)"
	}
	if res.Valid {
		fmt.Printf("OK %s — valid %s\n", res.File, typeLabel)
		return
	}
	fmt.Printf("FAIL %s — %d error(s) found [%s]\n", res.File, len(res.Errors), typeLabel)
	for _, e := range res.Errors {
		switch {
		case e.Line > 0 && e.Column > 0:
			fmt.Printf("  Line %d, Col %d: %s\n", e.Line, e.Column, e.Message)
		case e.Line > 0:
			fmt.Printf("  Line %d: %s\n", e.Line, e.Message)
		default:
			fmt.Printf("  %s\n", e.Message)
		}
	}
}
