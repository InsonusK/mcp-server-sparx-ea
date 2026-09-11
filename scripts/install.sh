#!/usr/bin/env bash
# Install mcp-server-sparx-ea from a GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/InsonusK/mcp-server-sparx-ea/master/scripts/install.sh | bash
#
# Options (flags or environment variables):
#   --version <v>       VERSION       release to install (default: latest)
#   --dir <path>        INSTALL_DIR   where to put the binary (default: /usr/local/bin)
#   --no-sudo           NO_SUDO=1     never call sudo; fail if INSTALL_DIR is not writable
#   --register <scope>  REGISTER      register with Claude Code: "project" writes
#                                     ./.mcp.json (committable), "user" adds it to
#                                     your user config, "no" skips (default: ask if
#                                     interactive, otherwise "no")
#   --no-register       REGISTER=no   never register; do not prompt
#
# Supports linux and darwin (macOS) on amd64 and arm64. On Windows use
# scripts/install.ps1.
set -euo pipefail

REPO="InsonusK/mcp-server-sparx-ea"
BINARY="mcp-server-sparx-ea"
MCP_NAME="${MCP_NAME:-sparx-ea}"
VERSION="${VERSION:-}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
NO_SUDO="${NO_SUDO:-}"
REGISTER="${REGISTER:-}"

while [ $# -gt 0 ]; do
	case "$1" in
		--version)     VERSION="${2:?--version needs a value}"; shift 2 ;;
		--dir)         INSTALL_DIR="${2:?--dir needs a value}"; shift 2 ;;
		--no-sudo)     NO_SUDO=1; shift ;;
		--register)    REGISTER="${2:?--register needs a value (project|user|no)}"; shift 2 ;;
		--no-register) REGISTER=no; shift ;;
		-h|--help) sed -n '2,17p' "$0" | sed 's/^#\( \|$\)//'; exit 0 ;;
		*) echo "unknown option: $1" >&2; exit 2 ;;
	esac
done

die() { echo "install.sh: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "need '$1' on PATH"; }

case "$REGISTER" in
	""|project|user|no|none) ;;
	*) die "--register must be project, user or no" ;;
esac

need curl
need tar

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
	x86_64|amd64) ARCH=amd64 ;;
	aarch64|arm64) ARCH=arm64 ;;
	*) die "unsupported architecture: $ARCH" ;;
esac
case "$OS" in
	linux|darwin) ;;
	*) die "unsupported OS: $OS (use the Windows .zip from the Releases page)" ;;
esac

api="https://api.github.com/repos/$REPO/releases"
if [ -z "$VERSION" ]; then
	VERSION="$(curl -fsSL "$api/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)"
	[ -n "$VERSION" ] || die "could not resolve the latest release tag"
fi
VERSION="${VERSION#v}"

archive="${BINARY}_v${VERSION}_${OS}_${ARCH}.tar.gz"
base="https://github.com/$REPO/releases/download/v${VERSION}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "downloading $archive ..."
curl -fsSL -o "$tmp/$archive" "$base/$archive" \
	|| die "download failed — check that release v${VERSION} has an asset for ${OS}/${ARCH}"

if curl -fsSL -o "$tmp/SHA256SUMS" "$base/SHA256SUMS" 2>/dev/null; then
	echo "verifying checksum ..."
	sum_tool=""
	command -v sha256sum >/dev/null 2>&1 && sum_tool="sha256sum"
	command -v shasum    >/dev/null 2>&1 && sum_tool="${sum_tool:-shasum -a 256}"
	if [ -n "$sum_tool" ]; then
		want="$(grep " $archive\$" "$tmp/SHA256SUMS" | awk '{print $1}')"
		got="$( $sum_tool "$tmp/$archive" | awk '{print $1}')"
		[ -n "$want" ] || die "no checksum for $archive in SHA256SUMS"
		[ "$want" = "$got" ] || die "checksum mismatch: want $want, got $got"
	else
		echo "  (no sha256sum/shasum — skipping)"
	fi
fi

tar -xzf "$tmp/$archive" -C "$tmp"
[ -f "$tmp/$BINARY" ] || die "archive did not contain $BINARY"
chmod +x "$tmp/$BINARY"

mkdir -p "$INSTALL_DIR" 2>/dev/null || true
if [ -w "$INSTALL_DIR" ]; then
	mv "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
