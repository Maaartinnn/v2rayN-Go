import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Wifi, WifiOff, RefreshCw, Trash2, Search, Layers, FolderOpen, Link, Edit3 } from 'lucide-react'
import { useStore } from '../store'
import type { Profile } from '../store'
import { profileApi, profileEnhancedApi, groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { DeleteConfirmBanner } from './DeleteConfirmBanner'
import { NodeEditForm } from './NodeEditForm'

interface NodeGroupItem {
  ID: number
  uuid: string
  alias: string
  is_subscription: boolean
  node_count: number
}

export function NodesView() {
  const { profiles, setProfiles, activeProfile, setActiveProfile } = useStore()
  const [loading, setLoading] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedGroupId, setSelectedGroupId] = useState<number>(0)
  const [groups, setGroups] = useState<NodeGroupItem[]>([])
  const [dedupResult, setDedupResult] = useState<string>('')
  const [deleteTargetId, setDeleteTargetId] = useState<number | null>(null)
  const [editId, setEditId] = useState<number | null>(null)
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
      setDeleteTargetId(null)
      await loadProfiles()
    } catch (err) {
      console.error('Delete failed:', err)
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
      const res = await profileEnhancedApi.dedup(selectedGroupId || undefined)
      const data = res.data
      setDedupResult(t('nodes.dedup_result', { removed: data.removed, total: data.total }))
      setTimeout(() => setDedupResult(''), 5000)
      await loadProfiles()
      await loadGroups()
    } catch (err) {
      console.error('Dedup failed:', err)
    }
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
    const matchesGroup = selectedGroupId === 0 || p.group_id === selectedGroupId
    return matchesSearch && matchesGroup
  })

  const displayName = (g: NodeGroupItem) => g.alias || t('groups.default_name')

  return (
    <div className="flex gap-6 max-w-5xl mx-auto">
      {/* Left: Node List */}
      <div className="flex-1 min-w-0">
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
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium cursor-pointer btn-ghost"
              style={{
                fontFamily: 'var(--font-heading)',
              }}
              whileTap={{ scale: 0.95 }}
            >
              <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
              {t('nodes.test_all')}
            </motion.button>
          </div>
        </div>

        {/* Toolbar: search + dedup */}
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

        {/* Node list */}
        <div className="space-y-1.5" key={selectedGroupId}>
          <AnimatePresence>
            {filteredProfiles.length === 0 ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="text-center py-20"
              >
                <Layers
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
              </motion.div>
            ) : (
              filteredProfiles.map((profile, index) => {
                const protoColor = getProtocolColor(profile.protocol)
                return (
                  <motion.div
                    key={profile.ID}
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, x: -16 }}
                    transition={{ duration: 0.2, delay: index * 0.02 }}
                  >
                    <div
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
                            onClick={(e) => { e.stopPropagation(); setEditId(editId === profile.ID ? null : profile.ID) }}
                            className="p-1 rounded-md transition-colors cursor-pointer"
                            style={{ color: 'var(--color-text-muted)' }}
                            onMouseEnter={(e) => {
                              e.currentTarget.style.color = 'var(--color-accent-warm)'
                              e.currentTarget.style.backgroundColor = 'var(--color-accent-dim)'
                            }}
                            onMouseLeave={(e) => {
                              e.currentTarget.style.color = 'var(--color-text-muted)'
                              e.currentTarget.style.backgroundColor = 'transparent'
                            }}
                          >
                            <Edit3 size={12} />
                          </button>
                          <button
                            onClick={(e) => { e.stopPropagation(); setDeleteTargetId(profile.ID) }}
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
                    </div>
                    {/* Delete confirm banner */}
                    <DeleteConfirmBanner
                      visible={deleteTargetId === profile.ID}
                      message={t('nodes.delete_confirm', { name: profile.name })}
                      onConfirm={() => handleDelete(profile.ID)}
                      onCancel={() => setDeleteTargetId(null)}
                    />
                    {/* Edit panel (inline, like GroupsView) */}
                    <AnimatePresence>
                      {editId === profile.ID && (
                        <NodeEditForm
                          editData={profile}
                          groupId={profile.group_id}
                          onClose={() => setEditId(null)}
                          onSaved={loadProfiles}
                        />
                      )}
                    </AnimatePresence>
                  </motion.div>
                )
              })
            )}
          </AnimatePresence>
        </div>
      </div>

      {/* Right: Group Selection Panel */}
      <div className="w-64 flex-shrink-0">
        <div className="sticky top-20">
          <div
            className="rounded-xl border overflow-hidden"
            style={{
              backgroundColor: 'var(--color-card)',
              borderColor: 'var(--color-border)',
              boxShadow: 'var(--shadow-card)',
            }}
          >
            <div className="p-2 space-y-1">
              {/* "All Groups" option */}
              <motion.button
                onClick={() => setSelectedGroupId(0)}
                className="w-full flex items-center gap-2 px-3 py-2 rounded-lg transition-colors text-left cursor-pointer"
                style={{
                  backgroundColor: selectedGroupId === 0 ? 'var(--color-accent-dim)' : 'transparent',
                  borderColor: selectedGroupId === 0 ? 'var(--color-primary)' : 'transparent',
                  borderWidth: selectedGroupId === 0 ? 1 : 0,
                  borderStyle: 'solid',
                }}
                whileTap={{ scale: 0.98 }}
              >
                <div
                  className="w-5 h-5 rounded flex items-center justify-center flex-shrink-0"
                  style={{ backgroundColor: 'var(--color-muted)' }}
                >
                  <Layers size={10} style={{ color: 'var(--color-muted-foreground)' }} />
                </div>
                <div className="min-w-0 flex-1">
                  <span
                    className="text-xs font-medium truncate block"
                    style={{
                      color: selectedGroupId === 0 ? 'var(--color-accent-warm)' : 'var(--color-foreground)',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    {t('nodes.all_groups')}
                  </span>
                </div>
                <span
                  className="text-[9px] flex-shrink-0"
                  style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-mono)' }}
                >
                  {profiles.length}
                </span>
              </motion.button>

              {groups.map((group) => {
                const isSelected = selectedGroupId === group.ID
                return (
                  <motion.button
                    key={group.ID}
                    onClick={() => setSelectedGroupId(group.ID)}
                    className="w-full flex items-center gap-2 px-3 py-2 rounded-lg transition-colors text-left cursor-pointer"
                    style={{
                      backgroundColor: isSelected ? 'var(--color-accent-dim)' : 'transparent',
                      borderColor: isSelected ? 'var(--color-primary)' : 'transparent',
                      borderWidth: isSelected ? 1 : 0,
                      borderStyle: 'solid',
                    }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <div
                      className="w-5 h-5 rounded flex items-center justify-center flex-shrink-0"
                      style={{
                        backgroundColor: group.is_subscription
                          ? 'rgba(217, 119, 87, 0.12)'
                          : 'var(--color-muted)',
                      }}
                    >
                      {group.is_subscription ? (
                        <Link size={10} style={{ color: '#D97757' }} />
                      ) : (
                        <FolderOpen size={10} style={{ color: 'var(--color-muted-foreground)' }} />
                      )}
                    </div>
                    <div className="min-w-0 flex-1">
                      <span
                        className="text-xs font-medium truncate block"
                        style={{
                          color: isSelected ? 'var(--color-accent-warm)' : 'var(--color-foreground)',
                          fontFamily: 'var(--font-heading)',
                        }}
                      >
                        {displayName(group)}
                      </span>
                    </div>
                    <span
                      className="text-[9px] flex-shrink-0"
                      style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-mono)' }}
                    >
                      {group.node_count}
                    </span>
                  </motion.button>
                )
              })}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}