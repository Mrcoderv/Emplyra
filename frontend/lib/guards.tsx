'use client'

import { useAuth } from '@/lib/auth'
import type { Scope } from '@/lib/api'

/**
 * Route-level scope guard. Renders children only when the authenticated session
 * has one of the required scopes; otherwise it renders the `fallback` (default: a
 * visible "wrong role" notice — never a silent switch of context).
 */
export function RequireScope({
  scopes,
  children,
  fallback,
}: {
  scopes: Scope[]
  children: React.ReactNode
  fallback?: React.ReactNode
}) {
  const { status, scope } = useAuth()
  if (status === 'loading') {
    return (
      <div className="grid min-h-screen place-items-center bg-background text-sm text-muted-foreground">
        Loading…
      </div>
    )
  }
  if (!scope || !scopes.includes(scope)) {
    return fallback ?? (
      <div className="grid min-h-screen place-items-center bg-background px-4">
        <div className="max-w-md rounded-2xl border border-border bg-card p-8 text-center">
          <h2 className="text-xl font-semibold">Access restricted</h2>
          <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
            Sign in with an account that is allowed in this area.
          </p>
          <a
            href="/"
            className="mt-6 inline-flex items-center justify-center gap-2 rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground"
          >
            Back to sign in
          </a>
        </div>
      </div>
    )
  }
  return <>{children}</>
}

export function requireScope(...scopes: Scope[]) {
  return function ScopeGuard({ children }: { children: React.ReactNode }) {
    return <RequireScope scopes={scopes}>{children}</RequireScope>
  }
}

export const TenantScope = requireScope('tenant')
export const PlatformScope = requireScope('platform')