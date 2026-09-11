'use client'

import { useCallback, useEffect, useState } from 'react'
import { BarChart3, CheckCircle2, RefreshCw, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api } from '@/lib/api'

type ReportTab = 'headcount' | 'attendance' | 'leaves' | 'payroll' | 'recruitment' | 'holidays'

const tabs: { key: ReportTab; label: string; description: string }[] = [
  { key: 'headcount', label: 'Headcount', description: 'Employees by department and status.' },
  { key: 'attendance', label: 'Attendance', description: 'Presence, absences, and average work hours.' },
  { key: 'leaves', label: 'Leaves', description: 'Leave applications, approvals, and usage.' },
  { key: 'payroll', label: 'Payroll', description: 'Runs, gross/net totals, and payment status.' },
  { key: 'recruitment', label: 'Recruitment', description: 'Jobs, candidates, and pipeline health.' },
  { key: 'holidays', label: 'Holidays', description: 'Company-wide holiday calendar.' },
]

function Row({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-border/60 py-2.5 text-sm last:border-0">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  )
}

function Panel({ title, data }: { title: string; data: unknown }) {
  const rows: { label: string; value: string | number }[] = []
  if (Array.isArray(data)) {
    data.slice(0, 15).forEach((item) => {
      if (item && typeof item === 'object') {
        const obj = item as Record<string, unknown>
        const label = (obj.name ?? obj.title ?? `${obj.first_name} ${obj.last_name}` ?? obj.employee_name ?? obj.job_title ?? obj.month ?? obj.date ?? 'Item').toString()
        const count = obj.count ?? obj.total ?? obj.employee_count ?? obj.total_days ?? obj.net_total ?? obj.gross_total ?? obj.volume ?? obj.status ?? obj.year
        rows.push({ label, value: count === undefined ? '' : String(count) })
      }
    })
  } else if (data && typeof data === 'object') {
    const obj = data as Record<string, unknown>
    Object.entries(obj).forEach(([key, value]) => {
      if (value !== null && typeof value !== 'object') rows.push({ label: key.replaceAll('_', ' '), value: String(value) })
    })
  }
  if (!rows.length) return <p className="p-6 text-sm text-muted-foreground">No data points in this report yet.</p>
  return <dl className="p-2">{rows.map((row, i) => <Row key={i} {...row} />)}</dl>
}

export default function ReportsPage() {
  const [tab, setTab] = useState<ReportTab>('headcount')
  const [data, setData] = useState<unknown>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')

  const load = useCallback(async (key: ReportTab = tab) => {
    setLoading(true)
    setError('')
    setNotice('')
    try {
      setData(await api.report(`/reports/${key}`))
      setNotice(`Loaded ${tabs.find((t) => t.key === key)?.label.toLowerCase()} report.`)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load report.')
    } finally {
      setLoading(false)
    }
  }, [tab])

  useEffect(() => { void load() }, [load])

  const active = tabs.find((t) => t.key === tab)!

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-2">
        <p className="text-sm font-medium text-primary">Insights</p>
        <h1 className="text-3xl font-semibold tracking-tight">Reports</h1>
        <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Live HRMS analytics for headcount, attendance, leave, payroll, hiring, and holidays.</p>
      </header>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3" role="tablist" aria-label="Report sections">
        {tabs.map((t) => (
          <button key={t.key} role="tab" aria-selected={t.key === tab} onClick={() => setTab(t.key)} className={`flex flex-col gap-1 rounded-2xl border p-4 text-left transition-colors ${t.key === tab ? 'border-primary bg-primary/5' : 'border-border bg-card hover:bg-muted/40'}`}>
            <p className="text-sm font-semibold">{t.label}</p>
            <p className="text-xs leading-5 text-muted-foreground">{t.description}</p>
          </button>
        ))}
      </div>

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="overflow-hidden rounded-2xl border border-border bg-card">
        <div className="flex items-center justify-between border-b border-border p-5">
          <div className="flex items-center gap-2"><BarChart3 className="text-primary" /><p className="font-medium">{active.label} report</p></div>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
        </div>
        {loading ? <p className="p-8 text-sm text-muted-foreground">Loading live data…</p> : <Panel title={active.label} data={data} />}
      </div>
    </section>
  )
}