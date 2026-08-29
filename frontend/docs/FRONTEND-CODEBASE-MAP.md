# Emplyra Frontend Codebase Map

## Current state
- Framework: Next.js 16 App Router with React 19 and TypeScript.
- Routing: only `app/page.tsx`; no nested routes or route groups.
- Layout/metadata: `app/layout.tsx` contains the root document, Analytics, and default v0 metadata.
- Styling: Tailwind CSS v4 through `app/globals.css`, shadcn theme tokens, dark-mode media support.
- Components: only `components/ui/button.tsx` exists; no shared shell, tables, forms, modals, charts, or notification components.
- API client: none. No fetch/axios integration or API environment variable is currently used.
- Authentication: none. No login, token storage, refresh, protected routes, or current-user state.
- State management: none beyond what the default page would provide.
- Forms/tables/modals/charts: none implemented.
- Existing functionality: placeholder page only; no working product feature is being replaced.
- Mock/hardcoded product data: none in the existing frontend.

## Backend contract source
The Go backend is in `Mrcoderv/Emplyra`, with the route registry at `backend/internal/routes/routes.go` and the authoritative contract at `docs/API-CONTRACT.md`. It exposes `/api/v1`, bearer JWT auth, rotating refresh tokens, an envelope `{success,message,data}` / `{success:false,message,errors}`, server-side pagination, and permission-scoped RBAC.

## Frontend implementation direction
- Add a centralized browser API client using `NEXT_PUBLIC_API_BASE_URL` and bearer access tokens.
- Keep all production data API-backed; show explicit loading, empty, unavailable, and error states when the backend is not reachable or a permission is missing.
- Build a reusable admin shell with responsive navigation, permission-aware module visibility, notification affordance, and profile menu.
- Implement the dashboard against `GET /dashboard/summary`, with an Employees view against `GET /employees` and the auth flow against `/auth/login`, `/auth/refresh`, `/auth/me`, `/auth/logout`.
- Represent the remaining backend-supported modules in the navigation and document their endpoint mappings; do not invent unsupported data.

## Known gaps
- The current Next.js scaffold has no route-level pages beyond `/`.
- The backend has no explicit dashboard endpoint for every chart requested in the brief; the frontend must render only fields returned by `/dashboard/summary` and identify missing analytics rather than fabricate them.
- The backend CORS configuration currently allows `*`; production deployment should restrict allowed origins separately.
- Backend documents and payroll values must remain authorization-controlled and should not be exposed through client-side secrets or fabricated client calculations.
