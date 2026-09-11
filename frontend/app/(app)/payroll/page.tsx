'use client'

import { useCallback, useEffect, useState } from 'react'
import { Banknote, CheckCircle2, Pencil, Plus, RefreshCw, Search, Trash2, XCircle, Zap } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { api, type Payroll, type SalaryStructure } from '@/lib/api'

const STATUS_VARIANT: Record<string, 'default' | 'primary' | 'warning' | 'success' | 'destructive' | 'outline'> = {
  DRAFT: 'outline', GENERATED: 'primary', PROCESSING: 'warning', PAID: 'success', CANCELLED: 'destructive',
}

export default function PayrollPage() {
  const [tab, setTab] = useState<'payroll' | 'structures'>('payroll')
  const [payrollRows, setPayrollRows] = useState<Payroll[]>([])
  const [salaryRows, setSalaryRows] = useState<SalaryStructure[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState('')
  const [runForm, setRunForm] = useState({ month: '', year: String(new Date().getFullYear()) })
  const [openSalaryForm, setOpenSalaryForm] = useState(false)
  const [salaryForm, setSalaryForm] = useState({ employee_id: '', basic_salary: '', allowances: '', bonus: '', deductions: '', effective_from: '' })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [payrolls, salaries] = await Promise.all([
        api.payroll({ page: 1, pageSize: 50, search: query }),
        api.salaries({ search: query }),
      ])
      setPayrollRows(payrolls.items ?? [])
      setSalaryRows(salaries.items ?? [])
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load payroll records.')
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
          <p className="text-sm font-medium text-primary">Compensation</p>
          <h1 className="text-3xl font-semibold tracking-tight">{tab === 'payroll' ? 'Payroll runs' : 'Salary structures'}</h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Generate pay runs, process payments, and manage base salaries.</p>
        </div>
        {tab === 'payroll' ? (
          <div className="flex flex-wrap items-end gap-2">
            <div className="flex flex-col gap-1"><label htmlFor="month" className="text-xs text-muted-foreground">Month</label><Select id="month" value={runForm.month} onChange={(e) => setRunForm({ ...runForm, month: e.target.value })} options={[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12].map((m) => ({ label: new Date(2000, m - 1).toLocaleString(undefined, { month: 'long' }), value: String(m) }))} placeholder="Period" /></div>
            <div className="flex flex-col gap-1"><label htmlFor="year" className="text-xs text-muted-foreground">Year</label><Input id="year" aria-label="Year" value={runForm.year} onChange={(e) => setRunForm({ ...runForm, year: e.target.value })} className="h-10 w-24 rounded-xl border border-input bg-background px-3 text-sm" /></div>
            <Button disabled={busy === 'generate' || !runForm.month} onClick={() => runForm.month && void runAction('generate', api.payrollGenerate({ month: Number(runForm.month), year: Number(runForm.year) || new Date().getFullYear() }), 'Payroll run generated.')}><Zap data-icon="inline-start" />{busy === 'generate' ? 'Generating…' : 'Generate run'}</Button>
          </div>
        ) : <Button onClick={() => setOpenSalaryForm((v) => !v)}><Plus data-icon="inline-start" />Add structure</Button>}
      </header>

      <div className="flex gap-1 rounded-xl bg-muted/60 p-1 text-sm sm:w-fit" role="tablist" aria-label="Payroll sections">
        <button role="tab" aria-selected={tab === 'payroll'} onClick={() => setTab('payroll')} className={`rounded-lg px-4 py-2 font-medium ${tab === 'payroll' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'}`}>Payroll runs</button>
        <button role="tab" aria-selected={tab === 'structures'} onClick={() => setTab('structures')} className={`rounded-lg px-4 py-2 font-medium ${tab === 'structures' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'}`}>Salary structures</button>
      </div>

      {tab === 'structures' && openSalaryForm && (
        <form onSubmit={async (e) => {
          e.preventDefault()
          await runAction('salary', api.salary.create(salaryForm), 'Salary structure saved.')
          setSalaryForm({ employee_id: '', basic_salary: '', allowances: '', bonus: '', deductions: '', effective_from: '' })
          setOpenSalaryForm(false)
        }} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="emp" className="text-sm font-medium">Employee</label><Input id="emp" onChange={(e) => setSalaryForm({ ...salaryForm, employee_id: e.target.value })} placeholder="Employee ID" /></div>
          <div className="flex flex-col gap-2"><label htmlFor="basic" className="text-sm font-medium">Basic salary</label><Input id="basic" type="number" min={0} onChange={(e) => setSalaryForm({ ...salaryForm, basic_salary: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="allow" className="text-sm font-medium">Allowances</label><Input id="allow" type="number" min={0} defaultValue="0" onChange={(e) => setSalaryForm({ ...salaryForm, allowances: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="bonus" className="text-sm font-medium">Bonus</label><Input id="bonus" type="number" min={0} defaultValue="0" onChange={(e) => setSalaryForm({ ...salaryForm, bonus: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="deduc" className="text-sm font-medium">Deductions</label><Input id="deduc" type="number" min={0} defaultValue="0" onChange={(e) => setSalaryForm({ ...salaryForm, deductions: e.target.value })} /></div>
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="eff" className="text-sm font-medium">Effective from</label><Input id="eff" type="date" onChange={(e) => setSalaryForm({ ...salaryForm, effective_from: e.target.value })} /></div>
          <div className="flex gap-2 sm:col-span-2"><Button type="submit" disabled={busy === 'salary'}>Save structure</Button><Button type="button" variant="outline" onClick={() => setOpenSalaryForm(false)}>Cancel</Button></div>
        </form>
      )}

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="rounded-2xl border border-border bg-card">
        <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input aria-label="Search payroll" value={query} onChange={(e) => setQuery(e.target.value)} placeholder={`Search ${tab === 'payroll' ? 'payroll runs' : 'salary structures'}`} className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
        </div>
        {loading ? <p className="p-8 text-sm text-muted-foreground">Loading…</p> :
        tab === 'payroll' && payrollRows.length === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><Banknote className="text-muted-foreground" /><p className="font-medium">No payroll runs yet.</p><p className="text-sm text-muted-foreground">Pick a month and year, then generate your first run.</p></div> :
        tab === 'structures' && salaryRows.length === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><p className="font-medium">No salary structures yet.</p><p className="text-sm text-muted-foreground">Add a structure to start compensating employees.</p></div> :
        tab === 'payroll' ? (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Period</th><th className="px-5 py-3">Employee</th><th className="px-5 py-3">Status</th><th className="px-5 py-3">Net</th><th className="px-5 py-3 text-right">Actions</th></tr></thead>
              <tbody>
                {payrollRows.map((row) => (
                  <tr key={row.id} className="border-t border-border hover:bg-muted/30">
                    <td className="px-5 py-4 font-medium">{row.month != null && row.year != null ? `${new Date(2000, Number(row.month) - 1).toLocaleString(undefined, { month: 'long' })} ${row.year}` : '—'}</td>
                    <td className="px-5 py-4">{empName(row.employee)}</td>
                    <td className="px-5 py-4"><Badge variant={STATUS_VARIANT[row.status ?? ''] ?? 'outline'}>{String(row.status ?? '—').replaceAll('_', ' ')}</Badge></td>
                    <td className="px-5 py-4 text-muted-foreground">{amount(row.net_salary)}</td>
                    <td className="px-5 py-4 text-right">
                      <div className="inline-flex gap-1">
                        {['DRAFT', 'GENERATED'].includes(String(row.status)) && <Button variant="ghost" size="sm" disabled={busy === 'process'} onClick={() => void runAction('process', api.payrollProcess(row.id!), 'Payroll processing started.')}>Process</Button>}
                        {['GENERATED', 'PROCESSING'].includes(String(row.status)) && <Button variant="ghost" size="sm" disabled={busy === 'paid'} onClick={() => void runAction('paid', api.payrollMarkPaid(row.id!), 'Payroll marked as paid.')}>Mark paid</Button>}
                        {!['PAID', 'CANCELLED'].includes(String(row.status)) && <Button variant="ghost" size="sm" disabled={busy === 'cancel'} onClick={() => { if (window.confirm(`Cancel payroll for ${row.month}/${row.year}?`)) void runAction('cancel', api.payrollCancel(row.id!), 'Payroll cancelled.') }}>Cancel</Button>}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Employee</th><th className="px-5 py-3">Basic</th><th className="px-5 py-3">Allowances</th><th className="px-5 py-3">Bonuses</th><th className="px-5 py-3">Deductions</th><th className="px-5 py-3">Effective</th><th className="px-5 py-3 text-right">Actions</th></tr></thead>
              <tbody>
                {salaryRows.map((row) => (
                  <tr key={row.id} className="border-t border-border hover:bg-muted/30">
                    <td className="px-5 py-4 font-medium">{empName(row.employee)}</td>
                    <td className="px-5 py-4 text-muted-foreground">{amount(row.basic_salary)}</td>
                    <td className="px-5 py-4 text-muted-foreground">{amount(row.allowances)}</td>
                    <td className="px-5 py-4 text-muted-foreground">{amount(row.bonus)}</td>
                    <td className="px-5 py-4 text-muted-foreground">{amount(row.deductions)}</td>
                    <td className="px-5 py-4 text-muted-foreground">{row.effective_from?.slice(0, 10) || '—'}</td>
                    <td className="px-5 py-4 text-right">
                      <div className="inline-flex gap-1">
                        <Button variant="ghost" size="icon-sm" aria-label="Edit structure"><Pencil /></Button>
                        <Button variant="ghost" size="icon-sm" aria-label="Delete structure" onClick={() => { if (window.confirm('Delete this salary structure?')) void runAction('delete', api.salary.remove(row.id!), 'Salary structure deleted.') }}><Trash2 /></Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  )
}

function amount(value: unknown) {
  if (value === null || value === undefined || value === '') return '—'
  const n = Number(value)
  if (Number.isNaN(n)) return String(value)
  return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(n)
}

function empName(employee: SalaryStructure['employee']) {
  if (!employee) return '—'
  if (typeof employee === 'string') return employee
  return `${employee.first_name ?? ''} ${employee.last_name ?? ''}`.trim() || employee.employee_code || '—'
}