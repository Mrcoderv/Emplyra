# Google Forms — Recruitment Integration Map

This document describes how the existing HRMS Recruitment module is extended with
Google Forms without touching or replacing any existing recruitment functionality.

## Existing recruitment mapping

Reusing the live system, not a copy:

| Concern | Location |
| --- | --- |
| Job posting model | `models.JobPost` (`internal/models/recruitment.go`) |
| Candidate model | `models.Candidate` (extended with `address/date_of_birth/education/experience/skills`) |
| Application model | `models.Application` (used as-is, `cover_letter` reused) |
| Interview system | `models.Interview` + `interview:*` endpoints |
| Candidate statuses | `NEW → SCREENING → SHORTLISTED → INTERVIEWING → OFFERED → HIRED`/`REJECTED` |
| Application statuses | `APPLIED → SHORTLISTED → INTERVIEWING → OFFERED → HIRED`/`REJECTED`/`WITHDRAWN` |
| APIs | `/api/v1/recruitment/jobs|candidates|applications|interviews|onboarding` |
| Service layer | `RecruitmentService.CreateCandidate/CreateApplication/Hire` |

External candidates enter the **existing** workflow: `APPLIED → SCREENING → SHORTLISTED →
INTERVIEW → SELECTED → HIRED`. No second workflow was introduced.

## End-to-end flow

```
HR/Admin ── create job ── create/connect Google Form (permission googleform:connect)
    │
    ▼
POST /recruitment/jobs/:id/google-form/connect  (form_url, sheet_id, sheet_name, field_mapping)
    │
    ▼
Applicant submits the public Google Form (no HRMS account)
    │
    ▼
Google Form Response -> Google Sheets
    │
    ▼   manual POST /sync  (incremental pointer, or full re-scan)
Sheets API (official, OAuth) -> Go backend sync engine
    │
    ▼
GoogleFormService.Sync → CreateCandidate (exists-by-email) → CreateApplication (APPLIED)
    │
    ▼
Candidate + Application land in the existing Recruitment dashboard/workflow
```

## Google Form configuration

Each job keeps exactly one integration (`OneToOne`), stored in
`google_form_integrations` (`models.GoogleFormIntegration`):

| Column | JSON | Purpose |
| --- | --- | --- |
| `job_id` | `job_id` | FK → job_posts (unique) |
| `provider` | `provider` | `google_forms` |
| `form_url` | `google_form_url` | admin-supplied URL (never hardcoded) |
| `spreadsheet_id` | `google_sheet_id` | responses spreadsheet |
| `response_sheet_name` | `response_sheet_name` | sheet tab (empty → first sheet) |
| `header_row` | `header_row` | 1-based header row |
| `field_mapping` | `field_mapping` | header→HRMS field rules (JSON) |
| `status` | `status` | `PENDING → CONNECTED → ERROR / DISCONNECTED` |
| `last_synced_at` | `last_synced_at` | last successful run |
| `synced_rows` | `synced_rows` | incremental pointer |
| `sync_error`, `status_detail` | same | surfaced to HR/Admin |

## Applicant fields & mapping

Google Form questions are **not** hardcoded. Sync resolves them against the live
sheeet **header row**, matching by column name — column order may change freely.

