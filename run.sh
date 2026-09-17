#!/usr/bin/env bash
# Emplyra one-click launcher (Linux / macOS / WSL).
#
# Starts all three pieces so you can develop with one command:
#   1. Database  — local PostgreSQL 16
#   2. Backend   — Go API server (backend/, reads backend/.env), port 8080
#   3. Frontend  — Next.js app (frontend/), port 3000
#
# Windows: use run.bat (or run.ps1) instead. WSL users can use this script.
# Ctrl+C stops the backend + frontend.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"
RUN_DIR="$ROOT/.run"
DB_USER="emplyra"
DB_NAME="emplyra"

C_GREEN='' C_YELLOW='' C_CYAN='' C_RED='' C_NC=''
if [ -t 1 ]; then
  C_GREEN='\033[0;32m'; C_YELLOW='\033[0;33m'; C_CYAN='\033[0;36m'; C_RED='\033[0;31m'; C_NC='\033[0m'
fi
step() { printf "${C_CYAN}==>${C_NC} %s\n" "$*"; }
info() { printf "${C_GREEN}[ok]${C_NC} %s\n" "$*"; }
warn() { printf "${C_YELLOW}[!]${C_NC} %s\n" "$*"; }
fail() { printf "${C_RED}[x]${C_NC} %s\n" "$*" >&2; exit 1; }

require() { command -v "$1" >/dev/null 2>&1 || fail "'$1' not found on PATH. Install it first: $2"; }

require go     "Go (https://go.dev/dl/)"
require node   "Node.js 18+ (https://nodejs.org/)"
require pnpm   "pnpm  (npm i -g pnpm)"
require psql   "psql — install PostgreSQL locally"

mkdir -p "$RUN_DIR"

# wait_port <host> <port>  — returns 0 once the TCP port accepts connections.
wait_port() {
  local i
  for i in $(seq 1 60); do
    if (exec 3<>"/dev/tcp/$1/$2") 2>/dev/null; then
      exec 3>&- 3<&- 2>/dev/null || true
      return 0
    fi
    sleep 1
  done
  return 1
}

CLEANED=0
cleanup() {
  [ "$CLEANED" = 1 ] && return
  CLEANED=1
  echo
  warn "Shutting down Emplyra..."
  if [ -n "${FRONTEND_PID:-}" ]; then
    kill "$FRONTEND_PID" 2>/dev/null || true
    pkill -P "$FRONTEND_PID" 2>/dev/null || true
  fi
  if [ -n "${BACKEND_PID:-}" ]; then
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
  info "All stopped."
}
interrupt() { exit 0; }
trap interrupt INT TERM
trap cleanup EXIT

# --- 1. Database ------------------------------------------------------------
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
step "Checking local Postgres on $DB_HOST:$DB_PORT..."
pg_isready -h "$DB_HOST" -p "$DB_PORT" >/dev/null 2>&1 \
  || fail "Postgres is not accepting connections on $DB_HOST:$DB_PORT. Is it running?"

if ! PGPASSWORD="${DB_PASSWORD:-emplyra_password}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c '\q' >/dev/null 2>&1; then
  psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "CREATE ROLE $DB_USER LOGIN PASSWORD '${DB_PASSWORD:-emplyra_password}';" 2>/dev/null \
    || warn "Could not create role '$DB_USER'. You may need to create it manually."
fi
if ! PGPASSWORD="${DB_PASSWORD:-emplyra_password}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; then
  createdb -h "$DB_HOST" -p "$DB_PORT" -O "$DB_USER" "$DB_NAME" 2>/dev/null \
    || createdb -h "$DB_HOST" -p "$DB_PORT" -U postgres -O "$DB_USER" "$DB_NAME" 2>/dev/null \
    || warn "Could not create database '$DB_NAME'. Create it manually."
fi
info "Database ready."

# --- 2. Backend -------------------------------------------------------------
step "Building backend..."
(cd "$BACKEND" && go build -o "$RUN_DIR/server" ./cmd/server)
step "Starting backend..."
"$RUN_DIR/server" &
BACKEND_PID=$!
wait_port 127.0.0.1 8080 || warn "Backend is still starting; watch its log output below."
info "Backend is listening on http://localhost:8080  (health check: /healthz)"

# --- 3. Frontend ------------------------------------------------------------
step "Installing frontend dependencies (if needed)..."
if [ ! -d "$FRONTEND/node_modules" ]; then
  (cd "$FRONTEND" && pnpm install)
else
  info "node_modules already present, skipping install."
fi
step "Starting frontend..."
(cd "$FRONTEND" && exec pnpm dev) &
FRONTEND_PID=$!
wait_port 127.0.0.1 3000 || warn "Frontend is still starting; give it a few more seconds."

echo
info "---------------- Emplyra is running ----------------"
info "Frontend  : http://localhost:3000"
info "Backend   : http://localhost:8080"
info "API       : http://localhost:8080/api/v1"
info "Health    : http://localhost:8080/healthz"
info "Admin user: admin@emplyra.local / ChangeMe123!"
echo
info "Logs for all services are visible right below. Press Ctrl+C to stop."
echo

# Block until the services stop.
set +e
wait "$BACKEND_PID" "$FRONTEND_PID"
STATUS=$?
set -e
exit "$STATUS"
