# Adding the server to your AI agent

This server is not a website or a background service. It is a small program that
your AI assistant (the "MCP client") starts on demand, talks to over its
standard input/output, and stops when done. "Installing" it means two things:

1. get the `mcp-server-sparx-ea` program onto your computer;
2. tell your AI client where that program is.

No ports, no accounts, no server to keep running.

---

## Step 1 — get the program

Three ways, pick one.

### a) Download a release binary

From the repository's **Releases** page, download the archive for your OS and
CPU (`linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64`), unpack it,
and put `mcp-server-sparx-ea` somewhere permanent.

The machine also needs the `mdbtools` runtime libraries:

```bash
sudo apt-get install -y mdbtools libglib2.0-0   # Debian / Ubuntu
brew install mdbtools glib                       # macOS
```

### b) Use the container image (no local build, libraries bundled)

```
ghcr.io/insonusk/mcp-server-sparx-ea:latest
```

Register it in your client as a `docker run` command — see
[the container section below](#option-b-run-it-as-a-container).

### c) Build from source

```bash
# Debian / Ubuntu
sudo apt-get install -y build-essential pkg-config libglib2.0-dev mdbtools-dev
# macOS
brew install mdbtools pkg-config glib

# then, in the project folder:
CGO_ENABLED=1 go build -o mcp-server-sparx-ea .
```

---

Whichever you chose, note the **full path** to the binary (or the image name),
for example:

```
/home/you/bin/mcp-server-sparx-ea        (Linux/macOS)
ghcr.io/insonusk/mcp-server-sparx-ea     (container)
```

> Windows: the server links the `mdbtools` C library, which is awkward on
> Windows. Use WSL (Ubuntu) or the container image.

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

### Option B: run it as a container

If you chose the container image, the `command` is `docker` and the arguments
run it in interactive (stdio) mode with your model files mounted:

```json
{
  "mcpServers": {
    "sparx-ea": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-v", "/path/to/your/models:/work",
        "ghcr.io/insonusk/mcp-server-sparx-ea:latest"
      ]
    }
  }
}
```

Pass file paths to the tools **under the mount point**, e.g.
`file=/work/TestProject.eapx`.

---

## Step 3 — check it works

Ask your assistant something that uses the server. For example, put a sample
Sparx EA file where the assistant can reach it and say:

> Using the sparx-ea tools, list the elements in `example/TestProject.eapx`.

The assistant should call `ea_query` and show you rows from the model. If you
don't have a project file yet, ask:

> What ArchiMate types does the sparx-ea server accept?

which calls `ea_archimate_types` and needs no file.

---

## Common problems

| Symptom | Cause | Fix |
| --- | --- | --- |
| Client shows the server as "failed" / "disconnected" | wrong path in the config | use the **absolute** path; run `which mcp-server-sparx-ea` (or `where` on Windows) to confirm it |
| `error while loading shared libraries: libmdb...` | the machine has the binary but not the `mdbtools` runtime | install `mdbtools` (`sudo apt-get install mdbtools` / `brew install mdbtools`) on the machine that runs the server |
| Tools don't appear after editing the config | the client wasn't restarted | fully quit and reopen the client (Claude Desktop especially) |
| `file not found` from `ea_query` | the path you gave points nowhere the server can see | give a path on the **same machine the server runs on**, absolute if unsure |
| The agent edited a model but "nothing changed in EA" | editing tools write a new `.xml`; they never touch the live project | import the tool's `output` file into EA: `File → Import → Package from XMI` — see [the editing workflow](workflow.md) |

---

## What the server can do once connected

- `ea_query` — read-only SQL against a `.eapx` project file
  ([reference](api/sql-query.md)).
- `ea_model_tree`, `ea_element`, `ea_package`, `ea_diagram`,
  `ea_archimate_types` — read an [ArchiMate](glossary/archimate.md) model the
  user exported to [XMI](glossary/xmi.md).
- `ea_create_element`, `ea_create_relationship`, … — edit a copy of that model
  and write a file you re-import into EA.

Full tool list and arguments: [docs/api/archimate.md](api/archimate.md). The
export/edit/import loop: [docs/workflow.md](workflow.md).
