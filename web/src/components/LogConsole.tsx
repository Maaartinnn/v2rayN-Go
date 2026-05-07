import { useEffect, useRef } from 'react'
import { motion } from 'framer-motion'
import { Trash2 } from 'lucide-react'
import { useStore } from '../store'

export function LogConsole() {
  const { logs, clearLogs } = useStore()
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [logs])

  const highlightLog = (content: string) => {
    // Syntax highlighting for log levels
    return content
      .replace(/\[INFO\]/g, '<span style="color: oklch(0.6 0.1 250)">[INFO]</span>')
      .replace(/\[WARN\]/g, '<span style="color: oklch(0.769 0.188 70)">[WARN]</span>')
      .replace(/\[ERROR\]/g, '<span style="color: oklch(0.577 0.245 27)">[ERROR]</span>')
      .replace(/\[xray\]/g, '<span style="color: oklch(0.6 0.15 280)">[xray]</span>')
      .replace(/\[sing-box\]/g, '<span style="color: oklch(0.6 0.15 160)">[sing-box]</span>')
  }

  return (
    <div className="max-w-3xl mx-auto py-6">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-medium">Logs</h1>
        <motion.button
          onClick={clearLogs}
          className="px-3 py-1.5 text-sm rounded-lg bg-muted hover:bg-accent text-muted-foreground hover:text-foreground transition-colors flex items-center gap-1.5"
          whileTap={{ scale: 0.95 }}
        >
          <Trash2 size={14} />
          Clear
        </motion.button>
      </div>

      <div className="bg-zinc-950 rounded-2xl border border-zinc-800 overflow-hidden">
        {/* macOS-style title bar */}
        <div className="flex items-center gap-2 px-4 py-2.5 border-b border-zinc-800">
          <div className="w-3 h-3 rounded-full bg-red-500/80" />
          <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
          <div className="w-3 h-3 rounded-full bg-green-500/80" />
          <span className="ml-2 text-xs text-zinc-500 font-mono">core.log</span>
        </div>

        {/* Log content */}
        <div
          ref={scrollRef}
          className="h-[500px] overflow-y-auto p-4 font-mono text-xs leading-relaxed"
        >
          {logs.length === 0 ? (
            <div className="text-zinc-600 text-center py-16">
              <p>No logs yet</p>
              <p className="mt-1 text-zinc-700">Start the core to see logs</p>
            </div>
          ) : (
            logs.map((log, i) => (
              <div key={i} className="flex gap-3 py-0.5">
                <span className="text-zinc-600 flex-shrink-0 select-none">
                  {new Date(log.time).toLocaleTimeString('en-US', { hour12: false })}
                </span>
                <span
                  className="text-zinc-300 break-all"
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