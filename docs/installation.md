# Installation

The server is a single pure-Go binary with no runtime dependencies.

## Pre-built binaries

The [Releases](https://github.com/InsonusK/mcp-server-sparx-ea/releases) page has
one archive per platform — `linux`, `darwin` (macOS) and `windows`, each for
`amd64` and `arm64` — plus a `SHA256SUMS` file. Unpack the one for your machine
and put `mcp-server-sparx-ea` somewhere permanent.

Each archive is `mcp-server-sparx-ea_v<version>_<os>_<arch>.tar.gz` (`.zip` for
Windows) and unpacks to a single `mcp-server-sparx-ea` binary.

### Install script (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/InsonusK/mcp-server-sparx-ea/master/install.sh | bash
```

[`install.sh`](../install.sh) detects your OS and CPU, downloads the newest
release, checks it against `SHA256SUMS`, and installs the binary to
`/usr/local/bin` (with `sudo` if that path needs it). Options — as flags or
environment variables:

| Flag | Env var | Default |
| --- | --- | --- |
| `--version <v>` | `VERSION` | latest release |
| `--dir <path>` | `INSTALL_DIR` | `/usr/local/bin` |
| `--no-sudo` | `NO_SUDO=1` | (uses `sudo` when needed) |

```bash
# a specific version, into a dir you own, no sudo
curl -fsSL https://raw.githubusercontent.com/InsonusK/mcp-server-sparx-ea/master/install.sh \
  | INSTALL_DIR="$HOME/.local/bin" bash -s -- --version 0.5.0 --no-sudo
```

Prefer to see the script before running it? Download it, read it, then run it.

Releases are cut by `.github/workflows/release.yml` whenever `mcpserver.Version`
is bumped on `master`.

## Build from source

Needs Go 1.23+ (see `go.mod` for the exact toolchain). No C toolchain, no
system libraries.

```bash
go build -o mcp-server-sparx-ea .
```

Verify:

```bash
go test ./...
```

## Register with an MCP client

The server speaks MCP over stdio — run it with no arguments and it reads MCP
requests on stdin, writes responses on stdout. Nothing is configured through
flags or environment variables.

For a step-by-step guide covering Claude Desktop, Claude Code and Cursor, plus
troubleshooting, see **[setup-with-an-agent.md](setup-with-an-agent.md)**.

## What the server reads

The server reads model files from **its own filesystem**, by the path you pass
in a tool argument. It never fetches anything over the network.

The ArchiMate tools need a model the user exported from EA
(`File → Export → Package to XMI`, XMI 2.1). Editing tools also take an `output`
path and write the edited model there; that directory must be writable and the
path must differ from the input `file`. See [the editing workflow](workflow.md).
