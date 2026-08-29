# Frontend Multi-Tenant Migration

Converts the existing single-company Emplyra frontend into a SaaS platform that
hosts many organizations, each with the same HRMS experience — while preserving the
current UI and existing functionality (see `frontend/docs/CURRENT-FRONTEND-ARCHITECTURE.md`).

Principled stance, from the brief:

- **ADD multi-tenancy, do NOT rebuild the HRMS.**
- **No mock data.** Every number comes from the backend; otherwise show
  loading/empty/unavailable.
- **No invented APIs.** The frontend only consumes the real contract
  (`docs/API-CONTRACT.md`). Because the backend has **no tenant concept today**, this
  migration has two workstreams: (A) backend tenant model/endpoints, (B) frontend
  consumption of them. Frontend steps that depend on (A) are gated.

## Target context model

```
Platform
 ├── Platform Admin (PLATFORM_OWNER/ADMIN/SUPPORT/AUDITOR)  → /platform/*
 ├── Company A ── HRMS  (tenant route group, unchanged URLs)
 ├── Company B ── HRMS
 └── Company C ── HRMS
```

Frontend always knows: `currentUser · currentRole(s) · currentPermissions ·
currentTenant {id, name, status, plan}`. **Tenant context comes only from the
authenticated backend session** (`GET /auth/me` once it returns tenant info). The
browser never accepts a tenant id from a URL/query/user input to define context.

## Architecture changes

### 1. Session & context layer (foundation)

- Extend `lib/api.ts` types (`lib/api.ts:3-7`): `CurrentUser` gains
  `tenant?: {id, name, status, plan}` and `roles?: string[]`; `MeResponse` gains
  `tenant` and `scope: 'platform' | 'tenant'`.
- New `lib/auth.tsx` — `AuthProvider`/`useAuth()` Context provider: owns session,
  `me()` hydration, login/logout, `401 → refresh → retry → else redirect`,
  `403` error surfacing, and **tenant suspended/inactive** state (banner, block
  mutations).
- Session persistence: preserve current in-memory behavior; add an explicit
  `restoreSession()` from `sessionStorage` as a compatibility improvement, and
  document a future HttpOnly-cookie/BFF swap (does not block tenant work).
- `app/layout.tsx` wraps the tree in `AuthProvider`.

### 2. Two route contexts

- **Tenant HRMS** (existing look & feel): add real route segments whose UI is the
  existing single-page shell, so URLs finally exist without visual change:
  `/` (dashboard) · `/employees` · `/departments` · `/designations` · `/attendance`
  · `/leaves` · `/payroll` · `/recruitment` · `/performance` · `/training` ·
  `/reports` · `/settings`. The current `EmplyraDashboard` tab-switching is kept as
  an inner app-shell (sidebar/header/dashboard/employees untouched); each nav item
  becomes a route that renders that module's panel.
- **Platform admin**: new group `app/(platform)/platform/*` — `/platform/dashboard`,
  `/platform/tenants`, `/platform/tenants/:id`, `/platform/users`,
  `/platform/audit-logs`, `/platform/settings`. Its own shell (same design tokens,
  new nav):
  `Platform Dashboard · Organizations · Platform Users · Plans · Usage ·
  Audit Logs · Settings`.
- Routing guard `lib/guards.tsx`: `requireScope('platform')` and
  `requireScope('tenant')`; both read from `useAuth()`. Unauthorized scope →
  visible redirect (never silently switch).

### 3. Tenant-aware API client (centralized)

- Keep one client. `request()` in `lib/api.ts:12-18` gains automatic injection of
  the tenant header (e.g. `X-Tenant-ID`) **sourced from `useAuth()` context**, and
  maps tenant lifecycle HTTP codes to typed errors; components never attach tenant
  data manually.
- Layering:
  ```
  Component → service module (lib/services/*) → api.request() → backend
  ```
  New `lib/services/` modules (tenant.ts, platform.ts, recruitment.ts, users.ts) wrap
  endpoints so components import `tenantApi.listOrganizations()` etc., never raw
  fetches. This kills the "component manually adds tenant_id" smell.
- `401` (expired) → refresh → redirect to login. `403` → hide action + message.
  **Tenant suspended/inactive** → banner + read-only mode. Session expired →
  dedicated screen offering "Sign in again".

### 4. Platform admin UI (new, on the existing design system)

- `/platform/dashboard`: organizations (total/active/trial/suspended), total users,
  total employees, system usage, recent platform activity — all from real platform
  endpoints; every card reuse the existing stat-card markup/DashboardSummary pattern.
- `/platform/tenants`: list (Company ABC | Status: Active | Plan: Professional |
  Users: 120 | Employees: 110 | View), Create/Edit Organization, Activate, Suspend,
  View Details/Usage, and **Create Initial Tenant Owner** (calls tenant-owner
  creation endpoint).
- `/platform/users`: manage PLATFORM_OWNER / PLATFORM_ADMIN / PLATFORM_SUPPORT /
  PLATFORM_AUDITOR — completely separate list+form from tenant users.
- `/platform/audit-logs`, `/platform/settings`: list + form pages against the
  platform contracts.
- No tenant HR/employee numbers appear on platform pages.

