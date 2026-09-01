#!/usr/bin/env bash
set -euo pipefail

# Emplyra one-click launcher for Linux, macOS, and WSL.
# Starts Postgres, the Go backend, and the Next.js frontend.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"
RUN_DIR="$ROOT/.run"
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

require docker "Docker Desktop / Docker Engine"
require go "Go (https://go.dev/dl/)"
require node "Node.js 18+ (https://nodejs.org/)"
require pnpm "pnpm (npm i -g pnpm)"

mkdir -p "$RUN_DIR"

port_open() {
  (exec 3<>"/dev/tcp/$1/$2") 2>/dev/null
}
wait_port() {
  for _ in $(seq 1 60); do
    if port_open "$1" "$2"; then return 0; fi
    sleep 1
  done
  return 1
}

# Select a free host port if local Postgres already uses 5432.
if port_open 127.0.0.1 5432; then
  warn "Port 5432 is already in use; finding a free Emplyra database port..."
  DB_HOST_PORT=""
  for candidate in $(seq 5433 5452); do
    if ! port_open 127.0.0.1 "$candidate"; then
      DB_HOST_PORT="$candidate"
      break
    fi
  done
  [ -n "$DB_HOST_PORT" ] || fail "No free port between 5433 and 5452."
  cat > "$RUN_DIR/compose.override.yml" <<EOF
services:
  db:
    ports: !override
      - "127.0.0.1:${DB_HOST_PORT}:5432"
EOF
  COMPOSE_OPTS=(-f "$BACKEND/docker-compose.yml" -f "$RUN_DIR/compose.override.yml" --project-directory "$BACKEND")
  export DB_PORT="$DB_HOST_PORT"
fi

CLEANED=0
cleanup() {
  [ "$CLEANED" = 1 ] && return
  CLEANED=1
  echo
  warn "Shutting down Emplyra..."
  kill "${FRONTEND_PID:-}" "${BACKEND_PID:-}" 2>/dev/null || true
  pkill -P "${FRONTEND_PID:-}" 2>/dev/null || true
  "${COMPOSE_OPTS[@]}" stop db >/dev/null 2>&1 || true
  info "All stopped. Database data is preserved."
}
trap cleanup EXIT INT TERM

step "Starting Postgres database..."
docker info >/dev/null 2>&1 || fail "Docker is not running. Start Docker Desktop/Engine and retry."
"${COMPOSE_OPTS[@]}" up -d db || fail "Could not start the Emplyra database."
for _ in $(seq 1 60); do
  if docker exec emplyra-db pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then break; fi
  sleep 1
done
docker exec emplyra-db pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1 || fail "Postgres did not become ready in time."
info "Database ready on localhost:$DB_HOST_PORT"

step "Building backend..."
(cd "$BACKEND" && go build -o "$RUN_DIR/server" ./cmd/server)
step "Starting backend on port 8080..."
"$RUN_DIR/server" >"$ROOT/backend.log" 2>&1 &
BACKEND_PID=$!
wait_port 127.0.0.1 8080 || warn "Backend health check is still starting."
info "Backend: http://localhost:8080"

step "Installing frontend dependencies if needed..."
if [ ! -d "$FRONTEND/node_modules" ]; then (cd "$FRONTEND" && pnpm install); else info "Frontend dependencies already installed."; fi
step "Starting frontend on port 3000..."
(cd "$FRONTEND" && exec pnpm dev) >"$ROOT/frontend.log" 2>&1 &
FRONTEND_PID=$!
wait_port 127.0.0.1 3000 || warn "Frontend is still compiling."

printf '\n'
info "---------------- Emplyra is running ----------------"
info "Frontend: http://localhost:3000"
info "Backend : http://localhost:8080"
info "API     : http://localhost:8080/api/v1"
info "Health  : http://localhost:8080/healthz"
info "Logs    : $ROOT/backend.log and $ROOT/frontend.log"
printf '\nPress Ctrl+C to stop all services.\n\n'

set +e
wait "$BACKEND_PID" "$FRONTEND_PID"
exit $?
