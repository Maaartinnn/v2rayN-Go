import { useEffect, useRef, useState } from 'react'
import { motion } from 'framer-motion'
import { Trash2, Monitor, Cpu } from 'lucide-react'
import { useStore } from '../store'
import { useT, useI18n } from '../lib/i18n'

const LOG_SOURCES = [
  { id: 'all', label: 'All', icon: Monitor },
  { id: 'system', label: 'System', icon: Monitor },
  { id: 'xray', label: 'Xray', icon: Cpu },
  { id: 'sing-box', label: 'Sing-box', icon: Cpu },
  { id: 'mihomo', label: 'Mihomo', icon: Cpu },
] as const

type SourceId = typeof LOG_SOURCES[number]['id']

export function LogConsole() {
  const { logs, clearLogs } = useStore()
  const scrollRef = useRef<HTMLDivElement>(null)
  const t = useT()
  const { theme } = useI18n()
  const [activeSource, setActiveSource] = useState<SourceId>('all')

  // Determine if dark mode is active
  const isDark = (() => {
    if (theme === 'dark') return true
    if (theme === 'light') return false
    return typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches
  })()

  // Filter logs by source
  const filteredLogs = activeSource === 'all'
    ? logs
    : logs.filter((log) => log.source === activeSource)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [filteredLogs])

  const highlightLog = (content: string) => {
    if (isDark) {
      return content
        .replace(/\[INFO\]/g, '<span style="color: #6A9BCC">[INFO]</span>')
        .replace(/\[WARN\]/g, '<span style="color: #C9943A">[WARN]</span>')
        .replace(/\[ERROR\]/g, '<span style="color: #C0453A">[ERROR]</span>')
    }
    return content
      .replace(/\[INFO\]/g, '<span style="color: #5A89B8">[INFO]</span>')
      .replace(/\[WARN\]/g, '<span style="color: #C9943A">[WARN]</span>')
      .replace(/\[ERROR\]/g, '<span style="color: #C0453A">[ERROR]</span>')
  }

  // Source color mapping
  const getSourceColor = (source: string) => {
    const colors: Record<string, { light: string; dark: string }> = {
      'system': { light: '#6B6860', dark: '#9D9A91' },
      'xray': { light: '#C96442', dark: '#D97757' },
      'sing-box': { light: '#6B8F47', dark: '#788C5D' },
      'mihomo': { light: '#5A89B8', dark: '#6A9BCC' },
    }
    const c = colors[source]
    if (!c) return isDark ? '#9D9A91' : '#6B6860'
    return isDark ? c.dark : c.light
  }

  // Count logs per source
  const sourceCounts: Record<string, number> = { all: logs.length }
  for (const log of logs) {
    sourceCounts[log.source] = (sourceCounts[log.source] || 0) + 1
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
        {/* Terminal-style title bar with source tabs */}
        <div
          className="flex items-center gap-1 px-4 py-2 border-b"
          style={{ borderColor: isDark ? '#2E2D27' : '#E8E6DC' }}
        >
          {/* Traffic lights */}
          <div className="flex items-center gap-1.5 mr-3">
            <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: '#C0453A', opacity: 0.8 }} />
            <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: '#C9943A', opacity: 0.8 }} />
            <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: '#6B8F47', opacity: 0.8 }} />
          </div>

          {/* Source filter tabs */}
          <div className="flex items-center gap-0.5">
            {LOG_SOURCES.map((src) => {
              const isActive = activeSource === src.id
              const count = sourceCounts[src.id] || 0
              return (
                <button
                  key={src.id}
                  onClick={() => setActiveSource(src.id)}
                  className="relative px-2.5 py-1 text-[11px] font-medium rounded-md transition-colors cursor-pointer"
                  style={{
                    backgroundColor: isActive
                      ? (isDark ? 'rgba(217, 119, 87, 0.15)' : 'rgba(217, 119, 87, 0.12)')
                      : 'transparent',
                    color: isActive
                      ? (isDark ? '#D97757' : '#C96442')
                      : (isDark ? '#5C5A54' : '#B0AEA5'),
                    fontFamily: 'var(--font-heading)',
                  }}
                >
                  {src.label}
                  {count > 0 && (
                    <span
                      className="ml-1 text-[9px] opacity-60"
                    >
                      {count}
                    </span>
                  )}
                </button>
              )
            })}
          </div>
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
          {filteredLogs.length === 0 ? (
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
            filteredLogs.map((log, i) => (
              <div key={i} className="flex gap-3 py-0.5">
                <span
                  className="flex-shrink-0 select-none"
                  style={{ color: isDark ? '#5C5A54' : '#B0AEA5' }}
                >
                  {new Date(log.time).toLocaleTimeString('en-US', { hour12: false })}
                </span>
                {activeSource === 'all' && (
                  <span
                    className="flex-shrink-0 select-none text-[10px] font-medium px-1 py-0.5 rounded"
                    style={{
                      color: getSourceColor(log.source),
                      backgroundColor: isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.04)',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    {log.source || 'system'}
                  </span>
                )}
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