# Installation and access — sparx-ea-mcp

> Agent-facing copy, kept self-contained because this skill is synced into
> `.claude/skills/` and `.agents/skills/`. The human version — install scripts,
> flags, troubleshooting — is [`docs/installation.md`](../../installation.md);
> keep the two in sync when either changes.

## Get the server

Single pure-Go binary, no dependencies.

- **Release binary** — the repository's Releases page has archives for
  `linux` / `darwin` / `windows` × `amd64` / `arm64`.
- **Install script** (Linux / macOS) — installs the binary and, on request,
  registers it:

  ```bash
  curl -fsSL https://raw.githubusercontent.com/InsonusK/mcp-server-sparx-ea/master/scripts/install.sh \
    | bash -s -- --register project
  ```

- **From source** — Go 1.23+: `go build -o mcp-server-sparx-ea .` (verify with
  `go test ./...`).

## Register with an MCP client

The server speaks MCP over stdio — no flags, no environment variables. Add one
entry keyed by a name (`sparx-ea`) with `command` set to the binary:

```json
{
  "mcpServers": {
    "sparx-ea": { "command": "/absolute/path/to/mcp-server-sparx-ea" }
  }
}
```

- **Claude Desktop / Cursor / other** — put that object in the client's config
  file.
- **Claude Code** — `claude mcp add --scope project sparx-ea -- mcp-server-sparx-ea`
  writes / merges a `.mcp.json` at the repository root (commit it to share with
  the team); `--scope user` adds it to the personal config instead. A committed
  `.mcp.json` should use the bare binary name, not an absolute path.

MCP servers load at client start — a running session does not pick up a new
registration until it restarts.

## Files the server reads

Every tool except `ea_new_model` takes a `file` — a model the user exported from
EA (`File → Export → Package to XMI`, XMI 2.1), a `.xml`. Editing tools also
write to `output` (a different, writable path). The server never opens a path you
did not pass in a tool argument, and never makes a network call.
