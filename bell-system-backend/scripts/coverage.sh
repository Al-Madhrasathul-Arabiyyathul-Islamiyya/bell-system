#!/usr/bin/env bash
# coverage.sh — Run unit + integration tests with coverage, merge profiles, and report.
#
# Designed to run in WSL where testcontainers works properly.
#
# Usage (from Windows):
#   wsl -e bash -c "cd /mnt/d/Stuff/projects/bell-system/bell-system-backend && bash scripts/coverage.sh"

set -euo pipefail

# --- Environment setup (no reliance on login shell / .bashrc) ---
export GOPATH="${GOPATH:-$HOME/go}"
export PATH="/usr/local/go/bin:$GOPATH/bin:/usr/local/bin:/usr/bin:/bin:$PATH"

# https://github.com/golang/go/issues/75031
export GOTOOLCHAIN=go1.26.1+auto

# x509negativeserial: testcontainers TLS certs use negative serial numbers
export GODEBUG=x509negativeserial=1

# Verify go is available
if ! command -v go &>/dev/null; then
  echo "ERROR: go not found on PATH" >&2
  exit 1
fi
echo "Using $(go version)"

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_DIR"

COVER_DIR="$PROJECT_DIR/coverage"
rm -rf "$COVER_DIR"
mkdir -p "$COVER_DIR"

# Only measure coverage for production packages (exclude tests/ and mocks)
COVERPKGS="arabiyya.edu.mv/bell-system-backend/cmd/...,arabiyya.edu.mv/bell-system-backend/config,arabiyya.edu.mv/bell-system-backend/internal/...,arabiyya.edu.mv/bell-system-backend/pkg/..."

echo ""
echo "=== Unit tests ==="
go test ./... \
  -count=1 \
  -coverprofile="$COVER_DIR/unit.out" \
  -coverpkg="$COVERPKGS"

echo ""
echo "=== Integration tests ==="
go test -tags integration \
  -count=1 \
  -timeout 300s \
  -coverprofile="$COVER_DIR/integration.out" \
  -coverpkg="$COVERPKGS" \
  ./tests/integration/

echo ""
echo "=== Merging coverage profiles ==="
head -1 "$COVER_DIR/unit.out" > "$COVER_DIR/merged.out"
tail -n +2 "$COVER_DIR/unit.out" >> "$COVER_DIR/merged.out"
tail -n +2 "$COVER_DIR/integration.out" >> "$COVER_DIR/merged.out"

echo ""
echo "--- Unit only ---"
go tool cover -func="$COVER_DIR/unit.out" | tail -1

echo ""
echo "--- Integration only ---"
go tool cover -func="$COVER_DIR/integration.out" | tail -1

echo ""
echo "--- Merged (unit + integration) ---"
go tool cover -func="$COVER_DIR/merged.out" | tail -1

echo ""
echo "=== Files below 100% (merged) ==="
go tool cover -func="$COVER_DIR/merged.out" | grep -v "100.0%" | grep -v "^total:" || echo "(all files at 100%)"

echo ""
echo "Profiles: $COVER_DIR/{unit,integration,merged}.out"
echo "HTML:     go tool cover -html=$COVER_DIR/merged.out -o $COVER_DIR/coverage.html"
