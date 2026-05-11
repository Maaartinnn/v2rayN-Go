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
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: 'auto' }}
      exit={{ opacity: 0, height: 0 }}
      transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
      // 空间补偿：p-1 + -m-1 撑开 4px 容纳焦点边框
      className="mb-4 overflow-hidden p-1 -m-1"
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