#!/usr/bin/env bash
set -euo pipefail

# ── Colours ──────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✔]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
err()  { echo -e "${RED}[✘]${NC} $*"; }

ROOT="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$ROOT/backend"
FRONTEND_DIR="$ROOT/frontend"
RUN_DIR="$ROOT/.run"
DB_CONTAINER="emplyra-db"

BACKEND_PORT=8080
FRONTEND_PORT=3000
DB_PORT=5432
DB_USER="emplyra"
DB_NAME="emplyra"
DB_PASSWORD="emplyra_password"
COMPOSE_OPTS=(-f "$BACKEND_DIR/docker-compose.yml" --project-directory "$BACKEND_DIR")

# ── 0. Parse flags ──────────────────────────────────
NO_DOCKER=0
for arg in "$@"; do
  case "$arg" in
    --no-docker) NO_DOCKER=1 ;;
  esac
done

# Derive the backend port from backend/.env so launch.sh and the backend always
# agree, even if a developer overrides PORT locally.
ENV_FILE="$BACKEND_DIR/.env"
if [ -f "$ENV_FILE" ]; then
  ENV_PORT="$(grep -E '^PORT=[0-9]+' "$ENV_FILE" | tail -n 1 | cut -d= -f2 2>/dev/null || true)"
  if [ -n "$ENV_PORT" ]; then
    BACKEND_PORT="$ENV_PORT"
    log "Using PORT=$BACKEND_PORT from backend/.env"
  fi
fi

mkdir -p "$RUN_DIR"

# ── 1. Check prerequisites ──────────────────────────
for cmd in go node pnpm; do
  command -v "$cmd" >/dev/null 2>&1 || { err "'$cmd' not found on PATH. Install it first."; exit 1; }
done
if [ "$NO_DOCKER" = 1 ]; then
  for cmd in psql createdb; do
    command -v "$cmd" >/dev/null 2>&1 || { err "'$cmd' not found. Postgres must be installed locally."; exit 1; }
  done
else
  command -v docker >/dev/null 2>&1 || { err "'docker' not found on PATH. Install Docker or use ./launch.sh --no-docker."; exit 1; }
fi

# ── 2. Start Postgres ───────────────────────────────
if [ "$NO_DOCKER" = 1 ]; then
  log "Using locally installed Postgres (no Docker)"
  DB_HOST="${DB_HOST:-localhost}"
  DB_PORT="${DB_PORT:-5432}"
  export DB_HOST DB_PORT DB_USER DB_NAME DB_PASSWORD

  # Ensure the role and database exist so the backend can connect.
  if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c '\q' >/dev/null 2>&1; then
    if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" >/dev/null 2>&1; then
      err "Postgres is not accepting connections on $DB_HOST:$DB_PORT. Is it running?"
      exit 1
    fi
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "CREATE ROLE $DB_USER LOGIN PASSWORD '$DB_PASSWORD';" >/dev/null 2>&1; then
      warn "Could not create role '$DB_USER'. Create it manually:"
      warn "  psql -h $DB_HOST -p $DB_PORT -U postgres -c \"CREATE ROLE $DB_USER LOGIN PASSWORD '$DB_PASSWORD';\""
    fi
  fi
  if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; then
    createdb -h "$DB_HOST" -p "$DB_PORT" -O "$DB_USER" "$DB_NAME" 2>/dev/null \
      || createdb -h "$DB_HOST" -p "$DB_PORT" -U postgres -O "$DB_USER" "$DB_NAME" 2>/dev/null \
      || warn "Could not create database '$DB_NAME'. Create it manually: createdb -h $DB_HOST -p $DB_PORT -O $DB_USER $DB_NAME"
  fi
  log "Postgres ready on $DB_HOST:$DB_PORT (role=$DB_USER db=$DB_NAME)"
else
  # Docker Postgres.
  docker info >/dev/null 2>&1 || { err "Docker is not running. Start Docker Desktop/Engine, or use ./launch.sh --no-docker."; exit 1; }

  # Select a free host port if local Postgres already uses 5432.
  if (echo > /dev/tcp/127.0.0.1/5432) 2>/dev/null; then
    warn "Port 5432 is already in use; finding a free Emplyra database port..."
    DB_PORT=""
    for candidate in $(seq 5433 5452); do
      if ! (echo > /dev/tcp/127.0.0.1/"$candidate") 2>/dev/null; then
        DB_PORT="$candidate"
        break
      fi
    done
    [ -n "$DB_PORT" ] || { err "No free port between 5433 and 5452."; exit 1; }
    cat > "$RUN_DIR/compose.override.yml" <<EOF
