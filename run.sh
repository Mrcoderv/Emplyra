#!/usr/bin/env bash
# Emplyra one-click launcher (Linux / macOS / WSL).
#
# Starts all three pieces so you can develop with one command:
#   1. Database  — Postgres 16 via Docker (backend/docker-compose.yml)
#   2. Backend   — Go API server (backend/, reads backend/.env), port 8080
#   3. Frontend  — Next.js app (frontend/), port 3000
#
# Windows: use run.bat (or run.ps1) instead. WSL users can use this script.
# Ctrl+C stops the backend + frontend and stops (not removes) the database
# container. Database data lives in the named docker volume, so it survives.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"
RUN_DIR="$ROOT/.run"
DB_CONTAINER="emplyra-db"
DB_USER="emplyra"
DB_NAME="emplyra"
DB_HOST_PORT="5432"
COMPOSE_OPTS=(-f "$BACKEND/docker-compose.yml" --project-directory "$BACKEND")

C_GREEN='' C_YELLOW='' C_CYAN='' C_RED='' C_NC=''
if [ -t 1 ]; then
  C_GREEN='\033[0;32m'; C_YELLOW='\033[0;33m'; C_CYAN='\033[0;36m'; C_RED='\033[0;31m'; C_NC='\033[0m'
fi
step() { printf "${C_CYAN}==>${C_NC} %s\n" "$*"; }
info() { printf "${C_GREEN}[ok]${C_NC} %s\n" "$*"; }
warn() { printf "${C_YELLOW}[!]${C_NC} %s\n" "$*"; }
fail() { printf "${C_RED}[x]${C_NC} %s\n" "$*" >&2; exit 1; }

require() { command -v "$1" >/dev/null 2>&1 || fail "'$1' not found on PATH. Install it first: $2"; }

require docker "Docker Desktop / Docker Engine (needed for the Postgres database)"
require go     "Go (https://go.dev/dl/)"
require node   "Node.js 18+ (https://nodejs.org/)"
require pnpm   "pnpm  (npm i -g pnpm)"

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
  step "Stopping database container..."
  "${COMPOSE_OPTS[@]}" stop db >/dev/null 2>&1 || true
  info "All stopped. Database data is preserved in the docker volume (emplyra_pgdata)."
}
interrupt() { exit 0; }
trap interrupt INT TERM
trap cleanup EXIT

# --- 1. Database ------------------------------------------------------------
step "Starting database (Postgres 16 via Docker)..."
docker info >/dev/null 2>&1 || fail "Docker is not running. Start Docker Desktop/Engine and retry."

# If host port 5432 is already taken (e.g. by a locally installed Postgres),
# pick the next free port for the database and point the backend at it.
if wait_port 127.0.0.1 5432; then
  warn "Port 5432 is already in use (a local Postgres?). Finding a free port for the Emplyra database..."
  DB_HOST_PORT=""
  for candidate in $(seq 5433 5452); do
    if ! wait_port 127.0.0.1 "$candidate"; then
      DB_HOST_PORT="$candidate"
      break
    fi
  done
  [ -n "$DB_HOST_PORT" ] || fail "No free port between 5433-5452 for the Emplyra database."
  cat > "$RUN_DIR/compose.override.yml" <<EOF
services:
  db:
    ports: !override
      - "127.0.0.1:${DB_HOST_PORT}:5432"
EOF
  COMPOSE_OPTS=(-f "$BACKEND/docker-compose.yml" -f "$RUN_DIR/compose.override.yml" --project-directory "$BACKEND")
  export DB_PORT="$DB_HOST_PORT"
  info "Using host port $DB_HOST_PORT for the Emplyra database (backend DB_PORT=$DB_HOST_PORT)."
fi

"${COMPOSE_OPTS[@]}" up -d db \
  || fail "Could not start the database container."

info "Waiting for Postgres to accept connections..."
tries=0
until docker exec "$DB_CONTAINER" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; do
  tries=$((tries + 1))
  [ "$tries" -ge 60 ] && fail "Postgres did not become ready in time."
  sleep 1
done
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

# Block until the services stop. If a service dies on its own, wait() returns
# non-zero and the EXIT trap shuts down whatever is still running.
set +e
wait "$BACKEND_PID" "$FRONTEND_PID"
STATUS=$?
set -e
exit "$STATUS"