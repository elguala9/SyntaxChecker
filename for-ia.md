# SyntaxChecker MCP — instructions for AI assistants

This document is intended for an AI assistant (Claude, etc.) that needs to install, configure and use the `syntaxchecker` MCP server.

## What it is

`syntaxchecker` is an MCP server that exposes a single tool, `check_syntax`, capable of validating files in JSON, XML, YAML and SQL (MySQL, PostgreSQL/ANSI, SQLite, SQL Server, Oracle). Use it when the user asks you to **verify the syntax** of a code/configuration file, or when you have just produced a file and want to confirm its validity before delivering it.

## When to use it

USE `check_syntax` when:
- The user explicitly asks to validate/check a file's syntax.
- You have just written or edited a JSON/XML/YAML/SQL file and want to verify it before considering the task complete.
- You are debugging a parsing error in one of these formats.

DO NOT use it for:
- Semantic validation (e.g. "does this SQL do the right thing?") — the tool checks syntax only.
- Schema validation (XSD, JSON Schema, etc.) — not supported.
- Languages not listed (Python, Go, TS, etc.).

## Installation

### 1. Build the binaries

From the repo root, with Go 1.22+:

```
make build
```

Produces:
- `dist/checker(.exe)` — validation CLI.
- `dist/syntaxchecker-mcp(.exe)` — MCP server (stdio transport).

If `make` is not available (Windows without make):

```
cd apps/checker     && go build -o ../../dist/checker.exe .
cd apps/mcp-server  && go build -o ../../dist/syntaxchecker-mcp.exe .
```

### 2. MCP client configuration

#### Claude Desktop / Claude Code

File: `%APPDATA%\Claude\claude_desktop_config.json` (Windows) or `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS).

```json
{
  "mcpServers": {
    "syntaxchecker": {
      "command": "C:\\absolute\\path\\dist\\syntaxchecker-mcp.exe",
      "env": { "CHECKER_BIN": "C:\\absolute\\path\\dist\\checker.exe" }
    }
  }
}
```

`CHECKER_BIN` is optional if both binaries live in the same folder: the server looks for `checker(.exe)` next to itself and then on `PATH`.

Restart the MCP client after editing the config.

#### VS Code (MCP extension)

Same JSON shape under the `mcpServers` key in the extension configuration.

#### OpenCode

File: `~/.config/opencode/opencode.json` (Linux/macOS) or `%APPDATA%\opencode\opencode.json` (Windows).

```json
{
  "mcp": {
    "syntaxchecker": {
      "type": "local",
      "command": ["C:\\absolute\\path\\dist\\syntaxchecker-mcp.exe"],
      "environment": { "CHECKER_BIN": "C:\\absolute\\path\\dist\\checker.exe" },
      "enabled": true
    }
  }
}
```

Restart OpenCode after editing the config.

### 3. Verify the installation

In the client, call the tool against a sample file:

```
check_syntax(file_path="C:/path/SyntaxChecker/test-samples/json/config.json")
```

Expected: `{"valid": true, ...}`.

## Using the tool

### Signature

```
check_syntax(file_path: string, type?: string, strict?: boolean)
```

### Parameters

- **`file_path`** (required) — path to the file. Absolute or relative to the server's working directory. Prefer absolute paths to avoid ambiguity.
- **`type`** (optional) — forces the type. If omitted, the checker auto-detects from the extension (`.json`, `.xml`, `.yml`/`.yaml`, `.sql`).
  - **Important for `.sql`**: auto-detect does NOT know which dialect to use. For SQL files **you must always pass `type`** with the correct dialect, otherwise the checker fails or uses an unwanted default.
  - Valid values: `json`, `xml`, `yaml`, `sql:mysql`, `sql:postgres` (alias `sql:postgresql`), `sql:ansi`, `sql:sqlite`, `sql:mssql` (alias `sql:tsql`), `sql:oracle` (alias `sql:plsql`).
- **`strict`** (optional, bool) — enables stricter checks. Known effect:
  - JSON: detects duplicate keys (otherwise silently ignored by the standard parser).
  - Other formats: no effect for now, but passing `true` is harmless.

### Call examples

```
check_syntax(file_path="/abs/path/config.json")
check_syntax(file_path="/abs/path/config.json", strict=true)
check_syntax(file_path="/abs/path/query.sql", type="sql:mysql")
check_syntax(file_path="/abs/path/migration.sql", type="sql:postgres")
check_syntax(file_path="/abs/path/deploy.yaml")
```

### Response

Structured content with this shape:

```json
{
  "file": "/abs/path/config.json",
  "type": "json",
  "valid": false,
  "errors": [
    { "line": 5, "column": 12, "message": "invalid character '}' looking for ..." }
  ]
}
```

- `valid: true` → file is syntactically correct. `errors` absent or empty.
- `valid: false` → at least one error. Every `errors[i]` always has `message`; `line`/`column` may be missing (e.g. YAML does not expose `column`; XML does not expose `column`).

## How to interpret results

1. If `valid: true`: the file is well-formed. This does NOT mean "correct at the logic/schema level" — only that the parser accepts it.
2. If `valid: false`:
   - Report errors to the user mentioning **file, line, column and message** (`file_path:line:column`).
   - If you generated the file yourself, **fix it** before returning it to the user.
   - For SQL: if the error looks inconsistent, double-check that you passed the right `type`. A construct valid in PostgreSQL may be invalid in MySQL.

## Common errors and troubleshooting

- **Tool not available in the client**: the server did not start. Check the `command` path, restart the client, and make sure `syntaxchecker-mcp.exe` exists.
- **"checker binary not found"**: the server cannot locate `checker(.exe)`. Set `CHECKER_BIN` or place it in the same folder as `syntaxchecker-mcp.exe`.
- **All SQL validations fail in odd ways**: you probably omitted `type` and auto-detect picked an unsuitable dialect. Pass the explicit dialect.
- **First `sql:postgres` call is slow (~700 ms)**: normal. That is the WASM module init, once per process.

## Best practices

- Always pass **absolute paths**.
- For SQL, always pass an **explicit `type`**.
- For JSON with sensitive configuration (e.g. config keys that must not repeat), pass **`strict=true`**.
- After generating a file for the user, validate before delivering. If invalid, fix and re-validate — do not hand over a broken file.
- Do not call the tool on large binary or non-textual files: respond sensibly.
