#!/usr/bin/env bash
# Install mcp-server-sparx-ea from a GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/InsonusK/mcp-server-sparx-ea/master/scripts/install.sh | bash
#
# Options (flags or environment variables):
#   --version <v>   VERSION       release to install (default: latest)
#   --dir <path>    INSTALL_DIR   where to put the binary (default: /usr/local/bin)
#   --no-sudo       NO_SUDO=1     never call sudo; fail if INSTALL_DIR is not writable
#
# Supports linux and darwin (macOS) on amd64 and arm64. On Windows use
# scripts/install.ps1.
set -euo pipefail

REPO="InsonusK/mcp-server-sparx-ea"
BINARY="mcp-server-sparx-ea"
VERSION="${VERSION:-}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
NO_SUDO="${NO_SUDO:-}"

while [ $# -gt 0 ]; do
	case "$1" in
		--version) VERSION="${2:?--version needs a value}"; shift 2 ;;
		--dir)     INSTALL_DIR="${2:?--dir needs a value}"; shift 2 ;;
		--no-sudo) NO_SUDO=1; shift ;;
		-h|--help) sed -n '2,12p' "$0" | sed 's/^#\( \|$\)//'; exit 0 ;;
		*) echo "unknown option: $1" >&2; exit 2 ;;
	esac
done

die() { echo "install.sh: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "need '$1' on PATH"; }

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
case ":$PATH:" in
	*":$INSTALL_DIR:"*) ;;
	*) echo "note: $INSTALL_DIR is not on your PATH" ;;
esac
