import { motion } from 'framer-motion'
import { Home, Server, FileText, Settings, Download } from 'lucide-react'
import { useStore } from '../store'

const navItems = [
  { id: 'home', icon: Home, label: 'Home' },
  { id: 'nodes', icon: Server, label: 'Nodes' },
  { id: 'logs', icon: FileText, label: 'Logs' },
  { id: 'settings', icon: Settings, label: 'Settings' },
  { id: 'updater', icon: Download, label: 'Cores' },
]

export function Sidebar() {
  const { currentView, setCurrentView, isConnected } = useStore()

  return (
    <nav className="fixed left-0 top-0 bottom-0 w-16 bg-card border-r border-border flex flex-col items-center py-4 z-50">
      {/* Logo */}
      <div className="mb-8">
        <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
          <span className="text-primary-foreground text-xs font-bold">V</span>
        </div>
      </div>

      {/* Status indicator */}
      <div className="mb-6">
        <div
          className={`w-2.5 h-2.5 rounded-full ${
            isConnected
              ? 'bg-emerald animate-ping'
              : 'bg-stone'
          }`}
        />
        <div
          className={`w-2.5 h-2.5 rounded-full absolute ${
            isConnected ? 'bg-emerald' : 'bg-stone'
          }`}
          style={{ marginTop: -10 }}
        />
      </div>

      {/* Nav items */}
      <div className="flex-1 flex flex-col items-center gap-1">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = currentView === item.id
          return (
            <motion.button
              key={item.id}
              onClick={() => setCurrentView(item.id)}
              className={`relative w-10 h-10 rounded-xl flex items-center justify-center transition-colors ${
                isActive
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
              }`}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              title={item.label}
            >
              <Icon size={18} strokeWidth={1.5} />
              {isActive && (
                <motion.div
                  layoutId="sidebar-indicator"
                  className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-5 bg-primary rounded-full"
                  transition={{ type: 'spring', stiffness: 500, damping: 30 }}
                />
              )}
            </motion.button>
          )
        })}
      </div>
    </nav>
  )
}