'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { ArrowLeft, Mail, UserRound } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api, type Employee } from '@/lib/api'

export default function EmployeeDetailPage({ params }: { params: { id: string } }) {
  const [employee, setEmployee] = useState<Employee | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  useEffect(() => { api.employee.get(params.id).then(setEmployee).catch(cause => setError(cause instanceof Error ? cause.message : 'Unable to load employee.')).finally(() => setLoading(false)) }, [params.id])
  return <section className="flex flex-col gap-6"><Link href="/employees" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"><ArrowLeft data-icon="inline-start" />Back to employees</Link>{error && <p role="alert" className="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">{error}</p>}{loading ? <p className="text-sm text-muted-foreground">Loading employee profile…</p> : employee && <><header className="flex flex-col gap-4 rounded-2xl border border-border bg-card p-6 sm:flex-row sm:items-center"><div className="grid size-16 place-items-center rounded-2xl bg-primary/10 text-primary"><UserRound /></div><div className="flex flex-col gap-1"><p className="text-sm text-muted-foreground">{employee.employee_code || 'Employee profile'}</p><h1 className="text-3xl font-semibold tracking-tight">{employee.first_name} {employee.last_name}</h1><p className="text-sm text-muted-foreground">{employee.email || 'No work email recorded'}</p></div></header><div className="grid gap-4 md:grid-cols-2"><div className="rounded-2xl border border-border bg-card p-6"><h2 className="font-semibold">Work details</h2><dl className="mt-4 flex flex-col gap-4 text-sm"><div className="flex justify-between gap-4"><dt className="text-muted-foreground">Department</dt><dd>{typeof employee.department === 'string' ? employee.department : employee.department?.name || '—'}</dd></div><div className="flex justify-between gap-4"><dt className="text-muted-foreground">Designation</dt><dd>{typeof employee.designation === 'string' ? employee.designation : employee.designation?.name || '—'}</dd></div><div className="flex justify-between gap-4"><dt className="text-muted-foreground">Status</dt><dd>{employee.status || employee.employment_status || '—'}</dd></div></dl></div><div className="rounded-2xl border border-border bg-card p-6"><h2 className="font-semibold">Contact</h2><a href={employee.email ? `mailto:${employee.email}` : undefined} className="mt-4 flex items-center gap-2 text-sm text-primary"><Mail />{employee.email || 'No email available'}</a><Button className="mt-6" variant="outline" asChild><Link href="/employees">Return to directory</Link></Button></div></div></>}</section>
}
