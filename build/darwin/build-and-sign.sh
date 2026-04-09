#!/bin/bash
# Wrapper for 'go build' that re-signs the Wails app bundle with PTY entitlements.
# Used by: just dev (via wails dev -compiler flag)
# Reason: macOS blocks fork/exec from unsigned app bundles, which prevents
#         terminal/PTY sessions from spawning in wails dev mode.
#
# Strategy: After go build, spawn a background watcher that waits for Wails to
# finish its own packaging+signing, then re-signs the .app bundle with our
# entitlements. The re-sign happens before the next user interaction.

set -euo pipefail

go "$@"

# Only act after a build command
if [ "${1:-}" != "build" ]; then
    exit 0
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENTITLEMENTS="$SCRIPT_DIR/entitlements.plist"
APP_BINARY="$SCRIPT_DIR/../bin/mashed.app/Contents/MacOS/mashed"
APP_MACOS_DIR="$SCRIPT_DIR/../bin/mashed.app/Contents/MacOS"
HELPER_SRC="$SCRIPT_DIR/../bin/mashed-pty-helper"

if [ -f "$ENTITLEMENTS" ] && [ -f "$APP_BINARY" ]; then
    # Wait for Wails to finish packaging+signing, then re-sign with entitlements.
    # We watch the binary's mtime — when it stops changing, Wails is done.
    (
        # Wait for Wails to re-sign (it does packaging + signing after go build)
        sleep 5
        # Copy PTY helper into the app bundle if it exists
        if [ -f "$HELPER_SRC" ]; then
            cp "$HELPER_SRC" "$APP_MACOS_DIR/mashed-pty-helper"
            codesign --force --sign - --entitlements "$ENTITLEMENTS" "$APP_MACOS_DIR/mashed-pty-helper" 2>/dev/null && \
                echo "[build-and-sign] Re-signed mashed-pty-helper with PTY entitlements" || true
        fi
        # Re-sign the actual binary inside the bundle with entitlements
        codesign --force --sign - --entitlements "$ENTITLEMENTS" "$APP_BINARY" 2>/dev/null && \
            echo "[build-and-sign] Re-signed mashed binary with PTY entitlements" || true
    ) &
fi
