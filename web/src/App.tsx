import { useState, useEffect, lazy, Suspense } from 'react'
import { Switch, Route, useLocation } from 'wouter'
import { Sidebar } from './components/Sidebar'
import { HomeView } from './components/HomeView'
import { NodesView } from './components/NodesView'
import { ErrorBoundary } from './components/ErrorBoundary'
import { useWebSocket } from './lib/useWebSocket'
import { useStore } from './store'
import { useT, initTheme } from './lib/i18n'
import { coreApi } from './lib/api'
import { motion, AnimatePresence } from 'framer-motion'

// 动态导入非首屏组件（代码分割）
const ImportView = lazy(() => import('./components/ImportView').then(m => ({ default: m.ImportView })))
const GroupsView = lazy(() => import('./components/GroupsView').then(m => ({ default: m.GroupsView })))
const LogConsole = lazy(() => import('./components/LogConsole').then(m => ({ default: m.LogConsole })))
const SettingsView = lazy(() => import('./components/SettingsView').then(m => ({ default: m.SettingsView })))
const CoresView = lazy(() => import('./components/CoresView').then(m => ({ default: m.CoresView })))
const RoutingView = lazy(() => import('./components/RoutingView').then(m => ({ default: m.RoutingView })))
const StrategyGroupView = lazy(() => import('./components/StrategyGroupView').then(m => ({ default: m.StrategyGroupView })))

function PageLoader() {
  return (
    <div className="flex items-center justify-center h-64">
      <div
        className="w-6 h-6 rounded-full border-2 animate-spin"
        style={{
          borderColor: 'var(--color-border)',
          borderTopColor: 'var(--color-primary)',
        }}
      />
    </div>
  )
}

export default function App() {
  useWebSocket()
  const { isConnected, activeProfile, setCoreStatuses } = useStore()
  const t = useT()
  const [location] = useLocation()
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

  // Map pathnames to page titles
  const pathTitles: Record<string, string> = {
    '/': t('nav.home'),
    '/nodes': t('nav.nodes'),
    '/import': t('nav.import'),
    '/groups': t('groups.title'),
    '/logs': t('nav.logs'),
    '/settings': t('nav.settings'),
    '/cores': t('nav.cores'),
    '/routing': t('nav.routing'),
    '/strategy': t('strategy.title'),
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
          {pathTitles[location] || ''}
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
              key={location}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
            >
              <ErrorBoundary>
                <Suspense fallback={<PageLoader />}>
                  <Switch location={location}>
                    <Route path="/" component={HomeView} />
                    <Route path="/nodes" component={NodesView} />
                    <Route path="/import" component={ImportView} />
                    <Route path="/groups" component={GroupsView} />
                    <Route path="/logs" component={LogConsole} />
                    <Route path="/settings" component={SettingsView} />
                    <Route path="/cores" component={CoresView} />
                    <Route path="/routing" component={RoutingView} />
                    <Route path="/strategy" component={StrategyGroupView} />
                    <Route component={HomeView} />
                  </Switch>
                </Suspense>
              </ErrorBoundary>
            </motion.div>
          </AnimatePresence>
        </div>
      </main>
    </div>
  )
}