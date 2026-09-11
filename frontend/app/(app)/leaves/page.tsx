'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { Check, CheckCircle2, Plus, RefreshCw, Search, X, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { api, type Leave, type LeaveBalance, type LeaveType } from '@/lib/api'

const STATUS_VARIANT: Record<string, 'warning' | 'success' | 'destructive' | 'outline'> = {
  PENDING: 'warning', APPROVED: 'success', REJECTED: 'destructive', CANCELLED: 'outline',
}

export default function LeavesPage() {
  const [rows, setRows] = useState<Leave[]>([])
  const [types, setTypes] = useState<LeaveType[]>([])
  const [balances, setBalances] = useState<LeaveBalance[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({ leave_type_id: '', start_date: '', end_date: '', reason: '' })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [leaveResult, leaveTypes, leaveBalances] = await Promise.all([
        api.leaves({ page: 1, pageSize: 50, search: query }),
        api.leaveTypes(),
        api.leaveBalances().catch(() => []),
      ])
      setRows(leaveResult.items ?? [])
      setTypes(leaveTypes ?? [])
      setBalances(leaveBalances ?? [])
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load leave requests.')
    } finally {
      setLoading(false)
    }
  }, [query])

  useEffect(() => { void load() }, [load])

  const typeName = useMemo(() => {
    const map = new Map(types.map((t) => [t.id, t.name]))
    return (id?: string) => (id && map.get(id)) || '—'
  }, [types])

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    setSaving(true)
    setError('')
    setNotice('')
    try {
      await api.createLeave(form)
      setForm({ leave_type_id: '', start_date: '', end_date: '', reason: '' })
      setFormOpen(false)
      setNotice('Leave request submitted successfully.')
      await load()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to submit leave request.')
    } finally {
      setSaving(false)
    }
  }

  async function decide(id: string, action: 'approve' | 'reject') {
    setError('')
    setNotice('')
    try {
      const note = action === 'approve' ? '' : window.prompt('Reason for rejection (optional):') ?? ''
      if (action === 'approve') await api.approveLeave(id)
      else await api.rejectLeave(id, note)
      setNotice(`Leave request ${action === 'approve' ? 'approved' : 'rejected'}.`)
      await load()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to update leave request.')
    }
  }

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium text-primary">Time off</p>
          <h1 className="text-3xl font-semibold tracking-tight">Leave requests</h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Track requests, balances, and approvals from one workspace.</p>
        </div>
        <Button onClick={() => { setFormOpen(!formOpen); setError(''); setNotice('') }}><Plus data-icon="inline-start" />Request leave</Button>
      </header>

      {formOpen && (
        <form onSubmit={submit} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          <div className="flex flex-col gap-2 sm:col-span-2">
            <label htmlFor="leave-type" className="text-sm font-medium">Leave type</label>
            <Select id="leave-type" required placeholder="Select a leave type" value={form.leave_type_id} onChange={(e) => setForm({ ...form, leave_type_id: e.target.value })} options={types.map((t) => ({ label: `${t.name}${t.is_paid ? '' : ' (unpaid)'}`, value: t.id }))} />
          </div>
          <div className="flex flex-col gap-2"><label htmlFor="start" className="text-sm font-medium">Start date</label><Input id="start" required type="date" value={form.start_date} onChange={(e) => setForm({ ...form, start_date: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="end" className="text-sm font-medium">End date</label><Input id="end" required type="date" value={form.end_date} onChange={(e) => setForm({ ...form, end_date: e.target.value })} /></div>
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="reason" className="text-sm font-medium">Reason</label><Textarea id="reason" required value={form.reason} onChange={(e) => setForm({ ...form, reason: e.target.value })} placeholder="Why are you taking this leave?" /></div>
          <div className="flex gap-2 sm:col-span-2">
            <Button type="submit" disabled={saving}>{saving ? 'Submitting…' : 'Submit request'}</Button>
            <Button type="button" variant="outline" onClick={() => setFormOpen(false)}>Cancel</Button>
          </div>
        </form>
      )}

      {balances.length > 0 && (
        <div className="rounded-2xl border border-border bg-card p-5">
          <h2 className="text-sm font-semibold">Your balances</h2>
          <div className="mt-3 flex flex-wrap gap-3">
            {balances.map((b) => (
              <div key={`${b.employee_id}-${b.leave_type_id}`} className="rounded-xl bg-muted/50 px-4 py-3">
                <p className="text-xs text-muted-foreground">{typeName(b.leave_type_id)}</p>
                <p className="text-lg font-semibold">{(b.entitlement ?? 0) - (b.used ?? 0)} <span className="text-xs font-normal text-muted-foreground">/ {b.entitlement ?? 0} days</span></p>
              </div>
            ))}
          </div>
        </div>
      )}

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="rounded-2xl border border-border bg-card">
        <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input aria-label="Search leave requests" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search leave requests" className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
        </div>
        {error && !rows.length ? <div role="alert" className="p-8 text-sm text-destructive">{error}</div> :
        loading ? <p className="p-8 text-sm text-muted-foreground">Loading leave requests…</p> :
        rows.length === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><p className="font-medium">No leave requests match this view.</p><p className="text-sm text-muted-foreground">Use the button above to request leave.</p></div> :
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Employee</th><th className="px-5 py-3">Type</th><th className="px-5 py-3">Dates</th><th className="px-5 py-3">Days</th><th className="px-5 py-3">Status</th><th className="px-5 py-3 text-right">Actions</th></tr></thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.id} className="border-t border-border hover:bg-muted/30">
                  <td className="px-5 py-4 font-medium">{fullName(row.employee)}<p className="text-xs text-muted-foreground">{row.reason || 'No reason provided'}</p></td>
                  <td className="px-5 py-4 text-muted-foreground">{typeName(row.leave_type_id)}</td>
                  <td className="px-5 py-4">{row.start_date} → {row.end_date}</td>
                  <td className="px-5 py-4">{row.days ?? '—'}</td>
                  <td className="px-5 py-4"><Badge variant={STATUS_VARIANT[row.status ?? ''] ?? 'outline'}>{row.status?.replaceAll('_', ' ') || '—'}</Badge></td>
                  <td className="px-5 py-4 text-right">
                    {row.status === 'PENDING' ? (
                      <div className="inline-flex gap-1">
                        <Button variant="ghost" size="icon-sm" aria-label="Approve leave" onClick={() => void decide(row.id!, 'approve')}><Check /></Button>
                        <Button variant="ghost" size="icon-sm" aria-label="Reject leave" onClick={() => void decide(row.id!, 'reject')}><X /></Button>
                      </div>
                    ) : <span className="text-xs text-muted-foreground">{row.review_note || '—'}</span>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>}
      </div>
    </section>
  )
}

function fullName(employee: Leave['employee']) {
  if (!employee) return '—'
  if (typeof employee === 'string') return employee
  return `${employee.first_name ?? ''} ${employee.last_name ?? ''}`.trim() || '—'
}