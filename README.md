# SyntaxChecker

Multi-format syntax validator, shipped as a CLI (`syntax-checker`) and as an MCP server (`syntaxchecker-mcp`) for integration with AI assistants (Claude Desktop, VS Code MCP, etc.).

## Supported formats

| Kind | `--type` | Notes |
|---|---|---|
| JSON | `json` | `--strict` detects duplicate keys; `--schema` validates against a JSON Schema |
| XML | `xml` | Well-formedness only (no DTD/XSD validation — see below) |
| YAML | `yaml` | `.yml` and `.yaml`; `--schema` validates against a JSON Schema |
| TOML | `toml` | `.toml`; reports line/column, including duplicate keys |
| INI | `ini` | `.ini`, `.cfg` |
| CSV | `csv` | `.csv`; `--strict` enforces a consistent field count per row |
| TSV | `tsv` | `.tsv`; tab-delimited, same checks as CSV |
| HCL | `hcl` | `.hcl`, `.tf` (Terraform); well-formedness only |
| Markdown | `markdown` | `.md`, `.markdown`; parse-only (Markdown has no invalid syntax) |
| .env | `env` | `.env`; KEY=VALUE assignments |
| Properties | `properties` | `.properties`; reports malformed `\uXXXX` escapes and circular `${...}` refs |
| Protobuf | `proto` | `.proto`; non-permissive well-formedness (no import/type resolution) |
| SQL MySQL | `sql:mysql` | TiDB parser |
| SQL PostgreSQL | `sql:postgres` / `sql:postgresql` | Official parser via WASM |
| SQL ANSI | `sql:ansi` | Mapped to the PostgreSQL parser |
| SQL SQLite | `sql:sqlite` | rqlite/sql |
| SQL SQL Server | `sql:mssql` / `sql:tsql` | ANTLR4 |
| SQL Oracle | `sql:oracle` / `sql:plsql` | ANTLR4 |

## Layout

```
SyntaxChecker/
├── apps/
│   ├── checker/       # CLI: dist/syntax-checker(.exe)
│   └── mcp-server/    # MCP server: dist/syntaxchecker-mcp(.exe)
├── pkg/result/        # Shared types (CheckResult, SyntaxError)
├── test-samples/      # Valid and invalid sample files (incl. schema/ for JSON Schema)
├── scripts/           # check-samples.ps1 (end-to-end tests)
└── Makefile
```

## Build

Requires Go 1.22+ (workspace already configured in `go.work`).

```bash
make build           # produces dist/syntax-checker(.exe) and dist/syntaxchecker-mcp(.exe)
make test            # runs the unit tests
make build-windows   # cross-build for Windows
make build-linux     # cross-build for Linux
```

Binaries built with `CGO_ENABLED=0`, `-trimpath`, `-ldflags "-s -w"`.

## CLI usage

```bash
syntax-checker --file <path> [--type <type>] [--schema <path>] [--format text|json|json-pretty] [--strict] [--quiet]
```

Examples:

```bash
syntax-checker -f config.json
syntax-checker -f query.sql -t sql:mysql
syntax-checker -f data.json --strict -o json-pretty
syntax-checker -f data.json --schema schema.json     # validate JSON against a JSON Schema
syntax-checker -f config.yaml --schema schema.json    # validate YAML against a JSON Schema
```

Exit codes: `0` valid, `1` syntax errors, `2` internal error / file not found.

### Schema validation

`--schema` validates the document against a [JSON Schema](https://json-schema.org/)
(drafts up to 2020-12, including draft-07, via `santhosh-tekuri/jsonschema`). It
is supported for `json` and `yaml` (YAML maps onto the same data model as JSON).
Schema violations are reported by their instance location (a JSON pointer such
as `/items/0/name`) rather than a source line, since validation runs on the
decoded value.

`format` keywords (e.g. `email`, `date`) are treated as annotations only, per
the JSON Schema 2020-12 default — they do not cause validation failures.
Structural keywords (`type`, `required`, `enum`, `const`, `pattern`,
`minimum`/`maximum`, `minItems`, `uniqueItems`, `additionalProperties`,
`oneOf`/`anyOf`, `$ref`, …) are fully enforced.

XSD validation for XML is **not supported**: there is no mature pure-Go XSD
validator, and the only production-grade option (cgo + libxml2) would require a
native library, breaking the self-contained static binary. Requesting `--schema`
for an XML file returns an error.

Sample schemas and documents (valid and invalid) live under
[`test-samples/schema/`](test-samples/schema). Each document `<topic>.<variant>.<ext>`
is paired with its sibling `<topic>.schema.json`; files whose name contains
`_not_correct` are expected to fail. They are exercised by the Go test
`apps/checker/checkers/schema_samples_test.go` and end-to-end by
`scripts/check-samples.ps1`.

## MCP usage

Exposes the `check_syntax(file_path, type?, schema?, strict?)` tool over stdio transport.

See [`apps/mcp-server/README.md`](apps/mcp-server/README.md) for details and [`for-ia.md`](for-ia.md) for usage instructions targeted at AI assistants.

### Claude Desktop / VS Code config

```json
{
  "mcpServers": {
    "syntaxchecker": {
      "command": "C:\\path\\to\\dist\\syntaxchecker-mcp.exe",
      "env": { "CHECKER_BIN": "C:\\path\\to\\dist\\syntax-checker.exe" }
    }
  }
}
```

If both binaries live in the same folder, `CHECKER_BIN` is optional.

## Performance notes

- JSON / XML / YAML / SQL MySQL / SQL SQLite: parse < 20 ms.
- SQL PostgreSQL / ANSI: ~700 ms WASM init on the first parse in each process (the WASM module is compiled on every start).
- SQL MSSQL / Oracle: ANTLR ATN built on first use in the process.

This is negligible for interactive use; for high throughput the MCP server could be evolved to use the checkers as an in-process library (the `apps/checker/checkers` package is already structured to be importable).

## Status

Phase 1 (CLI) and Phase 2 (MCP server) completed, plus JSON Schema validation
for JSON and YAML and the simple-format parsers (TOML, INI, CSV/TSV, HCL,
Markdown, .env, Properties). Out-of-scope backlog: XSD for XML, assertive
`format` validation, JSON5/JSONC, HTML, GraphQL, Protobuf, Dockerfile.
