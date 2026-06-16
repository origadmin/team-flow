#!/usr/bin/env bash
# team-flow v3 install script
# Usage: ./install.sh [DESTDIR]
#   Default DESTDIR: $GOPATH/bin (or %GOPATH%\bin on Windows)

set -e

DESTDIR="${1:-$GOPATH/bin}"
if [ -z "$DESTDIR" ]; then
    DESTDIR="$HOME/go/bin"
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/scripts/flow"

echo "[install] Building flow v3..."
(cd "$SCRIPT_DIR" && go build -o "$BIN" ./cmd/flow)

echo "[install] Installing to $DESTDIR/flow..."
if command -v install >/dev/null 2>&1; then
    install -m 0755 "$BIN" "$DESTDIR/flow"
else
    cp "$BIN" "$DESTDIR/flow"
    chmod +x "$DESTDIR/flow"
fi

echo "[install] Done."
echo ""
echo "Verify:  flow --version"
echo "Reinstall: $0"
echo "Uninstall: rm $DESTDIR/flow"
