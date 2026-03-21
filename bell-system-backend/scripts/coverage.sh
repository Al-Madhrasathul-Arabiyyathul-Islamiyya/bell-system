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

# Verify Docker is running (required for integration tests via testcontainers)
if ! docker info &>/dev/null; then
  echo ""
  echo "Docker is not running. Attempting to start Docker Desktop..."
  # Try to launch Docker Desktop from the Windows side
  DOCKER_DESKTOP="/mnt/c/Program Files/Docker/Docker/Docker Desktop.exe"
  if [[ -x "$DOCKER_DESKTOP" ]]; then
    "$DOCKER_DESKTOP" &>/dev/null &
    echo "Waiting for Docker to become ready (up to 60s)..."
    for i in $(seq 1 60); do
      if docker info &>/dev/null; then
        echo "Docker is ready."
        break
      fi
      if (( i == 60 )); then
        echo "ERROR: Docker did not become ready within 60s." >&2
        echo "       Start Docker Desktop manually, then re-run this script." >&2
        exit 1
      fi
      sleep 1
    done
  else
    echo "ERROR: Docker is not running and Docker Desktop was not found." >&2
    echo "       Start Docker manually, then re-run this script." >&2
    exit 1
  fi
fi

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_DIR"

COVER_DIR="$PROJECT_DIR/coverage"
rm -rf "$COVER_DIR"
mkdir -p "$COVER_DIR"

# Only measure coverage for production packages (exclude tests/, mocks, and cmd/migrate)
COVERPKGS="arabiyya.edu.mv/bell-system-backend/cmd/server,arabiyya.edu.mv/bell-system-backend/config,arabiyya.edu.mv/bell-system-backend/internal/...,arabiyya.edu.mv/bell-system-backend/pkg/..."

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

# Merge binary coverage if TestBinaryCoverage_Server produced it
if [[ -f "$COVER_DIR/binary.out" ]]; then
  tail -n +2 "$COVER_DIR/binary.out" >> "$COVER_DIR/merged.out"
  echo "(binary coverage profile merged)"
fi

echo ""
echo "--- Unit only ---"
go tool cover -func="$COVER_DIR/unit.out" | tail -1

echo ""
echo "--- Integration only ---"
go tool cover -func="$COVER_DIR/integration.out" | tail -1

if [[ -f "$COVER_DIR/binary.out" ]]; then
  echo ""
  echo "--- Binary only ---"
  go tool cover -func="$COVER_DIR/binary.out" | tail -1
fi

echo ""
echo "--- Merged (unit + integration + binary) ---"
go tool cover -func="$COVER_DIR/merged.out" | tail -1

echo ""
echo "=== Files below 100% (merged) ==="
go tool cover -func="$COVER_DIR/merged.out" | grep -v "100.0%" | grep -v "^total:" || echo "(all files at 100%)"

echo ""
echo "Profiles: $COVER_DIR/{unit,integration,merged}.out"
echo "HTML:     go tool cover -html=$COVER_DIR/merged.out -o $COVER_DIR/coverage.html"
