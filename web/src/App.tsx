import { useState, useEffect } from 'react'
import { Sidebar } from './components/Sidebar'
import { HomeView } from './components/HomeView'
import { NodesView } from './components/NodesView'
import { ImportView } from './components/ImportView'
import { GroupsView } from './components/GroupsView'
import { LogConsole } from './components/LogConsole'
import { SettingsView } from './components/SettingsView'
import { CoresView } from './components/CoresView'
import { RoutingView } from './components/RoutingView'
import { StrategyGroupView } from './components/StrategyGroupView'
import { ErrorBoundary } from './components/ErrorBoundary'
import { useWebSocket } from './lib/useWebSocket'
import { useStore } from './store'
import { useT, initTheme } from './lib/i18n'
import { coreApi } from './lib/api'
import { motion, AnimatePresence } from 'framer-motion'

const views: { [key: string]: React.FC } = {
  home: HomeView,
  nodes: NodesView,
  import: ImportView,
  groups: GroupsView,
  logs: LogConsole,
  settings: SettingsView,
  updater: CoresView,
  routing: RoutingView,
  strategy: StrategyGroupView,
}

export default function App() {
  useWebSocket()
  const { currentView, isConnected, activeProfile, setCoreStatuses } = useStore()
  const t = useT()
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  // Initialize theme on mount
  useEffect(() => {
    initTheme()
  }, [])

  // Fetch initial core status on mount
  useEffect(() => {
    coreApi.status().then((res) => {
      setCoreStatuses(res.data)
    }).catch(() => {
      // server may not be ready yet, WebSocket will handle updates
    })
  }, [setCoreStatuses])

  const View = views[currentView] || HomeView

  // Map view IDs to page titles
  const viewTitles: Record<string, string> = {
    home: t('nav.home'),
    nodes: t('nav.nodes'),
    import: t('nav.import'),
    groups: t('groups.title'),
    logs: t('nav.logs'),
    settings: t('nav.settings'),
    updater: t('nav.cores'),
    routing: t('nav.routing'),
    strategy: t('strategy.title'),
  }

  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--color-background)' }}>
      <Sidebar collapsed={sidebarCollapsed} onToggle={() => setSidebarCollapsed(!sidebarCollapsed)} />

      {/* Top Header Bar */}
      <header
        className="fixed top-0 right-0 h-14 flex items-center justify-between px-6 z-40 border-b"
        style={{
          left: sidebarCollapsed ? 64 : 200,
          backgroundColor: 'var(--color-background)',
          borderColor: 'var(--color-border)',
          transition: 'left 0.25s cubic-bezier(0.16, 1, 0.3, 1)',
        }}
      >
        {/* Page Title */}
        <h2
          className="text-sm font-medium"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {viewTitles[currentView] || ''}
        </h2>

        {/* Right side: minimal status indicator */}
        <div className="flex items-center gap-3">
          {activeProfile && (
            <span
              className="text-xs"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {activeProfile.name}
            </span>
          )}
          <div className="flex items-center gap-1.5">
            <div
              className="w-1.5 h-1.5 rounded-full"
              style={{
                backgroundColor: isConnected ? 'var(--color-success)' : 'var(--color-stone)',
              }}
            />
            <span
              className="text-[11px] font-medium"
              style={{
                color: isConnected ? 'var(--color-success)' : 'var(--color-text-muted)',
                fontFamily: 'var(--font-heading)',
              }}
            >
              {isConnected ? 'ON' : 'OFF'}
            </span>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main
        className="pt-14 min-h-screen"
        style={{
          marginLeft: sidebarCollapsed ? 64 : 200,
          transition: 'margin-left 0.25s cubic-bezier(0.16, 1, 0.3, 1)',
        }}
      >
        <div className="p-6">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentView}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
            >
              <ErrorBoundary>
                <View />
              </ErrorBoundary>
            </motion.div>
          </AnimatePresence>
        </div>
      </main>
    </div>
  )
}