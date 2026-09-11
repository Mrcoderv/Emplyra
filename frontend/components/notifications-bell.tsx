'use client'

import { useEffect, useRef, useState } from 'react'
import { Bell, BellRing, CheckCheck } from 'lucide-react'
import { api, type Notification } from '@/lib/api'

export default function NotificationsBell() {
  const [open, setOpen] = useState(false)
  const [items, setItems] = useState<Notification[]>([])
  const [unread, setUnread] = useState(0)
  const [loading, setLoading] = useState(true)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    let cancelled = false
    async function refresh() {
      try {
        const [list, count] = await Promise.all([
          api.notifications({ page: 1, pageSize: 12 }),
          api.unreadCount().catch(() => ({ unread: 0 })),
        ])
        if (cancelled) return
        setItems(list.items ?? [])
        setUnread(Number(count.unread ?? 0))
      } catch {
        if (!cancelled) setItems([])
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void refresh()
    const timer = window.setInterval(() => void refresh(), 45000)
    return () => { cancelled = true; window.clearInterval(timer) }
  }, [])

  useEffect(() => {
    function onClickOutside(event: MouseEvent) {
      if (ref.current && !ref.current.contains(event.target as Node)) setOpen(false)
    }
    function onKey(event: KeyboardEvent) { if (event.key === 'Escape') setOpen(false) }
    document.addEventListener('mousedown', onClickOutside)
    document.addEventListener('keydown', onKey)
    return () => { document.removeEventListener('mousedown', onClickOutside); document.removeEventListener('keydown', onKey) }
  }, [])

  async function markAllRead() {
    try { await api.markAllRead(); setUnread(0); setItems((prev) => prev.map((n) => ({ ...n, is_read: true }))) }
    catch { /* ignore */ }
  }

  return (
    <div className="relative" ref={ref}>
      <button aria-label={`Notifications, ${unread} unread`} onClick={() => setOpen((v) => !v)} className="relative grid size-9 place-items-center rounded-full text-muted-foreground hover:bg-muted hover:text-foreground">
        {unread > 0 ? <BellRing className="text-primary" /> : <Bell />}
        {unread > 0 && <span className="pointer-events-none absolute -right-0.5 -top-0.5 grid min-w-4 place-items-center rounded-full bg-destructive px-1 text-[10px] font-semibold text-destructive-foreground">{unread > 9 ? '9+' : unread}</span>}
      </button>
      {open && (
        <div className="absolute right-0 top-12 z-50 w-80 overflow-hidden rounded-2xl border border-border bg-card shadow-lg sm:w-96">
          <div className="flex items-center justify-between border-b border-border px-4 py-3">
            <p className="text-sm font-semibold">Notifications</p>
            <button onClick={() => void markAllRead()} className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"><CheckCheck data-icon="inline-start" />Mark all read</button>
          </div>
          <div className="max-h-96 overflow-y-auto">
            {loading ? <p className="p-6 text-center text-sm text-muted-foreground">Loading…</p> :
            items.length === 0 ? <div className="flex flex-col items-center gap-1 p-6 text-center"><Bell className="text-muted-foreground" /><p className="text-sm font-medium">You&apos;re all caught up</p><p className="text-xs text-muted-foreground">New notifications will appear here.</p></div> :
            <ul className="divide-y divide-border">
              {items.map((n) => (
                <li key={n.id} className={`flex flex-col gap-1 px-4 py-3 ${n.is_read ? '' : 'bg-primary/5'}`}>
                  <p className="text-sm font-medium">{n.title || 'Notification'}</p>
                  {n.message && <p className="text-xs leading-5 text-muted-foreground">{n.message}</p>}
                  <p className="text-xs text-muted-foreground/70">{n.created_at ? new Date(n.created_at).toLocaleString() : ''}</p>
                </li>
              ))}
            </ul>}
          </div>
        </div>
      )}
    </div>
  )
}