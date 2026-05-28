package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

func main() {
	os.Exit(run())
}

func run() int {
	var (
		file    string
		typ     string
		format  string
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
	flag.BoolVar(&strict, "strict", false, "enable stricter checks where supported")
	flag.BoolVar(&quiet, "quiet", false, "no output, exit code only")
	flag.BoolVar(&showVer, "version", false, "print version and exit")
	flag.Parse()

	if showVer {
		fmt.Printf("checker %s (commit %s, built %s)\n", version, commit, buildDate)
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

	if typ == "" {
		typ = detectType(file)
		if typ == "" {
			fmt.Fprintf(os.Stderr, "error: cannot auto-detect type for %q, use --type\n", file)
			return exitInternal
		}
	}

	v, err := validatorFor(typ)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitInternal
	}

	if !quiet && strings.ToLower(typ) == "sql:ansi" {
		fmt.Fprintln(os.Stderr, "warning: validated using PostgreSQL grammar (closest to ANSI standard)")
	}

	errs := v.Check(data, strict)
	res := result.CheckResult{
		File:   file,
		Type:   typ,
		Valid:  len(errs) == 0,
		Errors: errs,
	}

	if !quiet {
		printResult(res, format)
	}
	if res.Valid {
		return exitOK
	}
	return exitInvalid
}

// detectType maps a file extension to a checker type. Returns "" when unknown.
func detectType(file string) string {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".json":
		return "json"
	case ".xml", ".xhtml":
		return "xml"
	case ".yml", ".yaml":
		return "yaml"
	case ".sql":
		return "sql:ansi"
	default:
		return ""
	}
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
	if res.Valid {
		fmt.Printf("OK %s — valid %s\n", res.File, res.Type)
		return
	}
	fmt.Printf("FAIL %s — %d error(s) found\n", res.File, len(res.Errors))
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
