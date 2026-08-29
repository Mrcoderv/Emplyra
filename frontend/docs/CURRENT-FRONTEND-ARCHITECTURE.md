# Current Frontend Architecture

Snapshot taken before multi-tenant work. This documents what the existing frontend
actually is, not what the brief assumed it to be.

## Executive summary

The `frontend/` is a **single-page scaffold** (not a fully-implemented HRMS):

- One route (`/`) rendering one client component (`components/emplyra-dashboard.tsx`).
- Living, API-backed surfaces: **login modal**, **dashboard stat cards**,
  **employee directory** (list + search).
- Everything else (Departments, Attendance, Leave, Payroll, Recruitment,
  Performance, Training, Reports, Settings...) is a **navigation placeholder**
  with no page, no data, and no endpoint call.
- There is **no job application form builder** in the frontend today.
- The frontend is **single-tenant**: no organization concept anywhere.
- The backend it talks to is also **single-tenant** (no `tenant_id` /
  `organization_id` anywhere in `backend/`), so "tenant-aware API" does not exist
  yet — the migration plan must not invent it.

## Stack inventory

| Concern | Choice (from code) |
| --- | --- |
| Framework | Next.js `16.3.3` App Router |
| UI runtime | React `19`, `react-dom` `19`, server components (`RSC: true` in components.json) |
| Language | TypeScript `5.7.3`, `strict: true`, path alias `@/* → ./*` |
| Styling | Tailwind CSS `4.3.3` via `@tailwindcss/postcss`, shadcn "base-nova" tokens in `app/globals.css` (163 lines, `bg-background`/`text-foreground` surfaces) |
| Component basis | `@base-ui/react` `^1.5.0` + shadcn config `components.json`; only `components/ui/button.tsx` exists |
| Icons | `lucide-react` `^1.16.0` |
| Utilities | `clsx`, `tailwind-merge` (`lib/utils.ts` `cn()`) |
| Motion/theme | `tw-animate-css`, media-based light/dark (`colorScheme` viewport) |
| Analytics | `@vercel/analytics` (production only) |
| Package manager | `pnpm` (`pnpm-workspace.yaml`, `pnpm-lock.yaml`) |
| Build config | `next.config.mjs`: `typescript.ignoreBuildErrors: true`, `images.unoptimized: true` |

## Files (complete inventory)

```
app/globals.css              Tailwind v4 + shadcn theme tokens (surfaces/radii/borders)
app/layout.tsx               Root layout: html/body, metadata, icons, Analytics, viewport colorScheme
app/page.tsx                 Renders <EmplyraDashboard /> — the only route
components/emplyra-dashboard.tsx  All-in-one client component (sidebar, header, dashboard, employees,
                                  module placeholders, login modal) — ~3.5 KB of dense JSX
components/ui/button.tsx     shadcn-styled Button (variants/sizes/loading) — built, currently unused
components.json              shadcn config (base-nova, base-ui)
lib/api.ts                   Central API client (JWT + envelope + pagination + refresh-on-401)
lib/utils.ts                 cn() class combiner
docs/                        FRONTEND-CODEBASE-MAP.md, FRONTEND-BACKEND-MAPPING.md, INTEGRATION-AUDIT.md
public/                      icons, placeholders (apple-icon.png, icon.svg, placeholder*.png/svg/jpg)
```

## Routing

- Next.js App Router with **only** `app/page.tsx`.
- No route groups, no dynamic routes, no protected routes, no middleware.
- "Navigation" in the sidebar is **in-component tab state**:
  `const [active, setActive] = useState('Dashboard')`
  (`components/emplyra-dashboard.tsx`). Selecting "Employees" changes the rendered
  panel; it does **not** change the URL. There is no `/dashboard`, `/employees`,
  `/recruitment` etc. as URLs today.

## Authentication

All in `lib/api.ts` + the login modal inside `EmplyraDashboard`:

1. `api.login(identifier, password)` → `POST /auth/login` →
   `{access_token, refresh_token, expires_in, token_type}`.
2. `setSession(session)` stores tokens in **module-level variables**
   (`lib/api.ts:9-11`) — in-memory only, lost on reload, never persisted.
3. `api.me()` → `GET /auth/me` → `{user, permissions[]}`; used to hydrate the
   header avatar/name after sign-in.
4. `request()` attaches `Authorization: Bearer <access>`; on **401** retries once
   through `POST /auth/refresh` (rotating token) then re-runs the original request
   (`lib/api.ts:12-18`).
5. `api.logout()` → `POST /auth/logout` with refresh token.

Unhandled error classes today: no explicit handling for **403**, expired-session
(no redirect), or **tenant suspended/inactive** (backend has no such concept yet).

## Current user state, roles, permissions

- `CurrentUser` (`lib/api.ts:3`): `{id, email, username?, first_name, last_name,
  role?: string | {name}, permissions?: string[]}` — **no tenant field**.
- After login the component holds `user` in local state and `permissions` are
  merged in from `/auth/me`.
- Role is displayed nowhere and drives nothing in the UI today.
- `Module.permission` metadata exists (`modules[]`, `emplyra-dashboard.tsx:5-8`)
  but the nav renders **every** module regardless of permissions — no gating helper.

