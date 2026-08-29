# CODEBASE-MAP

> ⚠️ **Status: GREENFIELD → IMPLEMENTED.** This document was created as the "map first"
> deliverable before implementation began, so its sections describe the *plan*. All seven
> implementation phases (§12) are now **complete** and verified (`gofmt` clean, `go vet`
> clean, `go build ./...` passes, unit tests green). See **§13 Implementation Status** for
> the authoritative, up-to-date description of the finished backend, and
> `docs/API-CONTRACT.md` for the finalized endpoint inventory.
>
> The repository began empty (no Go module, no commits, no frontend, no database).

---

## 1. Current Architecture

- **Repository root:** `/media/mrrv/X/emplyra` (project: Emplyra)
  - `backend/` — Go source (`github.com/emplyra/backend`)
  - `database/migrations/` — versioned SQL
  - `frontend/` — web client (empty, planned)
  - `docs/` — this map + API contract
- **Existing code:** none. Verified empty directory.
- **Existing git history:** none (`git status` fails — directory is not a git repo).
- **Existing toolchain detected:**
  - Go `1.27.0` (linux/amd64)
  - Docker `29.7.2` + Docker Compose `v2.40.3`
  - Network access to `proxy.golang.org` confirmed

Because the project is new, the architecture below is the *planned* architecture, built
from scratch. There is no legacy code to preserve or integrate with.

## 2. Target Architecture

Layered backend following the required pattern:

```
HTTP Handler  →  Service  →  Repository  →  Database
```

- **Handlers** (`internal/handlers`) — HTTP concerns only: request parse, bind/validate,
  response formatting. No business logic.
- **Services** (`internal/services`) — business rules, validations, transactions
  (payroll processing, leave approval, onboarding, role changes).
- **Repositories** (`internal/repositories`) — GORM based database operations.
- **Models** (`internal/models`) — GORM domain models.
- **DTOs** (`internal/dto`) — API request/response payloads with `validator` tags.

### Package layout

```
cmd/server/            entrypoint
internal/
  config/              env-based configuration (.env)
  database/            PostgreSQL + GORM connection, AutoMigrate
  migrate/             versioned SQL migration runner
  models/              GORM models
  dto/                 request/response DTOs + validation
  repositories/        DB access per aggregate
  services/            business logic
  handlers/            HTTP layer
  middleware/          auth, RBAC, CORS, rate limit, request size, audit, logging
  routes/              /api/v1 router assembly
  auth/                JWT + password hashing + token store
  auditmanager/        audit log service
  notifications/       notification service (extensible for email/SMS)
  seed/                super admin, roles, permissions seeding
  responses/           unified JSON response helpers
  utils/               money (decimal), pagination, errors
database/migrations/    versioned *.sql files (applied by migrate runner)
docs/                  CODEBASE-MAP.md, API-CONTRACT.md, swagger.json
```

## 3. Technology Stack

| Concern              | Choice                                                     |
|----------------------|------------------------------------------------------------|
| Language / framework | Go 1.27, Gin                                             |
| Database             | PostgreSQL 16                                               |
| ORM                  | GORM (v2) + `gorm.io/driver/postgres`                       |
| Money precision      | `gorm.io/datatypes.Decimal` (Postgres `numeric`) + `math/big` — no float money |
| Auth                 | JWT (access) + opaque hashed refresh token (rotation, DB)   |
| Password hashing     | bcrypt (`golang.org/x/crypto`)                              |
| Validation           | `github.com/go-playground/validator/v10`                    |
| Config               | `.env` via `joho/godotenv`; never hardcoded secrets         |
| Logging              | `log/slog` (structured) + Gin request logging middleware    |
| Rate limiting        | `golang.org/x/time/rate` in-memory (login + general)        |
| Swagger              | swaggo (`swag`, `gin-swagger`)                              |
| Tests                | stdlib `testing` + `testify` where helpful                  |
| Containers           | `backend/Dockerfile` (multi-stage) + `backend/docker-compose.yml` |

