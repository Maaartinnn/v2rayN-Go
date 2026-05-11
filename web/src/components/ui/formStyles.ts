/**
 * Shared form style constants for all edit form components.
 * Import these to ensure consistent input/label/select styling across the app.
 */

export const inputStyle = {
  backgroundColor: 'var(--color-overlay)',
  borderColor: 'var(--color-border)',
  color: 'var(--color-foreground)',
  fontFamily: 'var(--font-mono)',
} as const

export const inputHeadingStyle = {
  backgroundColor: 'var(--color-overlay)',
  borderColor: 'var(--color-border)',
  color: 'var(--color-foreground)',
  fontFamily: 'var(--font-heading)',
} as const

export const labelStyle = {
  color: 'var(--color-muted-foreground)',
  fontFamily: 'var(--font-heading)' as const,
}

export const textareaStyle = {
  backgroundColor: 'var(--color-overlay)',
  borderColor: 'var(--color-border)',
  color: 'var(--color-foreground)',
  fontFamily: 'var(--font-heading)',
} as const