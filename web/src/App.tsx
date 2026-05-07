import React from 'react'
import { Sidebar } from './components/Sidebar'
import { HomeView } from './components/HomeView'
import { NodesView } from './components/NodesView'
import { LogConsole } from './components/LogConsole'
import { useWebSocket } from './lib/useWebSocket'
import { useStore } from './store'
import { motion, AnimatePresence } from 'framer-motion'

function SettingsView() {
  return (
    <div className="max-w-2xl mx-auto py-6">
      <h1 className="text-xl font-medium mb-6">Settings</h1>
      <div className="bg-card rounded-2xl border border-border p-6">
        <p className="text-sm text-muted-foreground">Settings panel coming soon...</p>
      </div>
    </div>
  )
}

function UpdaterView() {
  return (
    <div className="max-w-2xl mx-auto py-6">
      <h1 className="text-xl font-medium mb-6">Core Manager</h1>
      <div className="bg-card rounded-2xl border border-border p-6">
        <p className="text-sm text-muted-foreground">Core download & update panel coming soon...</p>
      </div>
    </div>
  )
}

const views: Record<string, React.FC> = {
  home: HomeView,
  nodes: NodesView,
  logs: LogConsole,
  settings: SettingsView,
  updater: UpdaterView,
}

export default function App() {
  useWebSocket()
  const { currentView } = useStore()
  const View = views[currentView] || HomeView

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <main className="ml-16 p-6">
        <AnimatePresence mode="wait">
          <motion.div
            key={currentView}
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.2 }}
          >
            <View />
          </motion.div>
        </AnimatePresence>
      </main>
    </div>
  )
}