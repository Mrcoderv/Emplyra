import { cn } from '@/lib/utils'

function Textarea({ className, ...props }: React.ComponentProps<'textarea'>) {
  return <textarea data-slot="textarea" className={cn('min-h-24 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm shadow-none outline-none placeholder:text-muted-foreground focus:ring-2 focus:ring-ring', className)} {...props} />
}

export { Textarea }