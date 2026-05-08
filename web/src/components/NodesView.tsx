import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Wifi, WifiOff, RefreshCw, Trash2, Plus, Clipboard, Search, ScanLine, Layers } from 'lucide-react'
import { useStore } from '../store'
import type { Profile } from '../store'
import { profileApi, profileEnhancedApi, groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'

export function NodesView() {
  const { profiles, setProfiles, activeProfile, setActiveProfile } = useStore()
  const [loading, setLoading] = useState(false)
  const [importText, setImportText] = useState('')
  const [showImport, setShowImport] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedGroup, setSelectedGroup] = useState<string>('')
  const [groups, setGroups] = useState<{ ID: number; name: string }[]>([])
  const [dedupResult, setDedupResult] = useState<string>('')
  const t = useT()

  useEffect(() => {
    loadProfiles()
    loadGroups()
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
    const colors: Record<string, { bg: string; text: string }> = {
      vmess: { bg: 'rgba(106, 155, 204, 0.12)', text: '#6A9BCC' },
      vless: { bg: 'rgba(217, 119, 87, 0.12)', text: '#D97757' },
      trojan: { bg: 'rgba(201, 148, 58, 0.12)', text: '#C9943A' },
      shadowsocks: { bg: 'rgba(107, 143, 71, 0.12)', text: '#6B8F47' },
      hysteria2: { bg: 'rgba(192, 69, 58, 0.12)', text: '#C0453A' },
    }
    return colors[protocol] || { bg: 'var(--color-muted)', text: 'var(--color-muted-foreground)' }
  }

  const getLatencyDot = (result: string) => {
    if (!result || result === 'timeout') return 'var(--color-error)'
    const ms = parseInt(result)
    if (ms < 100) return 'var(--color-success)'
    if (ms < 300) return 'var(--color-warning)'
    return 'var(--color-error)'
  }

  const loadGroups = async () => {
    try {
      const res = await groupsApi.list()
      setGroups(res.data || [])
    } catch {
      setGroups([])
    }
  }

  const handleDedup = async () => {
    try {
      const res = await profileEnhancedApi.dedup()
      const data = res.data
      setDedupResult(`Removed ${data.removed} duplicates from ${data.total} nodes`)
      setTimeout(() => setDedupResult(''), 5000)
      await loadProfiles()
    } catch (err) {
      console.error('Dedup failed:', err)
    }
  }

  const handleImageImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const formData = new FormData()
    formData.append('image', file)
    try {
      const res = await profileEnhancedApi.importImage(formData)
      if (res.data.imported > 0) {
        await loadProfiles()
      }
    } catch (err) {
      console.error('Image import failed:', err)
    }
    // Reset input
    e.target.value = ''
  }

  const filteredProfiles = profiles.filter((p) => {
    const matchesSearch = !searchQuery || (() => {
      const q = searchQuery.toLowerCase()
      return (
        p.name.toLowerCase().includes(q) ||
        p.address.toLowerCase().includes(q) ||
        p.protocol.toLowerCase().includes(q) ||
        p.group_name.toLowerCase().includes(q)
      )
    })()
    const matchesGroup = !selectedGroup || p.group_name === selectedGroup
    return matchesSearch && matchesGroup
  })

  return (
    <div className="max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <h1
          className="text-xl font-semibold"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('nodes.title')}
        </h1>
        <div className="flex gap-2">
          <motion.button
            onClick={handlePingAll}
            disabled={loading}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors cursor-pointer"
            style={{
              backgroundColor: 'var(--color-muted)',
              borderColor: 'var(--color-border)',
              color: 'var(--color-muted-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
            whileTap={{ scale: 0.95 }}
          >
            <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
            {t('nodes.test_all')}
          </motion.button>
          <motion.button
            onClick={() => setShowImport(!showImport)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
            style={{
              backgroundColor: 'var(--color-primary)',
              color: 'var(--color-primary-foreground)',
              boxShadow: 'var(--shadow-btn)',
              fontFamily: 'var(--font-heading)',
            }}
            whileTap={{ scale: 0.95 }}
          >
            <Plus size={13} />
            {t('nodes.import')}
          </motion.button>
        </div>
      </div>

      {/* Toolbar: search + group filter + actions */}
      <div className="flex items-center gap-2 mb-4">
        <div className="relative flex-1">
          <Search
            size={14}
            className="absolute left-3 top-1/2 -translate-y-1/2"
            style={{ color: 'var(--color-text-muted)' }}
          />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t('common.search') + '...'}
            className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border"
            style={{
              backgroundColor: 'var(--color-overlay)',
              borderColor: 'var(--color-border)',
              color: 'var(--color-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
          />
        </div>
        {/* Group filter */}
        {groups.length > 0 && (
          <select
            value={selectedGroup}
            onChange={(e) => setSelectedGroup(e.target.value)}
            className="px-3 py-2 text-xs rounded-lg border cursor-pointer"
            style={{
              backgroundColor: 'var(--color-overlay)',
              borderColor: 'var(--color-border)',
              color: 'var(--color-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            <option value="">{t('nodes.all_groups')}</option>
            {groups.map((g) => (
              <option key={g.ID} value={g.name}>{g.name}</option>
            ))}
          </select>
        )}
        {/* Dedup button */}
        <motion.button
          onClick={handleDedup}
          className="flex items-center gap-1 px-2.5 py-2 text-xs font-medium rounded-lg border transition-colors cursor-pointer"
          style={{
            backgroundColor: 'var(--color-muted)',
            borderColor: 'var(--color-border)',
            color: 'var(--color-muted-foreground)',
            fontFamily: 'var(--font-heading)',
          }}
          whileTap={{ scale: 0.95 }}
          title={t('nodes.dedup')}
        >
          <Layers size={13} />
        </motion.button>
        {/* QR Image import */}
        <motion.button
          onClick={() => {
            const input = document.createElement('input')
            input.type = 'file'
            input.accept = 'image/*'
            input.onchange = (e) => handleImageImport(e as unknown as React.ChangeEvent<HTMLInputElement>)
            input.click()
          }}
          className="flex items-center gap-1 px-2.5 py-2 text-xs font-medium rounded-lg border transition-colors cursor-pointer"
          style={{
            backgroundColor: 'var(--color-muted)',
            borderColor: 'var(--color-border)',
            color: 'var(--color-muted-foreground)',
            fontFamily: 'var(--font-heading)',
          }}
          whileTap={{ scale: 0.95 }}
          title={t('nodes.import_qr')}
        >
          <ScanLine size={13} />
        </motion.button>
      </div>

      {/* Dedup result toast */}
      <AnimatePresence>
        {dedupResult && (
          <motion.div
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            className="mb-3 px-4 py-2 rounded-lg text-xs font-medium"
            style={{
              backgroundColor: 'var(--color-success-dim)',
              color: 'var(--color-success)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            {dedupResult}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Import panel */}
      <AnimatePresence>
        {showImport && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="mb-4 overflow-hidden"
          >
            <div
              className="rounded-xl border p-4"
              style={{
                backgroundColor: 'var(--color-card)',
                borderColor: 'var(--color-border)',
                boxShadow: 'var(--shadow-card)',
              }}
            >
              <textarea
                value={importText}
                onChange={(e) => setImportText(e.target.value)}
                placeholder={t('nodes.import_placeholder')}
                className="w-full h-24 rounded-lg p-3 text-sm resize-none border"
                style={{
                  backgroundColor: 'var(--color-muted)',
                  borderColor: 'var(--color-border-subtle)',
                  color: 'var(--color-foreground)',
                  fontFamily: 'var(--font-mono)',
                }}
              />
              <p
                className="text-xs mt-2"
                style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
              >
                {t('nodes.import_base64')}
              </p>
              <div className="flex justify-end gap-2 mt-3">
                <button
                  onClick={() => { setShowImport(false); setImportText('') }}
                  className="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
                  style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                >
                  {t('nodes.cancel')}
                </button>
                <button
                  onClick={handleImport}
                  className="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
                  style={{
                    backgroundColor: 'var(--color-primary)',
                    color: 'var(--color-primary-foreground)',
                    fontFamily: 'var(--font-heading)',
                  }}
                >
                  {t('nodes.confirm')}
                </button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Node list */}
      <div className="space-y-1.5">
        <AnimatePresence mode="popLayout">
          {filteredProfiles.length === 0 ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center py-20"
            >
              <Clipboard
                size={32}
                className="mx-auto mb-3"
                style={{ color: 'var(--color-text-muted)' }}
              />
              <p
                className="text-sm"
                style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
              >
                {searchQuery ? t('common.no_data') : t('nodes.no_nodes')}
              </p>
              {!searchQuery && (
                <p
                  className="text-xs mt-1"
                  style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
                >
                  {t('nodes.import_hint')}
                </p>
              )}
            </motion.div>
          ) : (
            filteredProfiles.map((profile, index) => {
              const protoColor = getProtocolColor(profile.protocol)
              return (
                <motion.div
                  key={profile.ID}
                  layout
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, x: -16 }}
                  transition={{ duration: 0.2, delay: index * 0.02 }}
                  onClick={() => handleSelect(profile)}
                  className="rounded-xl border px-4 py-3 cursor-pointer transition-colors"
                  style={{
                    backgroundColor: activeProfile?.ID === profile.ID
                      ? 'var(--color-accent-dim)'
                      : 'var(--color-card)',
                    borderColor: activeProfile?.ID === profile.ID
                      ? 'var(--color-primary)'
                      : 'var(--color-border)',
                    boxShadow: 'var(--shadow-card)',
                  }}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3 min-w-0">
                      <div
                        className="w-2 h-2 rounded-full flex-shrink-0"
                        style={{ backgroundColor: getLatencyDot(profile.test_result) }}
                      />
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <span
                            className="text-sm font-medium truncate"
                            style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                          >
                            {profile.name}
                          </span>
                          <span
                            className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                            style={{
                              backgroundColor: protoColor.bg,
                              color: protoColor.text,
                              fontFamily: 'var(--font-heading)',
                            }}
                          >
                            {profile.protocol}
                          </span>
                        </div>
                        <p
                          className="text-xs mt-0.5 truncate"
                          style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                        >
                          {profile.address}:{profile.port}
                          {profile.group_name && (
                            <span style={{ fontFamily: 'var(--font-heading)' }}> · {profile.group_name}</span>
                          )}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 flex-shrink-0">
                      {profile.test_result && (
                        <span
                          className="text-xs"
                          style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                        >
                          {profile.test_result}
                        </span>
                      )}
                      {activeProfile?.ID === profile.ID ? (
                        <Wifi size={14} style={{ color: 'var(--color-success)' }} />
                      ) : (
                        <WifiOff size={14} style={{ color: 'var(--color-text-muted)' }} />
                      )}
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDelete(profile.ID) }}
                        className="p-1 rounded-md transition-colors cursor-pointer"
                        style={{ color: 'var(--color-text-muted)' }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.color = 'var(--color-error)'
                          e.currentTarget.style.backgroundColor = 'var(--color-error-dim)'
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.color = 'var(--color-text-muted)'
                          e.currentTarget.style.backgroundColor = 'transparent'
                        }}
                      >
                        <Trash2 size={12} />
                      </button>
                    </div>
                  </div>
                </motion.div>
              )
            })
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}