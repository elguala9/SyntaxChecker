# syntaxchecker (Claude Code plugin)

Adds the `check_syntax` MCP tool to Claude Code. No manual config editing and no
package manager required: the platform-specific binaries are downloaded from
[GitHub Releases](https://github.com/elguala9/SyntaxChecker/releases) on first
use and cached per version.

## Install

```text
/plugin marketplace add elguala9/SyntaxChecker
/plugin install syntaxchecker
```

On the first `check_syntax` call the launcher downloads
`syntaxchecker-<version>-<os>-<arch>` from the matching release, verifies its
SHA-256, and extracts it to a cache directory:

- Windows: `%LOCALAPPDATA%\syntaxchecker\v<version>\`
- Linux/macOS: `${XDG_CACHE_HOME:-$HOME/.cache}/syntaxchecker/v<version>/`

Subsequent runs reuse the cache; bumping the plugin version triggers a fresh
download.

## Platform support

| Platform | Status |
| --- | --- |
| Windows x64 | ✅ supported |
| Linux x64 | ✅ supported |
| **macOS** | ❌ **not supported** (no prebuilt binary published) |
| other (arm64, etc.) | ❌ not supported |

On unsupported platforms the launcher exits with a clear message. To use it
there, build from source (https://github.com/elguala9/SyntaxChecker) and point
the server at the binary via the `CHECKER_BIN` environment variable.

## Requirements

- Network access on first use (to reach GitHub Releases).

## How it works

`.mcp.json` runs `bin/launch.mjs` with Node (always present in Claude Code).
The launcher resolves the target release from `plugin.json`, ensures the
binaries are cached, then execs `syntaxchecker-mcp` with `CHECKER_BIN` set,
inheriting stdio so the MCP stdio transport passes straight through.

## Tool

`check_syntax(file_path, type?, schema?, strict?)` — see the
[server README](../apps/mcp-server/README.md) for the full parameter list and
supported types.
