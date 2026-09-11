'use client'

import { useCallback, useEffect, useState } from 'react'
import { CheckCircle2, Plus, RefreshCw, Search, Target, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { api, type Goal, type KPI, type PerformanceReview } from '@/lib/api'

const STATUS_VARIANT: Record<string, 'default' | 'primary' | 'warning' | 'success' | 'destructive' | 'outline'> = {
  DRAFT: 'outline', ACTIVE: 'success', ON_TRACK: 'success', AT_RISK: 'warning', COMPLETED: 'primary', ARCHIVED: 'outline',
  PENDING: 'warning', SUBMITTED: 'primary', APPROVED: 'success',
}

export default function PerformancePage() {
  const [tab, setTab] = useState<'goals' | 'kpis' | 'reviews'>('goals')
  const [goals, setGoals] = useState<Goal[]>([])
  const [kpis, setKpis] = useState<KPI[]>([])
  const [reviews, setReviews] = useState<PerformanceReview[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState('')
  const [openForm, setOpenForm] = useState(false)
  const [goalForm, setGoalForm] = useState({ title: '', description: '', target_date: '', weight: '' })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [gs, ks, rs] = await Promise.all([
        api.goals({ page: 1, pageSize: 50, search: query }),
        api.kpis({ page: 1, pageSize: 50, search: query }),
        api.review.list({ page: 1, pageSize: 50, search: query }),
      ])
      setGoals(gs.items ?? [])
      setKpis(ks.items ?? [])
      setReviews(rs.items ?? [])
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load performance data.')
    } finally {
      setLoading(false)
    }
  }, [query])

  useEffect(() => { void load() }, [load])

  async function runAction(name: string, promise: Promise<unknown>, successMessage: string, done?: () => void) {
    setBusy(name); setError(''); setNotice('')
    try { await promise; setNotice(successMessage); done?.(); await load() }
    catch (cause) { setError(cause instanceof Error ? cause.message : `${name} failed.`) }
    finally { setBusy('') }
  }

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium text-primary">People programs</p>
          <h1 className="text-3xl font-semibold tracking-tight">Performance</h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Set goals, measure with KPIs, and run structured reviews.</p>
        </div>
        <Button onClick={() => setOpenForm((v) => !v)} disabled={tab !== 'goals'}><Plus data-icon="inline-start" />New goal</Button>
      </header>

      <div className="flex gap-1 rounded-xl bg-muted/60 p-1 text-sm sm:w-fit" role="tablist" aria-label="Performance sections">
        {(['goals', 'kpis', 'reviews'] as const).map((key) => (
          <button key={key} role="tab" aria-selected={tab === key} onClick={() => setTab(key)} className={`rounded-lg px-4 py-2 font-medium capitalize ${tab === key ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'}`}>{key}</button>
        ))}
      </div>

      {tab === 'goals' && openForm && (
        <form onSubmit={(e) => {
          e.preventDefault()
          void runAction('goal', api.goal.create({ ...goalForm, weight: goalForm.weight ? Number(goalForm.weight) : undefined, status: 'IN_PROGRESS' }), 'Goal created.', () => { setGoalForm({ title: '', description: '', target_date: '', weight: '' }); setOpenForm(false) })
        }} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="gt" className="text-sm font-medium">Goal title</label><Input id="gt" required value={goalForm.title} onChange={(e) => setGoalForm({ ...goalForm, title: e.target.value })} /></div>
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="gd" className="text-sm font-medium">Description</label><Textarea id="gd" value={goalForm.description} onChange={(e) => setGoalForm({ ...goalForm, description: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="gtd" className="text-sm font-medium">Target date</label><Input id="gtd" type="date" value={goalForm.target_date} onChange={(e) => setGoalForm({ ...goalForm, target_date: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="gw" className="text-sm font-medium">Weight</label><Input id="gw" type="number" min={0} value={goalForm.weight} onChange={(e) => setGoalForm({ ...goalForm, weight: e.target.value })} placeholder="e.g. 1" /></div>
          <div className="flex items-end gap-2"><Button type="submit" disabled={busy === 'goal'}>Create goal</Button><Button type="button" variant="outline" onClick={() => setOpenForm(false)}>Cancel</Button></div>
        </form>
      )}

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="rounded-2xl border border-border bg-card">
        <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input aria-label="Search performance" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search performance" className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
        </div>
        {loading ? <p className="p-8 text-sm text-muted-foreground">Loading…</p> :
        (tab === 'goals' ? goals.length : tab === 'kpis' ? kpis.length : reviews.length) === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><Target className="text-muted-foreground" /><p className="font-medium">No {tab} yet.</p><p className="text-sm text-muted-foreground">{tab === 'goals' ? 'Create a goal to get started.' : 'New records will appear here.'}</p></div> :
        tab === 'goals' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Goal</th><th className="px-5 py-3">Owner</th><th className="px-5 py-3">Weight</th><th className="px-5 py-3">Target</th><th className="px-5 py-3">Status</th></tr></thead><tbody>
            {goals.map((g) => (<tr key={g.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{g.title}<p className="text-xs text-muted-foreground">{g.description}</p></td><td className="px-5 py-4 text-muted-foreground">{typeof g.employee === 'object' && g.employee ? `${g.employee.first_name ?? ''} ${g.employee.last_name ?? ''}`.trim() || '—' : '—'}</td><td className="px-5 py-4 text-muted-foreground">{g.weight ?? '—'}</td><td className="px-5 py-4 text-muted-foreground">{g.target_date || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[String(g.status ?? '')] ?? 'outline'}>{String(g.status ?? '—').replaceAll('_', ' ')}</Badge></td></tr>))}
          </tbody></table>
        ) : tab === 'kpis' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">KPI</th><th className="px-5 py-3">Actual</th><th className="px-5 py-3">Target</th><th className="px-5 py-3">Unit</th><th className="px-5 py-3">Owner</th></tr></thead><tbody>
            {kpis.map((k) => (<tr key={k.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{k.name}<p className="text-xs text-muted-foreground">{k.description}</p></td><td className="px-5 py-4 text-muted-foreground">{k.actual ?? '—'}</td><td className="px-5 py-4 text-muted-foreground">{k.target ?? '—'}</td><td className="px-5 py-4 text-muted-foreground">{k.unit || '—'}</td><td className="px-5 py-4 text-muted-foreground">{typeof k.employee === 'object' && k.employee ? `${k.employee.first_name ?? ''} ${k.employee.last_name ?? ''}`.trim() || '—' : '—'}</td></tr>))}
          </tbody></table>
        ) : (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Employee</th><th className="px-5 py-3">Period</th><th className="px-5 py-3">Status</th><th className="px-5 py-3">Score</th><th className="px-5 py-3 text-right">Actions</th></tr></thead><tbody>
            {reviews.map((r) => (<tr key={r.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{typeof r.employee === 'object' && r.employee ? `${r.employee.first_name ?? ''} ${r.employee.last_name ?? ''}`.trim() || '—' : '—'}</td><td className="px-5 py-4 text-muted-foreground">{r.period || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[String(r.status ?? '')] ?? 'outline'}>{String(r.status ?? '—').replaceAll('_', ' ')}</Badge></td><td className="px-5 py-4 text-muted-foreground">{r.score ?? '—'}</td><td className="px-5 py-4 text-right">{['PENDING'].includes(String(r.status)) && <Button variant="ghost" size="sm" disabled={busy === 'submit'} onClick={() => r.id && void runAction('submit', api.review.submit(r.id, { status: 'SELF_SUBMITTED' }), 'Review submitted.')}>Submit</Button>}</td></tr>))}
          </tbody></table>
        )}
      </div>
    </section>
  )
}