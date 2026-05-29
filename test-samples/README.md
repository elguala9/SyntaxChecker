# Test samples for `syntaxchecker`

This directory holds sample files used to exercise every parser exposed by the
`check_syntax` MCP tool. Each subfolder targets one format / language.

## Naming convention

- A file **without** `_not_correct` in its name MUST validate as `valid: true`.
- A file **with** `_not_correct` in its name MUST validate as `valid: false`
  (the parser is expected to report at least one syntax / validation error).

## How to run a check

```
check_syntax(file_path="<path>")                       # auto-detect from extension
check_syntax(file_path="<path>", type="<type>")        # force a type / dialect
check_syntax(file_path="<path>", strict=true)          # enable stricter checks
check_syntax(file_path="<doc>", schema="<schema/dtd>") # validate against a schema/DTD
```

`file_path` may be absolute or relative to the server's working directory.

## Samples that need extra parameters

Most samples auto-detect correctly with no parameters. The following are the
exceptions — running them without the listed parameter will NOT match the
expected result.

### `type` is required (SQL dialects)

The `.sql` extension does not encode a dialect, so auto-detect cannot pick one.
Always pass the matching `type`:

| Files                       | Required `type`  |
|-----------------------------|------------------|
| `sql/mysql_*.sql`           | `sql:mysql`      |
| `sql/postgres_*.sql`        | `sql:postgres`   |
| `sql/sqlite_*.sql`          | `sql:sqlite`     |
| `sql/mssql_*.sql`           | `sql:mssql`      |
| `sql/oracle_*.sql`          | `sql:oracle`     |

(`sql:ansi` is also available for generic ANSI SQL.)

### `strict=true` is required

These `_not_correct` files are well-formed in lenient mode and are only
rejected in strict mode — the failure they demonstrate is a strict-mode rule,
not a plain parse error:

| File                                          | What strict mode catches              |
|-----------------------------------------------|---------------------------------------|
| `json/duplicate_keys_not_correct.json`        | duplicate object keys                 |
| `json/nested_duplicate_keys_not_correct.json` | duplicate keys in nested objects      |
| `html/mismatched_tag_not_correct.html`        | mismatched closing tag                |
| `html/stray_close_not_correct.html`           | stray / unexpected closing tag        |
| `html/unclosed_tag_not_correct.html`          | unclosed elements                     |
| `yaml/custom_tag_not_correct.yaml`            | application-specific (non-portable) tag |

### `schema` (or DTD) is required

The `schema/` and `dtd/` folders test schema/DTD validation, not bare syntax.
The document is syntactically valid JSON/YAML/XML on its own; the error only
appears when validated against its schema.

| Document folder | Pass as `schema`                                   |
|-----------------|----------------------------------------------------|
| `schema/api_envelope.*.json`       | `schema/api_envelope.schema.json`     |
| `schema/blog_post.*.json`          | `schema/blog_post.schema.json`        |
| `schema/product.*.json`            | `schema/product.schema.json`          |
| `schema/weather_station.*.yaml`    | `schema/weather_station.schema.json`  |
| `dtd/note.*.xml`                   | `dtd/note.dtd`                        |
| `dtd/catalog.*.xml`                | `dtd/catalog.dtd`                     |

The `*.schema.json` files themselves are plain valid JSON and check with no
parameters.

## Coverage

Every other sample (CSV, TSV, Dockerfile, `.env`, GraphQL, Go, HCL, INI, jq,
JS/JSX, JSON, JSON5, JSONC, Lua, Markdown, Properties, Protobuf, Shell,
Starlark, TOML, TS/TSX, XML, YAML) auto-detects from its extension and needs no
extra parameters.
