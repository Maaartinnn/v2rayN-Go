import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { AlertTriangle, X } from 'lucide-react'

interface DeleteConfirmBannerProps {
  /** 是否正在显示确认横幅 */
  visible: boolean
  /** 横幅提示文字 */
  message: string
  /** 确认删除回调 */
  onConfirm: () => void
  /** 取消/关闭回调 */
  onCancel: () => void
  /** 超时时间（毫秒），默认 5000 */
  timeout?: number
}

export function DeleteConfirmBanner({
  visible,
  message,
  onConfirm,
  onCancel,
  timeout = 5000,
}: DeleteConfirmBannerProps) {
  const [timer, setTimer] = useState<ReturnType<typeof setTimeout> | null>(null)

  const handleCancel = useCallback(() => {
    if (timer) clearTimeout(timer)
    onCancel()
  }, [timer, onCancel])

  // 超时自动收回
  useEffect(() => {
    if (visible) {
      const t = setTimeout(() => {
        onCancel()
      }, timeout)
      setTimer(t)
      return () => clearTimeout(t)
    } else {
      setTimer(null)
    }
  }, [visible, timeout, onCancel])

  return (
    <AnimatePresence>
      {visible && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
          className="overflow-hidden"
        >
          <div
            className="flex items-center justify-between px-4 py-2.5 rounded-lg mt-2"
            style={{
              backgroundColor: 'var(--color-error-dim)',
              border: '1px solid var(--color-error)',
            }}
          >
            <div className="flex items-center gap-2">
              <AlertTriangle size={14} style={{ color: 'var(--color-error)' }} />
              <span
                className="text-xs font-medium"
                style={{ color: 'var(--color-error)', fontFamily: 'var(--font-heading)' }}
              >
                {message}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={onConfirm}
                className="px-2.5 py-1 text-[11px] font-medium rounded-md transition-colors cursor-pointer"
                style={{
                  backgroundColor: 'var(--color-error)',
                  color: '#fff',
                  fontFamily: 'var(--font-heading)',
                }}
              >
                删除
              </button>
              <button
                onClick={handleCancel}
                className="p-1 rounded-md transition-colors cursor-pointer"
                style={{ color: 'var(--color-error)' }}
              >
                <X size={13} />
              </button>
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}