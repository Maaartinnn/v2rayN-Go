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
          layout
          initial={{ height: 0, opacity: 0 }}
          animate={{ height: 'auto', opacity: 1 }}
          exit={{ height: 0, opacity: 0 }}
          transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
          // 核心：overflow-hidden 配合 p-1 和 -m-1 补偿焦点边框的渲染空间
          // flex flex-col 形成独立 BFC，防止 margin 坍塌导致动画高度计算不准
          className={`overflow-hidden flex flex-col ${className}`}
        >
          {children}
        </motion.div>
      )}
    </AnimatePresence>
  )
}