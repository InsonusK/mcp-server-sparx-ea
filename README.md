# mcp-server-sparx-ea

An [MCP](https://modelcontextprotocol.io) server that answers **SQL queries
against Sparx Enterprise Architect project files** (`.eapx` / `.eap`, which are
Microsoft Access JET databases) over stdio.

Queries run **in-process**: the server binds directly to the `mdbtools` C library
(`libmdb` / `libmdbsql`) through cgo — no `mdb-*` subprocess, no temporary files.

> Status: base connector. One tool (`ea_query`) is implemented. The
> domain-specific tools from [`task/base-task.md`](task/base-task.md)
> (`ea_search_objects`, `ea_get_connectors`) are not built yet.

## Layout

| Path | What |
|------|------|
| `eapx/` | The connector. `Open(path)` → `Connector`; `Connector.Query(sql)` → `ResultSet`. `cgo_mdb.go` is the cgo binding, `connector.go` the Go API. |
| `internal/mcpserver/` | Wires the connector to an MCP server and exposes the `ea_query` tool. |
| `main.go` | `server.ServeStdio` entry point. |
| `features/` | Cucumber (godog) specs — behaviour and architectural concerns. |
| `tools/testkit/` | Normalises Go test / coverage / gremlins output into the report contract. |
| `example/` | Sample projects: `TestProject.eapx` (populated), `EmptyProject.eapx`. |

## Build

Prerequisites (Debian/Ubuntu):

```bash
sudo apt-get install -y build-essential pkg-config libglib2.0-dev mdbtools-dev
```

```bash
CGO_ENABLED=1 go build -o mcp-server-sparx-ea .
```

## Use

The server speaks MCP over stdio. Register it with an MCP client, e.g.:

```json
{
  "mcpServers": {
    "sparx-ea": { "command": "/path/to/mcp-server-sparx-ea" }
  }
}
```

### Tool: `ea_query`

| Argument | Required | Description |
|----------|----------|-------------|
| `file` | yes | Path to the `.eapx` file on the server's filesystem. |
| `sql` | yes | A single read-only `SELECT` statement. |

Returns a text content block containing JSON:

```json
{ "columns": ["Object_ID", "Name"], "rows": [["2", "Stakeholder1"]], "rowCount": 1 }
```

Write/DDL statements (`INSERT`, `UPDATE`, `DELETE`, `DROP`, …) are rejected —
`mdbtools` is read-only. Text is decoded to UTF-8 (JET4 stores it as UCS-2LE).

Example against the sample project:

```
ea_query file=example/TestProject.eapx  sql="select Object_ID, Name, Object_Type from t_object"
```

Useful tables: `t_object` (elements), `t_connector` (relationships),
`t_package`, `t_diagram`, `t_attribute`, `t_operation`.

## Testing

Four targets, per
[solution-conformance-testing](.claude/skills/solution-conformance-testing/SKILL.md):

| Command | Does |
|---------|------|
| `make unit-test` | Cucumber scenarios + Go tests in one run, with `-race` and coverage. |
| `make mutation-test` | Mutation testing with [gremlins](docs/adr/0001-go-mutation-testing-tool.md). |
| `make test-report` | Assembles `public/` (per-kind reports + shields.io badges + landing page). |
| `make test-and-report` | All three, in order. |

Toggles: `WITH_CODE_COVERAGE=true` (emit the normalised coverage result + badge),
`ONLY_DELTA=true DELTA_BASE=<ref>` (mutate only changed code).

Normalised results land in `tmp/result/*.json`, native reports in
`tmp/report/<kind>/`. What is and isn't covered:
[docs/test-trace-matrix.md](docs/test-trace-matrix.md).

Current numbers: 45 tests green · **95.4%** line coverage · **97.1%** mutation score.

### Known gap

The Cyrillic / legacy-CP1251 decoding scenario
(`features/technical/text_encoding.feature`, `@needs-cyrillic-fixture`)
**self-skips**: authoring a Cyrillic `.eapx` needs Sparx EA itself, since
`mdbtools` cannot write rows. Add `example/CyrillicProject.eapx` to enable it.
The JET4 UCS-2LE → UTF-8 path is covered.
