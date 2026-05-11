import { motion } from 'framer-motion'

interface EditFormCardProps {
  children: React.ReactNode
}

/**
 * Animated floating card wrapper for edit forms.
 * Appears as an independent card below the target list card.
 * Wrap with AnimatePresence in the parent for exit animations.
 */
export function EditFormCard({ children }: EditFormCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, height: 0, overflow: 'hidden' }}
      animate={{ opacity: 1, height: 'auto', transitionEnd: { overflow: 'visible' } }}
      exit={{ opacity: 0, height: 0, overflow: 'hidden' }}
      transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
      className="mb-4"
    >
      <div
        className="rounded-xl border p-5"
        style={{
          backgroundColor: 'var(--color-card)',
          borderColor: 'var(--color-border)',
          boxShadow: 'var(--shadow-card)',
        }}
      >
        {children}
      </div>
    </motion.div>
  )
}