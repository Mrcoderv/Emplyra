'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { CheckCircle2, Plus, RefreshCw, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api, type ModuleRecord, type RequestOptions } from '@/lib/api'

type CoreModule = 'departments' | 'attendance' | 'leaves'

const config: Record<CoreModule, { title: string; eyebrow: string; description: string; path: string; empty: string; action?: string }> = {
  departments: { title: 'Departments', eyebrow: 'Organization', description: 'Manage the teams and reporting structure behind your people operations.', path: '/departments', empty: 'No departments have been created yet.', action: 'New department' },
  attendance: { title: 'Attendance', eyebrow: 'Time & attendance', description: 'Review today’s attendance and keep time records accurate.', path: '/attendance', empty: 'No attendance records match this view.', action: 'Check in' },
  leaves: { title: 'Leave requests', eyebrow: 'Time off', description: 'Track requests, balances, and approvals from one workspace.', path: '/leaves', empty: 'No leave requests match this view.', action: 'Request leave' },
}

function displayValue(value: unknown) {
  if (value === null || value === undefined || value === '') return '—'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

export default function CoreModulePage({ module }: { module: CoreModule }) {
  const meta = config[module]
  const [rows, setRows] = useState<ModuleRecord[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const options: RequestOptions = { page: 1, pageSize: 50, search: query }
      setRows(await api.module(meta.path, options).then((result) => result.items ?? []))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to connect to the Emplyra API.')
    } finally {
      setLoading(false)
    }
  }, [meta.path, query])

  useEffect(() => { void load() }, [load])

  const columns = useMemo(() => {
    const keys = new Set<string>()
    rows.slice(0, 8).forEach((row) => Object.keys(row).forEach((key) => keys.add(key)))
    return Array.from(keys).filter((key) => !['id', 'tenant_id', 'updated_at', 'created_at'].includes(key)).slice(0, 4)
  }, [rows])

  async function performAction() {
    if (module !== 'attendance') {
      setNotice(`${meta.action} is ready once the form fields are connected to your API policy.`)
      return
    }
    try {
      await api.attendanceCheckIn()
      setNotice('You are checked in. Attendance was updated successfully.')
      await load()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to check in.')
    }
  }

  return <section className="flex flex-col gap-6">
    <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
      <div className="flex flex-col gap-2"><p className="text-sm font-medium text-primary">{meta.eyebrow}</p><h1 className="text-3xl font-semibold tracking-tight text-balance">{meta.title}</h1><p className="max-w-2xl text-sm leading-6 text-muted-foreground">{meta.description}</p></div>
      <Button onClick={() => void performAction()}><Plus data-icon="inline-start" />{meta.action}</Button>
    </header>
    {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="text-primary" />{notice}</div>}
    <div className="rounded-2xl border border-border bg-card">
      <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between"><div className="relative max-w-sm flex-1"><Search className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" /><input aria-label={`Search ${meta.title}`} value={query} onChange={(event) => setQuery(event.target.value)} placeholder={`Search ${meta.title.toLowerCase()}`} className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" /></div><Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button></div>
      {error ? <div role="alert" className="flex flex-col gap-3 p-8"><p className="text-sm text-destructive">{error}</p><Button className="w-fit" variant="outline" size="sm" onClick={() => void load()}>Try again</Button></div> : loading ? <div className="p-8 text-sm text-muted-foreground">Connecting to the Emplyra API…</div> : rows.length === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><p className="font-medium">{meta.empty}</p><p className="text-sm text-muted-foreground">Records will appear here when they are available in your workspace.</p></div> : <div className="overflow-x-auto"><table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Record</th>{columns.map((column) => <th key={column} className="px-5 py-3">{column.replaceAll('_', ' ')}</th>)}</tr></thead><tbody>{rows.map((row, index) => <tr key={row.id ?? index} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{displayValue(row.name ?? row.title ?? row.employee_name ?? row.id ?? `Record ${index + 1}`)}</td>{columns.map((column) => <td key={column} className="max-w-56 truncate px-5 py-4 text-muted-foreground">{displayValue(row[column])}</td>)}</tr>)}</tbody></table></div>}
    </div>
  </section>
}
