import type {
  ApiEnvelope, Application, Candidate, CurrentUser, DashboardSummary, Department, Designation, Document,
  Employee, Goal, Holiday, Interview, JobPost, KPI, Leave, LeaveBalance, LeaveType, ListResponse,
  MeResponse, ModuleRecord, Notification, Onboarding, Payroll, PerformanceReview, RequestOptions,
  Role, SalaryStructure, Session, TrainingEnrollment, TrainingProgram, TrainingSchedule, User,
} from '@/lib/types'
export * from '@/lib/types'
const API_BASE_URL = (process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1').replace(/\/$/, '')
const SESSION_KEY = 'emplyra.session'
let accessToken: string | null = null
let refreshToken: string | null = null
let tenantId: string | null = null
let refreshFlight: Promise<Session> | null = null
function persist(session: Session | null) { accessToken = session?.access_token ?? null; refreshToken = session?.refresh_token ?? null; if (typeof window !== 'undefined') { if (session) window.sessionStorage.setItem(SESSION_KEY, JSON.stringify(session)); else window.sessionStorage.removeItem(SESSION_KEY) } }
export function setSession(session: Session | null) { persist(session); if (!session) setTenantId(null) }
export function restoreSession() { if (typeof window === 'undefined') return null; try { const value = JSON.parse(window.sessionStorage.getItem(SESSION_KEY) || 'null') as Session | null; if (value?.access_token) { persist(value); return value } } catch {} window.sessionStorage.removeItem(SESSION_KEY); return null }
export const hasSession = () => Boolean(accessToken)
export const setTenantId = (id: string | null) => { tenantId = id }
export const getTenantId = () => tenantId
async function refresh() { if (!refreshToken) throw new Error('Your session has expired. Please sign in again.'); if (!refreshFlight) refreshFlight = request<Session>('/auth/refresh', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) }, false).finally(() => { refreshFlight = null }); const session = await refreshFlight; persist(session); return session }
async function request<T>(path: string, init: RequestInit = {}, canRefresh = true): Promise<T> { const headers = new Headers(init.headers); if (!(init.body instanceof FormData)) headers.set('Content-Type', 'application/json'); if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`); if (tenantId) headers.set('X-Tenant-ID', tenantId); const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers, cache: 'no-store' }); const payload = await response.json().catch(() => null) as ApiEnvelope<T> | null; if (response.status === 401 && canRefresh && refreshToken) { await refresh(); return request(path, init, false) } if (!response.ok || payload?.success === false) { const errors = payload?.errors; const message = payload?.message || (Array.isArray(errors) ? errors.join(', ') : errors && Object.values(errors).join(', ')); throw new Error(message || `Request failed with status ${response.status}`) } if (!payload || !('data' in payload)) throw new Error(`Request failed with status ${response.status}`); return payload.data }
function withQuery(path: string, options: RequestOptions = {}) { const params = new URLSearchParams({ page: String(options.page ?? 1), page_size: String(options.pageSize ?? 20) }); if (options.search?.trim()) params.set('search', options.search.trim()); Object.entries(options.query ?? {}).forEach(([key, value]) => { if (value !== undefined) params.set(key, String(value)) }); return `${path}?${params}` }
const list = <T,>(path: string, options?: RequestOptions) => request<ListResponse<T>>(withQuery(path, options))
const plain = <T,>(path: string, options?: RequestOptions) => request<T>(options ? withQuery(path, options) : path)
const crud = <T>(path: string) => ({ list: (o?: RequestOptions) => list<T>(path, o), get: (id: string) => request<T>(`${path}/${encodeURIComponent(id)}`), create: (body: unknown) => request<T>(path, { method: 'POST', body: JSON.stringify(body) }), update: (id: string, body: unknown) => request<T>(`${path}/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(body) }), remove: (id: string) => request<unknown>(`${path}/${encodeURIComponent(id)}`, { method: 'DELETE' }) })
const action = <T>(method: string, path: string, body?: unknown) => request<T>(path, { method, ...(body !== undefined ? { body: JSON.stringify(body) } : {}) })

