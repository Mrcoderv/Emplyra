#!/usr/bin/env bash
# Emplyra project launcher.
# Starts the database, backend, and frontend through the canonical runner.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUNNER="$ROOT/run.sh"

if [[ ! -x "$RUNNER" ]]; then
  printf 'Error: expected executable runner at %s\n' "$RUNNER" >&2
  exit 1
fi

exec "$RUNNER" "$@"
