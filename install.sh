#!/bin/sh
# Install bd from GitHub Releases. No Go required.
#
#   curl -fsSL https://raw.githubusercontent.com/OWNER/REPO/main/install.sh | sh
#
# Options (environment variables):
#   BD_VERSION      release tag to install, e.g. v0.2.0 (default: latest)
#   BD_INSTALL_DIR  where to put the binary (default: ~/.local/bin)
set -eu

REPO="${BD_REPO:-OWNER/REPO}"
BASE_URL="${BD_DOWNLOAD_URL:-https://github.com/$REPO/releases}"
VERSION="${BD_VERSION:-latest}"
INSTALL_DIR="${BD_INSTALL_DIR:-$HOME/.local/bin}"

say() { echo "bd-install: $*"; }
die() { echo "bd-install: error: $*" >&2; exit 1; }

detect_platform() {
	case $(uname -s) in
	Darwin) os=darwin ;;
	Linux) os=linux ;;
	*) die "unsupported OS $(uname -s) (bd supports macOS and Linux)" ;;
	esac
	case $(uname -m) in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) die "unsupported CPU $(uname -m) (bd supports amd64 and arm64)" ;;
	esac
}

download() { # url dest
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL -o "$2" "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget -q -O "$2" "$1"
	else
		die "need curl or wget to download bd"
	fi
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d' ' -f1
	else
		shasum -a 256 "$1" | cut -d' ' -f1
	fi
}

detect_platform
asset="bd_${os}_${arch}.tar.gz"
if [ "$VERSION" = latest ]; then
	url="$BASE_URL/latest/download"
else
	url="$BASE_URL/download/$VERSION"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

say "downloading $asset ($VERSION)"
download "$url/$asset" "$tmp/$asset" || die "could not download $url/$asset"
download "$url/checksums.txt" "$tmp/checksums.txt" || die "could not download checksums.txt"

want=$(grep " $asset\$" "$tmp/checksums.txt" | cut -d" " -f1)
got=$(sha256 "$tmp/$asset")
[ -n "$want" ] || die "no checksum listed for $asset"
[ "$want" = "$got" ] || die "checksum mismatch for $asset (download corrupted or tampered with)"

tar -xzf "$tmp/$asset" -C "$tmp"
mkdir -p "$INSTALL_DIR"
chmod 755 "$tmp/bd"
mv "$tmp/bd" "$INSTALL_DIR/bd"
say "installed $("$INSTALL_DIR/bd" --version) to $INSTALL_DIR/bd"

case ":$PATH:" in
*":$INSTALL_DIR:"*) ;;
*) say "note: $INSTALL_DIR is not on your PATH; add this to your shell profile:"
   say "  export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac

if ! command -v fzf >/dev/null 2>&1; then
	say "note: bd's note picker needs fzf, which isn't installed:"
	say "  macOS: brew install fzf    Debian/Ubuntu: sudo apt install fzf"
	say "  (without it, dumps go straight to your daily note)"
fi

say "next: bd config set vault ~/path/to/your/vault"
