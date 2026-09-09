# mcp-server-sparx-ea

[![PR validation](https://github.com/InsonusK/mcp-server-sparx-ea/actions/workflows/pr.yml/badge.svg)](https://github.com/InsonusK/mcp-server-sparx-ea/actions/workflows/pr.yml)
[![Tests](https://img.shields.io/endpoint?url=https://insonusk.github.io/mcp-server-sparx-ea/tests-badge.json)](https://insonusk.github.io/mcp-server-sparx-ea/tests/)
[![Coverage](https://img.shields.io/endpoint?url=https://insonusk.github.io/mcp-server-sparx-ea/coverage-badge.json)](https://insonusk.github.io/mcp-server-sparx-ea/coverage/)
[![Mutation score](https://img.shields.io/endpoint?url=https://insonusk.github.io/mcp-server-sparx-ea/mutation-badge.json)](https://insonusk.github.io/mcp-server-sparx-ea/mutation/)

An [MCP](docs/glossary/model-context-protocol.md) server that lets an AI agent
**read and edit [ArchiMate](docs/glossary/archimate.md) models built in
[Sparx Enterprise Architect](docs/glossary/sparx-enterprise-architect.md)**,
over stdio.

## Why

Sparx EA stores its ArchiMate content as UML-with-stereotypes in an Access
database that only EA can write. This server works on a model the user exported
to [XMI](docs/glossary/xmi.md) (`File → Export → Package to XMI`) and gives an
agent 17 typed operations over it:

- **Read** — navigate the tree, read an element with its relationships, read a
  package or a diagram, list the accepted ArchiMate types.
- **Edit** — create / rename / move / delete elements, relationships and
  packages; copy a package; place elements on diagrams. Every edit is written to
  a **new** XMI file the user re-imports into EA — see
  [the editing workflow](docs/workflow.md).

Element types are ArchiMate types (`ArchiMate.Goal`), not `uml:Class` +
stereotype pairs; every mutation is validated against the ArchiMate 3.2
vocabulary and relationship rules before it is written.

The server is pure Go — a single static binary, no runtime dependencies.

## Installation

- **Release binary** — [Releases](https://github.com/InsonusK/mcp-server-sparx-ea/releases):
  `linux`, `darwin`, `windows` × `amd64`, `arm64`.
- **From source** — Go 1.23+:

  ```bash
  go build -o mcp-server-sparx-ea .
  ```

Details: [docs/installation.md](docs/installation.md).

## Quick start

New to MCP? Follow the step-by-step
[setup guide](docs/setup-with-an-agent.md) — it covers Claude Desktop, Claude
Code and Cursor. The short version: point your client at the binary.

```json
{
  "mcpServers": {
    "sparx-ea": { "command": "/path/to/mcp-server-sparx-ea" }
  }
}
```

Then, from the agent:

```
ea_model_tree      file=example/TestProject.xml
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
| ArchiMate tools | [docs/api/archimate.md](docs/api/archimate.md) | The 20 read and editing tools over an exported model |
| Model of the server | [docs/model.md](docs/model.md) | The ArchiMate model of the server itself (`docs/mcp-server-sparx-ea.xml`) and its roadmap |
| Glossary | [docs/glossary/](docs/glossary/README.md) | Sparx EA, XMI, ArchiMate, MCP |

For an **AI agent**, the executable instructions live in
[docs/skills/sparx-ea/sparx-ea-mcp.skill.md](docs/skills/sparx-ea/sparx-ea-mcp.skill.md).

## Layout

| Path | What |
| --- | --- |
| `client/eaxmi/` | The EA XMI 2.1 codec — parse, navigate, edit, copy, re-serialise. Package `eaxmi`, pure Go. |
| `internal/service/sparx/` | The ArchiMate service: ArchiMate vocabulary + validation over `eaxmi`. |
| `internal/mcpserver/` | The MCP server — 17 tools over the service. |
| `main.go` | `server.ServeStdio` entry point. |
| `docs/skills/cucumber-go-testing.md` | How the tests are written (every test is a Cucumber scenario). |

## Testing

Every test is a Cucumber scenario (godog); one `make unit-test` runs all of them.

| Command | Does |
| --- | --- |
| `make unit-test` | Every scenario, with coverage. |
| `make mutation-test` | Mutation testing with [gremlins](docs/adr/0001-go-mutation-testing-tool.md). |
| `make test-report` | Assembles `public/` (reports + badges + landing page). |
| `make test-and-report` | All three, in order. |

Conventions: [docs/skills/cucumber-go-testing.md](docs/skills/cucumber-go-testing.md).
