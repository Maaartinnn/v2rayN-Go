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
    <div className="flex items-center justify-center min-h-full">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
        className="w-[400px] bg-card rounded-2xl border border-border p-8"
      >
        {/* Status indicator */}
        <div className="flex items-center justify-center mb-8">
          <div className="relative">
            <div
              className={`w-3 h-3 rounded-full ${
                isConnected ? 'bg-emerald animate-ping' : 'bg-stone'
              }`}
            />
            <div
              className={`w-3 h-3 rounded-full absolute inset-0 ${
                isConnected ? 'bg-emerald' : 'bg-stone'
              }`}
            />
          </div>
          <span className="ml-3 text-sm text-muted-foreground">
            {isConnected ? t('home.connected') : t('home.disconnected')}
          </span>
        </div>

        {/* Power button */}
        <div className="flex justify-center mb-8">
          <motion.button
            onClick={handleToggle}
            className={`w-24 h-24 rounded-full flex items-center justify-center transition-colors ${
              isConnected
                ? 'bg-emerald/10 text-emerald hover:bg-emerald/20'
                : 'bg-muted text-muted-foreground hover:bg-accent'
            }`}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            <Power size={36} strokeWidth={1.5} />
          </motion.button>
        </div>

        {/* Current node info */}
        <div className="text-center mb-6">
          <p className="text-lg font-medium">
            {activeProfile ? activeProfile.name : t('home.no_node')}
          </p>
          {activeProfile && (
            <p className="text-sm text-muted-foreground mt-1">
              {activeProfile.protocol.toUpperCase()} · {activeProfile.address}:{activeProfile.port}
            </p>
          )}
        </div>

        {/* Metrics */}
        {isConnected && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="grid grid-cols-2 gap-4 pt-4 border-t border-border"
          >
            <div className="flex items-center gap-2">
              <ArrowUp size={14} className="text-emerald" />
              <div>
                <p className="text-xs text-muted-foreground">{t('home.upload')}</p>
                <p className="text-sm font-medium">{formatBytes(metrics.upload_speed)}</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <ArrowDown size={14} className="text-amber" />
              <div>
                <p className="text-xs text-muted-foreground">{t('home.download')}</p>
                <p className="text-sm font-medium">{formatBytes(metrics.download_speed)}</p>
              </div>
            </div>
          </motion.div>
        )}

        {/* Ping */}
        {activeProfile?.test_result && (
          <div className="flex items-center justify-center gap-2 mt-4 pt-4 border-t border-border">
            <Zap size={14} className="text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {t('home.latency')}: {activeProfile.test_result}
            </span>
          </div>
        )}
      </motion.div>
    </div>
  )
}