import { motion, AnimatePresence } from 'framer-motion'

interface SmoothCollapseProps {
  isOpen: boolean
  children: React.ReactNode
  className?: string
}

export function SmoothCollapse({ isOpen, children, className = '' }: SmoothCollapseProps) {
  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ height: 0, opacity: 0 }}
          animate={{ height: 'auto', opacity: 1 }}
          exit={{ height: 0, opacity: 0 }}
          transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
          // 核心：overflow-hidden 配合 p-1 和 -m-1 补偿焦点边框的渲染空间
          className={`overflow-hidden p-1 -m-1 ${className}`}
        >
          {children}
        </motion.div>
      )}
    </AnimatePresence>
  )
}