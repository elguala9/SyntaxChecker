# SyntaxChecker

Multi-format syntax validator, shipped as a CLI (`checker`) and as an MCP server (`mcp-server`) for integration with AI assistants (Claude Desktop, VS Code MCP, etc.).

## Supported formats

| Kind | `--type` | Notes |
|---|---|---|
| JSON | `json` | `--strict` detects duplicate keys |
| XML | `xml` | Well-formedness only (no DTD/XSD) |
| YAML | `yaml` | `.yml` and `.yaml` |
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
│   ├── checker/       # CLI: dist/checker(.exe)
│   └── mcp-server/    # MCP server: dist/mcp-server(.exe)
├── pkg/result/        # Shared types (CheckResult, SyntaxError)
├── test-samples/      # Valid and invalid sample files
├── scripts/           # check-samples.ps1 (end-to-end tests)
└── Makefile
```

## Build

Requires Go 1.22+ (workspace already configured in `go.work`).

```bash
make build           # produces dist/checker(.exe) and dist/mcp-server(.exe)
make test            # runs the unit tests
make build-windows   # cross-build for Windows
make build-linux     # cross-build for Linux
```

Binaries built with `CGO_ENABLED=0`, `-trimpath`, `-ldflags "-s -w"`.

## CLI usage

```bash
checker --file <path> [--type <type>] [--format text|json|json-pretty] [--strict] [--quiet]
```

Examples:

```bash
checker -f config.json
checker -f query.sql -t sql:mysql
checker -f data.json --strict -o json-pretty
```

Exit codes: `0` valid, `1` syntax errors, `2` internal error / file not found.

## MCP usage

Exposes the `check_syntax(file_path, type?, strict?)` tool over stdio transport.

See [`apps/mcp-server/README.md`](apps/mcp-server/README.md) for details and [`for-ia.md`](for-ia.md) for usage instructions targeted at AI assistants.

### Claude Desktop / VS Code config

```json
{
  "mcpServers": {
    "syntaxchecker": {
      "command": "C:\\path\\to\\dist\\mcp-server.exe",
      "env": { "CHECKER_BIN": "C:\\path\\to\\dist\\checker.exe" }
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

Phase 1 (CLI) and Phase 2 (MCP server) completed. Out-of-scope backlog: TOML, CSV, INI, ENV, HTML, Dockerfile, XHTML.
