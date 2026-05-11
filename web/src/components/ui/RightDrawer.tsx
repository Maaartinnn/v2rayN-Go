import { useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X } from 'lucide-react'

interface RightDrawerProps {
  isOpen: boolean
  onClose: () => void
  title: React.ReactNode
  subtitle?: React.ReactNode
  children: React.ReactNode
  width?: string
}

export function RightDrawer({ 
  isOpen, 
  onClose, 
  title, 
  subtitle, 
  children, 
  width = '480px' 
}: RightDrawerProps) {
  
  // 监听 ESC 键关闭
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) onClose()
    }
    window.addEventListener('keydown', handleEsc)
    return () => window.removeEventListener('keydown', handleEsc)
  }, [isOpen, onClose])

  // 锁定底部滚动
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [isOpen])

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* 背景遮罩 */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            onClick={onClose}
            className="fixed inset-0 bg-[#141413]/30 backdrop-blur-sm z-40"
          />
          
          {/* 抽屉主体 */}
          <motion.div
            initial={{ x: '100%', boxShadow: '-20px 0 40px rgba(0,0,0,0)' }}
            animate={{ x: 0, boxShadow: '-20px 0 40px rgba(0,0,0,0.1)' }}
            exit={{ x: '100%', boxShadow: '-20px 0 40px rgba(0,0,0,0)' }}
            transition={{ type: 'spring', bounce: 0, duration: 0.4 }}
            className="fixed top-0 right-0 h-full bg-background border-l border-border z-50 flex flex-col"
            style={{ width, maxWidth: '100vw' }}
          >
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-5 border-b border-border bg-(--color-card) shrink-0">
              <div>
                <h2 className="text-lg font-bold text-(--color-foreground) flex items-center gap-2">
                  {title}
                </h2>
                {subtitle && (
                  <p className="text-xs text-muted-foreground mt-1">
                    {subtitle}
                  </p>
                )}
              </div>
              <button 
                onClick={onClose}
                className="p-2 rounded-lg text-muted-foreground hover:bg-(--color-muted) hover:text-(--color-foreground) transition-colors"
              >
                <X size={20} />
              </button>
            </div>

            {/* Scrollable Content */}
            <div className="flex-1 overflow-y-auto p-6 scrollbar-thin">
              {children}
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}