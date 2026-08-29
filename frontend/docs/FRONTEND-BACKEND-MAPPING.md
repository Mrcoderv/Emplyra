# Emplyra Frontend ↔ Backend Mapping

Base URL: `NEXT_PUBLIC_API_BASE_URL` (expected to point to the Go API `/api/v1`; fallback is `http://localhost:8080/api/v1` for local development).

## Implemented in the frontend
| Frontend surface | Method | API endpoint | Permission / scope | Status |
|---|---:|---|---|---|
| Sign in modal | POST | `/auth/login` | Public | Implemented |
| Session refresh | POST | `/auth/refresh` | Public, rotating refresh token | Implemented in client |
| Current user | GET | `/auth/me` | Authenticated | Implemented in client |
| Logout | POST | `/auth/logout` | Authenticated | Implemented in client |
| Dashboard stat cards | GET | `/dashboard/summary` | `dashboard:read` | Implemented |
| Employee directory | GET | `/employees?page&page_size&search` | `employee:read` (employee scope applies) | Implemented |

## Navigation-ready backend modules
These routes exist in the Go router and are represented in the shell navigation. Module-specific pages should be added only against the exact request/response DTOs:
- Users: `/users` (`user:read/create/update/delete`)
- Roles and permissions: `/roles`, `/roles/permissions` (`role:*`, `permission:read`)
- Organization: `/departments`, `/designations`
- Attendance: `/attendance`, `/attendance/check-in`, `/attendance/check-out`
- Holidays: `/holidays`
- Leave: `/leaves`, `/leaves/types`, `/leaves/balances`, decision endpoints
- Payroll: `/salary`, `/payroll`, generation/process/payment/cancel/payslip endpoints
- Recruitment: `/recruitment/jobs`, `/candidates`, `/applications`, `/interviews`, `/onboarding`
- Performance: `/performance/goals`, `/kpis`, `/reviews`
- Training: `/training/programs`, `/schedules`, `/enrollments`
- Documents: `/documents` including multipart upload and authorized download
- Notifications: `/notifications`, `/notifications/unread-count`
- Reports: `/reports/headcount`, `/attendance`, `/leaves`, `/payroll`, `/recruitment`, `/holidays`
- Audit: `/audit/logs`

## Contract rules
- All successful responses use `{success:true,message,data}`; list data contains `items`, `total`, `page`, `page_size`, and `total_pages`.
- All protected requests use `Authorization: Bearer <jwt>`.
- The API client retries a 401 once through `/auth/refresh`, then surfaces the error.
- UI permission checks must remain additive; the Go backend remains authoritative.
- Payroll money values remain strings and are never recalculated in the browser.
- Private documents must be downloaded through `/documents/:id/download`, not exposed as public URLs.

## Verified gaps
- No explicit dashboard chart payload contract is currently guaranteed beyond `/dashboard/summary`; chart UI is intentionally not fabricated.
- Search/filter/sort query parameters beyond the documented pagination parameters need DTO/handler verification before use.
- The backend CORS wildcard should be restricted for production deployment.
