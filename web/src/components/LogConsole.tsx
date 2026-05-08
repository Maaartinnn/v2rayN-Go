import { useEffect, useRef } from 'react'
import { motion } from 'framer-motion'
import { Trash2 } from 'lucide-react'
import { useStore } from '../store'
import { useT, useI18n } from '../lib/i18n'

export function LogConsole() {
  const { logs, clearLogs } = useStore()
  const scrollRef = useRef<HTMLDivElement>(null)
  const t = useT()
  const { theme } = useI18n()

  // Determine if dark mode is active
  const isDark = (() => {
    if (theme === 'dark') return true
    if (theme === 'light') return false
    return typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches
  })()

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [logs])

  const highlightLog = (content: string) => {
    if (isDark) {
      return content
        .replace(/\[INFO\]/g, '<span style="color: #6A9BCC">[INFO]</span>')
        .replace(/\[WARN\]/g, '<span style="color: #C9943A">[WARN]</span>')
        .replace(/\[ERROR\]/g, '<span style="color: #C0453A">[ERROR]</span>')
        .replace(/\[xray\]/g, '<span style="color: #D97757">[xray]</span>')
        .replace(/\[sing-box\]/g, '<span style="color: #788C5D">[sing-box]</span>')
    }
    return content
      .replace(/\[INFO\]/g, '<span style="color: #5A89B8">[INFO]</span>')
      .replace(/\[WARN\]/g, '<span style="color: #C9943A">[WARN]</span>')
      .replace(/\[ERROR\]/g, '<span style="color: #C0453A">[ERROR]</span>')
      .replace(/\[xray\]/g, '<span style="color: #C96442">[xray]</span>')
      .replace(/\[sing-box\]/g, '<span style="color: #6B8F47">[sing-box]</span>')
  }

  return (
    <div className="max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-5">
        <h1
          className="text-xl font-semibold"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('logs.title')}
        </h1>
        <motion.button
          onClick={clearLogs}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors cursor-pointer"
          style={{
            backgroundColor: 'var(--color-muted)',
            borderColor: 'var(--color-border)',
            color: 'var(--color-muted-foreground)',
            fontFamily: 'var(--font-heading)',
          }}
          whileTap={{ scale: 0.95 }}
        >
          <Trash2 size={13} />
          {t('logs.clear')}
        </motion.button>
      </div>

      <div
        className="rounded-xl border overflow-hidden"
        style={{
          backgroundColor: isDark ? '#1A1916' : '#F5F3EC',
          borderColor: 'var(--color-border)',
          boxShadow: 'var(--shadow-card)',
        }}
      >
        {/* Terminal-style title bar */}
        <div
          className="flex items-center gap-2 px-4 py-2.5 border-b"
          style={{ borderColor: isDark ? '#2E2D27' : '#E8E6DC' }}
        >
          <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: '#C0453A', opacity: 0.8 }} />
          <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: '#C9943A', opacity: 0.8 }} />
          <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: '#6B8F47', opacity: 0.8 }} />
          <span
            className="ml-2 text-xs"
            style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-mono)' }}
          >
            core.log
          </span>
        </div>

        {/* Log content */}
        <div
          ref={scrollRef}
          className="h-[500px] overflow-y-auto p-4 text-xs leading-relaxed"
          style={{
            color: isDark ? '#EAE7DC' : '#141413',
            fontFamily: 'var(--font-mono)',
          }}
        >
          {logs.length === 0 ? (
            <div className="text-center py-20">
              <p style={{ color: isDark ? '#5C5A54' : '#B0AEA5' }}>{t('logs.no_logs')}</p>
              <p
                className="mt-1"
                style={{ color: isDark ? '#3A3830' : '#D8D5CC', fontFamily: 'var(--font-heading)' }}
              >
                {t('logs.start_hint')}
              </p>
            </div>
          ) : (
            logs.map((log, i) => (
              <div key={i} className="flex gap-3 py-0.5">
                <span
                  className="flex-shrink-0 select-none"
                  style={{ color: isDark ? '#5C5A54' : '#B0AEA5' }}
                >
                  {new Date(log.time).toLocaleTimeString('en-US', { hour12: false })}
                </span>
                <span
                  className="break-all"
                  dangerouslySetInnerHTML={{ __html: highlightLog(log.content) }}
                />
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}