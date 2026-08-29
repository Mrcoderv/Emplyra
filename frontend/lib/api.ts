export type ApiEnvelope<T> = { success: boolean; message?: string; data: T; errors?: Record<string, string> | string[] }
export type Session = { access_token: string; refresh_token: string; expires_in: number; token_type: string }
export type LoginResult = { tokens: Session; user: CurrentUser }
export type MeTenant = { id: string; name: string; status: string; plan: string }
export type Scope = 'platform' | 'tenant'
export type CurrentUser = { id: string; email: string; username?: string; first_name: string; last_name: string; role?: string | { name: string }; permissions?: string[] }
export type MeResponse = { user: CurrentUser; permissions: string[]; roles: string[]; scope: Scope; tenant?: MeTenant }
export type DashboardSummary = Record<string, unknown> & { total_employees?: number; active_employees?: number; employees_on_leave?: number; present_today?: number; pending_leave_requests?: number; open_positions?: number }
export type Employee = { id: string; employee_code?: string; first_name: string; last_name: string; email?: string; department?: { name: string } | string; designation?: { name: string } | string; status?: string; employment_status?: string; avatar_url?: string }
export type ListResponse<T> = { items: T[]; total: number; page: number; page_size: number; total_pages: number }

export const OPERATIONAL_TENANT_STATUSES = ['ACTIVE', 'TRIAL']

const API_BASE_URL = (process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1').replace(/\/$/, '')
const SESSION_KEY = 'emplyra.session'

let accessToken: string | null = null
let refreshToken: string | null = null
let tenantId: string | null = null

function persist(session: Session | null) {
  accessToken = session?.access_token ?? null
  refreshToken = session?.refresh_token ?? null
  if (typeof window === 'undefined') return
  if (session) window.sessionStorage.setItem(SESSION_KEY, JSON.stringify(session))
  else window.sessionStorage.removeItem(SESSION_KEY)
}

export function setSession(session: Session | null) {
  persist(session)
  if (!session) setTenantId(null)
}

// restoreSession rehydrates tokens from sessionStorage (compat improvement; the
// primary cache remains the in-memory variables, so SSR never sees a session).
export function restoreSession(): Session | null {
  if (typeof window === 'undefined') return null
  try {
    const raw = window.sessionStorage.getItem(SESSION_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Session
    if (parsed?.access_token) {
      accessToken = parsed.access_token
      refreshToken = parsed.refresh_token ?? null
      return parsed
    }
    window.sessionStorage.removeItem(SESSION_KEY)
  } catch {
    window.sessionStorage.removeItem(SESSION_KEY)
  }
  return null
}

export function hasSession() { return Boolean(accessToken) }

// Tenant context is set from /auth/me (never from URLs or user input) and attached
// as an opt-in header. Empty tenant id keeps today's platform-wide behavior.
export function setTenantId(id: string | null) { tenantId = id }
export function getTenantId() { return tenantId }

async function request<T>(path: string, init: RequestInit = {}, canRefresh = true): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Content-Type', 'application/json')
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`)
  if (tenantId) headers.set('X-Tenant-ID', tenantId)
  const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers, cache: 'no-store' })
  const payload = await response.json().catch(() => ({})) as ApiEnvelope<T>
  if (response.status === 401 && canRefresh && refreshToken) {
    const refreshed = await request<LoginResult>('/auth/refresh', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) }, false)
    persist(refreshed.tokens)
    return request<T>(path, init, false)
  }
  if (!response.ok || payload.success === false) throw new Error(payload.message || `Request failed with status ${response.status}`)
  return payload.data
}

const list = <T,>(path: string, page = 1, pageSize = 10, search = '') => request<ListResponse<T>>(`${path}?page=${page}&page_size=${pageSize}${search ? `&search=${encodeURIComponent(search)}` : ''}`)

export const api = {
  login: (identifier: string, password: string) => request<LoginResult>('/auth/login', { method: 'POST', body: JSON.stringify({ identifier, password }) }, false),
  me: () => request<MeResponse>('/auth/me'),
  logout: () => request<unknown>('/auth/logout', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) }, false),
  dashboard: () => request<DashboardSummary>('/dashboard/summary'),
  employees: (p = 1, s = 10, q = '') => list<Employee>('/employees', p, s, q),
  module: (path: string, p = 1, s = 10, q = '') => list<Record<string, unknown>>(path, p, s, q),
  notifications: () => list<Record<string, unknown>>('/notifications', 1, 20),
  unreadCount: () => request<{ count: number }>('/notifications/unread-count'),
  markAllRead: () => request<unknown>('/notifications/read-all', { method: 'PUT' }),
}

export function getApiBaseUrl() { return API_BASE_URL }