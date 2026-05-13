import React, { useEffect, useState, useCallback, useRef, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Wifi, WifiOff, RefreshCw, Trash2, Search, Layers, FolderOpen, Link, Edit3, GripVertical } from 'lucide-react'

import { useStore } from '../store'
import type { Profile } from '../store'
import { profileApi, profileEnhancedApi, groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { DeleteConfirmBanner } from './ui/DeleteConfirmBanner'
import { NodeEditForm } from './NodeEditForm'
import { RightDrawer } from './ui/RightDrawer'
import { VirtualSortableList } from './ui/VirtualSortableList'

// ==========================================
// Types & Helpers
// ==========================================

interface NodeGroupItem {
  ID: number
  uuid: string
  alias: string
  is_subscription: boolean
  node_count: number
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

const getLatencyDot = (result?: string) => {
  if (!result || result === 'timeout') return 'var(--color-error)'
  const ms = parseInt(result)
  if (ms < 100) return 'var(--color-success)'
  if (ms < 300) return 'var(--color-warning)'
  return 'var(--color-error)'
}

// 【已移除 ZIndexWrapper】
// 原 ZIndexWrapper 通过直接操作 DOM 修改 zIndex，会在 React 重渲染时被覆盖失效。
// 现改用 VirtualSortableList 的 isItemExpanded prop，由 React 状态驱动 zIndex 提升。

// ==========================================
// 主视图组件: NodesView
// ==========================================

export function NodesView() {
  const { profiles, setProfiles, activeProfile, setActiveProfile } = useStore()
  const [loading, setLoading] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedGroupId, setSelectedGroupId] = useState<number>(0)
  const [groups, setGroups] = useState<NodeGroupItem[]>([])
  const [dedupResult, setDedupResult] = useState<string>('')
  const [deleteTargetId, setDeleteTargetId] = useState<number | null>(null)
  const [editProfile, setEditProfile] = useState<Profile | null>(null)

  // 拖拽与多选状态
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [lastClickedId, setLastClickedId] = useState<number | null>(null)

  const scrollRef = useRef<HTMLDivElement>(null)
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

  const loadGroups = async () => {
    try {
      const res = await groupsApi.list()
      setGroups(res.data || [])
    } catch {
      setGroups([])
    }
  }

  // 节点操作
  const handleNodeSaved = (updatedProfile?: Profile) => {
    if (updatedProfile) {
      setProfiles(profiles.map(p => p.ID === updatedProfile.ID ? updatedProfile : p))
    } else {
      loadProfiles()
    }
    setEditProfile(null)
  }

  const handleActivate = useCallback(async (profile: Profile, e?: React.MouseEvent) => {
    if (e) e.stopPropagation()
    try {
      await profileApi.select(profile.ID)
      setActiveProfile(profile)
    } catch (err) {
      console.error('Activate failed:', err)
    }
  }, [setActiveProfile])

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

  const handleDelete = useCallback(async (id: number) => {
    try {
      await profileApi.delete(id)
      setDeleteTargetId(null)
      await loadProfiles()
    } catch (err) {
      console.error('Delete failed:', err)
    }
  }, [])

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

  // 计算筛选列表
  const filteredProfiles = useMemo(() => {
    return profiles.filter((p) => {
      const matchesSearch = !searchQuery || (() => {
        const q = searchQuery.toLowerCase()
        return (
          p.name.toLowerCase().includes(q) ||
          p.address.toLowerCase().includes(q) ||
          p.protocol.toLowerCase().includes(q) ||
          p.group_name?.toLowerCase().includes(q)
        )
      })()
      const matchesGroup = selectedGroupId === 0 || p.group_id === selectedGroupId
      return matchesSearch && matchesGroup
    })
  }, [profiles, searchQuery, selectedGroupId])

  // Ctrl/Shift 多选逻辑与双击逻辑（隔离了拖拽区）
  const isInteractiveTarget = (e: React.MouseEvent) => {
    const target = e.target as HTMLElement
    // 若点击部位在拖拽手柄、按钮元素内，判定为交互操作，应跳过行选中
    return target.closest('.drag-handle') || target.closest('button')
  }

  const handleRowClick = useCallback((profile: Profile, e: React.MouseEvent) => {
    if (isInteractiveTarget(e)) return

    if (e.shiftKey) {
      e.preventDefault()
      window.getSelection()?.removeAllRanges()
    }

    if (e.ctrlKey || e.metaKey) {
      setSelectedIds(prev => {
        const next = new Set(prev)
        if (next.has(profile.ID)) next.delete(profile.ID)
        else next.add(profile.ID)
        return next
      })
    } else if (e.shiftKey && lastClickedId !== null) {
      const ids = filteredProfiles.map(p => p.ID)
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
      setSelectedIds(new Set([profile.ID]))
    }
    setLastClickedId(profile.ID)
  }, [filteredProfiles, lastClickedId])

  const handleRowDoubleClick = useCallback((profile: Profile, e: React.MouseEvent) => {
    if (isInteractiveTarget(e)) return
    e.stopPropagation()
    handleActivate(profile, e)
  }, [handleActivate])

  const handleRowMouseDown = useCallback((e: React.MouseEvent) => {
    if (isInteractiveTarget(e)) return
    // 阻止 Shift 按下时浏览器的默认文本选中行为
    if (e.shiftKey) e.preventDefault()
  }, [])

  // Ctrl+A 快捷键
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'a') {
        const tag = (e.target as HTMLElement)?.tagName
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
        e.preventDefault()
        setSelectedIds(new Set(filteredProfiles.map(p => p.ID)))
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [filteredProfiles])

  // 拖拽排序逻辑
  const handleReorder = useCallback(async (newFilteredItems: Profile[]) => {
    const originalIndices = profiles
      .map((p, i) => filteredProfiles.find(fp => fp.ID === p.ID) ? i : -1)
      .filter(i => i !== -1)

    const newProfiles = [...profiles]
    originalIndices.forEach((origIdx, i) => {
      newProfiles[origIdx] = newFilteredItems[i]
    })

    setProfiles(newProfiles)

    try {
      await fetch('/api/profiles/reorder', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ uuids: newProfiles.map(p => p.uuid) })
      })
    } catch (err) {
      console.error('Failed to reorder profiles:', err)
    }
  }, [profiles, filteredProfiles, setProfiles])

  const handleDragStart = useCallback(() => {
    setDeleteTargetId(null)
    setEditProfile(null)
  }, [])

  // ==========================================
  // Render Item / Render Extra 
  // ==========================================

  const renderItem = useCallback(({ 
    item: profile, 
    isDragging, 
    dragListeners, 
    dragAttributes 
  }: {
    item: Profile
    isDragging: boolean
    isOverlay: boolean
    dragListeners: Record<string, any>
    dragAttributes: Record<string, any>
  }) => {
    const protoColor = getProtocolColor(profile.protocol)
    const isActive = activeProfile?.ID === profile.ID
    const isSelected = selectedIds.has(profile.ID)
    const disableDrag = !!searchQuery

    return (
      <div
        onClick={(e) => handleRowClick(profile, e)}
        onDoubleClick={(e) => handleRowDoubleClick(profile, e)}
        onMouseDown={handleRowMouseDown}
        className={`rounded-xl border px-4 py-3 cursor-pointer transition-colors select-none bg-white`}
        style={{
          backgroundColor: isActive
            ? 'var(--color-accent-dim)'
            : isSelected
              ? 'color-mix(in srgb, var(--color-primary) 6%, var(--color-card))'
              : 'var(--color-card)',
          borderColor: isActive
            ? 'var(--color-primary)'
            : isDragging
              ? 'var(--color-primary)'
              : isSelected
                ? 'color-mix(in srgb, var(--color-primary) 40%, transparent)'
                : 'var(--color-border)',
          boxShadow: isDragging ? '0 8px 24px rgba(0,0,0,0.12)' : 'var(--shadow-card)',
        }}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            {/* 拖拽手柄：带有 .drag-handle 标识供外部拦截，不加 stopPropagation 让 dnd-kit 原生捕获 */}
            {!disableDrag && (
              <div
                {...dragAttributes}
                {...dragListeners}
                className="drag-handle cursor-grab active:cursor-grabbing p-1 rounded-md shrink-0 text-gray-400 hover:text-gray-700 hover:bg-gray-100"
              >
                <GripVertical size={14} />
              </div>
            )}
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
              onClick={(e) => { e.stopPropagation(); handleActivate(profile, e) }}
              className="p-1 rounded-md transition-colors cursor-pointer"
              style={{
                color: isActive ? 'var(--color-success)' : 'var(--color-text-muted)',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.color = 'var(--color-success)'
                e.currentTarget.style.backgroundColor = 'var(--color-success-dim)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = isActive ? 'var(--color-success)' : 'var(--color-text-muted)'
                e.currentTarget.style.backgroundColor = 'transparent'
              }}
              whileHover={{ scale: 1.15 }}
              whileTap={{ scale: 0.9 }}
              title={isActive ? '当前激活' : '点击激活'}
            >
              {isActive ? <Wifi size={14} /> : <WifiOff size={14} />}
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
              onClick={(e) => { e.stopPropagation(); setDeleteTargetId(deleteTargetId === profile.ID ? null : profile.ID) }}
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
    )
  }, [activeProfile, selectedIds, searchQuery, handleRowClick, handleRowDoubleClick, handleRowMouseDown, handleActivate, deleteTargetId])

  const renderExtra = useCallback((profile: Profile) => {
    const isVisible = deleteTargetId === profile.ID
    return (
      <DeleteConfirmBanner
        visible={isVisible}
        message={t('nodes.delete_confirm', { name: profile.name })}
        onConfirm={() => handleDelete(profile.ID)}
        onCancel={() => setDeleteTargetId(null)}
      />
    )
  }, [deleteTargetId, t, handleDelete])

  // 【修复3】：判断 item 是否处于展开状态（删除确认面板可见），用于提升 zIndex
  const isItemExpanded = useCallback((profile: Profile) => {
    return deleteTargetId === profile.ID
  }, [deleteTargetId])

  const emptyContent = useMemo(() => (
    <div className="text-center py-20">
      <Layers size={32} className="mx-auto mb-3" style={{ color: 'var(--color-text-muted)' }} />
      <p className="text-sm" style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}>
        {searchQuery ? t('common.no_data') : t('nodes.no_nodes')}
      </p>
    </div>
  ), [searchQuery, t])

  const displayName = (g: NodeGroupItem) => g.alias || t('groups.default_name')

  return (
    <div className="flex gap-6 max-w-5xl mx-auto h-[calc(100vh-80px)]">
      {/* Left: Node List */}
      <div className="flex-1 min-w-0 flex flex-col h-full">
        {/* Header */}
        <div className="flex items-center justify-between mb-5 shrink-0">
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
              style={{ fontFamily: 'var(--font-heading)' }}
              whileTap={{ scale: 0.95 }}
            >
              <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
              {t('nodes.test_all')}
            </motion.button>
          </div>
        </div>

        {/* Toolbar: search + dedup */}
        <div className="flex items-center gap-2 mb-4 shrink-0">
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
              initial={{ opacity: 0, y: -8, height: 0, marginBottom: 0 }}
              animate={{ opacity: 1, y: 0, height: 'auto', marginBottom: 12 }}
              exit={{ opacity: 0, y: -8, height: 0, marginBottom: 0 }}
              className="px-4 py-2 rounded-lg text-xs font-medium overflow-hidden shrink-0"
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

        {/* Node list (Virtualized & Sortable) */}
        <div
          className="flex-1 min-h-0"
          onDoubleClick={(e) => {
            // Double-click blank area to deselect all
            if (e.target === e.currentTarget) setSelectedIds(new Set())
          }}
        >
          <VirtualSortableList
            key={selectedGroupId} // 当切换分组时强制重新实例化以重置滚动
            items={filteredProfiles}
            onItemsChange={handleReorder}
            renderItem={renderItem}
            renderExtra={renderExtra}
            isItemExpanded={isItemExpanded} // 【修复3】：展开面板时提升 zIndex
            estimateSize={74}
            overscan={5}
            className="h-full flex flex-col"
            disableDrag={!!searchQuery} // 仅在未搜索时允许拖拽
            onDragStart={handleDragStart}
            emptyContent={emptyContent}
          />
        </div>
      </div>

      {/* Right: Group Selection Panel */}
      <div className="w-64 shrink-0 h-full overflow-y-auto">
        <div className="sticky top-0">
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
                onClick={() => { setSelectedGroupId(0); scrollRef.current?.scrollTo({ top: 0 }) }}
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
                  className="w-5 h-5 rounded flex items-center justify-center shrink-0"
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
                  className="text-[9px] shrink-0"
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
                    onClick={() => { setSelectedGroupId(group.ID); scrollRef.current?.scrollTo({ top: 0 }) }}
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
        subtitle={editProfile ? `${editProfile.protocol.toUpperCase()} · ${editProfile.address}:${editProfile.port}` : ''}
      >
        {editProfile && (
          <NodeEditForm
            editData={editProfile}
            groupId={editProfile.group_id}
            onClose={() => setEditProfile(null)}
            onSaved={handleNodeSaved}
          />
        )}
      </RightDrawer>
    </div>
  )
}