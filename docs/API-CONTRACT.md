# HRMS API Contract

Base URL: `/api/v1` · Content-Type: `application/json` (multipart for document upload).

## Response envelope

```json
{ "success": true,  "message": "...", "data": { } }
{ "success": false, "message": "...", "errors": { } }
```

List-shaped `data`: `{ "items": [...], "total": 3, "page": 1, "page_size": 20, "total_pages": 1 }`.
Pagination via query params `page` (1-based) and `page_size` (default 20, max 100).

## Errors

`400` validation failed · `401` missing/invalid token · `403` missing permission or
cross-user access · `404` not found · `409` duplicate/conflict · `429` rate limited ·
`500` internal.

## Authentication

All routes except `/auth/login` and `/healthz` require `Authorization: Bearer <jwt>`.
Login returns `access_token`, `refresh_token`, `expires_in`, `token_type`. Refresh rotates
the token (old one revoked). RBAC permissions are checked per route.

| Method | Path | Permission | Notes |
|--------|------|-----------|-------|
| POST | `/auth/login` | — | email or username + password (rate-limited) |
| POST | `/auth/refresh` | — | `{refresh_token}` |
| POST | `/auth/logout` | — | body `{refresh_token}` |
| GET | `/auth/me` | — | own profile + granted permissions |

## Users

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/users` | user:read |
| POST | `/users` | user:create |
| GET | `/users/:id` | user:read |
| PUT | `/users/:id` | user:update |
| DELETE | `/users/:id` | user:delete |

## Roles

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/roles` | role:read |
| GET | `/roles/permissions` | permission:read |
| POST | `/roles` | role:create |
| GET | `/roles/:id` | role:read |
| PUT | `/roles/:id` | role:update |
| DELETE | `/roles/:id` | role:delete |

`POST/PUT /roles` accept `{name, description, permission_ids[]}`.

## Organization

| Method | Path | Permission |
|--------|------|-----------|
| GET/POST | `/departments` | department:read / department:create |
| GET/PUT/DELETE | `/departments/:id` | department:read / update / delete |
| GET/POST | `/designations` | designation:read / designation:create |
| GET/PUT/DELETE | `/designations/:id` | designation:read / update / delete |

## Employees

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/employees` | employee:read |
| POST | `/employees` | employee:create |
| GET | `/employees/me` | own (any authenticated user) |
| GET | `/employees/:id` | employee:read |
| PUT | `/employees/:id` | employee:update |
| DELETE | `/employees/:id` | employee:delete |

## Attendance

| Method | Path | Permission |
|--------|------|-----------|
| POST | `/attendance/check-in` | attendance:create |
| POST | `/attendance/check-out` | attendance:create |
| GET | `/attendance` | attendance:read (EMPLOYEE sees own) |
| GET | `/attendance/:id` | attendance:read |
| PUT | `/attendance/:id` | attendance:update |

## Holidays

| Method | Path | Permission |
|--------|------|-----------|
| GET/POST | `/holidays` | holiday:read / holiday:create |
| GET/PUT/DELETE | `/holidays/:id` | holiday:read / update / delete |

## Leaves

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/leaves/types` | leave:read |
| GET | `/leaves/balances` | leave:read |
| POST | `/leaves/balances/set` | leave:create (admin/HR set entitlement) |
| GET | `/leaves` | leave:read (EMPLOYEE sees own) |
| POST | `/leaves` | leave:create |
| GET | `/leaves/:id` | leave:read |
| PUT | `/leaves/:id/approve` | leave:approve |
| PUT | `/leaves/:id/reject` | leave:reject |

Create body: `{employee_id?, leave_type_id, start_date, end_date, reason}`.
Decision body: `{note?}`. Approval decrements the employee's yearly balance once.

## Payroll

| Method | Path | Permission |
|--------|------|-----------|
| GET/POST | `/salary` | salary:read / salary:create |
| GET/PUT/DELETE | `/salary/:id` | salary:read / update / delete |
| GET | `/payroll` | payroll:read |
| POST | `/payroll/generate` | payroll:create (body `{month, year}`) |
| POST | `/payroll/process` | payroll:process |
| POST | `/payroll/mark-paid` | payroll:pay |
| POST | `/payroll/cancel` | payroll:cancel |
| GET | `/payroll/:id` | payroll:read |
| GET | `/payroll/:id/payslip` | payroll:payslip |

Status flow: DRAFT → PROCESSING → PROCESSED → PAID. PAID cannot be cancelled.
Money fields are strings with 2 decimals (`"6750.50"`) — never floats.

## Recruitment

| Method | Path | Permission |
|--------|------|-----------|
| GET/POST | `/recruitment/jobs` | job:read / job:create |
| GET/PUT/DELETE | `/recruitment/jobs/:id` | job:read / update / delete |
| GET/POST | `/recruitment/candidates` | candidate:read / candidate:create |
| GET/PUT/DELETE | `/recruitment/candidates/:id` | candidate:read / update / delete |
| GET/POST | `/recruitment/applications` | application:read / application:create |
| GET | `/recruitment/applications/:id` | application:read |
| PUT | `/recruitment/applications/:id/status` | application:update |
| GET/POST | `/recruitment/interviews` | interview:read / interview:create |
| GET | `/recruitment/interviews/:id` | interview:read |
| PUT | `/recruitment/interviews/:id` | interview:update (complete + feedback) |
| GET/POST | `/recruitment/onboarding` | onboarding:read / onboarding:create |
| GET | `/recruitment/onboarding/:id` | onboarding:read |
| PUT | `/recruitment/onboarding/:id` | onboarding:update |
| POST | `/recruitment/onboarding/hire` | onboarding:create (hires candidate → employee) |