## 4. Database Structure (planned tables)

All tables: `id` UUID, `created_at`, `updated_at`, `deleted_at` (soft delete where noted).

- `users` — username, email, password_hash, names, `role_id`, status, last_login_at,
  profile
- `roles` — name (SUPER_ADMIN, HR_ADMIN, MANAGER, EMPLOYEE, RECRUITER, ACCOUNTANT),
  description
- `permissions` — name (`employee:create`, `leave:approve`, ...), description
- `role_permissions` — join table (m2m role ↔ permission)
- `refresh_tokens` — hashed token, user_id, expiry, revoked_at, replaced_by, ip, user_agent
- `audit_logs` — user_id, action, resource, resource_id, ip, user_agent, metadata(JSONB)
- `departments` — name, code(unique), description, manager_id, status
- `designations` — name, description, department_id, level, status
- `employees` — employee_code(unique), names, email(unique), phone, dob, gender, address,
  emergency_contact, joining_date, employment_type, department_id, designation_id,
  manager_id, status, user_id
- `salary_structures` — employee_id, basic_salary, allowances, bonus, tax, deductions,
  effective_from, status
- `attendance` — employee_id, date, check_in, check_out, working_hours, status,
  late_minutes, overtime, remarks
- `leave_types` — name, code, description, is_paid
- `leave_balances` — employee_id, leave_type_id, year, entitlement, used
- `leaves` — employee_id, leave_type_id, start_date, end_date, days, reason, status,
  reviewer_id, reviewed_at, review_note
- `holidays` — name, date, description, type, status
- `payroll` — month, year, employee_id, salary_structure_id, basic_salary, allowances,
  bonus, overtime, gross_salary, tax, deductions, net_salary, status(DRAFT/PROCESSING/
  PROCESSED/PAID/CANCELLED), processed_by, processed_at, paid_on, notes
- `job_posts` — title, department_id, description, requirements, vacancies, status,
  posted_by, deadline
- `candidates` — names, email(unique), phone, resume_path, source, status, hired_as_employee
- `applications` — job_post_id, candidate_id, cover_letter, status, applied_date
- `interviews` — application_id, interviewer_id, scheduled_at, type, status, feedback
- `onboardings` — employee_id, candidate_id, start_date, status, tasks(JSONB), notes
- `goals` — employee_id, title, description, target_date, weight, status
- `kpis` — employee_id, name, description, target, actual, unit, weight, period, score
- `performance_reviews` — employee_id, reviewer_id, period, self_evaluation,
  manager_feedback, score, status
- `training_programs` — title, description, provider, start_date, end_date, status
- `training_schedules` — program_id, date/start/end, trainer, location
- `training_enrollments` — program_id, employee_id, status, completed_at
- `documents` — employee_id, title, type, file_path, mime_type, size, status, uploaded_by
- `notifications` — user_id, title, message, type, is_read, metadata(JSONB)

Indexes on all FK columns, `employees.employee_code`, `employees.email`, `leaves(employee,
status)`, `attendance(employee,date)` unique, `payroll(employee,month,year)` unique.

## 5. Authentication Flow (planned)

1. `POST /api/v1/auth/login` — email/username + password → bcrypt verified → short-lived JWT
   access token + opaque refresh token (SHA-256 stored in `refresh_tokens`).
2. Middleware validates JWT on every protected route, loads the user + role + permissions
   (cached), attaches principal to context.
3. `POST /api/v1/auth/refresh` — validates stored refresh, rotates (old revoked, new issued).
4. `POST /api/v1/auth/logout` — revokes the presented refresh token.
5. `GET /api/v1/auth/me` — current user profile with role and granted permissions.
6. Account `status` checked at login and per request (inactive/suspended → 403).
7. Login rate-limited; account lockout disabled for now (tracked in reports).

## 6. Authorization Flow (planned)

