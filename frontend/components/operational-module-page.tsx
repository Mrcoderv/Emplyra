'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { RefreshCw, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api, type ModuleRecord, type RequestOptions } from '@/lib/api'

type ModuleConfig = {
  title: string
  eyebrow: string
  description: string
  path: string
  empty: string
  columns?: string[]
  report?: boolean
}

const configs: Record<string, ModuleConfig> = {
  payroll: { title: 'Payroll', eyebrow: 'Compensation', description: 'Review payroll runs, payment status, and employee compensation records.', path: '/payroll', empty: 'No payroll records are available yet.' },
  recruitment: { title: 'Recruitment', eyebrow: 'Talent acquisition', description: 'Track open roles, candidates, applications, and hiring activity.', path: '/recruitment/jobs', empty: 'No open roles have been created yet.' },
  performance: { title: 'Performance', eyebrow: 'People programs', description: 'Keep goals and reviews visible so managers and employees can act on them.', path: '/performance/goals', empty: 'No performance goals are available yet.' },
  training: { title: 'Training', eyebrow: 'People development', description: 'Manage learning programs and keep enrollment progress in one place.', path: '/training/programs', empty: 'No training programs have been created yet.' },
  reports: { title: 'Reports', eyebrow: 'Insights', description: 'Use live HRMS data to understand headcount, attendance, leave, payroll, and hiring.', path: '/reports/headcount', empty: 'No report data is available yet.', report: true },
}

function value(value: unknown) {
  if (value === null || value === undefined || value === '') return '—'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

export default function OperationalModulePage({ module }: { module: keyof typeof configs }) {
  const meta = configs[module]
  const [rows, setRows] = useState<ModuleRecord[]>([])
  const [report, setReport] = useState<unknown>(null)
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      if (meta.report) setReport(await api.report(meta.path))
      else {
        const result = await api.module(meta.path, { page: 1, pageSize: 50, search: query } satisfies RequestOptions)
        setRows(result.items ?? [])
      }
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load this workspace.')
    } finally { setLoading(false) }
  }, [meta.path, meta.report, query])

  useEffect(() => { void load() }, [load])

  const columns = useMemo(() => {
    if (meta.columns) return meta.columns
    const keys = new Set<string>()
    rows.slice(0, 10).forEach((row) => Object.keys(row).forEach((key) => keys.add(key)))
    return Array.from(keys).filter((key) => !['id', 'tenant_id', 'updated_at', 'created_at'].includes(key)).slice(0, 5)
  }, [meta.columns, rows])

  return <section className="flex flex-col gap-6">
    <header className="flex flex-col gap-2"><p className="text-sm font-medium text-primary">{meta.eyebrow}</p><h1 className="text-3xl font-semibold tracking-tight text-balance">{meta.title}</h1><p className="max-w-2xl text-sm leading-6 text-muted-foreground">{meta.description}</p></header>
    <div className="rounded-2xl border border-border bg-card">
      <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between"><div className="relative max-w-sm flex-1"><Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" /><input aria-label={`Search ${meta.title}`} value={query} onChange={(event) => setQuery(event.target.value)} placeholder={`Search ${meta.title.toLowerCase()}`} className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" /></div><Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button></div>
      {error ? <div role="alert" className="flex flex-col gap-3 p-8"><p className="text-sm text-destructive">{error}</p><Button variant="outline" size="sm" className="w-fit" onClick={() => void load()}>Try again</Button></div> : loading ? <p className="p-8 text-sm text-muted-foreground">Loading live records…</p> : meta.report ? <div className="p-6"><pre className="overflow-x-auto rounded-xl bg-muted/50 p-4 text-xs leading-6 text-foreground">{JSON.stringify(report, null, 2)}</pre></div> : !rows.length ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><p className="font-medium">{meta.empty}</p><p className="text-sm text-muted-foreground">Records will appear here when they are available in your workspace.</p></div> : <div className="overflow-x-auto"><table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Record</th>{columns.map((column) => <th key={column} className="px-5 py-3">{column.replaceAll('_', ' ')}</th>)}</tr></thead><tbody>{rows.map((row, index) => <tr key={row.id ?? index} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{value(row.name ?? row.title ?? row.employee_name ?? row.id ?? `Record ${index + 1}`)}</td>{columns.map((column) => <td key={column} className="max-w-56 truncate px-5 py-4 text-muted-foreground">{value(row[column])}</td>)}</tr>)}</tbody></table></div>}
    </div>
  </section>
}

export { configs }
