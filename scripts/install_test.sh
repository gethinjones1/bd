#!/bin/sh
# Tests install.sh against a fake release served from a local directory.
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
failures=0

pass() { echo "ok   - $1"; }
fail() { echo "FAIL - $1"; failures=$((failures + 1)); }

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case $(uname -m) in x86_64 | amd64) arch=amd64 ;; arm64 | aarch64) arch=arm64 ;; esac
asset="bd_${os}_${arch}.tar.gz"

# A release dir laid out like GitHub's: <base>/download/<tag>/<asset> and
# <base>/latest/download/<asset>.
make_release() {
	rel=$1
	mkdir -p "$WORK/build" "$rel/latest/download" "$rel/download/v9.9.9"
	(cd "$ROOT" && go build -ldflags "-X main.version=9.9.9" -o "$WORK/build/bd" ./cmd/bd)
	tar -czf "$rel/latest/download/$asset" -C "$WORK/build" bd
	(cd "$rel/latest/download" && shasum -a 256 "$asset" > checksums.txt)
	cp "$rel/latest/download/"* "$rel/download/v9.9.9/"
}

make_release "$WORK/rel"

# 1. installs the latest release into BD_INSTALL_DIR
out=$(BD_DOWNLOAD_URL="file://$WORK/rel" BD_INSTALL_DIR="$WORK/bin" sh "$ROOT/install.sh" 2>&1) || true
if [ -x "$WORK/bin/bd" ] && [ "$("$WORK/bin/bd" --version)" = "bd 9.9.9" ]; then
	pass "installs latest release"
else
	fail "installs latest release: $out"
fi

# 2. a pinned version is fetched from download/<tag>
rm -rf "$WORK/bin"
out=$(BD_VERSION=v9.9.9 BD_DOWNLOAD_URL="file://$WORK/rel" BD_INSTALL_DIR="$WORK/bin" sh "$ROOT/install.sh" 2>&1) || true
if [ -x "$WORK/bin/bd" ]; then pass "installs pinned version"; else fail "installs pinned version: $out"; fi

# 3. a checksum mismatch aborts without installing
rm -rf "$WORK/bin"
echo "0000000000000000000000000000000000000000000000000000000000000000  $asset" > "$WORK/rel/latest/download/checksums.txt"
if out=$(BD_DOWNLOAD_URL="file://$WORK/rel" BD_INSTALL_DIR="$WORK/bin" sh "$ROOT/install.sh" 2>&1); then
	fail "checksum mismatch should fail"
elif [ -e "$WORK/bin/bd" ]; then
	fail "checksum mismatch still installed bd"
else
	case $out in *checksum*) pass "checksum mismatch aborts" ;; *) fail "checksum mismatch message: $out" ;; esac
fi
make_release "$WORK/rel"

# 4. warns when fzf is not on PATH (bd still installs)
rm -rf "$WORK/bin"
mkdir -p "$WORK/nofzf"
for tool in sh uname tr curl tar mktemp shasum mkdir mv chmod rm cat grep awk cut dirname cp; do
	p=$(command -v "$tool") && ln -sf "$p" "$WORK/nofzf/$tool"
done
out=$(PATH="$WORK/nofzf" BD_DOWNLOAD_URL="file://$WORK/rel" BD_INSTALL_DIR="$WORK/bin" sh "$ROOT/install.sh" 2>&1) || true
case $out in
*fzf*) [ -x "$WORK/bin/bd" ] && pass "warns about missing fzf" || fail "missing fzf blocked install: $out" ;;
*) fail "no fzf warning: $out" ;;
esac

# 5. a checksums file that doesn't list the asset is a clear error
rm -rf "$WORK/bin"
echo "abc  some_other_file.tar.gz" > "$WORK/rel/latest/download/checksums.txt"
if out=$(BD_DOWNLOAD_URL="file://$WORK/rel" BD_INSTALL_DIR="$WORK/bin" sh "$ROOT/install.sh" 2>&1); then
	fail "unlisted asset should fail"
else
	case $out in *"no checksum listed"*) pass "unlisted asset explains itself" ;; *) fail "unlisted asset message: $out" ;; esac
fi

[ "$failures" -eq 0 ] && echo "all install tests passed" || { echo "$failures failed"; exit 1; }