export const api = {
  login: (identifier: string, password: string) => request<{ tokens: Session; user: CurrentUser }>('/auth/login', { method: 'POST', body: JSON.stringify({ identifier, password }) }, false),
  me: () => request<MeResponse>('/auth/me'),
  logout: () => request<unknown>('/auth/logout', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) }, false),
  dashboard: () => request<DashboardSummary>('/dashboard/summary'),
  module: (path: string, options?: RequestOptions) => list<ModuleRecord>(path, options),
  report: (path: string) => request<unknown>(path),

  employee: crud<Employee>('/employees'),
  employees: (page = 1, pageSize = 10, search = '', opts: RequestOptions = {}) => list<Employee>('/employees', { page, pageSize, search, query: opts.query }),
  employeeMe: () => request<Employee>('/employees/me'),

  department: crud<Department>('/departments'),
  departments: (opts?: RequestOptions) => list<Department>('/departments', { pageSize: 100, ...opts }),
  designation: crud<Designation>('/designations'),
  designations: (opts?: RequestOptions) => list<Designation>('/designations', { pageSize: 100, ...opts }),
  holiday: crud<Holiday>('/holidays'),
  holidays: (opts?: RequestOptions) => list<Holiday>('/holidays', { pageSize: 100, ...opts }),

  attendance: (opts?: RequestOptions) => list<ModuleRecord>('/attendance', { ...opts, pageSize: opts?.pageSize ?? 50 }),
  attendanceGet: (id: string) => request<ModuleRecord>(`/attendance/${encodeURIComponent(id)}`),
  attendanceUpdate: (id: string, body: unknown) => request<ModuleRecord>(`/attendance/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(body) }),
  attendanceCheckIn: (body?: Record<string, unknown>) => request<ModuleRecord>('/attendance/check-in', { method: 'POST', body: JSON.stringify(body ?? {}) }),
  attendanceCheckOut: (body?: Record<string, unknown>) => request<ModuleRecord>('/attendance/check-out', { method: 'POST', body: JSON.stringify(body ?? {}) }),

  leaveTypes: () => plain<LeaveType[]>('/leaves/types'),
  leaves: (opts?: RequestOptions) => list<Leave>('/leaves', { ...opts, pageSize: opts?.pageSize ?? 50 }),
  leaveGet: (id: string) => request<Leave>(`/leaves/${encodeURIComponent(id)}`),
  createLeave: (body: Record<string, unknown>) => request<Leave>('/leaves', { method: 'POST', body: JSON.stringify(body) }),
  approveLeave: (id: string, note = '') => action<unknown>('PUT', `/leaves/${encodeURIComponent(id)}/approve`, { note }),
  rejectLeave: (id: string, note = '') => action<unknown>('PUT', `/leaves/${encodeURIComponent(id)}/reject`, { note }),
  leaveBalances: (opts?: RequestOptions) => plain<LeaveBalance[]>('/leaves/balances', { page: 1, pageSize: 100, ...opts }),
  setLeaveBalance: (body: Record<string, unknown>) => action<unknown>('POST', '/leaves/balance', body),

  document: crud<Document>('/documents'),
  documents: (opts?: RequestOptions) => list<Document>('/documents', opts),
  uploadDocument: (form: FormData) => request<Document>('/documents', { method: 'POST', body: form }),
  deleteDocument: (id: string) => request<unknown>(`/documents/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  downloadDocument: (id: string) => `${API_BASE_URL}/documents/${encodeURIComponent(id)}/download`,

  salary: crud<SalaryStructure>('/salary'),
  salaries: (opts?: RequestOptions) => list<SalaryStructure>('/salary', { pageSize: 100, ...opts }),
  payroll: (opts?: RequestOptions) => list<Payroll>('/payroll', { ...opts, pageSize: opts?.pageSize ?? 50 }),
  payrollGet: (id: string) => request<Payroll>(`/payroll/${encodeURIComponent(id)}`),
  payrollGenerate: (body: { month: number; year: number }) => action<{ created: number; month: number; year: number }>('POST', '/payroll/generate', body),
  payrollProcess: (id: string, body?: Record<string, unknown>) => action<unknown>('POST', `/payroll/${encodeURIComponent(id)}/process`, body ?? {}),
  payrollMarkPaid: (id: string, body?: Record<string, unknown>) => action<unknown>('POST', `/payroll/${encodeURIComponent(id)}/mark-paid`, body ?? {}),
  payrollCancel: (id: string, body?: Record<string, unknown>) => action<unknown>('POST', `/payroll/${encodeURIComponent(id)}/cancel`, body ?? {}),

  job: crud<JobPost>('/recruitment/jobs'),
  jobs: (opts?: RequestOptions) => list<JobPost>('/recruitment/jobs', { ...opts, pageSize: opts?.pageSize ?? 50 }),
  candidate: crud<Candidate>('/recruitment/candidates'),
  candidates: (opts?: RequestOptions) => list<Candidate>('/recruitment/candidates', { ...opts, pageSize: opts?.pageSize ?? 50 }),
  application: { list: (opts?: RequestOptions) => list<Application>('/recruitment/applications', { ...opts, pageSize: opts?.pageSize ?? 50 }), get: (id: string) => request<Application>(`/recruitment/applications/${encodeURIComponent(id)}`), create: (body: Record<string, unknown>) => request<Application>('/recruitment/applications', { method: 'POST', body: JSON.stringify(body) }), updateStatus: (id: string, body: { status: string; note?: string }) => action<unknown>('PUT', `/recruitment/applications/${encodeURIComponent(id)}/status`, body) },
  interview: { list: (opts?: RequestOptions) => list<Interview>('/recruitment/interviews', { ...opts, pageSize: opts?.pageSize ?? 50 }), get: (id: string) => request<Interview>(`/recruitment/interviews/${encodeURIComponent(id)}`), create: (body: Record<string, unknown>) => request<Interview>('/recruitment/interviews', { method: 'POST', body: JSON.stringify(body) }), update: (id: string, body: Record<string, unknown>) => request<Interview>(`/recruitment/interviews/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(body) }) },
  onboarding: { list: (opts?: RequestOptions) => list<Onboarding>('/recruitment/onboarding', { ...opts, pageSize: opts?.pageSize ?? 50 }), get: (id: string) => request<Onboarding>(`/recruitment/onboarding/${encodeURIComponent(id)}`), create: (body: Record<string, unknown>) => request<Onboarding>('/recruitment/onboarding', { method: 'POST', body: JSON.stringify(body) }), update: (id: string, body: Record<string, unknown>) => request<Onboarding>(`/recruitment/onboarding/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(body) }) },
  hireCandidate: (body: Record<string, unknown>) => action<Employee>('POST', '/recruitment/onboarding/hire', body),

  goal: crud<Goal>('/performance/goals'),
  goals: (opts?: RequestOptions) => list<Goal>('/performance/goals', { ...opts, pageSize: opts?.pageSize ?? 50 }),
  kpi: crud<KPI>('/performance/kpis'),
  kpis: (opts?: RequestOptions) => list<KPI>('/performance/kpis', { ...opts, pageSize: opts?.pageSize ?? 50 }),
  review: { list: (opts?: RequestOptions) => list<PerformanceReview>('/performance/reviews', { ...opts, pageSize: opts?.pageSize ?? 50 }), get: (id: string) => request<PerformanceReview>(`/performance/reviews/${encodeURIComponent(id)}`), create: (body: Record<string, unknown>) => request<PerformanceReview>('/performance/reviews', { method: 'POST', body: JSON.stringify(body) }), submit: (id: string, body: Record<string, unknown>) => action<unknown>('PUT', `/performance/reviews/${encodeURIComponent(id)}/submit`, body) },

  trainingProgram: crud<TrainingProgram>('/training/programs'),
  trainingPrograms: (opts?: RequestOptions) => list<TrainingProgram>('/training/programs', { ...opts, pageSize: opts?.pageSize ?? 50 }),
  trainingSchedule: { list: (opts?: RequestOptions) => plain<TrainingSchedule[]>('/training/schedules', { page: 1, pageSize: 100, ...opts }), get: (id: string) => request<TrainingSchedule>(`/training/schedules/${encodeURIComponent(id)}`), create: (body: Record<string, unknown>) => request<TrainingSchedule>('/training/schedules', { method: 'POST', body: JSON.stringify(body) }), update: (id: string, body: Record<string, unknown>) => request<TrainingSchedule>(`/training/schedules/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(body) }), remove: (id: string) => request<unknown>(`/training/schedules/${encodeURIComponent(id)}`, { method: 'DELETE' }) },
  trainingEnrollment: { list: (opts?: RequestOptions) => list<TrainingEnrollment>('/training/enrollments', { ...opts, pageSize: opts?.pageSize ?? 50 }), create: (body: Record<string, unknown>) => request<TrainingEnrollment>('/training/enrollments', { method: 'POST', body: JSON.stringify(body) }), update: (id: string, body: Record<string, unknown>) => request<TrainingEnrollment>(`/training/enrollments/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(body) }), remove: (id: string) => request<unknown>(`/training/enrollments/${encodeURIComponent(id)}`, { method: 'DELETE' }) },

  user: crud<User>('/users'),
  users: (opts?: RequestOptions) => list<User>('/users', opts),
  role: crud<Role>('/roles'),
  roles: () => plain<Role[]>('/roles'),
  rolePermissions: () => plain<Role[]>('/roles/permissions'),

  notifications: (opts?: RequestOptions) => list<Notification>('/notifications', opts),
  unreadCount: () => request<{ unread: number }>('/notifications/unread-count'),
  markAllRead: () => request<unknown>('/notifications/read-all', { method: 'PUT' }),
  markRead: (id: string) => request<unknown>(`/notifications/${encodeURIComponent(id)}/read`, { method: 'PUT' }),

  auditLogs: (opts?: RequestOptions) => list<ModuleRecord>('/audit/logs', opts),
}
export const getApiBaseUrl = () => API_BASE_URL