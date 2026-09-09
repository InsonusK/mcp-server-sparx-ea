# mcp-server-sparx-ea

[![PR validation](https://github.com/InsonusK/mcp-server-sparx-ea/actions/workflows/pr.yml/badge.svg)](https://github.com/InsonusK/mcp-server-sparx-ea/actions/workflows/pr.yml)
[![Tests](https://img.shields.io/endpoint?url=https://insonusk.github.io/mcp-server-sparx-ea/tests-badge.json)](https://insonusk.github.io/mcp-server-sparx-ea/tests/)
[![Coverage](https://img.shields.io/endpoint?url=https://insonusk.github.io/mcp-server-sparx-ea/coverage-badge.json)](https://insonusk.github.io/mcp-server-sparx-ea/coverage/)
[![Mutation score](https://img.shields.io/endpoint?url=https://insonusk.github.io/mcp-server-sparx-ea/mutation-badge.json)](https://insonusk.github.io/mcp-server-sparx-ea/mutation/)

An [MCP](docs/glossary/model-context-protocol.md) server that lets an AI agent
**read and edit [Sparx Enterprise Architect](docs/glossary/sparx-enterprise-architect.md)
models** over stdio — both the raw project database and the [ArchiMate](docs/glossary/archimate.md)
model itself.

## Why

Sparx EA stores a project in a Microsoft Access database (`.eapx`) that no
standard tool can read on Linux, and its ArchiMate content is buried in
UML-with-stereotypes. This server gives an agent two ways in:

- **SQL** against the `.eapx` file, in-process (cgo binding to `mdbtools` — no
  subprocess, no temp files), read-only.
- **ArchiMate operations** against a model the user exported to
  [XMI](docs/glossary/xmi.md) (`File → Export → Package to XMI`): navigate the
  tree, read an element with its relationships, create/rename/move/delete
  elements, relationships and packages, place elements on diagrams. Edits are
  written to a **new** XMI file that the user re-imports into EA — see
  [the editing workflow](docs/workflow.md).

Element types are ArchiMate types (`ArchiMate.Goal`), not `uml:Class` +
stereotype pairs; every mutation is validated against the ArchiMate 3.2
vocabulary and relationship rules before it is written.

## Installation

- **Release binary** — [Releases](https://github.com/InsonusK/mcp-server-sparx-ea/releases):
  `linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64`. The machine
  needs `mdbtools` + `glib` installed.
- **From source** — Go 1.23+ and the `mdbtools` dev headers:

  ```bash
  sudo apt-get install -y build-essential pkg-config libglib2.0-dev mdbtools-dev
  CGO_ENABLED=1 go build -o mcp-server-sparx-ea .
  ```

Details: [docs/installation.md](docs/installation.md).

## Quick start

New to MCP? Follow the step-by-step
[setup guide](docs/setup-with-an-agent.md) — it covers Claude Desktop, Claude
Code and Cursor. The short version: point your client at the built binary.

```json
{
  "mcpServers": {
    "sparx-ea": { "command": "/path/to/mcp-server-sparx-ea" }
  }
}
```

Then, from the agent:

```
ea_query        file=example/TestProject.eapx  sql="select Object_ID, Name from t_object"
ea_model_tree   file=example/TestProject.xml
ea_create_element  file=example/TestProject.xml  output=/tmp/edited.xml \
                   parent="Model/Motivation_Package"  type=ArchiMate.Goal  name="Reduce cost"
```

`ea_create_element` writes `/tmp/edited.xml`; the user imports it back into EA
(`File → Import → Package from XMI`), which diffs by GUID and asks for
confirmation.

## Documentation

| Topic | Docs | Covers |
| --- | --- | --- |
| Setup with an agent | [docs/setup-with-an-agent.md](docs/setup-with-an-agent.md) | Registering the server with Claude Desktop / Claude Code / Cursor, and troubleshooting |
| Editing workflow | [docs/workflow.md](docs/workflow.md) | The export → edit → re-import loop, `report.xml`, identity |
| SQL query tool | [docs/api/sql-query.md](docs/api/sql-query.md) | `ea_query` — read-only SQL against `.eapx` |
| ArchiMate tools | [docs/api/archimate.md](docs/api/archimate.md) | The 17 read and editing tools over an exported model |
| Glossary | [docs/glossary/](docs/glossary/README.md) | Sparx EA, XMI, ArchiMate, MCP |

For an **AI agent**, the executable instructions live in
[docs/skills/sparx-ea/sparx-ea-mcp.skill.md](docs/skills/sparx-ea/sparx-ea-mcp.skill.md).

## Layout

| Path | What |
| --- | --- |
| `client/eapx/` | The `.eapx` SQL connector (cgo → `mdbtools`). Package `eapx`. |
| `client/eaxmi/` | The EA XMI 2.1 codec — parse, navigate, edit, copy, re-serialise. Package `eaxmi`, pure Go. |
| `internal/service/sparx/` | The ArchiMate service: ArchiMate vocabulary + validation over `eaxmi`. |
| `internal/mcpserver/` | The MCP server — wires both to 18 tools. |
| `main.go` | `server.ServeStdio` entry point. |
| `docs/skills/cucumber-go-testing.md` | How the tests are written (every test is a Cucumber scenario). |

## Testing

Every test is a Cucumber scenario (godog); one `make unit-test` runs all 176.

| Command | Does |
| --- | --- |
| `make unit-test` | Every scenario, with `-race` and coverage. |
| `make mutation-test` | Mutation testing with [gremlins](docs/adr/0001-go-mutation-testing-tool.md). |
| `make test-report` | Assembles `public/` (reports + badges + landing page). |
| `make test-and-report` | All three, in order. |

Conventions: [docs/skills/cucumber-go-testing.md](docs/skills/cucumber-go-testing.md).
