#!/usr/bin/env bash
# ============================================================
#  1bot build script (Linux / macOS)
#  Produces: gateway, plugin-codebuddy, plugin-codex
#  Usage:    ./build.sh          -> build all
#            ./build.sh clean    -> delete binaries then build
# ============================================================

# --- locate Go: PATH first, fall back to common install paths ---
GO="$(command -v go 2>/dev/null || true)"
if [ -z "$GO" ]; then
    for p in /usr/local/go/bin/go /usr/lib/go/bin/go "$HOME/go/bin/go"; do
        if [ -x "$p" ]; then
            GO="$p"
            break
        fi
    done
fi
if [ -z "$GO" ]; then
    echo "[FAIL] go not found in PATH or common locations" >&2
    exit 1
fi

# --- pure Go, no CGO (per DESIGN.md ADR-6) ---
export CGO_ENABLED=0

# --- run from this script's directory ---
cd "$(dirname "$0")" || exit 1

# --- optional clean ---
if [ "${1:-}" = "clean" ]; then
    echo "cleaning old binaries ..."
    rm -f gateway plugin-codebuddy plugin-codex
fi

echo
echo "[1/3] building gateway ..."
"$GO" build -trimpath -o gateway ./cmd/gateway || { echo "[FAIL] gateway build failed."; exit 1; }

echo "[2/3] building plugin-codebuddy ..."
"$GO" build -trimpath -o plugin-codebuddy ./cmd/plugin-codebuddy || { echo "[FAIL] plugin-codebuddy build failed."; exit 1; }

echo "[3/3] building plugin-codex ..."
"$GO" build -trimpath -o plugin-codex ./cmd/plugin-codex || { echo "[FAIL] plugin-codex build failed."; exit 1; }

echo
echo "build ok. binaries:"
ls -1 gateway plugin-codebuddy plugin-codex

echo
echo "run:"
echo "  ./gateway -plugin ./plugin-codebuddy -model gpt-6.1-sol -y"
echo "  ./gateway -plugin ./plugin-codex -model gpt-5.4"
echo
exit 0
