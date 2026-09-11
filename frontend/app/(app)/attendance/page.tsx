'use client'

import { useCallback, useEffect, useState } from 'react'
import { CheckCircle2, LogIn, LogOut, Pencil, RefreshCw, Search, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { api, type ModuleRecord } from '@/lib/api'

const STATUS_VARIANT: Record<string, 'success' | 'warning' | 'destructive' | 'outline' | 'primary'> = {
  PRESENT: 'success', HALF_DAY: 'warning', LATE: 'warning', ABSENT: 'destructive', ON_LEAVE: 'primary', HOLIDAY: 'outline',
}

const inputClass = 'h-10 w-full rounded-xl border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring'

export default function AttendancePage() {
  const [rows, setRows] = useState<ModuleRecord[]>([])
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [status, setStatus] = useState('')
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState<'' | 'in' | 'out' | 'update'>('')
  const [editing, setEditing] = useState<ModuleRecord | null>(null)
  const [updateForm, setUpdateForm] = useState({ check_in: '', check_out: '', status: '', late_minutes: '', remarks: '' })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const result = await api.attendance({ page: 1, pageSize: 50, search: query, query: { from: from || undefined, to: to || undefined, status: status || undefined } })
      setRows(result.items ?? [])
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load attendance records.')
    } finally {
      setLoading(false)
    }
  }, [query, from, to, status])

  useEffect(() => { void load() }, [load])

  async function checkIn() {
    setBusy('in'); setError(''); setNotice('')
    try { await api.attendanceCheckIn(); setNotice('Checked in successfully.'); await load() }
    catch (cause) { setError(cause instanceof Error ? cause.message : 'Unable to check in.') }
    finally { setBusy('') }
  }

  async function checkOut() {
    setBusy('out'); setError(''); setNotice('')
    try { await api.attendanceCheckOut(); setNotice('Checked out successfully.'); await load() }
    catch (cause) { setError(cause instanceof Error ? cause.message : 'Unable to check out.') }
    finally { setBusy('') }
  }

  function openEdit(row: ModuleRecord) {
    setEditing(row)
    setError('')
    const ci = row.check_in as string | undefined
    const co = row.check_out as string | undefined
    setUpdateForm({
      check_in: ci ? ci.slice(0, 16) : '',
      check_out: co ? co.slice(0, 16) : '',
      status: String(row.status ?? ''),
      late_minutes: String(row.late_minutes ?? ''),
      remarks: String(row.remarks ?? ''),
    })
  }

  async function saveEdit(event: React.FormEvent) {
    event.preventDefault()
    if (!editing?.id) return
    setBusy('update'); setError(''); setNotice('')
    try {
      const body: Record<string, unknown> = { status: updateForm.status || undefined, remarks: updateForm.remarks || undefined }
      if (updateForm.check_in) body.check_in = new Date(updateForm.check_in).toISOString()
      if (updateForm.check_out) body.check_out = new Date(updateForm.check_out).toISOString()
      if (updateForm.late_minutes !== '') body.late_minutes = Number(updateForm.late_minutes)
      await api.attendanceUpdate(editing.id, body)
      setEditing(null)
      setNotice('Attendance record updated.')
      await load()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to update record.')
    } finally {
      setBusy('')
    }
  }

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium text-primary">Time &amp; attendance</p>
          <h1 className="text-3xl font-semibold tracking-tight">Attendance</h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Review today&apos;s attendance and keep time records accurate.</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button onClick={() => void checkIn()} disabled={busy === 'in'}><LogIn data-icon="inline-start" />{busy === 'in' ? 'Checking in…' : 'Check in'}</Button>
          <Button variant="outline" onClick={() => void checkOut()} disabled={busy === 'out'}><LogOut data-icon="inline-start" />{busy === 'out' ? 'Checking out…' : 'Check out'}</Button>
        </div>
      </header>

      {editing && (
        <form onSubmit={saveEdit} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          <div className="flex flex-col gap-2"><label htmlFor="ci" className="text-sm font-medium">Check-in</label><Input id="ci" type="datetime-local" value={updateForm.check_in} onChange={(e) => setUpdateForm({ ...updateForm, check_in: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="co" className="text-sm font-medium">Check-out</label><Input id="co" type="datetime-local" value={updateForm.check_out} onChange={(e) => setUpdateForm({ ...updateForm, check_out: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="st" className="text-sm font-medium">Status</label><Select id="st" value={updateForm.status} onChange={(e) => setUpdateForm({ ...updateForm, status: e.target.value })} options={['PRESENT', 'HALF_DAY', 'LATE', 'ABSENT', 'ON_LEAVE', 'HOLIDAY']} placeholder="Select status" /></div>
          <div className="flex flex-col gap-2"><label htmlFor="lm" className="text-sm font-medium">Late minutes</label><Input id="lm" type="number" min={0} value={updateForm.late_minutes} onChange={(e) => setUpdateForm({ ...updateForm, late_minutes: e.target.value })} /></div>
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="rm" className="text-sm font-medium">Remarks</label><Textarea id="rm" value={updateForm.remarks} onChange={(e) => setUpdateForm({ ...updateForm, remarks: e.target.value })} /></div>
          <div className="flex gap-2 sm:col-span-2">
            <Button type="submit" disabled={busy === 'update'}>{busy === 'update' ? 'Saving…' : 'Save changes'}</Button>
            <Button type="button" variant="outline" onClick={() => setEditing(null)}>Cancel</Button>
          </div>
        </form>
      )}

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="rounded-2xl border border-border bg-card">
        <div className="flex flex-col gap-3 border-b border-border p-5 lg:flex-row lg:items-center lg:justify-between">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input aria-label="Search attendance" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search attendance" className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <div className="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:items-center">
            <Input type="date" aria-label="From date" value={from} onChange={(e) => setFrom(e.target.value)} className="h-10 w-full rounded-xl border border-input bg-background px-3 text-sm sm:w-36" />
            <Input type="date" aria-label="To date" value={to} onChange={(e) => setTo(e.target.value)} className="h-10 w-full rounded-xl border border-input bg-background px-3 text-sm sm:w-36" />
            <Select aria-label="Status filter" value={status} onChange={(e) => setStatus(e.target.value)} options={['PRESENT', 'HALF_DAY', 'LATE', 'ABSENT', 'ON_LEAVE', 'HOLIDAY']} placeholder="All statuses" />
            <Button variant="outline" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
          </div>
        </div>
        {loading ? <p className="p-8 text-sm text-muted-foreground">Loading attendance records…</p> :
        rows.length === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><p className="font-medium">No attendance records match this view.</p><p className="text-sm text-muted-foreground">Use Check in to create today&apos;s first record.</p></div> :
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Employee</th><th className="px-5 py-3">Date</th><th className="px-5 py-3">Check-in</th><th className="px-5 py-3">Check-out</th><th className="px-5 py-3">Hours</th><th className="px-5 py-3">Status</th><th className="px-5 py-3 text-right">Actions</th></tr></thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.id} className="border-t border-border hover:bg-muted/30">
                  <td className="px-5 py-4 font-medium">{row.employee && typeof row.employee === 'object' ? `${(row.employee as { first_name?: string }).first_name ?? ''} ${(row.employee as { last_name?: string }).last_name ?? ''}`.trim() || 'Employee' : '—'}</td>
                  <td className="px-5 py-4">{String(row.date ?? '—')}</td>
                  <td className="px-5 py-4 text-muted-foreground">{fmtTime(row.check_in)}</td>
                  <td className="px-5 py-4 text-muted-foreground">{fmtTime(row.check_out)}</td>
                  <td className="px-5 py-4 text-muted-foreground">{row.working_hours != null ? `${Number(row.working_hours).toFixed(2)}h` : '—'}</td>
                  <td className="px-5 py-4"><Badge variant={STATUS_VARIANT[String(row.status ?? '')] ?? 'outline'}>{String(row.status ?? '—').replaceAll('_', ' ')}</Badge></td>
                  <td className="px-5 py-4 text-right"><Button variant="ghost" size="icon-sm" aria-label="Edit record" onClick={() => openEdit(row)}><Pencil /></Button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>}
      </div>
    </section>
  )
}

function fmtTime(value: unknown) {
  if (!value) return '—'
  const s = String(value)
  if (!s.includes('T')) return s
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s
  return d.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}