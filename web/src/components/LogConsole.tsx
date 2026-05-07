import { useEffect, useRef } from 'react'
import { motion } from 'framer-motion'
import { Trash2 } from 'lucide-react'
import { useStore } from '../store'
import { useT } from '../lib/i18n'
import { useDarkMode } from '../lib/useDarkMode'

export function LogConsole() {
  const { logs, clearLogs } = useStore()
  const scrollRef = useRef<HTMLDivElement>(null)
  const t = useT()
  const isDark = useDarkMode()

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [logs])

  const highlightLog = (content: string, isDark: boolean) => {
    if (isDark) {
      return content
        .replace(/\[INFO\]/g, '<span style="color: oklch(0.6 0.1 250)">[INFO]</span>')
        .replace(/\[WARN\]/g, '<span style="color: oklch(0.769 0.188 70)">[WARN]</span>')
        .replace(/\[ERROR\]/g, '<span style="color: oklch(0.577 0.245 27)">[ERROR]</span>')
        .replace(/\[xray\]/g, '<span style="color: oklch(0.6 0.15 280)">[xray]</span>')
        .replace(/\[sing-box\]/g, '<span style="color: oklch(0.6 0.15 160)">[sing-box]</span>')
    }
    // Light mode colors
    return content
      .replace(/\[INFO\]/g, '<span style="color: #2563eb">[INFO]</span>')
      .replace(/\[WARN\]/g, '<span style="color: #d97706">[WARN]</span>')
      .replace(/\[ERROR\]/g, '<span style="color: #dc2626">[ERROR]</span>')
      .replace(/\[xray\]/g, '<span style="color: #7c3aed">[xray]</span>')
      .replace(/\[sing-box\]/g, '<span style="color: #059669">[sing-box]</span>')
  }

  return (
    <div className="max-w-3xl mx-auto py-6">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-medium">{t('logs.title')}</h1>
        <motion.button
          onClick={clearLogs}
          className="px-3 py-1.5 text-sm rounded-lg bg-muted hover:bg-accent text-muted-foreground hover:text-foreground transition-colors flex items-center gap-1.5"
          whileTap={{ scale: 0.95 }}
        >
          <Trash2 size={14} />
          {t('logs.clear')}
        </motion.button>
      </div>

      <div className={`rounded-2xl border overflow-hidden transition-colors ${
        isDark
          ? 'bg-zinc-950 border-zinc-800'
          : 'bg-zinc-50 border-zinc-200'
      }`}>
        {/* macOS-style title bar */}
        <div className={`flex items-center gap-2 px-4 py-2.5 border-b ${
          isDark ? 'border-zinc-800' : 'border-zinc-200'
        }`}>
          <div className="w-3 h-3 rounded-full bg-red-500/80" />
          <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
          <div className="w-3 h-3 rounded-full bg-green-500/80" />
          <span className={`ml-2 text-xs font-mono ${
            isDark ? 'text-zinc-500' : 'text-zinc-400'
          }`}>core.log</span>
        </div>

        {/* Log content */}
        <div
          ref={scrollRef}
          className={`h-[500px] overflow-y-auto p-4 font-mono text-xs leading-relaxed ${
            isDark ? 'text-zinc-300' : 'text-zinc-700'
          }`}
        >
          {logs.length === 0 ? (
            <div className={`text-center py-16 ${isDark ? 'text-zinc-600' : 'text-zinc-400'}`}>
              <p>{t('logs.no_logs')}</p>
              <p className={`mt-1 ${isDark ? 'text-zinc-700' : 'text-zinc-300'}`}>{t('logs.start_hint')}</p>
            </div>
          ) : (
            logs.map((log, i) => (
              <div key={i} className="flex gap-3 py-0.5">
                <span className={`flex-shrink-0 select-none ${
                  isDark ? 'text-zinc-600' : 'text-zinc-400'
                }`}>
                  {new Date(log.time).toLocaleTimeString('en-US', { hour12: false })}
                </span>
                <span
                  className="break-all"
                  dangerouslySetInnerHTML={{ __html: highlightLog(log.content, isDark) }}
                />
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}