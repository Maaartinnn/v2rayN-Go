import { labelStyle } from './formStyles'

interface FormFieldProps {
  label: string
  /** Grid column span for this field, defaults to full width */
  cols?: 'full' | '1/2' | '1/3' | '2/3'
  /** Optional hint text shown below the field */
  hint?: string
  children: React.ReactNode
}

/**
 * Standardized form field wrapper: label + input/select + optional hint.
 * Use within a grid or space-y container.
 */
export function FormField({ label, cols = 'full', hint, children }: FormFieldProps) {
  const colClass =
    cols === '1/2'
      ? 'col-span-1'
      : cols === '1/3'
        ? 'col-span-1'
        : cols === '2/3'
          ? 'col-span-2'
          : 'col-span-full'

  return (
    <div className={colClass}>
      <label className="text-xs font-medium block mb-1" style={labelStyle}>
        {label}
      </label>
      {children}
      {hint && (
        <span
          className="text-[10px] mt-0.5 block"
          style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
        >
          {hint}
        </span>
      )}
    </div>
  )
}