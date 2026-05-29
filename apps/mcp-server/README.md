# syntaxchecker MCP server

Exposes the `syntax-checker` as the MCP tool `check_syntax` over stdio transport. The
server invokes the `syntax-checker` executable as a subprocess and returns its JSON
output as structured content.

## Tool

`check_syntax(file_path, type?, schema?, strict?)`

- `file_path` — path to the file (absolute or relative to the server's working directory).
- `type` — optional; forces the type. If omitted, auto-detected from the extension.
  Values: `json`, `xml`, `yaml`, `toml`, `ini`, `csv`, `tsv`, `hcl`, `markdown`,
  `env`, `properties`, `proto`, `graphql`, `gql`, `sql:mysql`, `sql:postgres`,
  `sql:ansi`, `sql:sqlite`, `sql:mssql`, `sql:oracle`.
- `schema` — optional; path to a JSON Schema to validate the document against.
  Supported for `json` and `yaml` only (XSD validation for XML is not supported).
- `strict` — optional; enables stricter checks (e.g. duplicate JSON keys).

## Build

```
make build        # produces dist/syntax-checker(.exe) and dist/syntaxchecker-mcp(.exe)
```

## Resolving the syntax-checker executable

The server looks for `syntax-checker` in this order:

1. `CHECKER_BIN` environment variable;
2. `syntax-checker`/`syntax-checker.exe` binary next to `syntaxchecker-mcp`;
3. `syntax-checker` on `PATH`.

## Claude Desktop / VS Code config

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

If `syntax-checker.exe` is in the same folder as `syntaxchecker-mcp.exe`, `CHECKER_BIN` is optional.

## Performance note

Each `check_syntax` call spawns a `syntax-checker` process. For `sql:postgres`/`sql:ansi`
you pay the WASM module init (~700 ms) on every call; the ANTLR4 dialects
(`sql:mssql`, `sql:oracle`) build the ATN on first use in the process. This is
acceptable for interactive use; for high throughput it is worth evolving the
checkers into an in-process library.
