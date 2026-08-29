# Emplyra

HR management system (HRMS).

```
emplyra/
├── backend/    Go REST API (Gin + PostgreSQL + GORM + JWT)
├── frontend/   web client (planned)
├── database/   versioned SQL migrations
├── docs/       CODEBASE-MAP.md, API-CONTRACT.md
└── README.md
```

## Backend

- Go `1.27`, Gin, GORM, PostgreSQL 16, JWT (access + refresh), bcrypt.
- Role-based access control with granular permissions.
- Modules: auth, users, roles, departments, designations, employees, attendance,
  holidays, leaves, salary, payroll, recruitment, performance, training, documents,
  notifications, reports, dashboard, audit logs.
- JSON envelope: `{success, message, data|errors}`.

### Run

```bash
cd backend
cp .env.example .env      # set JWT_SECRET + super admin credentials
docker compose up -d db   # start PostgreSQL 16
go run ./cmd/server       # migrates + seeds on startup
```

Server listens on `:8080` (`GET /healthz` for health).

### Tests

```bash
cd backend
go test ./...
```

Integration tests (leave approval, attendance check-in/out) require
`TEST_DB_URL` and are skipped without it.

## Docs

- [`docs/CODEBASE-MAP.md`](docs/CODEBASE-MAP.md) — architecture, modules, run/test guide
- [`docs/API-CONTRACT.md`](docs/API-CONTRACT.md) — full endpoint + permission inventory