import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const inputVariants = cva(
  'h-10 w-full rounded-xl border border-input bg-background px-3 text-sm shadow-none transition-colors outline-none placeholder:text-muted-foreground focus:ring-2 focus:ring-ring file:border-0 file:bg-transparent file:text-sm',
  {
    variants: {
      invalid: { true: 'border-destructive focus:ring-destructive/20' },
    },
  },
)

function Input({ className, invalid, type = 'text', ...props }: React.ComponentProps<'input'> & VariantProps<typeof inputVariants>) {
  return <input type={type} data-slot="input" className={cn(inputVariants({ invalid }), className)} {...props} />
}

export { Input, inputVariants }