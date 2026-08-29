'use client'

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

import {
  api,
  OPERATIONAL_TENANT_STATUSES,
  restoreSession,
  setSession,
  setTenantId,
  type CurrentUser,
  type MeResponse,
  type MeTenant,
  type Scope,
} from '@/lib/api'

export type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated'

export type AuthContextValue = {
  status: AuthStatus
  user: CurrentUser | null
  me: MeResponse | null
  roles: string[]
  permissions: string[]
  scope: Scope | null
  tenant: MeTenant | null
  tenantSuspended: boolean
  login: (identifier: string, password: string) => Promise<void>
  logout: () => Promise<void>
  can: (permission: string) => boolean
  hasRole: (role: string) => boolean
  isScope: (scope: Scope) => boolean
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>('loading')
  const [me, setMe] = useState<MeResponse | null>(null)

  useEffect(() => {
    let cancelled = false
    if (!restoreSession()) {
      setStatus('unauthenticated')
      return
    }
    api
      .me()
      .then((r) => {
        if (cancelled) return
        setMe(r)
        setTenantId(r.tenant?.id ?? null)
        setStatus('authenticated')
      })
      .catch(() => {
        if (cancelled) return
        setSession(null)
        setStatus('unauthenticated')
      })
    return () => {
      cancelled = true
    }
  }, [])

  const login = useCallback(async (identifier: string, password: string) => {
    const result = await api.login(identifier, password)
    setSession(result.tokens)
    const r = await api.me()
    setMe(r)
    setTenantId(r.tenant?.id ?? null)
    setStatus('authenticated')
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.logout()
    } finally {
      setMe(null)
      setSession(null)
      setStatus('unauthenticated')
    }
  }, [])

  const value = useMemo<AuthContextValue>(() => {
    const permissions = me?.permissions ?? []
    const roles = me?.roles ?? []
    const scope = me?.scope ?? null
    const tenant = me?.tenant ?? null
    return {
      status,
      user: me?.user ?? null,
      me,
      roles,
      permissions,
      scope,
      tenant,
      tenantSuspended: !!tenant && !OPERATIONAL_TENANT_STATUSES.includes(tenant.status),
      login,
      logout,
      can: (p) => permissions.includes(p),
      hasRole: (r) => roles.includes(r),
      isScope: (s) => scope === s,
    }
  }, [status, me, login, logout])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}