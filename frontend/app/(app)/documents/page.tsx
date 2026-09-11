'use client'

import Link from 'next/link'
import { useCallback, useEffect, useRef, useState } from 'react'
import { CheckCircle2, Download, FileText, RefreshCw, Search, Trash2, Upload, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { api, type Document, type Employee } from '@/lib/api'

const darknikVariant: Record<string, 'default' | 'success' | 'warning' | 'destructive' | 'outline'> = {
  ACTIVE: 'success', ARCHIVED: 'outline', EXPIRED: 'warning', REVOKED: 'destructive',
}

export default function DocumentsPage() {
  const fileRef = useRef<HTMLInputElement>(null)
  const [rows, setRows] = useState<Document[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [uploading, setUploading] = useState(false)
  const [deletingId, setDeletingId] = useState('')
  const [employees, setEmployees] = useState<Employee[]>([])
  const [ownerId, setOwnerId] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [docs, emps] = await Promise.all([
        api.documents({ page: 1, pageSize: 60, search: query }),
        api.employees({ page: 1, pageSize: 100 }).catch(() => null),
      ])
      setRows(docs.items ?? [])
      const all = emps?.items ?? []
      setEmployees(all)
      setOwnerId((prev) => prev || all[0]?.id || '')
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to load documents.')
    } finally {
      setLoading(false)
    }
  }, [query])

  useEffect(() => { void load() }, [load])

  async function onUpload(file: File | null) {
    if (!file) return
    if (!ownerId) { setError('Select an employee to attach the document to.'); return }
    const form = new FormData()
    form.append('file', file)
    form.append('title', file.name)
    form.append('type', 'OTHER')
    form.append('employee_id', ownerId)
    setUploading(true); setError(''); setNotice('')
    try {
      const doc = await api.uploadDocument(form)
      setNotice(`Uploaded “${doc.title ?? file.name}”.`)
      fileRef.current!.value = ''
      await load()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Upload failed.')
    } finally {
      setUploading(false)
    }
  }

  async function remove(id: string) {
    setDeletingId(id); setError(''); setNotice('')
    try {
      await api.deleteDocument(id)
      setNotice('Document deleted.')
      await load()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to delete document.')
    } finally {
      setDeletingId('')
    }
  }

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium text-primary">People documents</p>
          <h1 className="text-3xl font-semibold tracking-tight">Documents</h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">Store and share files such as contracts, IDs, and certificates.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Select aria-label="Attach to employee" value={ownerId} onChange={(e) => setOwnerId(e.target.value)} options={employees.map((emp) => ({ label: `${emp.first_name ?? ''} ${emp.last_name ?? ''}`.trim() || emp.email || (emp.id ?? ''), value: emp.id ?? '' }))} placeholder="Attach to…" className="h-10 w-56" />
          <input ref={fileRef} type="file" className="hidden" onChange={(e) => void onUpload(e.target.files?.[0] ?? null)} />
          <Button onClick={() => fileRef.current?.click()} disabled={uploading}><Upload data-icon="inline-start" />{uploading ? 'Uploading…' : 'Upload document'}</Button>
        </div>
      </header>

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="rounded-2xl border border-border bg-card">
        <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input aria-label="Search documents" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search documents" className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
        </div>
        {loading ? <p className="p-8 text-sm text-muted-foreground">Loading documents…</p> :
        rows.length === 0 ? <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center"><p className="font-medium">No documents yet.</p><p className="text-sm text-muted-foreground">Upload a contract, ID, or certificate to get started.</p></div> :
        <ul className="divide-y divide-border">
          {rows.map((doc) => (
            <li key={doc.id} className="flex flex-col gap-3 p-5 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex min-w-0 items-center gap-3">
                <div className="grid size-10 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary"><FileText /></div>
                <div className="min-w-0">
                  <p className="truncate font-medium">{doc.title ?? 'Untitled document'}</p>
                  <p className="truncate text-xs text-muted-foreground">{doc.type || 'Other'} · {doc.mime_type || 'file'} {doc.size_bytes ? `· ${(doc.size_bytes / 1024).toFixed(0)} KB` : ''}</p>
                </div>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                {doc.status && <Badge variant={darknikVariant[doc.status] ?? 'outline'}>{doc.status.replaceAll('_', ' ')}</Badge>}
                <Button variant="outline" size="sm" render={<Link href={api.downloadDocument(doc.id!)} target="_blank" rel="noreferrer" download />}><Download data-icon="inline-start" />Download</Button>
                <Button variant="ghost" size="icon-sm" aria-label="Delete document" disabled={deletingId === doc.id} onClick={() => { if (window.confirm('Delete this document?')) void remove(doc.id!) }}><Trash2 /></Button>
              </div>
            </li>
          ))}
        </ul>}
      </div>
    </section>
  )
}