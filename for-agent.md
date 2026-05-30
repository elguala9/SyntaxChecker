# Setting up `AGENTS.md` / `CLAUDE.md` in your projects

This document explains how to configure an instructions file for AI agents (Claude Code, OpenCode, etc.) in a project, so they automatically use the `syntaxchecker` MCP to validate JSON / JSON5 / JSONC / XML / HTML / YAML / TOML / INI / CSV / HCL / Markdown / `.env` / Properties / Protobuf / GraphQL / Dockerfile / jq / Go / TypeScript / JavaScript / Shell / Lua / Starlark / SQL files.

The installer copies this file together with `for-ia.md` into the installation folder, so it is always available as a reference.

## Which file to use

AI agents read, at the start of a session, an instructions file in the project root. The convention depends on the client:

| Client                | File read automatically |
|-----------------------|----------------------------|
| Claude Code           | `CLAUDE.md` (repo root, also nested in subfolders) |
| OpenCode              | `AGENTS.md` (repo root) |
| Cursor / other agents | `AGENTS.md` (emerging de-facto standard) |

**Practical tip**: create a single `AGENTS.md` with the instructions and then, in the projects where you use Claude Code, add a `CLAUDE.md` that just includes the content of `AGENTS.md` (or keep them identical). This way any agent finds the same rules.

## What to write inside

The file must tell the agent **when** to invoke `check_syntax` and **how** to call it. Below is a minimal template, ready to copy-paste into the project root.

### `AGENTS.md` (and/or `CLAUDE.md`) template

```markdown
# Instructions for AI agents in this project

## Syntax validation of files

This project provides the MCP server **`syntaxchecker`**, which exposes
the `check_syntax` tool for JSON, JSON5, JSONC, XML, HTML, YAML, TOML, INI,
CSV/TSV, HCL, Markdown, `.env`, Properties, Protobuf, GraphQL, Dockerfile, jq,
the programming languages Go, TypeScript/JavaScript (incl. JSX/TSX),
Shell/Bash, Lua and Starlark, and SQL (MySQL, PostgreSQL, ANSI, SQLite, SQL Server,
Oracle). The programming-language checks are **parse-only**
(no type-checking, no import/name resolution).

**Mandatory rules:**

1. After creating or editing a `.json`/`.json5`/`.jsonc`, `.xml`,
   `.html`/`.htm`, `.yml`/`.yaml`, `.toml`, `.ini`/`.cfg`, `.csv`/`.tsv`,
   `.hcl`/`.tf`, `.md`, `.env`, `.properties`, `.proto`, `.graphql`/`.gql`,
   `Dockerfile`, `.jq`, `.go`, `.ts`/`.tsx`, `.js`/`.mjs`/`.cjs`/`.jsx`,
   `.lua`, `.sh`/`.bash`, `.star`/`.bzl` or `.sql` file, ALWAYS invoke `check_syntax`
   on the file before considering the task complete.
2. For `.sql` files ALWAYS pass the `type` parameter with the correct dialect
   (`sql:mysql`, `sql:postgres`, `sql:ansi`, `sql:sqlite`, `sql:mssql`,
   `sql:oracle`). Auto-detect does not pick SQL dialects.
3. For configuration JSON files pass `strict=true` to catch
   duplicate keys.
4. ALWAYS use absolute paths in `file_path`.
5. If `valid: false`, fix the file and re-validate before delivering it
   to the user. Never return a broken file.

**Call examples:**

```
check_syntax(file_path="C:/proj/config.json", strict=true)
check_syntax(file_path="C:/proj/migrations/001_init.sql", type="sql:postgres")
check_syntax(file_path="C:/proj/deploy/k8s.yaml")
```

For the complete tool reference see `for-ia.md` (installed together with the
MCP server).
```

## Step-by-step setup in a new project

1. Make sure the `syntaxchecker` MCP is already configured in the AI client (see `for-ia.md`, section *MCP client configuration*). If it is not, configure it first — the `AGENTS.md` file alone does not install anything.
2. Go to the project root.
3. Create `AGENTS.md` by pasting the template above. Adapt the examples to the paths/file types actually present in the project (e.g. if you have no SQL, drop the dialect rule).
4. If you use Claude Code, also create `CLAUDE.md` with the same content (or a single line `See AGENTS.md.` if the agent follows references — Claude Code generally prefers direct content).
5. Commit both files to the repo: this way the rules apply to anyone who clones the project, not just to your machine.
6. Restart the agent session (new conversation) so it re-reads the file.

## Verify it works

Open a new conversation with the agent in the project root and ask something like:

> "Create a `test.json` file with a sample configuration."

After writing the file, the agent should spontaneously invoke `check_syntax` on it. If it does not, check:

- that the file is named exactly `AGENTS.md` or `CLAUDE.md` (case-sensitive on Linux/macOS);
- that it is in the repo root (or in a folder the agent is exploring);
- that the `syntaxchecker` MCP is active in the client (in Claude Code: `/mcp`; in OpenCode: the list of available tools).

## Notes for whoever prepares the installer

The installer should:

- Copy `for-agent.md` (this file) and `for-ia.md` into the installation folder, `docs/` section.
- NOT automatically create `AGENTS.md` or `CLAUDE.md` in the user's projects: they are per-repo files and must be added manually where needed. The installer may, however, offer a command or a menu *"Copy `AGENTS.md` template into the current folder"* for convenience.