- Roles: `SUPER_ADMIN`, `HR_ADMIN`, `MANAGER`, `RECRUITER`, `ACCOUNTANT`, `EMPLOYEE`.
- Permissions are granular strings, e.g. `employee:create`, `employee:update`,
  `leave:approve`, `payroll:process`, `audit:read`.
- Permission set is resolved on the **backend** from role_permissions in the database.
  The frontend never asserts authorization.
- `RBAC()` middleware rejects with 403 when the principal lacks the required permission.
- Ownership rule: `EMPLOYEE` can read/update only their own records for personal modules
  (attendance, leave, profile); cross-access requires a manager/HR permission.

## 7. API Structure (planned)

Base path `/api/v1`. All responses use the envelope:

```json
{ "success": true,  "message": "...", "data": {} }
{ "success": false, "message": "...", "errors": {} }
```

Modules and route groups are listed exhaustively in `docs/API-CONTRACT.md`.

## 8. Important Dependencies

None yet — the module is empty. Planned direct dependencies are listed in §3 and will be
pinned by `go mod tidy`. No unnecessary dependencies will be added (e.g. no big ORM-level
caches, no unused web frameworks).

## 9. Existing Business Logic

None — new build. Key business rules that must live in services (not handlers/repos):

- Leave overlap + balance validation, and auto leave-balance decrement on approval.
- Attendance single check-in per day, check-out after check-in, working-hours calc.
- Payroll: Gross = Basic + Allowances + Bonus + Overtime; Net = Gross − Tax − Deductions;
  exact decimal arithmetic via `math/big` / `numeric`.
- Prevent duplicate candidate vs employee (email match). Prevent duplicate job applications
  per candidate per job.
- Onboarding ties a hired candidate to a new employee record without duplicates.
- Permission checks for every protected endpoint.

## 10. Missing Features

Everything — the entire HRMS backend will be built in phases (see §12).

## 11. Potential Conflicts / Risks

- **Money precision:** floats corrupt payroll. Mitigation: `numeric` columns + `math/big`.
- **Model explosion:** 30+ tables. Mitigation: consolidated files by module; generic list
  helper; DTO-driven validation shared via base paging/date filters.
- **Swagger annotation volume:** full annotations for every handler is a large cost;
  mitigation: annotate public contract surface (auth, employees, departments,
  attendance, leaves, payroll) and keep the complete machine-readable contract in
  `docs/API-CONTRACT.md`.
- **Test DB dependency:** repo integration tests need PostgreSQL; mitigation: unit tests for
  all pure business logic (payroll math, leave validation, tokens, RBAC), integration tests
  skipped when `TEST_DB_URL` is unset.
- **Seed data in production path:** avoided; seeding is driven by env + idempotent DB upserts.

## 12. Recommended Integration Points & Phases

1. **Phase 1** — project map, database, config, auth, users, roles, permissions, RBAC, seed.
2. **Phase 2** — employees, departments, designations.
3. **Phase 3** — attendance, leave, holidays.
4. **Phase 4** — salary, payroll, payslips.
5. **Phase 5** — recruitment, candidates, interviews, onboarding.
6. **Phase 6** — performance, goals, KPIs, training.
7. **Phase 7** — documents, notifications, reports, audit logs, dashboard APIs.
8. **Hardening** — Swagger, API-CONTRACT.md, tests, Docker, migrations, gofmt/vet/test.

Each phase ends with: `gofmt` → `go vet` → `go test ./...` → route check → migration check →
doc updates → report.

## 13. Implementation Status (post-implementation)

All phases are shipped. Deviations from the plan above:

- **No `internal/migrate` runner.** Schema is applied via `database.Migrate` (GORM
  `AutoMigrate` over all models) on startup. A supplementary, idempotent
  `database/migrations/000001_init.sql` is provided for teams that want versioned SQL.
- **Money type is a custom `models.Decimal`** (string-backed, `math/big`, mapped to PG
  `numeric`) — `gorm.io/datatypes.Decimal` does not exist; `datatypes` is used only for
  `Date` and `JSON`.