services:
  db:
    ports: !override
      - "127.0.0.1:${DB_PORT}:5432"
EOF
    COMPOSE_OPTS=(-f "$BACKEND_DIR/docker-compose.yml" -f "$RUN_DIR/compose.override.yml" --project-directory "$BACKEND_DIR")
    export DB_PORT="$DB_PORT"
  fi

  warn "Starting Postgres via Docker..."
  docker compose "${COMPOSE_OPTS[@]}" up -d db || { err "Could not start the Emplyra database."; exit 1; }
  for _ in $(seq 1 60); do
    if docker exec "$DB_CONTAINER" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then
      break
    fi
    sleep 1
  done
  docker exec "$DB_CONTAINER" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1 || { err "Postgres did not become ready in time."; exit 1; }
  log "Postgres ready via Docker on localhost:$DB_PORT"
fi

# ── 3. Build backend ────────────────────────────────
log "Building backend..."
(cd "$BACKEND_DIR" && go build -o "$RUN_DIR/server" ./cmd/server)

# ── 4. Kill any existing processes on our ports ──────
for port in $BACKEND_PORT $FRONTEND_PORT; do
  pid=$(lsof -ti :"$port" 2>/dev/null || true)
  if [ -n "$pid" ]; then
    warn "Killing process on port $port (PID $pid)"
    kill -9 $pid 2>/dev/null || true
    sleep 1
  fi
done

# ── 5. Start backend ────────────────────────────────
log "Starting backend on port $BACKEND_PORT..."
"$RUN_DIR/server" >"$ROOT/backend.log" 2>&1 &
BACKEND_PID=$!
sleep 2

if kill -0 $BACKEND_PID 2>/dev/null; then
  if curl -sf "http://localhost:$BACKEND_PORT/healthz" &>/dev/null; then
    log "Backend running (PID $BACKEND_PID) → http://localhost:$BACKEND_PORT"
  else
    warn "Backend started but health check pending..."
  fi
else
  err "Backend failed to start. Check $ROOT/backend.log"
  tail -20 "$ROOT/backend.log" 2>/dev/null
  exit 1
fi

# ── 6. Install + start frontend ─────────────────────
log "Installing frontend dependencies if needed..."
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
  warn "Installing frontend dependencies..."
  (cd "$FRONTEND_DIR" && pnpm install)
else
  log "Frontend dependencies already installed"
fi

log "Starting frontend on port $FRONTEND_PORT..."
(cd "$FRONTEND_DIR" && exec pnpm dev) >"$ROOT/frontend.log" 2>&1 &
FRONTEND_PID=$!
sleep 4

if kill -0 $FRONTEND_PID 2>/dev/null; then
  if curl -sf "http://localhost:$FRONTEND_PORT" &>/dev/null; then
    log "Frontend running (PID $FRONTEND_PID) → http://localhost:$FRONTEND_PORT"
  else
    warn "Frontend started but still compiling..."
  fi
else
  err "Frontend failed to start. Check $ROOT/frontend.log"
  tail -20 "$ROOT/frontend.log" 2>/dev/null
  exit 1
fi

# ── 7. Print summary ────────────────────────────────
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Emplyra — App is running!${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  Frontend:  ${GREEN}http://localhost:$FRONTEND_PORT${NC}"
echo -e "  Backend:   ${GREEN}http://localhost:$BACKEND_PORT${NC}"
echo -e "  API:       ${GREEN}http://localhost:$BACKEND_PORT/api/v1${NC}"
echo -e "  Health:    ${GREEN}http://localhost:$BACKEND_PORT/healthz${NC}"
echo -e "  Database:  ${GREEN}localhost:$DB_PORT${NC}"
echo ""
echo -e "  ${YELLOW}Default Admin:${NC}  admin@emplyra.local / ChangeMe123!"
echo ""
echo -e "  Logs:      ${NC}$ROOT/backend.log"
echo -e "             ${NC}$ROOT/frontend.log"
echo ""
echo -e "  ${YELLOW}Press Ctrl+C to stop all services${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# ── 8. Cleanup on exit ──────────────────────────────
cleanup() {
  echo ""
  warn "Shutting down..."
  kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
  wait $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
  if [ "$NO_DOCKER" != 1 ]; then
    docker compose "${COMPOSE_OPTS[@]}" stop db >/dev/null 2>&1 || true
  fi
  log "Stopped. Database data is preserved."
  exit 0
}
trap cleanup SIGINT SIGTERM

wait
# cd /media/mrrv/X/emplyra./launch.sh
