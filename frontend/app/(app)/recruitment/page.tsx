'use client'

import { useCallback, useEffect, useState } from 'react'
import { CheckCircle2, Plus, RefreshCw, Search, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { api, type Application, type Candidate, type Interview, type JobPost, type Onboarding } from '@/lib/api'

const STATUS_VARIANT: Record<string, 'default' | 'warning' | 'success' | 'destructive' | 'outline' | 'primary'> = {
  OPEN: 'success', CLOSED: 'outline', DRAFT: 'warning', ON_HOLD: 'destructive',
  NEW: 'primary', SCREENING: 'warning', INTERVIEWING: 'warning', OFFERED: 'success', HIRED: 'success', REJECTED: 'destructive',
  SCHEDULED: 'primary', COMPLETED: 'success', CANCELLED: 'destructive', NO_SHOW: 'outline',
  NOT_STARTED: 'outline', IN_PROGRESS: 'warning', DONE: 'success',
}

const appStatusOptions = ['SHORTLISTED', 'INTERVIEWING', 'OFFERED', 'REJECTED']

export default function RecruitmentPage() {
  const [tab, setTab] = useState<'jobs' | 'candidates' | 'applications' | 'interviews' | 'onboarding'>('jobs')
  const [jobs, setJobs] = useState<JobPost[]>([])
  const [candidates, setCandidates] = useState<Candidate[]>([])
  const [applications, setApplications] = useState<Application[]>([])
  const [interviews, setInterviews] = useState<Interview[]>([])
  const [onboardings, setOnboardings] = useState<Onboarding[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState('')
  const [openJobForm, setOpenJobForm] = useState(false)
  const [jobForm, setJobForm] = useState({ title: '', department_id: '', description: '', requirements: '', vacancies: '', deadline: '' })
  const [hireForm, setHireForm] = useState({ application_id: '', joining_date: '' })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [js, cs, as, is, os] = await Promise.all([
        api.jobs({ page: 1, pageSize: 50, search: query }),
        api.candidates({ page: 1, pageSize: 100, search: query }),
        api.application.list({ page: 1, pageSize: 100, search: query }),
        api.interview.list({ page: 1, pageSize: 100, search: query }),
        api.onboarding.list({ page: 1, pageSize: 100, search: query }),
      ])
      setJobs(js.items ?? [])
      setCandidates(cs.items ?? [])
      setApplications(as.items ?? [])
      setInterviews(is.items ?? [])
      setOnboardings(os.items ?? [])
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load recruitment data.')
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

  const candidateName = useCallback((id?: string) => {
    if (!id) return '—'
    const hit = candidates.find((c) => c.id === id)
    return hit ? `${hit.first_name ?? ''} ${hit.last_name ?? ''}`.trim() || hit.email || '—' : id
  }, [candidates])

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium text-primary">Talent acquisition</p>
          <h1 className="text-3xl font-semibold tracking-tight">Recruitment</h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Cover job posts, candidates, applications, interviews, and onboarding.</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {tab === 'jobs' && <Button onClick={() => setOpenJobForm((v) => !v)}><Plus data-icon="inline-start" />Post a job</Button>}
        </div>
      </header>

      <div className="flex flex-wrap gap-1 rounded-xl bg-muted/60 p-1 text-sm sm:w-fit" role="tablist" aria-label="Recruitment sections">
        {(['jobs', 'candidates', 'applications', 'interviews', 'onboarding'] as const).map((key) => (
          <button key={key} role="tab" aria-selected={tab === key} onClick={() => setTab(key)} className={`rounded-lg px-3 py-2 font-medium capitalize ${tab === key ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'}`}>{key}</button>
        ))}
      </div>

      {tab === 'jobs' && openJobForm && (
        <form onSubmit={(e) => {
          e.preventDefault()
          const body = { ...jobForm, vacancies: jobForm.vacancies ? Number(jobForm.vacancies) : undefined, department_id: jobForm.department_id || undefined, deadline: jobForm.deadline || undefined, status: 'OPEN' }
          void runAction('job', api.job.create(body), 'Job posted.', () => setJobForm({ title: '', department_id: '', description: '', requirements: '', vacancies: '', deadline: '' }))
        }} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="jt" className="text-sm font-medium">Job title</label><Input id="jt" required value={jobForm.title} onChange={(e) => setJobForm({ ...jobForm, title: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="jd" className="text-sm font-medium">Department</label><Input id="jd" value={jobForm.department_id} onChange={(e) => setJobForm({ ...jobForm, department_id: e.target.value })} placeholder="Department ID" /></div>
          <div className="flex flex-col gap-2"><label htmlFor="jv" className="text-sm font-medium">Vacancies</label><Input id="jv" type="number" min={1} value={jobForm.vacancies} onChange={(e) => setJobForm({ ...jobForm, vacancies: e.target.value })} /></div>
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="jr" className="text-sm font-medium">Requirements</label><Textarea id="jr" value={jobForm.requirements} onChange={(e) => setJobForm({ ...jobForm, requirements: e.target.value })} /></div>
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="jdec" className="text-sm font-medium">Description</label><Textarea id="jdec" value={jobForm.description} onChange={(e) => setJobForm({ ...jobForm, description: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="jdl" className="text-sm font-medium">Deadline</label><Input id="jdl" type="date" value={jobForm.deadline} onChange={(e) => setJobForm({ ...jobForm, deadline: e.target.value })} /></div>
          <div className="flex items-end gap-2"><Button type="submit" disabled={busy === 'job'}>Post job</Button><Button type="button" variant="outline" onClick={() => setOpenJobForm(false)}>Cancel</Button></div>
        </form>
      )}

      {tab === 'onboarding' && (
        <form onSubmit={(e) => {
          e.preventDefault()
          if (!hireForm.application_id) return
          const app = applications.find((a) => a.id === hireForm.application_id)
          if (!app) return
          const body = { job_post_id: app.job_post_id, application_id: app.id, joining_date: hireForm.joining_date || undefined }
          void runAction('hire', api.hireCandidate(body), 'Candidate hired and employee record created.', () => setHireForm({ application_id: '', joining_date: '' }))
        }} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          <div className="flex flex-col gap-2 sm:col-span-2"><label htmlFor="happ" className="text-sm font-medium">Application</label><Select id="happ" required value={hireForm.application_id} onChange={(e) => setHireForm({ ...hireForm, application_id: e.target.value })} options={applications.map((a) => ({ label: `${candidateName(a.candidate_id)} — ${a.job_post?.title ?? 'Job'}`, value: a.id ?? '' }))} placeholder="Select an application to hire from" /></div>
          <div className="flex flex-col gap-2"><label htmlFor="hdate" className="text-sm font-medium">Joining date</label><Input id="hdate" type="date" value={hireForm.joining_date} onChange={(e) => setHireForm({ ...hireForm, joining_date: e.target.value })} /></div>
          <div className="flex items-end gap-2"><Button type="submit" disabled={busy === 'hire' || !hireForm.application_id}>{busy === 'hire' ? 'Hiring…' : 'Hire candidate'}</Button></div>
        </form>
      )}

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="rounded-2xl border border-border bg-card">
        <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input aria-label="Search recruitment" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search recruitment" className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
        </div>
        {loading ? <p className="p-8 text-sm text-muted-foreground">Loading…</p> :
        (tab === 'jobs' ? jobs.length : tab === 'candidates' ? candidates.length : tab === 'applications' ? applications.length : tab === 'interviews' ? interviews.length : onboardings.length) === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><p className="font-medium">No {tab} yet.</p><p className="text-sm text-muted-foreground">New records will appear here.</p></div> :
        tab === 'jobs' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Role</th><th className="px-5 py-3">Department</th><th className="px-5 py-3">Vacancies</th><th className="px-5 py-3">Deadline</th><th className="px-5 py-3">Status</th></tr></thead><tbody>
            {jobs.map((j) => (<tr key={j.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{j.title}</td><td className="px-5 py-4 text-muted-foreground">{typeof j.department === 'object' ? j.department?.name : j.department || '—'}</td><td className="px-5 py-4 text-muted-foreground">{j.vacancies ?? '—'}</td><td className="px-5 py-4 text-muted-foreground">{j.deadline || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[String(j.status ?? '')] ?? 'outline'}>{String(j.status ?? '—').replaceAll('_', ' ')}</Badge></td></tr>))}
          </tbody></table>
        ) : tab === 'candidates' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Candidate</th><th className="px-5 py-3">Contact</th><th className="px-5 py-3">Source</th><th className="px-5 py-3">Status</th></tr></thead><tbody>
            {candidates.map((c) => (<tr key={c.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{c.first_name} {c.last_name}<p className="text-xs text-muted-foreground">{c.education || c.experience || '—'}</p></td><td className="px-5 py-4 text-muted-foreground">{c.email || '—'}<p className="text-xs">{c.phone || ''}</p></td><td className="px-5 py-4 text-muted-foreground">{c.source || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[String(c.status ?? '')] ?? 'outline'}>{String(c.status ?? '—').replaceAll('_', ' ')}</Badge></td></tr>))}
          </tbody></table>
        ) : tab === 'applications' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Candidate</th><th className="px-5 py-3">Role</th><th className="px-5 py-3">Applied</th><th className="px-5 py-3">Status</th><th className="px-5 py-3 text-right">Update</th></tr></thead><tbody>
            {applications.map((a) => (<tr key={a.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{candidateName(a.candidate_id)}</td><td className="px-5 py-4 text-muted-foreground">{a.job_post?.title || '—'}</td><td className="px-5 py-4 text-muted-foreground">{a.applied_date || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[String(a.status ?? '')] ?? 'outline'}>{String(a.status ?? '—').replaceAll('_', ' ')}</Badge></td><td className="px-5 py-4 text-right"><Select aria-label="Update application status" value="" onChange={(ev) => { const val = String(ev.target.value); if (val && a.id) void runAction('app', api.application.updateStatus(a.id, { status: val }), `Application marked ${val.toLowerCase()}.`) }} options={appStatusOptions} placeholder="Move to…" className="h-9 w-36" /></td></tr>))}
          </tbody></table>
        ) : tab === 'interviews' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Interviewer</th><th className="px-5 py-3">Candidate</th><th className="px-5 py-3">Scheduled</th><th className="px-5 py-3">Type</th><th className="px-5 py-3">Status</th></tr></thead><tbody>
            {interviews.map((i) => (<tr key={i.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{typeof i.interviewer === 'object' && i.interviewer ? `${i.interviewer.first_name ?? ''} ${i.interviewer.last_name ?? ''}`.trim() || '—' : '—'}</td><td className="px-5 py-4 text-muted-foreground">{candidateName(i.application?.candidate_id)}</td><td className="px-5 py-4 text-muted-foreground">{i.scheduled_at ? i.scheduled_at.slice(0, 16) : '—'}</td><td className="px-5 py-4 text-muted-foreground">{i.type || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[String(i.status ?? '')] ?? 'outline'}>{String(i.status ?? '—').replaceAll('_', ' ')}</Badge></td></tr>))}
          </tbody></table>
        ) : (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Employee</th><th className="px-5 py-3">Candidate</th><th className="px-5 py-3">Started</th><th className="px-5 py-3">Status</th><th className="px-5 py-3 text-right">Action</th></tr></thead><tbody>
            {onboardings.map((o) => (<tr key={o.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{typeof o.employee === 'object' && o.employee ? `${o.employee.first_name ?? ''} ${o.employee.last_name ?? ''}`.trim() || '—' : '—'}</td><td className="px-5 py-4 text-muted-foreground">{typeof o.candidate === 'object' && o.candidate ? `${o.candidate.first_name ?? ''} ${o.candidate.last_name ?? ''}`.trim() || '—' : '—'}</td><td className="px-5 py-4 text-muted-foreground">{o.start_date || '—'}</td><td className="px-5 py-4"><Badge variant={STATUS_VARIANT[String(o.status ?? '')] ?? 'outline'}>{String(o.status ?? '—').replaceAll('_', ' ')}</Badge></td><td className="px-5 py-4 text-right">{o.status === 'PENDING' || o.status === 'IN_PROGRESS' ? <Select aria-label="Advance onboarding" value="" onChange={(ev) => { const val = String(ev.target.value); if (val && o.id) void runAction('ob', api.onboarding.update(o.id, { status: val }), 'Onboarding updated.') }} options={['IN_PROGRESS', 'COMPLETED']} placeholder="Set status…" className="h-9 w-36" /> : <span className="text-xs text-muted-foreground">{o.status}</span>}</td></tr>))}
          </tbody></table>
        )}
      </div>
    </section>
  )
}