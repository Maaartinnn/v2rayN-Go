import { useEffect, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Wifi, WifiOff, RefreshCw, Trash2, Search, Layers, FolderOpen, Link, Edit3 } from 'lucide-react'
import { useStore } from '../store'
import type { Profile, ProfileListItem } from '../store'
import { profileApi, profileEnhancedApi, groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { isStrategyGroup } from '../lib/constants'
import { DeleteConfirmBanner } from './ui/DeleteConfirmBanner'
import { NodeEditForm } from './NodeEditForm'
import { StrategyEditForm } from './StrategyEditForm'
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
  // 列表使用精简数据（ProfileListItem），编辑时按需获取完整 Profile
  const { profileList, setProfileList, activeProfileUUID, setActiveProfileUUID, addToast } = useStore()
  const [loading, setLoading] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedGroupUUID, setSelectedGroupUUID] = useState<string>('')
  const [groups, setGroups] = useState<NodeGroupItem[]>([])
  const [dedupResult, setDedupResult] = useState<string>('')
  const [deleteTargetUUID, setDeleteTargetUUID] = useState<string | null>(null)
  const [editProfile, setEditProfile] = useState<Profile | null>(null)
  const [editLoading, setEditLoading] = useState(false)

  // Multi-selection state（统一使用 uuid 标识）
  const [selectedUUIDs, setSelectedUUIDs] = useState<Set<string>>(new Set())
  const [lastClickedUUID, setLastClickedUUID] = useState<string | null>(null)

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

  // loadProfiles 请求后端获取精简节点列表（ProfileListItem），含后端计算的颜色。
  const loadProfiles = async (groupUuid?: string, q?: string) => {
    try {
      const res = await profileApi.list(groupUuid, q)
      setProfileList(res.data || [])
    } catch {
      setProfileList([])
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
      // 仍属于当前视图：刷新列表（重新请求以获取最新的颜色等后端计算数据）
      refreshProfiles()
    } else {
      // 已脱离视图：从列表中移除（AnimatePresence 的 exit 动画自动接管）
      setProfileList(prev => prev.filter(p => p.uuid !== updatedProfile.uuid))
    }
  }, [selectedGroupUUID, debouncedSearch, groups, refreshProfiles])

  // handleActivate 激活节点为当前代理
  const handleActivate = async (item: ProfileListItem, e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await profileApi.select(item.uuid)
      setActiveProfileUUID(item.uuid)
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
      setDeleteTargetUUID(null)
      // 如果删除的是当前激活节点，清除激活状态
      if (activeProfileUUID === uuid) {
        setActiveProfileUUID(null)
      }
      // 如果删除的节点在多选集中，清除它
      setSelectedUUIDs(prev => {
        const next = new Set(prev)
        next.delete(uuid)
        return next
      })
      refreshProfiles()
    } catch (err) {
      console.error('Delete failed:', err)
    }
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

  // handleEditClick 点击编辑按钮：先通过 API 获取完整 Profile 数据，再打开编辑抽屉
  const handleEditClick = async (item: ProfileListItem, e: React.MouseEvent) => {
    e.stopPropagation()
    setEditLoading(true)
    try {
      const res = await profileApi.get(item.uuid)
      setEditProfile(res.data)
    } catch (err) {
      console.error('Failed to load profile for editing:', err)
      addToast(t('nodes.edit_load_failed'), 'error', { duration: 5000 })
    } finally {
      setEditLoading(false)
    }
  }

  // handleRowClick 行点击：选择节点（支持 Ctrl/Shift 多选）
  const handleRowClick = useCallback((item: ProfileListItem, e: React.MouseEvent) => {
    // Prevent text selection on Shift+click
    if (e.shiftKey) {
      e.preventDefault()
      window.getSelection()?.removeAllRanges()
    }

    if (e.ctrlKey || e.metaKey) {
      // Ctrl+click: toggle selection
      setSelectedUUIDs(prev => {
        const next = new Set(prev)
        if (next.has(item.uuid)) next.delete(item.uuid)
        else next.add(item.uuid)
        return next
      })
    } else if (e.shiftKey && lastClickedUUID !== null) {
      // Shift+click: range selection
      const uuids = profileList.map(p => p.uuid)
      const from = uuids.indexOf(lastClickedUUID)
      const to = uuids.indexOf(item.uuid)
      if (from !== -1 && to !== -1) {
        const [start, end] = from < to ? [from, to] : [to, from]
        setSelectedUUIDs(prev => {
          const next = new Set(prev)
          for (let i = start; i <= end; i++) next.add(uuids[i])
          return next
        })
      }
    } else {
      // Normal click: single select
      setSelectedUUIDs(new Set([item.uuid]))
    }
    setLastClickedUUID(item.uuid)
  }, [profileList, lastClickedUUID])

  // Ctrl+A: select all visible nodes
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'a') {
        // Only capture if not focused on an input
        const tag = (e.target as HTMLElement)?.tagName
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
        e.preventDefault()
        setSelectedUUIDs(new Set(profileList.map(p => p.uuid)))
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [profileList])

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
          if (e.target === e.currentTarget) setSelectedUUIDs(new Set())
        }}>
          <AnimatePresence>
            {profileList.length === 0 ? (
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
              profileList.map((item) => (
                  <motion.div
                    key={item.uuid}
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }}
                  >
                    <div
                      onClick={(e) => handleRowClick(item, e)}
                      onDoubleClick={(e) => { e.stopPropagation(); handleActivate(item, e) }}
                      onMouseDown={(e) => { if (e.shiftKey) e.preventDefault() }}
                      className="rounded-xl border px-4 py-3 cursor-pointer transition-colors select-none"
                      style={{
                        backgroundColor: activeProfileUUID === item.uuid
                          ? 'var(--color-accent-dim)'
                          : selectedUUIDs.has(item.uuid)
                            ? 'color-mix(in srgb, var(--color-primary) 6%, var(--color-card))'
                            : 'var(--color-card)',
                        borderColor: activeProfileUUID === item.uuid
                          ? 'var(--color-primary)'
                          : selectedUUIDs.has(item.uuid)
                            ? 'color-mix(in srgb, var(--color-primary) 40%, transparent)'
                            : 'var(--color-border)',
                        boxShadow: 'var(--shadow-card)',
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3 min-w-0">
                          <div
                            className="w-2 h-2 rounded-full shrink-0"
                            style={{ backgroundColor: item.latency_color }}
                          />
                          <div className="min-w-0">
                            <div className="flex items-center gap-2">
                              <span
                                className="text-sm font-medium truncate"
                                style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                              >
                                {item.name}
                              </span>
                              <span
                                className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                                style={{
                                  backgroundColor: item.protocol_color.bg,
                                  color: item.protocol_color.text,
                                  fontFamily: 'var(--font-heading)',
                                }}
                              >
                                {item.proxy_protocol}
                              </span>
                              {/* 内核徽标：仅在手动指定内核时显示 */}
                              {item.core_type && (
                                <span
                                  className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                                  style={{
                                    backgroundColor: item.core_color.bg,
                                    color: item.core_color.text,
                                    fontFamily: 'var(--font-heading)',
                                  }}
                                >
                                  {item.core_type}
                                </span>
                              )}
                            </div>
                            <p
                              className="text-xs mt-0.5 truncate"
                              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                            >
                              {item.address}
                              {item.group_uuid && (() => {
                                const g = groups.find(gr => gr.uuid === item.group_uuid)
                                return g ? <span style={{ fontFamily: 'var(--font-heading)' }}> · {g.alias}</span> : null
                              })()}
                            </p>
                          </div>
                        </div>

                        <div className="flex items-center gap-2 shrink-0">
                          {item.test_result && (
                            <span
                              className="text-xs"
                              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                            >
                              {item.test_result}
                            </span>
                          )}
                          <motion.button
                            onClick={(e) => handleActivate(item, e)}
                            className="p-1 rounded-md transition-colors cursor-pointer"
                            style={{
                              color: activeProfileUUID === item.uuid
                                ? 'var(--color-success)'
                                : 'var(--color-text-muted)',
                            }}
                            onMouseEnter={(e) => {
                              e.currentTarget.style.color = 'var(--color-success)'
                              e.currentTarget.style.backgroundColor = 'var(--color-success-dim)'
                            }}
                            onMouseLeave={(e) => {
                              e.currentTarget.style.color = activeProfileUUID === item.uuid
                                ? 'var(--color-success)'
                                : 'var(--color-text-muted)'
                              e.currentTarget.style.backgroundColor = 'transparent'
                            }}
                            whileHover={{ scale: 1.15 }}
                            whileTap={{ scale: 0.9 }}
                            title={activeProfileUUID === item.uuid ? t('nodes.activated') : t('nodes.click_to_activate')}
                          >
                            {activeProfileUUID === item.uuid ? (
                              <Wifi size={14} />
                            ) : (
                              <WifiOff size={14} />
                            )}
                          </motion.button>
                          <button
                            onClick={(e) => handleEditClick(item, e)}
                            disabled={editLoading}
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
                            onClick={(e) => { e.stopPropagation(); setDeleteTargetUUID(item.uuid) }}
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
                      visible={deleteTargetUUID === item.uuid}
                      message={t('nodes.delete_confirm', { name: item.name })}
                      onConfirm={() => handleDelete(item.uuid)}
                      onCancel={() => setDeleteTargetUUID(null)}
                    />
                  </motion.div>
                ))
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
                  {profileList.length}
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

      {/* Edit Drawer — 单抽屉，根据节点类型动态渲染不同表单 */}
      <RightDrawer
        isOpen={editProfile !== null}
        onClose={() => setEditProfile(null)}
        title={editProfile ? `${t('nodes.edit')}: ${editProfile.name}` : ''}
        subtitle={editProfile ? editProfile.proxy_protocol.toUpperCase() : ''}
      >
        {editProfile && (
          isStrategyGroup(editProfile.proxy_protocol) ? (
            <StrategyEditForm
              editData={editProfile}
              onClose={() => setEditProfile(null)}
              onSaved={handleNodeSaved}
            />
          ) : (
            <NodeEditForm
              editData={editProfile}
              groupUUID={editProfile.group_uuid}
              onClose={() => setEditProfile(null)}
              onSaved={handleNodeSaved}
            />
          )
        )}
      </RightDrawer>
    </div>
  )
}