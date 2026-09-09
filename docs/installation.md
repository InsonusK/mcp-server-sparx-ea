# Installation

## Pre-built

- **Release binaries** — the [Releases](https://github.com/InsonusK/mcp-server-sparx-ea/releases)
  page has one archive per platform (`linux_amd64`, `linux_arm64`,
  `darwin_amd64`, `darwin_arm64`), each with a `.sha256`. They link `mdbtools`
  dynamically, so the target machine still needs the runtime libraries:
  `sudo apt-get install -y mdbtools libglib2.0-0` / `brew install mdbtools glib`.
- **Container** — `ghcr.io/insonusk/mcp-server-sparx-ea:<version>` (and
  `:latest`), `linux/amd64` and `linux/arm64`, with the libraries bundled.

Both are published by `.github/workflows/release.yml` whenever `mcpserver.Version`
is bumped on `master`.

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

The server speaks MCP over stdio — run it with no arguments and it reads MCP
requests on stdin, writes responses on stdout. Nothing is configured through
flags or environment variables.

For a step-by-step guide covering Claude Desktop, Claude Code and Cursor, plus
troubleshooting, see **[setup-with-an-agent.md](setup-with-an-agent.md)**.

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
