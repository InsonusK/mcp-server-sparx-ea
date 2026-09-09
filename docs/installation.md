# Installation

The server is a single pure-Go binary with no runtime dependencies.

## Pre-built binaries

The [Releases](https://github.com/InsonusK/mcp-server-sparx-ea/releases) page has
one archive per platform — `linux`, `darwin` (macOS) and `windows`, each for
`amd64` and `arm64` — plus a `SHA256SUMS` file. Unpack the one for your machine
and put `mcp-server-sparx-ea` somewhere permanent.

Each archive is `mcp-server-sparx-ea_v<version>_<os>_<arch>.tar.gz` (`.zip` for
Windows) and unpacks to a single `mcp-server-sparx-ea` binary.

### Download the latest release in one command (Linux / macOS)

Grabs the right archive for your OS and CPU from the newest release and unpacks
the binary into the current directory:

```bash
REPO=InsonusK/mcp-server-sparx-ea
OS=$(uname -s | tr '[:upper:]' '[:lower:]')                       # linux | darwin
ARCH=$(uname -m); case $ARCH in x86_64) ARCH=amd64;; aarch64|arm64) ARCH=arm64;; esac
curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep -o "https://[^\"]*_${OS}_${ARCH}\.tar\.gz" \
  | xargs curl -fsSL \
  | tar -xz
```

Then move it onto your `PATH`, e.g. `sudo install mcp-server-sparx-ea /usr/local/bin/`.

With the [GitHub CLI](https://cli.github.com/) instead:

```bash
gh release download --repo InsonusK/mcp-server-sparx-ea \
  --pattern "*_$(uname -s | tr '[:upper:]' '[:lower:]')_amd64.tar.gz"
tar -xzf mcp-server-sparx-ea_*.tar.gz
```

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
