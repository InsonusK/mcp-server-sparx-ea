# Installation

## Build from source

The server links the `mdbtools` C library through cgo, so it needs a C toolchain
and the `mdbtools` development headers.

### Debian / Ubuntu

```bash
sudo apt-get install -y build-essential pkg-config libglib2.0-dev mdbtools-dev
CGO_ENABLED=1 go build -o mcp-server-sparx-ea .
```

### macOS (Homebrew)

```bash
brew install mdbtools pkg-config glib
CGO_ENABLED=1 go build -o mcp-server-sparx-ea .
```

`pkg-config` must find `libmdbsql` and `glib-2.0`. Check with:

```bash
pkg-config --cflags --libs libmdbsql glib-2.0
```

### Verify

```bash
CGO_ENABLED=1 go test ./...
```

All 176 Cucumber scenarios should pass.

## Register with an MCP client

The server speaks MCP over stdio. Point your client at the built binary.

### Claude Code / Claude Desktop

```json
{
  "mcpServers": {
    "sparx-ea": {
      "command": "/absolute/path/to/mcp-server-sparx-ea"
    }
  }
}
```

### Any MCP client

Run `mcp-server-sparx-ea` with no arguments; it reads MCP requests on stdin and
writes responses on stdout. Nothing is configured through flags or environment
variables — every input is a tool argument.

## What the server needs on disk

The server reads files from **its own filesystem**, by the path you pass in a
tool argument. It never fetches anything over the network.

- `ea_query` needs a Sparx EA project file: `.eapx` (Access 2000 / JET4) or
  `.eap` (JET 3.5).
- The ArchiMate tools need a model the user exported from EA:
  `File → Export → Package to XMI`, XMI 2.1. See
  [the editing workflow](workflow.md).

Editing tools also take an `output` path and write the edited model there; that
directory must be writable and the path must differ from the input `file`.
