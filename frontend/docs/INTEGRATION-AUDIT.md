# Emplyra Integration Audit

## Status
| Module | Status | Frontend route | Backend contract |
|---|---|---|---|
| Authentication | PARTIALLY IMPLEMENTED | `/` sign-in modal | `/auth/login`, `/auth/refresh`, `/auth/me`, `/auth/logout` |
| Dashboard | PARTIALLY IMPLEMENTED | `/` | `/dashboard/summary` |
| Employees | PARTIALLY IMPLEMENTED | `/` directory panel | `GET /employees` |
| Departments | NOT IMPLEMENTED | navigation placeholder | `/departments` |
| Designations | NOT IMPLEMENTED | not yet in shell | `/designations` |
| Attendance | NOT IMPLEMENTED | navigation placeholder | `/attendance` and check-in/out |
| Leave | NOT IMPLEMENTED | navigation placeholder | `/leaves` |
| Holidays | NOT IMPLEMENTED | not yet in shell | `/holidays` |
| Payroll | NOT IMPLEMENTED | navigation placeholder | `/salary`, `/payroll` |
| Recruitment | NOT IMPLEMENTED | navigation placeholder | `/recruitment/*` |
| Performance | NOT IMPLEMENTED | navigation placeholder | `/performance/*` |
| Training | NOT IMPLEMENTED | navigation placeholder | `/training/*` |
| Documents | NOT IMPLEMENTED | not yet in shell | `/documents` |
| Notifications | UI affordance only | header bell | `/notifications` |
| Reports | NOT IMPLEMENTED | navigation placeholder | `/reports/*` |
| Users / Roles / Permissions | NOT IMPLEMENTED | not yet in shell | `/users`, `/roles`, `/roles/permissions` |
| Audit Logs | NOT IMPLEMENTED | not yet in shell | `/audit/logs` |
| Settings | UI placeholder | sidebar | no dedicated backend settings route in contract |

## Request path
Implemented dashboard and employee reads follow `Frontend → lib/api.ts → Go REST API → JWT middleware/RBAC → service/repository → envelope → React state → UI`.

## Known issues / next steps
- The current implementation intentionally does not fabricate data when the backend is unavailable.
- The in-memory browser token session is suitable for the current client surface but should be replaced with an HttpOnly-cookie/BFF session before production if the backend deployment permits it.
- Add dedicated route segments and DTO-specific forms after verifying each handler DTO, rather than guessing request fields.
- Restrict backend CORS from `*` to the deployed frontend origin.
- Add automated contract tests for login, refresh rotation, dashboard summary, and employee pagination.