## API client layering

`Component → lib/api.ts request() → fetch(API_BASE_URL + path) → backend`.

- `API_BASE_URL = NEXT_PUBLIC_API_BASE_URL || http://localhost:8080/api/v1`.
- Envelope handling: `{success, message, data, errors}`; list responses
  `{items,total,page,page_size,total_pages}` via `list<T>()`.
- No tenant header, no context object, no interceptors — tenant support is absent
  by design; every component that calls the API reaches the client directly.

## State management

`useState`/`useEffect`/`useMemo` in a single component. No Context provider, no
Redux/Zustand/React Query/TanStack, no SWR, no server-state library. No persisted
session.

## Layout, sidebar, header, dashboard

All inline in `components/emplyra-dashboard.tsx`:

- **Sidebar**: fixed 256px, `logo "e"`, "Emplyra / People operations"; module list
  from `modules[]`; Leave badge (hardcoded `3` — fabricated); footer card
  "Connected workspace / JWT auth and permission-scoped Go API". Mobile: hidden
  drawer + overlay (`translate-x`), toggle button in header.
- **Header**: breadcrumb `Workspace / {active}`, title = active module; notification
  bell (icon with a static dot — no data); user avatar button toggles Sign-in,
  else logout.
- **Welcome block**: hardcoded date string `Friday, August 29, 2026` and
  `Good morning, {first_name}` (`emplyra-dashboard.tsx` render). Both are
  single-company/static fabrications.
- **Dashboard**: 4 stat cards from `/dashboard/summary`
  (`total_employees, active_employees, employees_on_leave, pending_leave_requests`),
  employee directory table w/ search, "Today's pulse" list, and a promo card.
- **ModuleView** for non-Employees modules: header + `Add record` button (stubbed)
  + "workspace" placeholder.
- **Login modal**: fixed overlay, email-or-username + password, no validation
  rules beyond `required`.
- `EmployeeTable`: loading rows (`Loading…`), empty state (`No employees found.`),
  renders `first_name/last_name`, department, status. Uses `initials()`.

## Error / empty / loading states

- On API failure the component shows an inline banner "Live data is unavailable"
  with the message and a hint to set `NEXT_PUBLIC_API_BASE_URL` — honest, no
  fabricated numbers.
- Stats render `—` when the key is missing; employee table has a true empty state.
- These are the patterns the multi-tenant build will reuse everywhere.

## Responsive behavior

- Tailwind `sm:`/`lg:` breakpoints; sidebar collapses on small screens with a
  drawer; header stacks content; dashboard grid `sm:grid-cols-2 xl:grid-cols-4`.
- No dedicated tablet optimization, no visual-regression or browser testing setup.

## Testing

- None. No unit/integration/E2E test files, no test runner configured, no CI.
- `npm run build` is the only gate, and `typescript.ignoreBuildErrors: true`
  weakens it.

## Module-by-module reality check

| Module | Real implementation | Backend endpoint in contract |
| --- | --- | --- |
| Auth | Login modal, refresh-on-401, me, logout | `/auth/*` |
| Dashboard | Stat cards + pulse (live) | `/dashboard/summary` |
| Employees | Directory table + search (live) | `/employees` |
| Departments | Nav placeholder | `/departments` |
| Designations | Not in nav | `/designations` |
| Attendance | Nav placeholder | `/attendance`, `/attendance/check-in`, `/attendance/check-out` |
| Leave | Nav placeholder (hardcoded badge `3`) | `/leaves`, `/leaves/types`, `/leaves/balances` |
| Holidays | Not in nav | `/holidays` |
| Payroll | Nav placeholder | `/salary`, `/payroll` |
| Recruitment | Nav placeholder | `/recruitment/*` (incl. Google Forms sync endpoints) |
| Performance | Nav placeholder | `/performance/*` |
| Training | Nav placeholder | `/training/*` |
| Reports | Nav placeholder | `/reports/*` |
| Settings | Nav placeholder | none dedicated |
| Notifications | Bell affordance only | `/notifications*` |
| Job application form builder | **Does not exist** | none |

## Constraints for the migration (must not be violated)

1. **No mock data** — the dashboard/summary and employee reads are the reference:
   loading / empty / unavailable until the backend serves data.
2. **No invented APIs** — the frontend must consume the contract in `docs/API-CONTRACT.md`
   (and `frontend/docs/FRONTEND-BACKEND-MAPPING.md`); tenant-aware endpoints must
   exist in the backend before the frontend calls them.
3. **Backend is the authority** — UI permission checks are additive.
4. **Existing single-company UI must keep working** — the shell, colors, typography,
   and layout stay; multi-tenancy is layered on, not a redesign.

## Known gaps vs. the multi-tenant brief

- The brief assumes implemented HRMS modules (employees/attendance/leave/payroll/
  recruitment/reports pages) — those are placeholders today, so "preserve existing
  module pages" collapses to "preserve the shell + dashboard + employees".
- The brief assumes a job application form builder — none exists; building one is
  net-new work, with tenant-awareness designed in from the start.
- The brief assumes tenant-aware APIs — none exist in the backend; this is the
  critical dependency for the whole migration.