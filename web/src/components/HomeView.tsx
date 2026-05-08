import { motion } from 'framer-motion'
import { Power, Zap, ArrowUp, ArrowDown } from 'lucide-react'
import { useStore } from '../store'
import { coreApi } from '../lib/api'
import { useT } from '../lib/i18n'

export function HomeView() {
  const { isConnected, activeProfile, metrics } = useStore()
  const t = useT()

  const handleToggle = async () => {
    try {
      if (isConnected) {
        await coreApi.stop('xray')
      } else {
        await coreApi.start('xray', '')
      }
    } catch (err) {
      console.error('Toggle failed:', err)
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B/s`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB/s`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB/s`
  }

  return (
    <div className="flex items-center justify-center min-h-[calc(100vh-8rem)]">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
        className="w-[420px] rounded-2xl border p-8"
        style={{
          backgroundColor: 'var(--color-card)',
          borderColor: 'var(--color-border)',
          boxShadow: 'var(--shadow-elevated)',
        }}
      >
        {/* Status indicator */}
        <div className="flex items-center justify-center mb-8">
          <div className="relative">
            <div
              className="w-3 h-3 rounded-full"
              style={{
                backgroundColor: isConnected ? 'var(--color-success)' : 'var(--color-stone)',
                animation: isConnected ? 'pulse-glow 2s infinite' : 'none',
              }}
            />
          </div>
          <span
            className="ml-3 text-sm"
            style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {isConnected ? t('home.connected') : t('home.disconnected')}
          </span>
        </div>

        {/* Power button */}
        <div className="flex justify-center mb-8">
          <motion.button
            onClick={handleToggle}
            className="w-24 h-24 rounded-full flex items-center justify-center transition-colors cursor-pointer"
            style={{
              backgroundColor: isConnected ? 'var(--color-success-dim)' : 'var(--color-muted)',
              color: isConnected ? 'var(--color-success)' : 'var(--color-muted-foreground)',
            }}
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
          >
            <Power size={36} strokeWidth={1.5} />
          </motion.button>
        </div>

        {/* Current node info */}
        <div className="text-center mb-6">
          <p
            className="text-lg font-medium"
            style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {activeProfile ? activeProfile.name : t('home.no_node')}
          </p>
          {activeProfile && (
            <p
              className="text-sm mt-1"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {activeProfile.protocol.toUpperCase()} · {activeProfile.address}:{activeProfile.port}
            </p>
          )}
        </div>

        {/* Metrics */}
        {isConnected && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="grid grid-cols-2 gap-4 pt-5 border-t"
            style={{ borderColor: 'var(--color-border)' }}
          >
            <div className="flex items-center gap-2">
              <ArrowUp size={14} style={{ color: 'var(--color-success)' }} />
              <div>
                <p
                  className="text-xs"
                  style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                >
                  {t('home.upload')}
                </p>
                <p
                  className="text-sm font-medium"
                  style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-mono)' }}
                >
                  {formatBytes(metrics.upload_speed)}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <ArrowDown size={14} style={{ color: 'var(--color-warning)' }} />
              <div>
                <p
                  className="text-xs"
                  style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                >
                  {t('home.download')}
                </p>
                <p
                  className="text-sm font-medium"
                  style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-mono)' }}
                >
                  {formatBytes(metrics.download_speed)}
                </p>
              </div>
            </div>
          </motion.div>
        )}

        {/* Ping */}
        {activeProfile?.test_result && (
          <div
            className="flex items-center justify-center gap-2 mt-4 pt-4 border-t"
            style={{ borderColor: 'var(--color-border)' }}
          >
            <Zap size={14} style={{ color: 'var(--color-muted-foreground)' }} />
            <span
              className="text-sm"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('home.latency')}: {activeProfile.test_result}
            </span>
          </div>
        )}
      </motion.div>
    </div>
  )
}