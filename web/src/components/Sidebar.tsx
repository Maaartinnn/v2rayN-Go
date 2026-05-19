import { motion, AnimatePresence } from 'framer-motion'
import { Link, useLocation } from 'wouter'
import { Home, Server, FileText, Settings, Download, Route, ChevronLeft, ChevronRight, Shuffle, FolderOpen, ArrowUpFromLine } from 'lucide-react'
import { useStore } from '../store'
import { useT } from '../lib/i18n'

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const { isConnected } = useStore()
  const t = useT()
  const [location] = useLocation()

  const navItems = [
    { href: '/', icon: Home, label: t('nav.home') },
    { href: '/nodes', icon: Server, label: t('nav.nodes') },
    { href: '/import', icon: ArrowUpFromLine, label: t('nav.import') },
    { href: '/groups', icon: FolderOpen, label: t('groups.title') },
    { href: '/routing', icon: Route, label: t('nav.routing') },
    { href: '/strategy', icon: Shuffle, label: t('strategy.title') },
    { href: '/cores', icon: Download, label: t('nav.cores') },
    { href: '/logs', icon: FileText, label: t('nav.logs') },
    { href: '/settings', icon: Settings, label: t('nav.settings') },
  ]

  return (
    <motion.nav
      className="fixed left-0 top-0 bottom-0 flex flex-col z-50 border-r"
      style={{
        backgroundColor: 'var(--color-sidebar)',
        borderColor: 'var(--color-border)',
      }}
      initial={false}
      animate={{ width: collapsed ? 64 : 200 }}
      transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
    >
      {/* Logo & Brand */}
      <div className="flex items-center h-14 px-4 border-b" style={{ borderColor: 'var(--color-border)' }}>
        <div
          className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0"
          style={{ backgroundColor: 'var(--color-primary)' }}
        >
          <span className="text-sm font-bold" style={{ color: 'var(--color-primary-foreground)', fontFamily: 'var(--font-heading)' }}>V</span>
        </div>
        <AnimatePresence>
          {!collapsed && (
            <motion.span
              initial={{ opacity: 0, width: 0 }}
              animate={{ opacity: 1, width: 'auto' }}
              exit={{ opacity: 0, width: 0 }}
              className="ml-3 text-sm font-semibold whitespace-nowrap overflow-hidden"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              v2rayN-Go
            </motion.span>
          )}
        </AnimatePresence>
      </div>

      {/* Status Indicator */}
      <div className="flex items-center h-10 px-4 mx-3 mt-3 rounded-lg" style={{ backgroundColor: 'var(--color-card)' }}>
        <div className="relative shrink-0">
          <div
            className="w-2 h-2 rounded-full"
            style={{
              backgroundColor: isConnected ? 'var(--color-success)' : 'var(--color-stone)',
              animation: isConnected ? 'pulse-glow 2s infinite' : 'none',
            }}
          />
        </div>
        <AnimatePresence>
          {!collapsed && (
            <motion.span
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="ml-2.5 text-xs font-medium whitespace-nowrap"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {isConnected ? t('home.connected') : t('home.disconnected')}
            </motion.span>
          )}
        </AnimatePresence>
      </div>

      {/* Nav Items */}
      <div className="flex-1 flex flex-col gap-0.5 px-3 mt-4 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = location === item.href
          return (
            <Link key={item.href} href={item.href}>
              <motion.div
                className="relative flex items-center h-9 rounded-lg transition-colors cursor-pointer"
                style={{
                  backgroundColor: isActive ? 'var(--color-accent-dim)' : 'transparent',
                  color: isActive ? 'var(--color-accent-warm)' : 'var(--color-muted-foreground)',
                  justifyContent: collapsed ? 'center' : 'flex-start',
                  paddingLeft: collapsed ? 0 : '12px',
                  paddingRight: collapsed ? 0 : '12px',
                }}
                whileHover={{
                  backgroundColor: isActive ? undefined : 'var(--color-muted)',
                }}
                whileTap={{ scale: 0.97 }}
                title={collapsed ? item.label : undefined}
              >
                {isActive && (
                  <motion.div
                    layoutId="sidebar-active"
                    className="absolute left-0 top-1/2 -translate-y-1/2 w-0.75 h-4 rounded-r-full"
                    style={{ backgroundColor: 'var(--color-primary)' }}
                    transition={{ type: 'spring', stiffness: 500, damping: 35 }}
                  />
                )}
                <Icon size={17} strokeWidth={1.6} className="shrink-0" />
                <AnimatePresence>
                  {!collapsed && (
                    <motion.span
                      initial={{ opacity: 0, width: 0 }}
                      animate={{ opacity: 1, width: 'auto' }}
                      exit={{ opacity: 0, width: 0 }}
                      className="ml-2.5 text-[13px] font-medium whitespace-nowrap overflow-hidden"
                      style={{ fontFamily: 'var(--font-heading)' }}
                    >
                      {item.label}
                    </motion.span>
                  )}
                </AnimatePresence>
              </motion.div>
            </Link>
          )
        })}
      </div>

      {/* Collapse Toggle */}
      <div className="px-3 pb-4">
        <button
          onClick={onToggle}
          className="w-full flex items-center justify-center h-8 rounded-lg transition-colors cursor-pointer"
          style={{
            color: 'var(--color-text-muted)',
            backgroundColor: 'transparent',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.backgroundColor = 'var(--color-muted)'
            e.currentTarget.style.color = 'var(--color-muted-foreground)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.backgroundColor = 'transparent'
            e.currentTarget.style.color = 'var(--color-text-muted)'
          }}
        >
          {collapsed ? <ChevronRight size={14} /> : <ChevronLeft size={14} />}
        </button>
      </div>
    </motion.nav>
  )
}