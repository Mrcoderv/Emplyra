'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import CoreModulePage from '@/components/core-module-page'
import { api } from '@/lib/api'

export default function LeavesPage() {
  const [open, setOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')
  const [form, setForm] = useState({ leave_type_id: '', start_date: '', end_date: '', reason: '' })

  async function submit(event: React.FormEvent) {
    event.preventDefault(); setSaving(true); setMessage('')
    try { await api.createLeave(form); setMessage('Leave request submitted successfully.'); setOpen(false); setForm({ leave_type_id: '', start_date: '', end_date: '', reason: '' }) }
    catch (cause) { setMessage(cause instanceof Error ? cause.message : 'Unable to submit leave request.') }
    finally { setSaving(false) }
  }

  return <div className="flex flex-col gap-6"><div className="flex justify-end"><Button onClick={() => setOpen(true)}>Request leave</Button></div>{message && <p role="status" className="rounded-xl border border-border bg-card p-4 text-sm">{message}</p>}{open && <form onSubmit={submit} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2"><div className="flex flex-col gap-2"><label htmlFor="leave_type_id" className="text-sm font-medium">Leave type ID</label><input required id="leave_type_id" value={form.leave_type_id} onChange={e => setForm({ ...form, leave_type_id: e.target.value })} className="h-10 rounded-xl border border-input bg-background px-3 text-sm" placeholder="Provided by your HR team" /></div><div className="flex flex-col gap-2"><label htmlFor="start_date" className="text-sm font-medium">Start date</label><input required type="date" id="start_date" value={form.start_date} onChange={e => setForm({ ...form, start_date: e.target.value })} className="h-10 rounded-xl border border-input bg-background px-3 text-sm" /></div><div className="flex flex-col gap-2"><label htmlFor="end_date" className="text-sm font-medium">End date</label><input required type="date" id="end_date" value={form.end_date} onChange={e => setForm({ ...form, end_date: e.target.value })} className="h-10 rounded-xl border border-input bg-background px-3 text-sm" /></div><div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="reason" className="text-sm font-medium">Reason</label><textarea required id="reason" value={form.reason} onChange={e => setForm({ ...form, reason: e.target.value })} className="min-h-24 rounded-xl border border-input bg-background p-3 text-sm" /></div><div className="flex gap-2 sm:col-span-2"><Button type="submit" disabled={saving}>{saving ? 'Submitting…' : 'Submit request'}</Button><Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button></div></form>}<CoreModulePage module="leaves" /></div>
}
