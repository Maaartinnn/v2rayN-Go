import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Wifi, WifiOff, RefreshCw, Trash2, Plus, Clipboard } from 'lucide-react'
import { useStore } from '../store'
import type { Profile } from '../store'
import { profileApi } from '../lib/api'

export function NodesView() {
  const { profiles, setProfiles, activeProfile, setActiveProfile } = useStore()
  const [loading, setLoading] = useState(false)
  const [importText, setImportText] = useState('')
  const [showImport, setShowImport] = useState(false)

  useEffect(() => {
    loadProfiles()
  }, [])

  const loadProfiles = async () => {
    try {
      const res = await profileApi.list()
      setProfiles(res.data || [])
    } catch {
      setProfiles([])
    }
  }

  const handleSelect = async (profile: Profile) => {
    try {
      await profileApi.select(profile.ID)
      setActiveProfile(profile)
    } catch (err) {
      console.error('Select failed:', err)
    }
  }

  const handlePingAll = async () => {
    setLoading(true)
    try {
      await profileApi.pingAll()
      await loadProfiles()
    } catch (err) {
      console.error('Ping failed:', err)
    }
    setLoading(false)
  }

  const handleDelete = async (id: number) => {
    try {
      await profileApi.delete(id)
      await loadProfiles()
    } catch (err) {
      console.error('Delete failed:', err)
    }
  }

  const handleImport = async () => {
    if (!importText.trim()) return
    try {
      await profileApi.importLinks(importText)
      setImportText('')
      setShowImport(false)
      await loadProfiles()
    } catch (err) {
      console.error('Import failed:', err)
    }
  }

  const getProtocolColor = (protocol: string) => {
    const colors: Record<string, string> = {
      vmess: 'bg-blue-500/10 text-blue-500',
      vless: 'bg-purple-500/10 text-purple-500',
      trojan: 'bg-orange-500/10 text-orange-500',
      shadowsocks: 'bg-green-500/10 text-green-500',
      hysteria2: 'bg-pink-500/10 text-pink-500',
    }
    return colors[protocol] || 'bg-muted text-muted-foreground'
  }

  const getLatencyDot = (result: string) => {
    if (!result || result === 'timeout') return 'bg-red-400'
    const ms = parseInt(result)
    if (ms < 100) return 'bg-emerald'
    if (ms < 300) return 'bg-amber'
    return 'bg-red-400'
  }

  return (
    <div className="max-w-2xl mx-auto py-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-medium">Nodes</h1>
        <div className="flex gap-2">
          <motion.button
            onClick={handlePingAll}
            disabled={loading}
            className="px-3 py-1.5 text-sm rounded-lg bg-muted hover:bg-accent text-muted-foreground hover:text-foreground transition-colors flex items-center gap-1.5"
            whileTap={{ scale: 0.95 }}
          >
            <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
            Test All
          </motion.button>
          <motion.button
            onClick={() => setShowImport(!showImport)}
            className="px-3 py-1.5 text-sm rounded-lg bg-primary text-primary-foreground hover:opacity-90 transition-opacity flex items-center gap-1.5"
            whileTap={{ scale: 0.95 }}
          >
            <Plus size={14} />
            Import
          </motion.button>
        </div>
      </div>

      {/* Import panel */}
      <AnimatePresence>
        {showImport && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="mb-4 overflow-hidden"
          >
            <div className="bg-card rounded-xl border border-border p-4">
              <textarea
                value={importText}
                onChange={(e) => setImportText(e.target.value)}
                placeholder="Paste share links here (vmess://, vless://, trojan://, ss://, ...)&#10;One link per line"
                className="w-full h-24 bg-muted rounded-lg p-3 text-sm resize-none focus:outline-none focus:ring-1 focus:ring-border"
              />
              <div className="flex justify-end gap-2 mt-2">
                <button
                  onClick={() => { setShowImport(false); setImportText('') }}
                  className="px-3 py-1.5 text-sm rounded-lg text-muted-foreground hover:text-foreground"
                >
                  Cancel
                </button>
                <button
                  onClick={handleImport}
                  className="px-3 py-1.5 text-sm rounded-lg bg-primary text-primary-foreground hover:opacity-90"
                >
                  Import
                </button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Node list */}
      <div className="space-y-2">
        <AnimatePresence mode="popLayout">
          {profiles.length === 0 ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center py-16 text-muted-foreground"
            >
              <Clipboard size={32} className="mx-auto mb-3 opacity-50" />
              <p className="text-sm">No nodes yet</p>
              <p className="text-xs mt-1">Import share links or add a subscription</p>
            </motion.div>
          ) : (
            profiles.map((profile) => (
              <motion.div
                key={profile.ID}
                layout
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, x: -20 }}
                onClick={() => handleSelect(profile)}
                className={`bg-card rounded-xl border border-border p-4 cursor-pointer transition-colors hover:bg-accent/50 ${
                  activeProfile?.ID === profile.ID ? 'ring-1 ring-primary/20' : ''
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3 min-w-0">
                    {/* Latency dot */}
                    <div className={`w-2 h-2 rounded-full flex-shrink-0 ${getLatencyDot(profile.test_result)}`} />

                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium truncate">{profile.name}</span>
                        <span className={`text-[10px] px-1.5 py-0.5 rounded-md ${getProtocolColor(profile.protocol)}`}>
                          {profile.protocol}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground mt-0.5 truncate">
                        {profile.address}:{profile.port}
                        {profile.group_name && ` · ${profile.group_name}`}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 flex-shrink-0">
                    {profile.test_result && (
                      <span className="text-xs text-muted-foreground">{profile.test_result}</span>
                    )}
                    {activeProfile?.ID === profile.ID ? (
                      <Wifi size={14} className="text-emerald" />
                    ) : (
                      <WifiOff size={14} className="text-muted-foreground opacity-30" />
                    )}
                    <button
                      onClick={(e) => { e.stopPropagation(); handleDelete(profile.ID) }}
                      className="p-1 rounded-md hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                </div>
              </motion.div>
            ))
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}