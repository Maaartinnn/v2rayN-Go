import { Check, X } from 'lucide-react'

interface FormActionsProps {
  onCancel: () => void
  onSubmit: () => void
  /** Cancel button label key / text, defaults to a standard cancel text */
  cancelLabel?: string
  /** Submit button label key / text */
  submitLabel: string
  /** Disable submit when true */
  submitDisabled?: boolean
}

/**
 * Standardized action buttons row for edit forms.
 * Renders a right-aligned cancel + submit pair.
 */
export function FormActions({
  onCancel,
  onSubmit,
  cancelLabel = 'Cancel',
  submitLabel,
  submitDisabled = false,
}: FormActionsProps) {
  return (
    <div className="flex justify-end gap-2 mt-5">
      <button
        onClick={onCancel}
        className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium cursor-pointer btn-ghost"
        style={{ fontFamily: 'var(--font-heading)' }}
      >
        <X size={13} />
        {cancelLabel}
      </button>
      <button
        onClick={onSubmit}
        disabled={submitDisabled}
        className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium cursor-pointer btn-primary"
        style={{ fontFamily: 'var(--font-heading)' }}
      >
        <Check size={13} />
        {submitLabel}
      </button>
    </div>
  )
}