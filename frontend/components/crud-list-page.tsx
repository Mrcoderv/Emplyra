'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { CheckCircle2, Pencil, Plus, RefreshCw, Search, Trash2, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'

export type FieldDef = {
  name: string
  label: string
  type?: 'text' | 'number' | 'date' | 'select' | 'textarea' | 'email'
  required?: boolean
  options?: { label: string; value: string }[]
  placeholder?: string
  span2?: boolean
}

function toFormValue(v: unknown): string {
  if (v === null || v === undefined) return ''
  if (typeof v === 'object' && 'id' in (v as Record<string, unknown>)) return String((v as { id?: unknown }).id ?? '')
  return String(v)
}

export type CrudPageProps<T extends { id?: string }> = {
  title: string
  eyebrow: string
  description: string
  actionLabel?: string
  empty: string
  fields: FieldDef[]
  columns: string[]
  load: (query: string, signal?: AbortSignal) => Promise<T[]>
  create: (body: Record<string, unknown>) => Promise<unknown>
  update: (id: string, body: Record<string, unknown>) => Promise<unknown>
  remove: (id: string) => Promise<unknown>
  toRow?: (item: T) => Record<string, unknown>
  rowLabel?: (item: T) => string
  render?: (item: T) => React.ReactNode
  extraHeader?: React.ReactNode
  editable?: boolean
  deletable?: boolean
  fieldOptions?: Record<string, () => Promise<{ label: string; value: string }[]>>
}