- **Swagger/swaggo were not added**; the complete machine-readable endpoint contract lives
  in `docs/API-CONTRACT.md`.
- **Notifications** read/write through `notifications.Service` (DB); email/SMS plug in via
  a `Provider` without changing callers. Candidates without accounts use `NotifyByEmail`.

### Implemented modules (route → handler → service → repository)

| Module | Routes (base `/api/v1`) |
|--------|-------------------------|
| Auth | `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/me` |
| Users | `/users` (CRUD) + `/auth/me` returns granted permissions |
| Roles | `/roles` (CRUD + `/roles/permissions` catalog) |
| Org | `/departments`, `/designations` (CRUD) |
| Employees | `/employees` (CRUD + `/employees/me`) |
| Attendance | `/attendance` check-in/out, list, get, update |
| Holidays | `/holidays` (CRUD) |
| Leaves | `/leaves` (apply, list, approve/reject), `/leaves/types`, `/leaves/balances`, `/leaves/balances/set` |
| Payroll | `/salary` (structures), `/payroll` (generate/process/mark-paid/cancel/payslip) |
| Recruitment | `/recruitment/{jobs,candidates,applications,interviews,onboarding}` |
| Performance | `/performance/{goals,kpis,reviews}` |
| Training | `/training/{programs,schedules,enrollments}` |
| Documents | `/documents` (upload, list, get, download, delete) |
| Notifications | `/notifications` (list, unread-count, read, read-all — own scope) |
| Reports | `/reports/{headcount,attendance,leaves,payroll,recruitment,holidays}` |
| Dashboard | `/dashboard/summary` |
| Audit | `/audit/logs` |
| Health | `GET /healthz` |

### Key implemented business rules

- Payroll: `Gross = Basic + Allowances + Bonus + Overtime`,
  `Net = Gross − (Tax + Deductions)`; exact decimal math end-to-end. State machine
  DRAFT → PROCESSING → PROCESSED → PAID (PAID cannot be cancelled).
- Leave: business-day (weekday-only) counting, overlap rejection, balance check on apply,
  automatic `used` decrement on approval inside a DB transaction; manager + employee
  notifications.
- Attendance: one check-in per employee per day, check-out must follow check-in,
  09:00 start with 15-minute grace computing `late_minutes`, working-hours tracked.
- RBAC: granular permission strings resolved from `role_permissions` with a 30s TTL cache;
  403 on missing permission. `EMPLOYEE` requests are self-scoped per module.
- Recruitment: dedupes candidates by email and applications per job; `Hire` creates the
  employee + optional user account + onboarding plan without duplicates.
- Seed: idempotent — permission catalog, six system roles with mapped permissions,
  configurable super admin from env.
- Audit: every mutation records `user_id, action, resource, resource_id, ip, user_agent,
  metadata` via `auditmanager.Service` (which also receives login/role-change events).

### Test strategy

- Unit tests (always run, no DB): JWT round-trip/expiry/wrong-secret, bcrypt have/verify,
  refresh-token hashing, email normalization, `Decimal` arithmetic/JSON/scan, business-day
  counting, payroll gross/net math, late-minute computation, RBAC allow/deny/unauthenticated.
- Integration tests (`internal/services/*_integration_test.go`): full leave
  create→approve→balance-decrement + overlap + insufficient-balance flows, and attendance
  check-in→duplicate→check-out. Skipped automatically unless `TEST_DB_URL` is set.
- Every phase ended with `gofmt -w .` → `go build ./...` → `go vet ./...` →
  `go test ./...`; final pass is clean.

### Run

```bash
cp backend/.env.example backend/.env   # set JWT_SECRET + super admin credentials
docker compose -f backend/docker-compose.yml up -d db # or point DB_HOST at an existing PostgreSQL 16+
go run ./cmd/server     # migrates + seeds on startup (run from backend/)
# integration tests:
TEST_DB_URL='postgres://emplyra:emplyra_password@localhost:5432/emplyra' go test ./internal/services/ -run Integration
```