Default rules (used when the admin doesn't configure a mapping):

| Google Form/Sheets header | HRMS field |
| --- | --- |
| Full Name / Name / First Name | `first_name` (full name auto-split to `last_name`) |
| Last Name | `last_name` |
| Email / Email Address | `email` (required for import) |
| Phone / Phone Number / Contact | `phone` |
| Address | `address` |
| Date of Birth / DOB / Birthdate | `date_of_birth` |
| Education / Qualification | `education` |
| Experience / Years of Experience | `experience` |
| Skills / Technical Skills | `skills` |
| Resume/CV / Resume URL / Resume Link | `resume_path` |
| Cover Letter | `application.cover_letter` |
| Response ID / ID | external dedup id (`gsr:<id>`) |
| Timestamp | `submitted_at` |

Custom mappings are store via `PUT .../google-form` with
`field_mapping: [{ "source": "Header Name", "target": "email" }, ...]`.
Allowed targets: `first_name, last_name, email, phone, address, date_of_birth,
education, experience, skills, resume_path, source, notes, status, cover_letter,
response_id, submitted_at`. Missing configured columns fail the sync loudly
(errors surface to HR; nothing is imported).

## Synchronization

`POST /recruitment/jobs/:id/google-form/sync` with body `{ "mode": "incremental"|"full" }`:

- **Initial sync** — first run imports every response row.
- **Incremental** (default) — continues from `synced_rows`; already-seen rows are
  never re-read.
- **Manual** — the same endpoint, invoked by an HR/admin.
- **Automatic** — same engine, callable by an internal scheduler/cron (idempotent;
  document an example `curl` in the ops section).

Deduplication is guaranteed two ways:

1. `google_form_responses.external_response_id` (unique index) — the response
   ledger; `gsr:<response id>` when the sheet has a Response ID column, else
   `gsrow:<integration_id>:<row>`.
2. Existing application uniqueness — a candidate already applied to the same job
   is skipped (`ErrDuplicateApplication`) and recorded as `DUPLICATE`.

Each response row is recorded in `google_form_responses` as
`IMPORTED / DUPLICATE / ERROR` with `error_message`, `candidate_id`, `application_id`,
`raw_response`, `submitted_at`, `imported_at`.

## Google API integration (official, no scraping)

- `internal/google` implements the **official Google Sheets + OAuth HTTP APIs**
  (no form scraping, no hardcoded credentials).
- Credentials come from env: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`,
  `GOOGLE_PROJECT_ID`, `GOOGLE_REDIRECT_URL`, `GOOGLE_REFRESH_TOKEN`,
  `GOOGLE_TOKEN_ENCRYPTION_KEY`, `GOOGLE_OAUTH_SUCCESS_REDIRECT`.
- Two authorization modes:
  - **Env refresh token**: `GOOGLE_REFRESH_TOKEN` — zero-interaction for CI/cron.
  - **Interactive OAuth**: `POST /api/v1/integrations/google/oauth/authorize`
    returns a state-bound consent URL; the browser returns to the public callback
    `GET /api/v1/integrations/google/oauth/callback`, and the refresh token is
    stored server-side **encrypted at rest** (AES-256-GCM, key derived from
    `GOOGLE_TOKEN_ENCRYPTION_KEY` or `JWT_SECRET`).
- Access tokens are refreshed automatically when near expiry and cached in memory;
  refresh tokens are never exposed to the frontend.

Reads use `GET /v4/spreadsheets/{id}/values/{sheet}` and spreadsheet metadata;
read-only scope `https://www.googleapis.com/auth/spreadsheets.readonly`.

## API design

New endpoints (all under existing `/api/v1` conventions, JSON envelope):

| Method | Path | Permission |
| --- | --- | --- |
| POST | `/recruitment/jobs/:id/google-form/connect` | `googleform:connect` |
| GET | `/recruitment/jobs/:id/google-form` | `googleform:read` |
| PUT | `/recruitment/jobs/:id/google-form` | `googleform:connect` |
| DELETE | `/recruitment/jobs/:id/google-form` | `googleform:connect` |
| POST | `/recruitment/jobs/:id/google-form/sync` | `googleform:sync` |
| GET | `/recruitment/jobs/:id/google-form/sync-status` | `googleform:read` |
| GET | `/recruitment/jobs/:id/google-form/responses` | `googleform:read` |
| POST | `/integrations/google/oauth/authorize` | `googleform:connect` |
| GET | `/integrations/google/oauth/callback` | public (state-validated) |

Sync response:

```json
{ "success": true, "message": "google form synchronized",
  "data": { "imported": 12, "duplicates": 1, "failed": 0, "total_rows": 13 } }
```

## Database

```sql
google_form_integrations  -- one row per job (job_id UNIQUE, FK job_posts CASCADE)
google_form_responses     -- import ledger (external_response_id UNIQUE, FK integration CASCADE)
google_oauth_tokens       -- encrypted token store (keyed: "account", "state:<state>")
```

Candidate table gained five nullable profile columns
(`address, date_of_birth, education, experience, skills`). No existing column or
relationship was removed or redefined. Migration in `database/migrations/000001_init.sql`
(additive, idempotent), also applied by GORM `AutoMigrate` on startup.

## Error handling

| Scenario | Behavior |
| --- | --- |
| Invalid form URL | 422 at connect/update |
| Sheet/spreadsheet not found (404) | `ErrGoogleInvalidSpreadsheet`, integration `ERROR` + `sync_error` |
| Permission denied (403) | `ErrGooglePermissionDenied` → 403 |
| Expired/invalid token | auto-refresh; else `ErrGoogleNotAuthorized` → 401 |
| Rate limit (429) | `ErrGoogleRateLimit` → 429 |
| Network failure | `ErrGoogleNetwork` → 502 |
| Missing configured column | strict 422, nothing imported |
| Missing email on a row | row recorded `ERROR`, failed count incremented |
| Duplicate response | row recorded `DUPLICATE`, never re-imported |

Nothing fails silently: the integration row stores the last error, `sync-status`
exposes counters, and the actor who triggered `sync` gets a notification when a
run reports failures/duplicates.

## Security

- Only `googleform:*` holders (HR admin, recruiters, super admin) can connect,
  configure, sync, or view integrations.
- Applicants use the public Google Form; no HRMS credentials requested.
- Google credentials/tokens live in env or in the encrypted DB store — never in
  source and never returned by any endpoint.

## Config (backend/.env.example)

```
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_PROJECT_ID=
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/integrations/google/oauth/callback
GOOGLE_REFRESH_TOKEN=
GOOGLE_TOKEN_ENCRYPTION_KEY=
GOOGLE_OAUTH_SUCCESS_REDIRECT=/
```

## Tests

- Unit (`internal/services/google_forms_test.go`): header mapping, strict-vs-default
  rules, row value extraction, name split, external response id, email validation,
  timestamp parsing, blank-row detection.
- Integration (gated by `TEST_DB_URL`, `internal/services/google_forms_integration_test.go`)
  with a fake `SheetsReader`: full import flow, duplicate prevention, incremental
  append, sheet-error propagation, no-op incremental re-sync.

## Ops: automatic sync example

```bash
# cron every 5 minutes; manual trigger is idempotent
*/5 * * * * curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  https://hrms.example.com/api/v1/recruitment/jobs/JOB_ID/google-form/sync \
  -d '{"mode":"incremental"}'
```