### 5. Tenant admin & current-company identity

- The existing tenant shell gains an explicit **company block** in the header:
  `Company ABC` / `HR Management System` (the vendored "Company ABC" card pattern),
  driven by `currentTenant` from `useAuth()` — never user-editable.
- The existing modules (Employees, Departments, Designations, Attendance, Leave,
  Payroll, Recruitment, Performance, Training, Documents, Reports) keep their UI
  and endpoints; they become tenant-scoped by the client header + backend. Data
  shown is always the current tenant's.

### 6. Platform admin tenant access ("support mode")

- `/platform/tenants/:id` → **[Open Tenant]** sets `supportMode = {tenantId}` in an
  `AuthProvider` action (not a URL param change of identity).
- While active, the tenant shell shows a persistent banner: `Viewing: Company ABC ·
  Mode: Platform Support / Admin Access` and **[Exit Tenant]** restores normal scope.
- Entering requires a confirmation; exiting is explicit; the banner is non-dismissable.

### 7. User & role management (tenant-aware)

- Tenant `/users` and `/roles`: role picker filters to roles valid **within the
  tenant**; `PLATFORM_OWNER / PLATFORM_ADMIN / PLATFORM_SUPPORT / PLATFORM_AUDITOR`
  never appear for tenant administrators.
- Platform users rotate only through `/platform/users`.
- Permission-gated UI: new `usePermission()`/`can('permission')` helper; filters
  nav items, buttons, and pages. UI checks remain additive — backend is
  authoritative.

### 8. Recruitment

- Jobs, candidates, applications, interviews, Google-Forms sync stay wired to
  `/recruitment/*`; tenant scoping comes from the central client, so Company A data
  can never render in Company B's session (backend must also enforce this).
- **Job application form builder** does not exist in the frontend today; it is
  net-new. When built it ships tenant-aware by design: Create/Edit/Copy Template,
  Add/Remove/Reorder Fields, Publish/Unpublish; global templates are offered as
  starting points and a tenant always copies before editing. (Gated on the backend
  template/field endpoints existing.)

### 9. Error handling & lifecycle states

- Route-level error/empty/loading components reuse the existing banner/table
  patterns. A suspended tenant shows a full-screen notice with contact/support
  copy, not a generic error.

### 10. Responsive

- Preserve existing breakpoint behavior; validate Platform Dashboard, Tenant
  Dashboard, Organization Management, User Management, and support-mode banner on
  desktop / tablet / mobile.

### 11. Cleanup of existing fabrications (during migration, small)

- Remove hardcoded Leave badge `3`, hardcoded date `Friday, August 29, 2026`, and
  static "Good morning" greeting, replacing them with real data or neutral copy —
  they are single-tenant assumptions that would mislead in a SaaS context.

## Sequencing & gates

| Phase | Work | Gate |
| --- | --- | --- |
| 0 | Map frontend; this document | done here |
| 1 | `lib/api.ts` types + `AuthProvider`/`useAuth` (tenant from `/auth/me`); guard; tenant client header (opt-in, no-op when tenant absent) | backend `/auth/me` returns `tenant` |
| 2 | Route groups: split shell into tenant context + platform context; port existing dashboard/employees UI unchanged | none (structural, disables nothing) |
| 3 | Platform admin pages against real endpoints | platform endpoints exist |
| 4 | Support-mode (Open/Exit Tenant) | tenant-switch/impersonation endpoint |
| 5 | Tenant-aware user/role management | tenant-roles endpoint |
| 6 | Job-application form builder (tenant-aware) | backend template/field endpoints |
| 7 | Audit + tests (role matrix, A/B isolation, support-mode, navigation restrictions, full integration trace) | features complete |

Backend dependency: `GET /auth/me` must return `{user, roles, permissions, tenant, scope}`;
a tenant-scoped registration/sign-in; organizations CRUD + usage + lifecycle;
platform user management; tenant roles isolation. Per brief §18, phases ahead of
their backend gate render **empty/loading/unavailable states** rather than faked data.

## Deliverables to create when implementation starts

- `frontend/docs/FRONTEND-MULTI-TENANT-AUDIT.md` — per page/module:
  `FULLY MIGRATED / PARTIALLY MIGRATED / NOT MIGRATED / BROKEN / NOT VERIFIED`.
- Update `frontend/docs/FRONTEND-BACKEND-MAPPING.md` — per page: `Frontend Page →
  API Endpoint → Tenant Scope → Permission`.

## Success criteria (from the brief §22-23)

- Platform Owner/Admin logins → platform UI only; Tenant Owner/HR Admin/Manager/
  Employee logins → their tenant HRMS only.
- Employee sees employee-scoped functionality and nothing platform-related.
- Tenant A sees Tenant A data; Tenant B sees Tenant B data; a browser in Tenant B
  can never load Tenant A records.
- Navigation restrictions enforced by scope guard + permission helper.
- Full trace passes: `Login → Backend Auth → User → Tenant → Role → Permissions →
  Frontend Context → Dashboard → API → Tenant Data`.
- Existing single-company workflows still pass (dashboard, employees, auth flow).
- `gofmt`-style hygiene for frontend: `npm run build` + `tsc --noEmit` + a test
  runner added in Phase 1.