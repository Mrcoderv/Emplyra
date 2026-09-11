'use client'

import { useCallback, useEffect, useState } from 'react'
import { BookOpen, CheckCircle2, Plus, RefreshCw, Search, Trash2, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { api, type TrainingEnrollment, type TrainingProgram, type TrainingSchedule } from '@/lib/api'

const STATUS_VARIANT: Record<string, 'default' | 'primary' | 'warning' | 'success' | 'destructive' | 'outline'> = {
  PLANNED: 'outline', ACTIVE: 'success', SCHEDULED: 'outline', ONGOING: 'warning', COMPLETED: 'primary', CANCELLED: 'destructive', DRAFT: 'warning',
  ENROLLED: 'primary', IN_PROGRESS: 'warning', PASSED: 'success', FAILED: 'destructive', DROPPED: 'outline',
}

export default function TrainingPage() {
  const [tab, setTab] = useState<'programs' | 'schedules' | 'enrollments'>('programs')
  const [programs, setPrograms] = useState<TrainingProgram[]>([])
  const [schedules, setSchedules] = useState<TrainingSchedule[]>([])
  const [enrollments, setEnrollments] = useState<TrainingEnrollment[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState('')
  const [openForm, setOpenForm] = useState(false)
  const [progForm, setProgForm] = useState({ title: '', description: '', provider: '', start_date: '', end_date: '', location: '', max_seats: '' })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [ps, ss, es] = await Promise.all([
        api.trainingPrograms({ page: 1, pageSize: 50, search: query }),
        api.trainingSchedule.list({ search: query }).catch(() => []),
        api.trainingEnrollment.list({ page: 1, pageSize: 50, search: query }).catch(() => ({ items: [] })),
      ])
      setPrograms(ps.items ?? [])
      setSchedules(ss ?? [])
      setEnrollments(es.items ?? [])
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load training data.')
    } finally {
      setLoading(false)
    }
  }, [query])

  useEffect(() => { void load() }, [load])

  async function runAction(name: string, promise: Promise<unknown>, successMessage: string) {
    setBusy(name); setError(''); setNotice('')
    try { await promise; setNotice(successMessage); await load() }
    catch (cause) { setError(cause instanceof Error ? cause.message : `${name} failed.`) }
    finally { setBusy('') }
  }

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium text-primary">People development</p>
          <h1 className="text-3xl font-semibold tracking-tight">Training</h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Plan learning programs, schedule sessions, and track enrollment.</p>
        </div>
        <Button onClick={() => setOpenForm((v) => !v)} disabled={tab !== 'programs'}><Plus data-icon="inline-start" />New program</Button>
      </header>

      <div className="flex gap-1 rounded-xl bg-muted/60 p-1 text-sm sm:w-fit" role="tablist" aria-label="Training sections">
        {(['programs', 'schedules', 'enrollments'] as const).map((key) => (
          <button key={key} role="tab" aria-selected={tab === key} onClick={() => setTab(key)} className={`rounded-lg px-4 py-2 font-medium capitalize ${tab === key ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'}`}>{key}</button>
        ))}
      </div>

      {tab === 'programs' && openForm && (
        <form onSubmit={(e) => {
          e.preventDefault()
          void runAction('program', api.trainingProgram.create({ ...progForm, max_seats: progForm.max_seats ? Number(progForm.max_seats) : undefined }), 'Training program created.')
          setProgForm({ title: '', description: '', provider: '', start_date: '', end_date: '', location: '', max_seats: '' })
          setOpenForm(false)
        }} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="t" className="text-sm font-medium">Program title</label><Input id="t" required value={progForm.title} onChange={(e) => setProgForm({ ...progForm, title: e.target.value })} /></div>
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="d" className="text-sm font-medium">Description</label><Textarea id="d" value={progForm.description} onChange={(e) => setProgForm({ ...progForm, description: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="p" className="text-sm font-medium">Provider</label><Input id="p" value={progForm.provider} onChange={(e) => setProgForm({ ...progForm, provider: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="loc" className="text-sm font-medium">Location</label><Input id="loc" value={progForm.location} onChange={(e) => setProgForm({ ...progForm, location: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="sd" className="text-sm font-medium">Start date</label><Input id="sd" type="date" required value={progForm.start_date} onChange={(e) => setProgForm({ ...progForm, start_date: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="ed" className="text-sm font-medium">End date</label><Input id="ed" type="date" required value={progForm.end_date} onChange={(e) => setProgForm({ ...progForm, end_date: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="ms" className="text-sm font-medium">Max seats</label><Input id="ms" type="number" min={0} value={progForm.max_seats} onChange={(e) => setProgForm({ ...progForm, max_seats: e.target.value })} /></div>
          <div className="flex items-end gap-2"><Button type="submit" disabled={busy === 'program'}>Create program</Button><Button type="button" variant="outline" onClick={() => setOpenForm(false)}>Cancel</Button></div>
        </form>
      )}

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="rounded-2xl border border-border bg-card">
        <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input aria-label="Search training" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search training" className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
        </div>
        {loading ? <p className="p-8 text-sm text-muted-foreground">Loading…</p> :
        (tab === 'programs' ? programs.length : tab === 'schedules' ? schedules.length : enrollments.length) === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><BookOpen className="text-muted-foreground" /><p className="font-medium">No {tab} yet.</p><p className="text-sm text-muted-foreground">{tab === 'programs' ? 'Create a program to get started.' : `Programs will appear here once scheduled.`}</p></div> :
        tab === 'programs' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Program</th><th className="px-5 py-3">Provider</th><th className="px-5 py-3">Dates</th><th className="px-5 py-3">Location</th><th className="px-5 py-3">Status</th><th className="px-5 py-3 text-right">Actions</th></tr></thead><tbody>
            {programs.map((p) => (<tr key={p.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{p.title}<p className="text-xs text-muted-foreground">{p.description}</p></td><td className="px-5 py-4 text-muted-foreground">{p.provider || '—'}</td><td className="px-5 py-4 text-muted-foreground">{p.start_date ? `${p.start_date} → ${p.end_date ?? '…'}` : '—'}</td><td className="px-5 py-4 text-muted-foreground">{p.location || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[p.status ?? ''] ?? 'outline'}>{String(p.status ?? '—').replaceAll('_', ' ')}</Badge></td><td className="px-5 py-4 text-right"><Button variant="ghost" size="icon-sm" aria-label="Delete program" onClick={() => { if (window.confirm('Delete this program?')) void runAction('delete', api.trainingProgram.remove(p.id!), 'Program deleted.') }}><Trash2 /></Button></td></tr>))}
          </tbody></table>
        ) : tab === 'schedules' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Program</th><th className="px-5 py-3">Date</th><th className="px-5 py-3">Time</th><th className="px-5 py-3">Trainer</th><th className="px-5 py-3">Location</th></tr></thead><tbody>
            {schedules.map((s) => (<tr key={s.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{s.program?.title || 'Program session'}</td><td className="px-5 py-4 text-muted-foreground">{s.date || '—'}</td><td className="px-5 py-4 text-muted-foreground">{[s.start_time, s.end_time].filter(Boolean).join(' – ') || '—'}</td><td className="px-5 py-4 text-muted-foreground">{s.trainer || '—'}</td><td className="px-5 py-4 text-muted-foreground">{s.location || '—'}</td></tr>))}
          </tbody></table>
        ) : (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Employee</th><th className="px-5 py-3">Program</th><th className="px-5 py-3">Status</th><th className="px-5 py-3 text-right">Update</th></tr></thead><tbody>
            {enrollments.map((e) => (<tr key={e.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{typeof e.employee === 'object' && e.employee ? `${e.employee.first_name ?? ''} ${e.employee.last_name ?? ''}`.trim() || 'Employee' : '—'}</td><td className="px-5 py-4 text-muted-foreground">{e.program?.title || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[e.status ?? ''] ?? 'outline'}>{String(e.status ?? '—').replaceAll('_', ' ')}</Badge></td><td className="px-5 py-4 text-right"><Select aria-label="Update enrollment status" value="" onChange={(ev) => { const val = String(ev.target.value); if (val && e.id) void runAction('enroll', api.trainingEnrollment.update(e.id, { status: val }), 'Enrollment updated.') }} options={['ENROLLED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED']} placeholder="Set status…" className="h-9 w-40" /></td></tr>))}
          </tbody></table>
        )}
      </div>
    </section>
  )
}