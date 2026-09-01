import AppShell from '@/components/app-shell'
import { AuthGuard } from '@/lib/guards'
export default function AppLayout({ children }: { children: React.ReactNode }) { return <AuthGuard><AppShell>{children}</AppShell></AuthGuard> }
