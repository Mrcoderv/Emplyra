import { cn } from '@/lib/utils'

type SelectOption = { label: string; value: string }
type Props = React.ComponentProps<'select'> & { options: (SelectOption | string)[]; placeholder?: string }

function toOption(option: SelectOption | string): SelectOption {
  return typeof option === 'string' ? { label: option, value: option } : option
}

function Select({ className, options, placeholder, ...props }: Props) {
  return (
    <select data-slot="select" className={cn('h-10 w-full appearance-none rounded-xl border border-input bg-background px-3 text-sm outline-none transition-colors focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50', className)} {...props}>
      {placeholder ? <option value="">{placeholder}</option> : null}
      {options.map((option, i) => {
        const { label, value } = toOption(option)
        return <option key={`${value}-${i}`} value={value}>{label}</option>
      })}
    </select>
  )
}

export { Select }