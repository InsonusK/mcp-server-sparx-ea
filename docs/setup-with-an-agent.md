# Adding the server to your AI agent

This server is not a website or a background service. It is a small program that
your AI assistant (the "MCP client") starts on demand, talks to over its
standard input/output, and stops when done. "Installing" it means two things:

1. get the `mcp-server-sparx-ea` program onto your computer;
2. tell your AI client where that program is.

No ports, no accounts, no server to keep running.

---

## Step 1 — get the program

It is a single file with no dependencies. Two ways, pick one.

### a) Download a release binary

From the repository's **Releases** page, download the archive for your OS and
CPU (`linux` / `darwin` / `windows`, `amd64` or `arm64`), unpack it, and put
`mcp-server-sparx-ea` (or `.exe`) somewhere permanent.

### b) Build from source

```bash
go build -o mcp-server-sparx-ea .
```

---

Either way, note the **full path** to the binary, for example:

```
/home/you/bin/mcp-server-sparx-ea        (Linux/macOS)
C:\Tools\mcp-server-sparx-ea.exe          (Windows)
```

---

## Step 2 — tell your client where it is

Pick your client below. In every case you are adding one entry: a name you
choose (`sparx-ea` here) and the path from Step 1.

### Claude Desktop

Edit the config file (create it if it does not exist):

| OS | File |
| --- | --- |
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |

```json
{
  "mcpServers": {
    "sparx-ea": {
      "command": "/home/you/bin/mcp-server-sparx-ea"
    }
  }
}
```

If the file already has other servers under `"mcpServers"`, add `"sparx-ea"`
next to them — do not replace the whole file. Save it and **fully quit and
reopen Claude Desktop** (not just close the window).

### Claude Code (CLI)

```bash
claude mcp add sparx-ea /home/you/bin/mcp-server-sparx-ea
```

Check it registered:

```bash
claude mcp list
```

### Cursor

Settings → **MCP** → **Add new MCP server**:

- Name: `sparx-ea`
- Type: `command`
- Command: `/home/you/bin/mcp-server-sparx-ea`

### Any other MCP client

The pattern is always the same — a JSON object keyed by a name, with a
`command` that is the full path to the binary and no arguments:

```json
{ "mcpServers": { "sparx-ea": { "command": "<full path>" } } }
```

---

## Step 3 — check it works

Ask your assistant something that uses the server. The simplest check needs no
file:

> What ArchiMate types does the sparx-ea server accept?

which calls `ea_archimate_types`. To try a real model, export a package from EA
(`File → Export → Package to XMI`), put the `.xml` where the assistant can reach
it, and say:

> Using the sparx-ea tools, show me the model tree of `model.xml`.

---

## Common problems

| Symptom | Cause | Fix |
| --- | --- | --- |
| Client shows the server as "failed" / "disconnected" | wrong path in the config | use the **absolute** path; run `which mcp-server-sparx-ea` (or `where` on Windows) to confirm it |
| Tools don't appear after editing the config | the client wasn't restarted | fully quit and reopen the client (Claude Desktop especially) |
| `sparx: no ... for "<path>"` | the `ref` you gave does not match the model | run `ea_model_tree` first to see the exact paths |
| `file not found` / a parse error | the path you gave points nowhere the server can see, or is not an XMI export | give a path on the **same machine the server runs on**; make sure it is `File → Export → Package to XMI` output, not the `.eapx` |
| The agent edited a model but "nothing changed in EA" | editing tools write a new `.xml`; they never touch the live project | import the tool's `output` file into EA: `File → Import → Package from XMI` — see [the editing workflow](workflow.md) |

---

## What the server can do once connected

- `ea_model_tree`, `ea_element`, `ea_package`, `ea_diagram`,
  `ea_archimate_types` — read an [ArchiMate](glossary/archimate.md) model the
  user exported to [XMI](glossary/xmi.md).
- `ea_create_element`, `ea_create_relationship`, … — edit a copy of that model
  and write a file you re-import into EA.

Full tool list and arguments: [docs/api/archimate.md](api/archimate.md). The
export/edit/import loop: [docs/workflow.md](workflow.md).