### Recruitment ↔ Google Forms

| Method | Path | Permission |
|--------|------|-----------|
| POST | `/recruitment/jobs/:id/google-form/connect` | googleform:connect (body: form_url, sheet_id, sheet_name, header_row, field_mapping) |
| GET | `/recruitment/jobs/:id/google-form` | googleform:read |
| PUT | `/recruitment/jobs/:id/google-form` | googleform:connect (update form/sheet/mapping) |
| DELETE | `/recruitment/jobs/:id/google-form` | googleform:connect |
| POST | `/recruitment/jobs/:id/google-form/sync` | googleform:sync (body: `{"mode":"incremental"\|"full"}`) |
| GET | `/recruitment/jobs/:id/google-form/sync-status` | googleform:read |
| GET | `/recruitment/jobs/:id/google-form/responses` | googleform:read (import ledger + pagination) |
| POST | `/integrations/google/oauth/authorize` | googleform:connect (returns consent URL) |
| GET | `/integrations/google/oauth/callback` | public (state-validated OAuth redirect, returns tokens to HRMS) |

Sync response data:

```json
{ "imported": 12, "duplicates": 1, "failed": 0, "total_rows": 13 }
```

## Performance

| Method | Path | Permission |
|--------|------|-----------|
| GET/POST | `/performance/goals` | goal:read / goal:create |
| GET/PUT/DELETE | `/performance/goals/:id` | goal:read / update / delete |
| GET/POST | `/performance/kpis` | kpi:read / kpi:create |
| GET/PUT/DELETE | `/performance/kpis/:id` | kpi:read / update / delete |
| GET/POST | `/performance/reviews` | review:read / review:create |
| GET | `/performance/reviews/:id` | review:read |
| PUT | `/performance/reviews/:id/submit` | review:submit (self + manager evaluation) |

## Training

| Method | Path | Permission |
|--------|------|-----------|
| GET/POST | `/training/programs` | training:read / training:create |
| GET/PUT/DELETE | `/training/programs/:id` | training:read / update / delete |
| GET/POST | `/training/schedules` | training:read / training:create |
| PUT/DELETE | `/training/schedules/:id` | training:update / delete |
| GET/POST | `/training/enrollments` | enrollment:read / enrollment:create |
| PUT | `/training/enrollments/:id` | enrollment:update |

## Documents

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/documents` | document:read |
| POST | `/documents` | document:create (multipart: `file`, `employee_id`, `title`, `type`) |
| GET | `/documents/:id` | document:read |
| GET | `/documents/:id/download` | document:download |
| DELETE | `/documents/:id` | document:delete |

## Notifications (own scope)

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/notifications` | notification:read (?unread=true) |
| GET | `/notifications/unread-count` | notification:read |
| PUT | `/notifications/:id/read` | notification:update |
| PUT | `/notifications/read-all` | notification:update |

## Reports, dashboard, audit

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/reports/headcount` | report:read |
| GET | `/reports/attendance` | report:read (?from, ?to) |
| GET | `/reports/leaves` | report:read |
| GET | `/reports/payroll` | report:read (?month, ?year) |
| GET | `/reports/recruitment` | report:read |
| GET | `/reports/holidays` | report:read |
| GET | `/dashboard/summary` | dashboard:read |
| GET | `/audit/logs` | audit:read (?user_id, ?resource, ?action) |

## Health

| Method | Path |
|--------|------|
| GET | `/healthz` → `{"status":"ok"}` |

## System roles → permission map

| Role | Scope |
|------|-------|
| SUPER_ADMIN | every permission |
| HR_ADMIN | user/employee/org/time-off/payroll/recruitment/performance/training/docs/reports/audit + googleform:connect/sync/read |
| MANAGER | read-most; approve leaves; manage direct reports' goals/KPIs/reviews; respond to leave requests |
| RECRUITER | all recruitment actions + training read + reports + googleform:read/sync |
| ACCOUNTANT | salary/payroll + attendance read + reports |
| EMPLOYEE | own-scope: attendance, leave apply, goals (read/update own), self review, enrollment, documents, notifications, dashboard + googleform:read |

## Google Forms integration env (backend/.env)

Refer to `docs/GOOGLE-FORMS-INTEGRATION-MAP.md` for the full integration contract.
Environment variables consumed at startup (see `google.ConfigFromEnv` in `internal/google/oauth.go`):

```
GOOGLE_CLIENT_ID
GOOGLE_CLIENT_SECRET
GOOGLE_PROJECT_ID
GOOGLE_REDIRECT_URL
GOOGLE_REFRESH_TOKEN
GOOGLE_TOKEN_ENCRYPTION_KEY   (falls back to JWT_SECRET)
GOOGLE_OAUTH_SUCCESS_REDIRECT
```