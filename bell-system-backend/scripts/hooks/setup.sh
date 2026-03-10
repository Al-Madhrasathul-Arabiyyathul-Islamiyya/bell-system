#!/usr/bin/env bash
set -e

REPO_ROOT="$(git rev-parse --show-toplevel)"
HOOKS_DIR="$REPO_ROOT/bell-system-backend/scripts/hooks"

git config core.hooksPath "$HOOKS_DIR"
chmod +x "$HOOKS_DIR/pre-commit" "$HOOKS_DIR/pre-push"

echo "Git hooks configured. Hooks path set to: $HOOKS_DIR"
