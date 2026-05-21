import { useEffect, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Wifi, WifiOff, RefreshCw, Trash2, Search, Layers, FolderOpen, Link, Edit3 } from 'lucide-react'
import { useStore } from '../store'
import type { Profile } from '../store'
import { profileApi, profileEnhancedApi, groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { DeleteConfirmBanner } from './ui/DeleteConfirmBanner'
import { NodeEditForm } from './NodeEditForm'
import { RightDrawer } from './ui/RightDrawer'

// useDebounce 防抖 Hook：延迟更新值，避免频繁触发后端请求。
// 用于搜索输入框，用户停止输入 250ms 后才发起 API 请求。
function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}

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
  const [selectedGroupUUID, setSelectedGroupUUID] = useState<string>('')
  const [groups, setGroups] = useState<NodeGroupItem[]>([])
  const [dedupResult, setDedupResult] = useState<string>('')
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null)
  const [editProfile, setEditProfile] = useState<Profile | null>(null)

  // Multi-selection state
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [lastClickedId, setLastClickedId] = useState<number | null>(null)

  const t = useT()

  // 防抖搜索：用户停止输入 250ms 后才触发后端请求，避免频繁 API 调用。
  const debouncedSearch = useDebounce(searchQuery, 250)

  // 初始加载：获取分组列表
  useEffect(() => {
    loadGroups()
  }, [])

  // 后端驱动筛选：分组或搜索关键词变化时重新请求后端。
  // 后端返回什么，前端就显示什么，不再做客户端过滤。
  useEffect(() => {
    loadProfiles(selectedGroupUUID || undefined, debouncedSearch || undefined)
  }, [selectedGroupUUID, debouncedSearch])

  // loadProfiles 请求后端获取节点列表，支持按分组和关键词筛选。
  // 预留排序扩展点：未来可通过 sortBy 参数支持按名称、延迟等排序。
  const loadProfiles = async (groupUuid?: string, q?: string) => {
    try {
      const res = await profileApi.list(groupUuid, q)
      setProfiles(res.data || [])
    } catch {
      setProfiles([])
    }
  }

  // refreshProfiles 统一刷新入口：始终携带当前 UI 视图状态请求后端，确保数据源唯一。
  // 用于：新建、删除、批量去重等影响数据总量的操作。
  const refreshProfiles = useCallback(() => {
    loadProfiles(selectedGroupUUID || undefined, debouncedSearch || undefined)
  }, [selectedGroupUUID, debouncedSearch])

  // handleNodeSaved 节点保存回调：采用 Visual Culling 策略。
  //   - 新建模式：数据总量变化，交由后端决定最终列表（refreshProfiles）。
  //   - 编辑模式（仍在视图内）：局部替换，零网络延迟。
  //   - 编辑模式（脱离视图）：平滑剔除，AnimatePresence exit 动画自动接管。
  const handleNodeSaved = useCallback((updatedProfile?: Profile) => {
    setEditProfile(null)

    if (!updatedProfile) {
      // 新建模式：数据总量变化，交由后端决定最终列表
      refreshProfiles()
      return
    }

    // 编辑模式：Visual Culling — 前端快速校验节点是否仍属于当前视图
    let stillInView = true

    // 分组越界校验：节点是否被移到了其他分组？
    if (selectedGroupUUID && updatedProfile.group_uuid !== selectedGroupUUID) {
      stillInView = false
    }

    // 搜索脱离校验：节点改名后是否已不匹配搜索关键词？
    if (stillInView && debouncedSearch) {
      const q = debouncedSearch.toLowerCase()
      const groupAlias = (groups.find(g => g.uuid === updatedProfile.group_uuid)?.alias || '').toLowerCase()
      const matchesSearch =
        updatedProfile.name.toLowerCase().includes(q) ||
        updatedProfile.proxy_address.toLowerCase().includes(q) ||
        updatedProfile.proxy_protocol.toLowerCase().includes(q) ||
        groupAlias.includes(q)
      if (!matchesSearch) {
        stillInView = false
      }
    }

    if (stillInView) {
      // 仍属于当前视图：局部替换，零网络延迟
      setProfiles(profiles.map(p => p.ID === updatedProfile.ID ? updatedProfile : p))
    } else {
      // 已脱离视图：平滑剔除（AnimatePresence 的 exit 动画自动接管）
      setProfiles(profiles.filter(p => p.ID !== updatedProfile.ID))
    }
  }, [selectedGroupUUID, debouncedSearch, groups, profiles, refreshProfiles])

  // Activate a node as proxy (clicking WiFi icon)
  const handleActivate = async (profile: Profile, e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await profileApi.select(profile.uuid)
      setActiveProfile(profile)
    } catch (err) {
      console.error('Activate failed:', err)
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

  const handleDelete = async (uuid: string) => {
    try {
      await profileApi.delete(uuid)
      setDeleteTargetId(null)
      refreshProfiles()
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
      const res = await profileEnhancedApi.dedup(selectedGroupUUID || undefined)
      const data = res.data
      setDedupResult(t('nodes.dedup_result', { removed: data.removed, total: data.total }))
      setTimeout(() => setDedupResult(''), 5000)
      // 去重后统一刷新（数据总量已变化）
      refreshProfiles()
      await loadGroups()
    } catch (err) {
      console.error('Dedup failed:', err)
    }
  }

  // Row click: select node (Ctrl/Shift for multi-select)
  const handleRowClick = useCallback((profile: Profile, e: React.MouseEvent) => {
    // Prevent text selection on Shift+click
    if (e.shiftKey) {
      e.preventDefault()
      window.getSelection()?.removeAllRanges()
    }

    if (e.ctrlKey || e.metaKey) {
      // Ctrl+click: toggle selection
      setSelectedIds(prev => {
        const next = new Set(prev)
        if (next.has(profile.ID)) next.delete(profile.ID)
        else next.add(profile.ID)
        return next
      })
    } else if (e.shiftKey && lastClickedId !== null) {
      // Shift+click: range selection
      const ids = profiles.map(p => p.ID)
      const from = ids.indexOf(lastClickedId)
      const to = ids.indexOf(profile.ID)
      if (from !== -1 && to !== -1) {
        const [start, end] = from < to ? [from, to] : [to, from]
        setSelectedIds(prev => {
          const next = new Set(prev)
          for (let i = start; i <= end; i++) next.add(ids[i])
          return next
        })
      }
    } else {
      // Normal click: single select
      setSelectedIds(new Set([profile.ID]))
    }
    setLastClickedId(profile.ID)
  }, [profiles, lastClickedId])

  // Ctrl+A: select all visible nodes
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'a') {
        // Only capture if not focused on an input
        const tag = (e.target as HTMLElement)?.tagName
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
        e.preventDefault()
        setSelectedIds(new Set(profiles.map(p => p.ID)))
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [profiles])

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
        <div className="space-y-1.5" key={selectedGroupUUID} onDoubleClick={(e) => {
          // Double-click blank area to deselect all
          if (e.target === e.currentTarget) setSelectedIds(new Set())
        }}>
          <AnimatePresence>
            {profiles.length === 0 ? (
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
              profiles.map((profile) => {
                const protoColor = getProtocolColor(profile.proxy_protocol)
                return (
                  <motion.div
                    key={profile.ID}
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }}
                  >
                    <div
                      onClick={(e) => handleRowClick(profile, e)}
                      onDoubleClick={(e) => { e.stopPropagation(); handleActivate(profile, e) }}
                      onMouseDown={(e) => { if (e.shiftKey) e.preventDefault() }}
                      className="rounded-xl border px-4 py-3 cursor-pointer transition-colors select-none"
                      style={{
                        backgroundColor: activeProfile?.ID === profile.ID
                          ? 'var(--color-accent-dim)'
                          : selectedIds.has(profile.ID)
                            ? 'color-mix(in srgb, var(--color-primary) 6%, var(--color-card))'
                            : 'var(--color-card)',
                        borderColor: activeProfile?.ID === profile.ID
                          ? 'var(--color-primary)'
                          : selectedIds.has(profile.ID)
                            ? 'color-mix(in srgb, var(--color-primary) 40%, transparent)'
                            : 'var(--color-border)',
                        boxShadow: 'var(--shadow-card)',
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3 min-w-0">
                          <div
                            className="w-2 h-2 rounded-full shrink-0"
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
                                {profile.proxy_protocol}
                              </span>
                            </div>
                            <p
                              className="text-xs mt-0.5 truncate"
                              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                            >
                              {profile.proxy_address}:{profile.proxy_port}
                              {profile.group_uuid && (() => {
                                const g = groups.find(gr => gr.uuid === profile.group_uuid)
                                return g ? <span style={{ fontFamily: 'var(--font-heading)' }}> · {g.alias}</span> : null
                              })()}
                            </p>
                          </div>
                        </div>

                        <div className="flex items-center gap-2 shrink-0">
                          {profile.test_result && (
                            <span
                              className="text-xs"
                              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                            >
                              {profile.test_result}
                            </span>
                          )}
                          <motion.button
                            onClick={(e) => handleActivate(profile, e)}
                            className="p-1 rounded-md transition-colors cursor-pointer"
                            style={{
                              color: activeProfile?.ID === profile.ID
                                ? 'var(--color-success)'
                                : 'var(--color-text-muted)',
                            }}
                            onMouseEnter={(e) => {
                              e.currentTarget.style.color = 'var(--color-success)'
                              e.currentTarget.style.backgroundColor = 'var(--color-success-dim)'
                            }}
                            onMouseLeave={(e) => {
                              e.currentTarget.style.color = activeProfile?.ID === profile.ID
                                ? 'var(--color-success)'
                                : 'var(--color-text-muted)'
                              e.currentTarget.style.backgroundColor = 'transparent'
                            }}
                            whileHover={{ scale: 1.15 }}
                            whileTap={{ scale: 0.9 }}
                            title={activeProfile?.ID === profile.ID ? '当前激活' : '点击激活'}
                          >
                            {activeProfile?.ID === profile.ID ? (
                              <Wifi size={14} />
                            ) : (
                              <WifiOff size={14} />
                            )}
                          </motion.button>
                          <button
                            onClick={(e) => { e.stopPropagation(); setEditProfile(profile) }}
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
                            onClick={(e) => { e.stopPropagation(); setDeleteTargetId(profile.uuid) }}
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
                      visible={deleteTargetId === profile.uuid}
                      message={t('nodes.delete_confirm', { name: profile.name })}
                      onConfirm={() => handleDelete(profile.uuid)}
                      onCancel={() => setDeleteTargetId(null)}
                    />
                  </motion.div>
                )
              })
            )}
          </AnimatePresence>
        </div>
      </div>

      {/* Right: Group Selection Panel */}
      <div className="w-64 shrink-0">
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
                onClick={() => setSelectedGroupUUID('')}
                className="w-full flex items-center gap-2 px-3 py-2 rounded-lg transition-colors text-left cursor-pointer"
                style={{
                  backgroundColor: !selectedGroupUUID ? 'var(--color-accent-dim)' : 'transparent',
                  borderColor: !selectedGroupUUID ? 'var(--color-primary)' : 'transparent',
                  borderWidth: !selectedGroupUUID ? 1 : 0,
                  borderStyle: 'solid',
                }}
                whileTap={{ scale: 0.98 }}
              >
                <div
                  className="w-5 h-5 rounded flex items-center justify-center shrink-0"
                  style={{ backgroundColor: 'var(--color-muted)' }}
                >
                  <Layers size={10} style={{ color: 'var(--color-muted-foreground)' }} />
                </div>
                <div className="min-w-0 flex-1">
                  <span
                    className="text-xs font-medium truncate block"
                    style={{
                      color: !selectedGroupUUID ? 'var(--color-accent-warm)' : 'var(--color-foreground)',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    {t('nodes.all_groups')}
                  </span>
                </div>
                <span
                  className="text-[9px] shrink-0"
                  style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-mono)' }}
                >
                  {profiles.length}
                </span>
              </motion.button>

              {groups.map((group) => {
                const isSelected = selectedGroupUUID === group.uuid
                return (
                  <motion.button
                    key={group.ID}
                    onClick={() => setSelectedGroupUUID(group.uuid)}
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
                      className="w-5 h-5 rounded flex items-center justify-center shrink-0"
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
                      className="text-[9px] shrink-0"
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

      {/* Edit Node Drawer */}
      <RightDrawer
        isOpen={editProfile !== null}
        onClose={() => setEditProfile(null)}
        title={editProfile ? `${t('nodes.edit') || '编辑'}: ${editProfile.name}` : ''}
        subtitle={editProfile ? `${editProfile.proxy_protocol.toUpperCase()} · ${editProfile.proxy_address}:${editProfile.proxy_port}` : ''}
      >
        {editProfile && (
          <NodeEditForm
            editData={editProfile}
              groupUUID={editProfile.group_uuid}
            onClose={() => setEditProfile(null)}
            onSaved={handleNodeSaved}
          />
        )}
      </RightDrawer>
    </div>
  )
}
