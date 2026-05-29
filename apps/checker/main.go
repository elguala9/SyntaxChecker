package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
		typ = checkers.DetectByPath(file, data)
		if typ == "" {
			fmt.Fprintf(os.Stderr, "error: cannot auto-detect type for %q, use --type\n", file)
			return exitInternal
		}
		autoDetected = true
	}

	v, ok := checkers.Lookup(typ)
	if !ok {
		fmt.Fprintf(os.Stderr, "error: unsupported type %q\n", typ)
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