export default function CrudListPage<T extends { id?: string }>({
  title, eyebrow, description, actionLabel = 'New record', empty, fields, columns,
  load, create, update, remove, toRow, rowLabel, render, extraHeader, editable = true, deletable = true,
  fieldOptions,
}: CrudPageProps<T>) {
  const emptyForm = useMemo(() => Object.fromEntries(fields.map((f) => [f.name, ''])), [fields])
  const [rows, setRows] = useState<T[]>([])
  const [options, setOptions] = useState<Record<string, { label: string; value: string }[]>>({})
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState<Record<string, string>>(emptyForm)

  useEffect(() => {
    if (!fieldOptions) return
    let cancelled = false
    Promise.all(Object.entries(fieldOptions).map(async ([field, fetcher]) => {
      try { return [field, await fetcher()] as const } catch { return [field, []] as const }
    })).then((entries) => {
      if (cancelled) return
      setOptions(Object.fromEntries(entries))
    })
    return () => { cancelled = true }
  }, [fieldOptions])

  const loadRows = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const controller = new AbortController()
      const items = await load(query, controller.signal)
      setRows(items)
    } catch (cause) {
      if (cause instanceof Error && cause.name === 'AbortError') return
      setError(cause instanceof Error ? cause.message : 'Unable to load records.')
    } finally {
      setLoading(false)
    }
  }, [load, query])

  useEffect(() => {
    void loadRows()
  }, [loadRows])

  const openCreate = () => {
    setEditingId(null)
    setForm(emptyForm)
    setError('')
    setNotice('')
    setFormOpen(true)
  }

  const openEdit = (item: T) => {
    setEditingId(item.id ?? null)
    setForm((prev) => ({
      ...emptyForm,
      ...Object.fromEntries(Object.entries(item).map(([k, v]) => [k, toFormValue(v)])),
    }))
    setError('')
    setNotice('')
    setFormOpen(true)
  }

  const closeForm = () => { setFormOpen(false); setEditingId(null) }

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    setSaving(true)
    setError('')
    setNotice('')
    try {
      const body = Object.fromEntries(Object.entries(form).filter(([, v]) => v !== ''))
      if (editingId) {
        await update(editingId, body)
        setNotice('Record updated successfully.')
      } else {
        await create(body)
        setNotice('Record created successfully.')
      }
      closeForm()
      await loadRows()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to save record.')
    } finally {
      setSaving(false)
    }
  }

  async function removeRow(item: T) {
    if (!item.id) return
    if (!window.confirm('Delete this record? This action cannot be undone.')) return
    setError('')
    setNotice('')
    try {
      await remove(item.id)
      setNotice('Record deleted.')
      await loadRows()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to delete record.')
    }
  }

  const tableRow = useMemo(() => {
    if (toRow) return toRow
    return (item: T): Record<string, unknown> => Object.fromEntries(Object.entries(item).filter(([k]) => columns.includes(k)))
  }, [toRow, columns])

  const dataColumns = useMemo(() => columns.filter((c) => !['id', 'tenant_id', 'updated_at', 'created_at', 'deleted_at'].includes(c)).slice(0, 8), [columns])

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium text-primary">{eyebrow}</p>
          <h1 className="text-3xl font-semibold tracking-tight text-balance">{title}</h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">{description}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {extraHeader}
          <Button onClick={openCreate}><Plus data-icon="inline-start" />{actionLabel}</Button>
        </div>
      </header>

      {formOpen && (
        <form onSubmit={submit} className="grid gap-4 rounded-2xl border border-border bg-card p-5 sm:grid-cols-2">
          {fields.map((field) => (
            <div key={field.name} className={cn('flex flex-col gap-2', field.span2 && 'sm:col-span-2')}>
              <label htmlFor={`field-${field.name}`} className="text-sm font-medium">{field.label}{field.required ? ' *' : ''}</label>
              {field.type === 'textarea' ? (
                <Textarea id={`field-${field.name}`} required={field.required} value={form[field.name] ?? ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} placeholder={field.placeholder} />
              ) : field.type === 'select' ? (
                <select id={`field-${field.name}`} required={field.required} value={form[field.name] ?? ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} className="h-10 w-full appearance-none rounded-xl border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring">
                  <option value="">{field.placeholder ?? 'Select…'}</option>
                  {(fieldOptions?.[field.name] ? options[field.name] ?? [] : field.options ?? []).map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
                </select>
              ) : (
                <Input id={`field-${field.name}`} type={field.type ?? 'text'} required={field.required} value={form[field.name] ?? ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} placeholder={field.placeholder} />
              )}
            </div>
          ))}
          <div className="flex gap-2 sm:col-span-2">
            <Button type="submit" disabled={saving}>{saving ? 'Saving…' : editingId ? 'Save changes' : 'Save record'}</Button>
            <Button type="button" variant="outline" onClick={closeForm}>Cancel</Button>
          </div>
        </form>
      )}

      {notice && <div className="flex items-center gap-2 rounded-xl border border-primary/25 bg-primary/10 p-4 text-sm text-foreground"><CheckCircle2 className="shrink-0 text-primary" />{notice}</div>}
      {error && <div role="alert" className="flex items-center gap-2 rounded-xl border border-destructive/25 bg-destructive/10 p-4 text-sm text-destructive"><XCircle className="shrink-0" />{error}</div>}

      <div className="rounded-2xl border border-border bg-card">
        <div className="flex flex-col gap-3 border-b border-border p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input aria-label={`Search ${title}`} value={query} onChange={(e) => setQuery(e.target.value)} placeholder={`Search ${title.toLowerCase()}`} className="h-10 w-full rounded-xl border border-input bg-background pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => void loadRows()} disabled={loading}><RefreshCw data-icon="inline-start" />Refresh</Button>
          </div>
        </div>
        {loading ? (
          <p className="p-8 text-sm text-muted-foreground">Loading records…</p>
        ) : rows.length === 0 ? (
          <div className="flex min-h-56 flex-col items-center justify-center gap-2 p-8 text-center">
            <p className="font-medium">{empty}</p>
            <p className="text-sm text-muted-foreground">Records will appear here once they are created in your workspace.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-muted/50 text-xs uppercase text-muted-foreground">
                <tr>
                  <th className="px-5 py-3">Record</th>
                  {dataColumns.map((column) => <th key={column} className="px-5 py-3">{column.replaceAll('_', ' ')}</th>)}
                  {(editable || deletable) && <th className="px-5 py-3 text-right">Actions</th>}
                </tr>
              </thead>
              <tbody>
                {rows.map((item, index) => {
                  const display = render ? render(item) : (
                    <>
                      <td className="px-5 py-4 font-medium">
                        {rowLabel ? rowLabel(item) : String(tableRow(item).name ?? tableRow(item).title ?? tableRow(item).employee_name ?? item.id ?? `Record ${index + 1}`)}
                      </td>
                      {dataColumns.map((column) => (
                        <td key={column} className="max-w-56 truncate px-5 py-4 text-muted-foreground">
                          {value(column === 'status' ? statusValue(tableRow(item)[column]) : tableRow(item)[column])}
                        </td>
                      ))}
                    </>
                  )
                  return (
                    <tr key={item.id ?? index} className="border-t border-border hover:bg-muted/30">
                      {Array.isArray(display) ? display : display}
                      {(editable || deletable) && (
                        <td className="px-5 py-4 text-right">
                          <div className="inline-flex gap-1">
                            {editable && <Button variant="ghost" size="icon-xs" onClick={() => openEdit(item)} aria-label="Edit record"><Pencil /></Button>}
                            {deletable && <Button variant="ghost" size="icon-xs" onClick={() => void removeRow(item)} aria-label="Delete record"><Trash2 /></Button>}
                          </div>
                        </td>
                      )}
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  )
}

function value(value: unknown) {
  if (value === null || value === undefined || value === '') return '—'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function statusValue(value: unknown) {
  if (typeof value === 'string') return value.replaceAll('_', ' ')
  return value
}