# SyntaxChecker

Multi-format syntax validator, shipped as a CLI (`syntax-checker`) and as an MCP server (`syntaxchecker-mcp`) for integration with AI assistants (Claude Desktop, VS Code MCP, etc.).

## Supported formats

| Kind | `--type` | Notes |
|---|---|---|
| JSON | `json` | `--strict` detects duplicate keys; `--schema` validates against a JSON Schema |
| JSON5 | `json5` | `.json5`; comments, trailing commas, single quotes, unquoted keys, hex/Infinity/NaN |
| JSONC | `jsonc` | `.jsonc`; JSON with comments and trailing commas (stripped, then validated as JSON) |
| XML | `xml` | Well-formedness; `--schema` validates against a DTD (no XSD/RelaxNG — see below) |
| HTML | `html` / `htm` | `.html`, `.htm`; lenient parse (tolerant, like Markdown); `--strict` enforces XHTML-style balanced/nested tags |
| YAML | `yaml` | `.yml` and `.yaml`; `--strict` rejects custom tags; `--schema` validates against a JSON Schema |
| TOML | `toml` | `.toml`; reports line/column, including duplicate keys |
| INI | `ini` | `.ini`, `.cfg` |
| CSV | `csv` | `.csv`; `--strict` enforces a consistent field count per row |
| TSV | `tsv` | `.tsv`; tab-delimited, same checks as CSV |
| HCL | `hcl` | `.hcl`, `.tf` (Terraform); well-formedness only |
| Markdown | `markdown` | `.md`, `.markdown`; parse-only (Markdown has no invalid syntax) |
| .env | `env` | `.env`; KEY=VALUE assignments |
| Properties | `properties` | `.properties`; reports malformed `\uXXXX` escapes and circular `${...}` refs |
| Protobuf | `proto` | `.proto`; non-permissive well-formedness (no import/type resolution) |
| GraphQL | `graphql` / `gql` | `.graphql`, `.gql`; valid as either an SDL schema or an executable document |
| Dockerfile | `dockerfile` | `Dockerfile`, `Containerfile`, `*.dockerfile`; structural parse + unknown-instruction/argument checks |
| jq | `jq` | `.jq`; jq program syntax (parse-only) |
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

For `json` and `yaml`, `--schema` validates against a [JSON Schema](https://json-schema.org/)
(drafts up to 2020-12, including draft-07, via `santhosh-tekuri/jsonschema`;
YAML maps onto the same data model as JSON). Schema violations are reported by
their instance location (a JSON pointer such as `/items/0/name`) rather than a
source line, since validation runs on the decoded value.

`format` keywords (e.g. `email`, `date`) are treated as annotations only, per
the JSON Schema 2020-12 default — they do not cause validation failures.
Structural keywords (`type`, `required`, `enum`, `const`, `pattern`,
`minimum`/`maximum`, `minItems`, `uniqueItems`, `additionalProperties`,
`oneOf`/`anyOf`, `$ref`, …) are fully enforced.

For `xml`, `--schema` validates against a **DTD**. The validator checks that
every element is declared, that child elements are permitted by the parent's
content model (membership; `EMPTY` and element-content vs `#PCDATA`), that
`#REQUIRED`/`#FIXED`/enumerated attributes are satisfied, and that attributes
are declared. It does not enforce content-model order/cardinality, expand
entities, or resolve ID/IDREF. **XSD and RelaxNG are not supported**: there is no
mature pure-Go implementation, and the only production-grade option
(cgo + libxml2) would require a native library, breaking the self-contained
static binary.

Sample JSON Schema documents live under
[`test-samples/schema/`](test-samples/schema) (`<topic>.<variant>.<ext>` paired
with `<topic>.schema.json`); DTD samples live under
[`test-samples/dtd/`](test-samples/dtd) (`<topic>.<variant>.xml` paired with
`<topic>.dtd`). Files whose name contains `_not_correct` are expected to fail.
They are exercised by `apps/checker/checkers/schema_samples_test.go` and
`xml_dtd_test.go`, and end-to-end by `scripts/check-samples.ps1`.

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
for JSON and YAML, DTD validation for XML, the simple-format parsers (TOML, INI,
CSV/TSV, HCL, Markdown, .env, Properties) and the medium-format parsers
(JSON5/JSONC, Protobuf, GraphQL, HTML, Dockerfile, jq). Out-of-scope:
XSD/RelaxNG for XML (cgo/libxml2 would break the static binary), assertive
`format` validation, JSONPath.
