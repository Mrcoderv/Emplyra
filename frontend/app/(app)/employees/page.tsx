'use client'

import Link from 'next/link'
import { useCallback, useEffect, useState } from 'react'
import { Plus, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api, type Employee } from '@/lib/api'

export default function EmployeesPage() {
  const [rows, setRows] = useState<Employee[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({ first_name: '', last_name: '', email: '', employee_code: '' })

  const load = useCallback(async () => {
    setLoading(true); setError('')
    try { const result = await api.employees(1, 50, query); setRows(result.items ?? []) }
    catch (cause) { setError(cause instanceof Error ? cause.message : 'Unable to load employees.') }
    finally { setLoading(false) }
  }, [query])
  useEffect(() => { void load() }, [load])

  async function createEmployee(event: React.FormEvent) {
    event.preventDefault(); setSaving(true); setError('')
    try { await api.employee.create(form); setForm({ first_name: '', last_name: '', email: '', employee_code: '' }); setFormOpen(false); await load() }
    catch (cause) { setError(cause instanceof Error ? cause.message : 'Unable to create employee.') }
    finally { setSaving(false) }
  }

  return <section className="flex flex-col gap-6"><header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between"><div className="flex flex-col gap-2"><p className="text-sm font-medium text-primary">Directory</p><h1 className="text-3xl font-semibold tracking-tight">Employees</h1><p className="text-sm text-muted-foreground">Find people, roles, and reporting context.</p></div><Button onClick={() => setFormOpen(true)}><Plus data-icon="inline-start" />Add employee</Button></header>
    {formOpen && <form onSubmit={createEmployee} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2"><div className="flex flex-col gap-2"><label htmlFor="first_name" className="text-sm font-medium">First name</label><input required id="first_name" value={form.first_name} onChange={e => setForm({ ...form, first_name: e.target.value })} className="h-10 rounded-xl border border-input bg-background px-3 text-sm" /></div><div className="flex flex-col gap-2"><label htmlFor="last_name" className="text-sm font-medium">Last name</label><input required id="last_name" value={form.last_name} onChange={e => setForm({ ...form, last_name: e.target.value })} className="h-10 rounded-xl border border-input bg-background px-3 text-sm" /></div><div className="flex flex-col gap-2"><label htmlFor="email" className="text-sm font-medium">Work email</label><input required type="email" id="email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} className="h-10 rounded-xl border border-input bg-background px-3 text-sm" /></div><div className="flex flex-col gap-2"><label htmlFor="employee_code" className="text-sm font-medium">Employee code</label><input id="employee_code" value={form.employee_code} onChange={e => setForm({ ...form, employee_code: e.target.value })} className="h-10 rounded-xl border border-input bg-background px-3 text-sm" /></div><div className="flex gap-2 sm:col-span-2"><Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save employee'}</Button><Button type="button" variant="outline" onClick={() => setFormOpen(false)}>Cancel</Button></div></form>}
    <div className="rounded-2xl border border-border bg-card"><div className="border-b border-border p-5"><div className="relative max-w-sm"><Search className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" /><input aria-label="Search employees" value={query} onChange={e => setQuery(e.target.value)} placeholder="Search employees" className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" /></div></div>{error && <p role="alert" className="p-5 text-sm text-destructive">{error}</p>}{loading ? <p className="p-8 text-sm text-muted-foreground">Loading employee records…</p> : !error && !rows.length ? <p className="p-8 text-sm text-muted-foreground">No employees found.</p> : !error && <div className="overflow-x-auto"><table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Employee</th><th className="px-5 py-3">Department</th><th className="px-5 py-3">Status</th></tr></thead><tbody>{rows.map(e => <tr key={e.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4"><Link className="font-medium hover:text-primary" href={`/employees/${e.id}`}>{e.first_name} {e.last_name}</Link><p className="text-xs text-muted-foreground">{e.email || e.employee_code || 'Employee record'}</p></td><td className="px-5 py-4 text-muted-foreground">{typeof e.department === 'string' ? e.department : e.department?.name || '—'}</td><td className="px-5 py-4 text-muted-foreground">{e.status || e.employment_status || '—'}</td></tr>)}</tbody></table></div>}</div></section>
}
