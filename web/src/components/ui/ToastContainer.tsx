import { useEffect, useRef, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-react'
import { useStore } from '../../store'
import type { Toast } from '../../store'

// ==================== ToastItem ====================
// 单个 Toast 项，管理自身的定时器生命周期。
// 定时器逻辑放在组件而非 Store 中，利用 useEffect 管理和清理，
// 避免"手动关闭时定时器仍触发"的竞态条件。

interface ToastItemProps {
  toast: Toast
}

function ToastItem({ toast }: ToastItemProps) {
  const removeToast = useStore((s) => s.removeToast)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // 手动关闭：清除定时器后移除
  const handleClose = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
    removeToast(toast.id)
  }, [toast.id, removeToast])

  // 自动消失定时器：仅在 duration > 0 时生效
  useEffect(() => {
    if (toast.duration && toast.duration > 0) {
      timerRef.current = setTimeout(() => {
        timerRef.current = null
        removeToast(toast.id)
      }, toast.duration)
      // 组件卸载或依赖变化时清理定时器
      return () => {
        if (timerRef.current) {
          clearTimeout(timerRef.current)
          timerRef.current = null
        }
      }
    }
  }, [toast.id, toast.duration, removeToast])

  // 类型对应的默认颜色和图标
  const typeConfig = {
    success: { icon: CheckCircle, bg: 'var(--color-success-dim)', color: 'var(--color-success)' },
    error:   { icon: AlertCircle,  bg: 'var(--color-error-dim)',   color: 'var(--color-error)' },
    warning: { icon: AlertTriangle, bg: 'var(--color-warning-dim)', color: 'var(--color-warning)' },
    info:    { icon: Info,          bg: 'var(--color-muted)',       color: 'var(--color-muted-foreground)' },
  }
  const config = typeConfig[toast.type] || typeConfig.info
  const Icon = config.icon

  // 优先使用自定义颜色，否则使用类型默认色
  const bgColor = toast.color?.bg || config.bg
  const textColor = toast.color?.text || config.color

  return (
    <motion.div
      layout
      initial={{ opacity: 0, x: 60, scale: 0.95 }}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: 60, scale: 0.95 }}
      transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
      role="alert"
      className="flex items-center gap-3 px-4 py-3 rounded-xl border shadow-lg min-w-70 max-w-105"
      style={{
        backgroundColor: 'var(--color-card)',
        borderColor: 'var(--color-border)',
        boxShadow: 'var(--shadow-elevated)',
      }}
    >
      {/* 类型图标 */}
      <div
        className="w-7 h-7 rounded-lg flex items-center justify-center shrink-0"
        style={{ backgroundColor: bgColor }}
      >
        <Icon size={14} style={{ color: textColor }} />
      </div>

      {/* 消息文本 */}
      <span
        className="flex-1 text-xs font-medium"
        style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
      >
        {toast.message}
      </span>

      {/* 操作按钮（可选） */}
      {toast.action && (
        <button
          onClick={() => {
            toast.action!.onClick()
            handleClose()
          }}
          className="px-2.5 py-1 text-[11px] font-medium rounded-md transition-colors cursor-pointer shrink-0"
          style={{
            backgroundColor: bgColor,
            color: textColor,
            fontFamily: 'var(--font-heading)',
          }}
        >
          {toast.action.label}
        </button>
      )}

      {/* 关闭按钮 */}
      <button
        onClick={handleClose}
        className="p-1 rounded-md transition-colors cursor-pointer shrink-0"
        style={{ color: 'var(--color-muted-foreground)' }}
        onMouseEnter={(e) => {
          e.currentTarget.style.color = 'var(--color-foreground)'
          e.currentTarget.style.backgroundColor = 'var(--color-muted)'
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.color = 'var(--color-muted-foreground)'
          e.currentTarget.style.backgroundColor = 'transparent'
        }}
        aria-label="关闭通知"
      >
        <X size={13} />
      </button>
    </motion.div>
  )
}

// ==================== ToastContainer ====================
// 固定在右上角的 Toast 容器，负责循环渲染所有 toast 项。
// 窄屏（≤640px）自动切换为顶部居中 + 拉伸宽度。
// aria-live + role 确保屏幕阅读器播报动态插入的通知。

export function ToastContainer() {
  const toasts = useStore((s) => s.toasts)

  return (
    <div
      className="fixed z-50 flex flex-col gap-2
        right-4 top-16
        max-sm:right-0 max-sm:left-0 max-sm:top-16 max-sm:items-center max-sm:px-4"
      aria-live="polite"
      role="status"
    >
      <AnimatePresence mode="popLayout">
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} />
        ))}
      </AnimatePresence>
    </div>
  )
}