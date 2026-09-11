'use client'

import { useCallback, useEffect, useState } from 'react'
import { CheckCircle2, Settings2, Trash2, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { api, type Designation, type Holiday, type Role, type User } from '@/lib/api'

const TABS = [
  { key: 'users', label: 'Users', description: 'Who can sign in and their roles.' },
  { key: 'roles', label: 'Roles & permissions', description: 'Role-based access control.' },
  { key: 'holidays', label: 'Holidays', description: 'Company-wide days off.' },
  { key: 'designations', label: 'Designations', description: 'Job titles used across the org.' },
] as const

type TabKey = typeof TABS[number]['key']

const inputClass = 'h-10 w-full rounded-xl border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring'

export default function SettingsPage() {
  const [tab, setTab] = useState<TabKey>('users')
  const [users, setUsers] = useState<User[]>([])
  const [roles, setRoles] = useState<Role[]>([])
  const [holidays, setHolidays] = useState<Holiday[]>([])
  const [designations, setDesignations] = useState<Designation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState('')
  const [newUser, setNewUser] = useState({ first_name: '', last_name: '', email: '', password: '', role: '' })
  const [newHoliday, setNewHoliday] = useState({ name: '', date: '', type: '' })
  const [newDesignation, setNewDesignation] = useState({ name: '', description: '' })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [us, rs, hs, ds] = await Promise.all([
        api.users({ page: 1, pageSize: 100 }),
        api.roles().catch(() => []),
        api.holidays({ page: 1, pageSize: 200 }),
        api.designations({ page: 1, pageSize: 200 }),
      ])
      setUsers(us.items ?? [])
      setRoles(rs ?? [])
      setHolidays(hs.items ?? [])
      setDesignations(ds.items ?? [])
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load workspace settings.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  async function runAction(name: string, promise: Promise<unknown>, successMessage: string, done?: () => void) {
    setBusy(name); setError(''); setNotice('')
    try { await promise; setNotice(successMessage); done?.(); await load() }
    catch (cause) { setError(cause instanceof Error ? cause.message : `${name} failed.`) }
    finally { setBusy('') }
  }

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-2">
        <p className="text-sm font-medium text-primary">Workspace preferences</p>
        <h1 className="text-3xl font-semibold tracking-tight">Settings</h1>
        <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Manage access, roles, holidays, and job designations.</p>
      </header>

      <div className="flex flex-wrap gap-1 rounded-xl bg-muted/60 p-1 text-sm sm:w-fit" role="tablist" aria-label="Settings sections">
        {TABS.map((t) => (
          <button key={t.key} role="tab" aria-selected={tab === t.key} onClick={() => setTab(t.key)} className={`rounded-lg px-4 py-2 font-medium ${tab === t.key ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'}`}>{t.label}</button>
        ))}
      </div>

      {tab === 'users' && (
        <form onSubmit={(e) => {
          e.preventDefault()
          if (!newUser.email || !newUser.password) return
          void runAction('user', api.user.create({ ...newUser, role: newUser.role || undefined }), 'User invited.', () => setNewUser({ first_name: '', last_name: '', email: '', password: '', role: '' }))
        }} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          <div className="flex flex-col gap-2 sm:col-span-2"><p className="text-sm font-semibold">Invite a user</p></div>
          <div className="flex flex-col gap-2"><label htmlFor="uf" className="text-sm font-medium">First name</label><Input id="uf" value={newUser.first_name} onChange={(e) => setNewUser({ ...newUser, first_name: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="ul" className="text-sm font-medium">Last name</label><Input id="ul" value={newUser.last_name} onChange={(e) => setNewUser({ ...newUser, last_name: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="ue" className="text-sm font-medium">Email</label><Input id="ue" type="email" required value={newUser.email} onChange={(e) => setNewUser({ ...newUser, email: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="up" className="text-sm font-medium">Temporary password</label><Input id="up" type="password" required value={newUser.password} onChange={(e) => setNewUser({ ...newUser, password: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="ur" className="text-sm font-medium">Role</label><Select id="ur" value={newUser.role} onChange={(e) => setNewUser({ ...newUser, role: e.target.value })} options={roles.map((r) => ({ label: r.name ?? '', value: r.id ?? '' }))} placeholder="Default role" /></div>
          <div className="flex items-end gap-2"><Button type="submit" disabled={busy === 'user'}>Invite</Button></div>
        </form>
      )}

      {tab === 'holidays' && (
        <form onSubmit={(e) => {
          e.preventDefault()
          if (!newHoliday.name || !newHoliday.date) return
          void runAction('holiday', api.holiday.create(newHoliday), 'Holiday added.', () => setNewHoliday({ name: '', date: '', type: '' }))
        }} className="flex flex-wrap items-end gap-3 rounded-2xl border border-border bg-card p-5">
          <div className="flex flex-col gap-2"><label htmlFor="hn" className="text-sm font-medium">Holiday</label><Input id="hn" required value={newHoliday.name} onChange={(e) => setNewHoliday({ ...newHoliday, name: e.target.value })} placeholder="e.g. Dashain" /></div>
          <div className="flex flex-col gap-2"><label htmlFor="hd" className="text-sm font-medium">Date</label><Input id="hd" type="date" required value={newHoliday.date} onChange={(e) => setNewHoliday({ ...newHoliday, date: e.target.value })} /></div>
          <div className="flex flex-col gap-2"><label htmlFor="ht" className="text-sm font-medium">Type</label><Select id="ht" value={newHoliday.type} onChange={(e) => setNewHoliday({ ...newHoliday, type: e.target.value })} options={['PUBLIC', 'COMPANY', 'OPTIONAL']} placeholder="Type" /></div>
          <Button type="submit" disabled={busy === 'holiday'}>Add holiday</Button>
        </form>
      )}

      {tab === 'designations' && (
        <form onSubmit={(e) => {
          e.preventDefault()
          if (!newDesignation.name) return
          void runAction('designation', api.designation.create(newDesignation), 'Designation added.', () => setNewDesignation({ name: '', description: '' }))
        }} className="flex flex-wrap items-end gap-3 rounded-2xl border border-border bg-card p-5">
          <div className="flex flex-col gap-2"><label htmlFor="dn" className="text-sm font-medium">Name</label><Input id="dn" required value={newDesignation.name} onChange={(e) => setNewDesignation({ ...newDesignation, name: e.target.value })} placeholder="e.g. Senior Engineer" /></div>
          <div className="flex flex-col gap-2"><label htmlFor="dd" className="text-sm font-medium">Description</label><Input id="dd" value={newDesignation.description} onChange={(e) => setNewDesignation({ ...newDesignation, description: e.target.value })} placeholder="Optional" className={inputClass} /></div>
          <Button type="submit" disabled={busy === 'designation'}>Add designation</Button>
        </form>
      )}

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="overflow-hidden rounded-2xl border border-border bg-card">
        {loading ? <p className="p-8 text-sm text-muted-foreground">Loading {TABS.find((t) => t.key === tab)?.label.toLowerCase()}…</p> :
        tab === 'users' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">User</th><th className="px-5 py-3">Email</th><th className="px-5 py-3">Roles</th><th className="px-5 py-3 text-right">Actions</th></tr></thead><tbody>
            {users.length === 0 ? <tr><td colSpan={4} className="px-5 py-8 text-center text-sm text-muted-foreground">No users yet.</td></tr> : users.map((u) => (
              <tr key={u.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{u.first_name ?? ''} {u.last_name ?? ''}</td><td className="px-5 py-4 text-muted-foreground">{u.email}</td><td className="px-5 py-4">{(u.role ? [u.role] : []).map((r) => <Badge key={r} variant="outline" className="mr-1">{r}</Badge>)}</td><td className="px-5 py-4 text-right"><Button variant="ghost" size="icon-sm" aria-label="Remove user" onClick={() => { if (window.confirm(`Remove ${u.email}?`)) void runAction('del-user', api.user.remove(u.id!), 'User removed.') }}><Trash2 /></Button></td></tr>
            ))}
          </tbody></table>
        ) : tab === 'roles' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Role</th><th className="px-5 py-3">Description</th><th className="px-5 py-3">Permissions</th></tr></thead><tbody>
            {roles.length === 0 ? <tr><td colSpan={3} className="px-5 py-8 text-center text-sm text-muted-foreground">No roles found.</td></tr> : roles.map((r) => (
              <tr key={r.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium"><p className="flex items-center gap-2"><Settings2 className="text-muted-foreground" />{r.name}</p></td><td className="px-5 py-4 text-muted-foreground">{r.description || '—'}</td><td className="px-5 py-4"><div className="flex flex-wrap gap-1">{(r.permissions ?? []).slice(0, 8).map((p) => <Badge key={typeof p === 'string' ? p : p.id} variant="outline">{typeof p === 'string' ? p : p.name}</Badge>)}{(r.permissions?.length ?? 0) > 8 && <Badge variant="outline">+{r.permissions!.length - 8}</Badge>}</div></td></tr>
            ))}
          </tbody></table>
        ) : tab === 'holidays' ? (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Holiday</th><th className="px-5 py-3">Date</th><th className="px-5 py-3">Type</th><th className="px-5 py-3 text-right">Actions</th></tr></thead><tbody>
            {holidays.length === 0 ? <tr><td colSpan={4} className="px-5 py-8 text-center text-sm text-muted-foreground">No holidays yet.</td></tr> : holidays.map((h) => (
              <tr key={h.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{h.name}</td><td className="px-5 py-4 text-muted-foreground">{h.date || '—'}</td><td className="px-5 py-4 text-muted-foreground">{h.type || '—'}</td><td className="px-5 py-4 text-right"><Button variant="ghost" size="icon-sm" aria-label="Delete holiday" onClick={() => { if (window.confirm(`Delete ${h.name}?`)) void runAction('del-holiday', api.holiday.remove(h.id!), 'Holiday deleted.') }}><Trash2 /></Button></td></tr>
            ))}
          </tbody></table>
        ) : (
          <table className="w-full text-left text-sm"><thead className="bg-muted/50 text-xs uppercase text-muted-foreground"><tr><th className="px-5 py-3">Designation</th><th className="px-5 py-3">Description</th><th className="px-5 py-3 text-right">Actions</th></tr></thead><tbody>
            {designations.length === 0 ? <tr><td colSpan={3} className="px-5 py-8 text-center text-sm text-muted-foreground">No designations yet.</td></tr> : designations.map((d) => (
              <tr key={d.id} className="border-t border-border hover:bg-muted/30"><td className="px-5 py-4 font-medium">{d.name}</td><td className="px-5 py-4 text-muted-foreground">{d.description || '—'}</td><td className="px-5 py-4 text-right"><Button variant="ghost" size="icon-sm" aria-label="Delete designation" onClick={() => { if (window.confirm(`Delete ${d.name}?`)) void runAction('del-designation', api.designation.remove(d.id!), 'Designation deleted.') }}><Trash2 /></Button></td></tr>
            ))}
          </tbody></table>
        )}
      </div>
    </section>
  )
}