elif [ -z "$NO_SUDO" ] && command -v sudo >/dev/null 2>&1; then
	echo "installing to $INSTALL_DIR (needs sudo) ..."
	sudo install -m 0755 "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
else
	die "$INSTALL_DIR is not writable (set --dir to a writable path, or drop --no-sudo)"
fi

echo "installed $BINARY v${VERSION} -> $INSTALL_DIR/$BINARY"
on_path=""
case ":$PATH:" in
	*":$INSTALL_DIR:"*) on_path=1 ;;
	*) echo "note: $INSTALL_DIR is not on your PATH" ;;
esac

# --- register with Claude Code -----------------------------------------------
# A "project" registration writes ./.mcp.json in the current directory, which you
# commit so the whole team gets the server. A "user" registration adds it to your
# personal Claude Code config. See docs/setup-with-an-agent.md.

# The command to record: bare name if INSTALL_DIR is on PATH (portable across a
# team), otherwise the absolute path.
if [ -n "$on_path" ]; then mcp_cmd="$BINARY"; else mcp_cmd="$INSTALL_DIR/$BINARY"; fi

if [ -z "$REGISTER" ]; then
	if [ -e /dev/tty ]; then
		printf 'Register "%s" with Claude Code? [p]roject (./.mcp.json), [u]ser, [N]o: ' "$MCP_NAME" > /dev/tty
		read -r ans < /dev/tty || ans=""
		case "$ans" in
			p|P|project) REGISTER=project ;;
			u|U|user)    REGISTER=user ;;
			*)           REGISTER=no ;;
		esac
	else
		REGISTER=no
	fi
fi

write_project_mcp_json() {
	file="./.mcp.json"
	if command -v python3 >/dev/null 2>&1; then
		python3 - "$file" "$MCP_NAME" "$mcp_cmd" <<'PY'
import json, os, sys
path, name, cmd = sys.argv[1:4]
data = {}
if os.path.exists(path):
    try:
        with open(path) as f:
            data = json.load(f)
    except ValueError:
        data = {}
if not isinstance(data, dict):
    data = {}
data.setdefault("mcpServers", {})[name] = {"command": cmd, "args": []}
with open(path, "w") as f:
    json.dump(data, f, indent=2)
    f.write("\n")
PY
		return 0
	fi
	if command -v jq >/dev/null 2>&1; then
		[ -f "$file" ] || echo '{}' > "$file"
		tmpf="$(mktemp)"
		jq --arg n "$MCP_NAME" --arg c "$mcp_cmd" \
			'.mcpServers = (.mcpServers // {}) | .mcpServers[$n] = {command: $c, args: []}' \
			"$file" > "$tmpf" && mv "$tmpf" "$file"
		return 0
	fi
	if [ -f "$file" ]; then
		echo "install.sh: $file already exists and neither python3 nor jq is available to merge" >&2
		echo "  add this by hand:  \"$MCP_NAME\": { \"command\": \"$mcp_cmd\", \"args\": [] }" >&2
		return 1
	fi
	cat > "$file" <<EOF
{
  "mcpServers": {
    "$MCP_NAME": {
      "command": "$mcp_cmd",
      "args": []
    }
  }
}
EOF
}

case "$REGISTER" in
	project)
		if command -v claude >/dev/null 2>&1; then
			claude mcp add --scope project "$MCP_NAME" -- "$mcp_cmd" \
				&& echo "registered \"$MCP_NAME\" in ./.mcp.json (commit it to share)" \
				|| echo "note: 'claude mcp add' failed — see docs/setup-with-an-agent.md"
		else
			write_project_mcp_json \
				&& echo "wrote ./.mcp.json — commit it so the team gets \"$MCP_NAME\"" \
				|| true
		fi
		;;
	user)
		if command -v claude >/dev/null 2>&1; then
			claude mcp add --scope user "$MCP_NAME" -- "$mcp_cmd" \
				&& echo "registered \"$MCP_NAME\" in your Claude Code user config" \
				|| echo "note: 'claude mcp add' failed — see docs/setup-with-an-agent.md"
		else
			echo "note: the 'claude' CLI is not on PATH — install it, then run:"
			echo "  claude mcp add --scope user $MCP_NAME -- $mcp_cmd"
		fi
		;;
	*) ;;
esac
