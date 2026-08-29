'use client'

// Rendered at the absolute top level (replacing the root layout) when an
// uncaught error bubbles past the nearest error boundary. It must not depend on
// layout-level providers — keep this component self-contained.
export const dynamic = 'force-dynamic'

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <html lang="en" className="bg-background">
      <body className="antialiased">
        <div className="grid min-h-screen place-items-center bg-background px-4 text-foreground">
          <div className="max-w-md rounded-2xl border border-border bg-card p-8 text-center">
            <h2 className="text-xl font-semibold">Something went wrong</h2>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
              An unexpected error occurred. You can try again, or contact support.
            </p>
            <button
              onClick={reset}
              className="mt-6 inline-flex items-center justify-center gap-2 rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground"
            >
              Try again
            </button>
          </div>
        </div>
      </body>
    </html>
  )
}