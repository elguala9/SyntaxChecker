# syntaxchecker MCP server

Exposes the `checker` as the MCP tool `check_syntax` over stdio transport. The
server invokes the `checker` executable as a subprocess and returns its JSON
output as structured content.

## Tool

`check_syntax(file_path, type?, strict?)`

- `file_path` — path to the file (absolute or relative to the server's working directory).
- `type` — optional; forces the type. If omitted, auto-detected from the extension.
  Values: `json`, `xml`, `yaml`, `sql:mysql`, `sql:postgres`, `sql:ansi`,
  `sql:sqlite`, `sql:mssql`, `sql:oracle`.
- `strict` — optional; enables stricter checks (e.g. duplicate JSON keys).

## Build

```
make build        # produces dist/checker(.exe) and dist/syntaxchecker-mcp(.exe)
```

## Resolving the checker executable

The server looks for `checker` in this order:

1. `CHECKER_BIN` environment variable;
2. `checker`/`checker.exe` binary next to `syntaxchecker-mcp`;
3. `checker` on `PATH`.

## Claude Desktop / VS Code config

```json
{
  "mcpServers": {
    "syntaxchecker": {
      "command": "C:\\path\\to\\dist\\syntaxchecker-mcp.exe",
      "env": { "CHECKER_BIN": "C:\\path\\to\\dist\\checker.exe" }
    }
  }
}
```

If `checker.exe` is in the same folder as `syntaxchecker-mcp.exe`, `CHECKER_BIN` is optional.

## Performance note

Each `check_syntax` call spawns a `checker` process. For `sql:postgres`/`sql:ansi`
you pay the WASM module init (~700 ms) on every call; the ANTLR4 dialects
(`sql:mssql`, `sql:oracle`) build the ATN on first use in the process. This is
acceptable for interactive use; for high throughput it is worth evolving the
checkers into an in-process library.
