import { useEffect, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import {
  Plus,
  Trash2,
  Edit3,
  Check,
  X,
  FolderOpen,
  GripVertical,
  Link,
  RefreshCw,
  Globe,
} from 'lucide-react'
import { groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { useStore } from '../store'

interface NodeGroup {
  ID: number
  uuid: string
  alias: string
  is_subscription: boolean
  url: string
  enabled: boolean
  enable_update: boolean
  update_interval: number
  alias_regex: string
  user_agent: string
  notes: string
  sort_order: number
  node_count: number
  last_update_time: string
  color: string
}

// ========== Sortable Group Card ==========
function SortableGroupCard({
  group,
  isEditing,
  onEdit,
  onDelete,
  onSave,
  onCancel,
  onRefresh,
  onRefreshProxy,
  canDelete,
  t,
}: {
  group: NodeGroup
  isEditing: boolean
  onEdit: () => void
  onDelete: () => void
  onSave: (data: Partial<NodeGroup>) => void
  onCancel: () => void
  onRefresh: () => void
  onRefreshProxy: () => void
  canDelete: boolean
  t: (key: any, params?: Record<string, any>) => string
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: group.uuid })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    zIndex: isDragging ? 50 : 'auto',
    opacity: isDragging ? 0.8 : 1,
  }

  // Edit form state
  const [formAlias, setFormAlias] = useState(group.alias)
  const [formIsSub, setFormIsSub] = useState(group.is_subscription)
  const [formUrl, setFormUrl] = useState(group.url || '')
  const [formEnabled, setFormEnabled] = useState(group.enabled)
  const [formEnableUpdate, setFormEnableUpdate] = useState(group.enable_update)
  const [formInterval, setFormInterval] = useState(String(group.update_interval || 0))
  const [formAliasRegex, setFormAliasRegex] = useState(group.alias_regex || '')
  const [formUserAgent, setFormUserAgent] = useState(group.user_agent || '')
  const [formNotes, setFormNotes] = useState(group.notes || '')

  // Reset form when edit starts
  useEffect(() => {
    if (isEditing) {
      setFormAlias(group.alias)
      setFormIsSub(group.is_subscription)
      setFormUrl(group.url || '')
      setFormEnabled(group.enabled)
      setFormEnableUpdate(group.enable_update)
      setFormInterval(String(group.update_interval || 0))
      setFormAliasRegex(group.alias_regex || '')
      setFormUserAgent(group.user_agent || '')
      setFormNotes(group.notes || '')
    }
  }, [isEditing, group])

  const handleSave = () => {
    onSave({
      alias: formAlias,
      is_subscription: formIsSub,
      url: formUrl,
      enabled: formEnabled,
      enable_update: formEnableUpdate,
      update_interval: parseInt(formInterval) || 0,
      alias_regex: formAliasRegex,
      user_agent: formUserAgent,
      notes: formNotes,
    })
  }

  const displayName = group.alias || t('groups.default_name')

  return (
    <div ref={setNodeRef} style={style}>
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        className="rounded-xl border"
        style={{
          backgroundColor: 'var(--color-card)',
          borderColor: isDragging ? 'var(--color-primary)' : 'var(--color-border)',
          boxShadow: isDragging ? '0 8px 24px rgba(0,0,0,0.12)' : 'var(--shadow-card)',
        }}
      >
        {/* Card Header */}
        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            {/* Drag Handle */}
            <div
              {...attributes}
              {...listeners}
              className="cursor-grab active:cursor-grabbing p-1 rounded-md flex-shrink-0"
              style={{ color: 'var(--color-text-muted)' }}
            >
              <GripVertical size={14} />
            </div>

            {/* Group Icon */}
            <div
              className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
              style={{
                backgroundColor: group.is_subscription
                  ? 'rgba(217, 119, 87, 0.12)'
                  : 'var(--color-accent-dim)',
              }}
            >
              {group.is_subscription ? (
                <Link size={14} style={{ color: '#D97757' }} />
              ) : (
                <FolderOpen size={14} style={{ color: 'var(--color-accent-warm)' }} />
              )}
            </div>

            {/* Name & Info */}
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <span
                  className="text-sm font-medium truncate"
                  style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                >
                  {displayName}
                </span>
                {group.is_subscription && (
                  <span
                    className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                    style={{
                      backgroundColor: 'rgba(217, 119, 87, 0.12)',
                      color: '#D97757',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    {t('groups.is_subscription')}
                  </span>
                )}
              </div>
              <div className="flex items-center gap-3 mt-0.5">
                <span
                  className="text-[10px]"
                  style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
                >
                  {t('groups.nodes_count', { count: group.node_count })}
                </span>
                {group.is_subscription && group.notes && (
                  <span
                    className="text-[10px] truncate max-w-[200px]"
                    style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
                  >
                    {group.notes}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-1 flex-shrink-0 ml-3">
            {/* Refresh buttons (subscription only) */}
            {group.is_subscription && (
              <>
                <button
                  onClick={onRefresh}
                  className="p-1.5 rounded-md transition-colors cursor-pointer"
                  style={{ color: 'var(--color-text-muted)' }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.color = 'var(--color-accent-warm)'
                    e.currentTarget.style.backgroundColor = 'var(--color-accent-dim)'
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.color = 'var(--color-text-muted)'
                    e.currentTarget.style.backgroundColor = 'transparent'
                  }}
                  title={t('groups.update_no_proxy')}
                >
                  <RefreshCw size={13} />
                </button>
                <button
                  onClick={onRefreshProxy}
                  className="p-1.5 rounded-md transition-colors cursor-pointer"
                  style={{ color: 'var(--color-text-muted)' }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.color = 'var(--color-accent-warm)'
                    e.currentTarget.style.backgroundColor = 'var(--color-accent-dim)'
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.color = 'var(--color-text-muted)'
                    e.currentTarget.style.backgroundColor = 'transparent'
                  }}
                  title={t('groups.update_with_proxy')}
                >
                  <Globe size={13} />
                </button>
              </>
            )}
            <button
              onClick={onEdit}
              className="p-1.5 rounded-md transition-colors cursor-pointer"
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
              <Edit3 size={13} />
            </button>
            <button
              onClick={onDelete}
              disabled={!canDelete}
              className="p-1.5 rounded-md transition-colors cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
              style={{ color: 'var(--color-text-muted)' }}
              onMouseEnter={(e) => {
                if (canDelete) {
                  e.currentTarget.style.color = 'var(--color-error)'
                  e.currentTarget.style.backgroundColor = 'var(--color-error-dim)'
                }
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = 'var(--color-text-muted)'
                e.currentTarget.style.backgroundColor = 'transparent'
              }}
            >
              <Trash2 size={13} />
            </button>
          </div>
        </div>

        {/* Edit Panel (expands below the card) */}
        <AnimatePresence>
          {isEditing && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
              className="overflow-hidden"
            >
              <div
                className="px-4 pb-4 pt-1 border-t"
                style={{ borderColor: 'var(--color-border)' }}
              >
                <div className="space-y-3">
                  {/* Alias + Subscription Toggle */}
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label
                        className="text-xs font-medium block mb-1"
                        style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                      >
                        {t('groups.alias')}
                      </label>
                      <input
                        type="text"
                        value={formAlias}
                        onChange={(e) => setFormAlias(e.target.value)}
                        placeholder={t('groups.default_name')}
                        className="w-full px-3 py-2 text-sm rounded-lg border"
                        style={{
                          backgroundColor: 'var(--color-overlay)',
                          borderColor: 'var(--color-border)',
                          color: 'var(--color-foreground)',
                          fontFamily: 'var(--font-heading)',
                        }}
                      />
                    </div>
                    <div>
                      <label
                        className="text-xs font-medium block mb-1"
                        style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                      >
                        {t('groups.is_subscription')}
                      </label>
                      <button
                        onClick={() => setFormIsSub(!formIsSub)}
                        className="w-full px-3 py-2 text-sm rounded-lg border text-left transition-colors cursor-pointer"
                        style={{
                          backgroundColor: 'var(--color-overlay)',
                          borderColor: formIsSub ? 'var(--color-primary)' : 'var(--color-border)',
                          color: 'var(--color-foreground)',
                          fontFamily: 'var(--font-heading)',
                        }}
                      >
                        {formIsSub ? t('common.yes') : t('common.no')}
                      </button>
                    </div>
                  </div>

                  {/* Subscription fields (conditional) */}
                  <AnimatePresence>
                    {formIsSub && (
                      <motion.div
                        initial={{ opacity: 0, height: 0 }}
                        animate={{ opacity: 1, height: 'auto' }}
                        exit={{ opacity: 0, height: 0 }}
                        className="space-y-3 overflow-hidden"
                      >
                        {/* URL */}
                        <div>
                          <label
                            className="text-xs font-medium block mb-1"
                            style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                          >
                            {t('groups.url')}
                          </label>
                          <input
                            type="text"
                            value={formUrl}
                            onChange={(e) => setFormUrl(e.target.value)}
                            placeholder={t('groups.url_placeholder')}
                            className="w-full px-3 py-2 text-sm rounded-lg border"
                            style={{
                              backgroundColor: 'var(--color-overlay)',
                              borderColor: 'var(--color-border)',
                              color: 'var(--color-foreground)',
                              fontFamily: 'var(--font-mono)',
                            }}
                          />
                        </div>

                        {/* Enable Update + Interval */}
                        <div className="grid grid-cols-2 gap-3">
                          <div>
                            <label
                              className="text-xs font-medium block mb-1"
                              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                            >
                              {t('groups.enable_update')}
                            </label>
                            <button
                              onClick={() => setFormEnableUpdate(!formEnableUpdate)}
                              className="w-full px-3 py-2 text-sm rounded-lg border text-left transition-colors cursor-pointer"
                              style={{
                                backgroundColor: 'var(--color-overlay)',
                                borderColor: formEnableUpdate ? 'var(--color-primary)' : 'var(--color-border)',
                                color: 'var(--color-foreground)',
                                fontFamily: 'var(--font-heading)',
                              }}
                            >
                              {formEnableUpdate ? t('common.enabled') : t('common.disabled')}
                            </button>
                          </div>
                          <div>
                            <label
                              className="text-xs font-medium block mb-1"
                              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                            >
                              {t('groups.update_interval')}
                            </label>
                            <input
                              type="number"
                              value={formInterval}
                              onChange={(e) => setFormInterval(e.target.value)}
                              className="w-full px-3 py-2 text-sm rounded-lg border"
                              style={{
                                backgroundColor: 'var(--color-overlay)',
                                borderColor: 'var(--color-border)',
                                color: 'var(--color-foreground)',
                                fontFamily: 'var(--font-mono)',
                              }}
                            />
                            <span
                              className="text-[10px] mt-0.5 block"
                              style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
                            >
                              {t('groups.update_interval_hint')}
                            </span>
                          </div>
                        </div>

                        {/* Alias Regex + User Agent */}
                        <div className="grid grid-cols-2 gap-3">
                          <div>
                            <label
                              className="text-xs font-medium block mb-1"
                              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                            >
                              {t('groups.alias_regex')}
                            </label>
                            <input
                              type="text"
                              value={formAliasRegex}
                              onChange={(e) => setFormAliasRegex(e.target.value)}
                              placeholder={t('groups.alias_regex_placeholder')}
                              className="w-full px-3 py-2 text-sm rounded-lg border"
                              style={{
                                backgroundColor: 'var(--color-overlay)',
                                borderColor: 'var(--color-border)',
                                color: 'var(--color-foreground)',
                                fontFamily: 'var(--font-mono)',
                              }}
                            />
                          </div>
                          <div>
                            <label
                              className="text-xs font-medium block mb-1"
                              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                            >
                              {t('groups.user_agent')}
                            </label>
                            <input
                              type="text"
                              value={formUserAgent}
                              onChange={(e) => setFormUserAgent(e.target.value)}
                              placeholder={t('groups.user_agent_placeholder')}
                              className="w-full px-3 py-2 text-sm rounded-lg border"
                              style={{
                                backgroundColor: 'var(--color-overlay)',
                                borderColor: 'var(--color-border)',
                                color: 'var(--color-foreground)',
                                fontFamily: 'var(--font-mono)',
                              }}
                            />
                          </div>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>

                  {/* Notes */}
                  <div>
                    <label
                      className="text-xs font-medium block mb-1"
                      style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                    >
                      {t('groups.notes')}
                    </label>
                    <textarea
                      value={formNotes}
                      onChange={(e) => setFormNotes(e.target.value)}
                      placeholder={t('groups.notes_placeholder')}
                      rows={2}
                      className="w-full px-3 py-2 text-sm rounded-lg border resize-none"
                      style={{
                        backgroundColor: 'var(--color-overlay)',
                        borderColor: 'var(--color-border)',
                        color: 'var(--color-foreground)',
                        fontFamily: 'var(--font-heading)',
                      }}
                    />
                  </div>
                </div>

                {/* Save / Cancel */}
                <div className="flex justify-end gap-2 mt-4">
                  <button
                    onClick={onCancel}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
                    style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
                  >
                    <X size={13} />
                    {t('nodes.cancel')}
                  </button>
                  <button
                    onClick={handleSave}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
                    style={{
                      backgroundColor: 'var(--color-primary)',
                      color: 'var(--color-primary-foreground)',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    <Check size={13} />
                    {t('nodes.save')}
                  </button>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </div>
  )
}

// ========== Main GroupsView ==========
export function GroupsView() {
  const [groups, setGroups] = useState<NodeGroup[]>([])
  const [editId, setEditId] = useState<number | null>(null)
  const t = useT()
  const { addToast } = useStore()

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )

  const loadGroups = useCallback(async () => {
    try {
      const res = await groupsApi.list()
      setGroups(res.data || [])
    } catch {
      setGroups([])
    }
  }, [])

  useEffect(() => {
    loadGroups()
  }, [loadGroups])

  const handleAdd = async () => {
    try {
      await groupsApi.create({
        alias: '',
        is_subscription: false,
        enabled: true,
        notes: '',
      })
      await loadGroups()
    } catch (err) {
      console.error('Add group failed:', err)
      addToast(t('groups.save_failed'), 'error')
    }
  }

  const handleSave = async (id: number, data: Partial<NodeGroup>) => {
    try {
      const res = await groupsApi.update(id, data)
      setEditId(null)
      // Update locally without reloading the full list to avoid visual reorder
      setGroups(prev => prev.map(g => g.ID === id ? { ...g, ...res.data } : g))
    } catch (err) {
      console.error('Update group failed:', err)
      addToast(t('groups.save_failed'), 'error')
    }
  }

  const handleDelete = async (id: number) => {
    if (groups.length <= 1) {
      addToast(t('groups.cannot_delete'), 'error')
      return
    }
    if (!window.confirm(t('groups.delete_confirm'))) return
    try {
      await groupsApi.delete(id)
      if (editId === id) setEditId(null)
      await loadGroups()
    } catch (err) {
      console.error('Delete group failed:', err)
      addToast(t('groups.cannot_delete'), 'error')
    }
  }

  const handleRefresh = async (id: number) => {
    try {
      await groupsApi.refresh(id)
      addToast(t('groups.update_success'), 'success')
    } catch (err) {
      console.error('Refresh failed:', err)
      addToast(t('groups.update_failed'), 'error')
    }
  }

  const handleRefreshProxy = async (id: number) => {
    try {
      await groupsApi.refreshProxy(id)
      addToast(t('groups.update_success'), 'success')
    } catch (err) {
      console.error('Refresh via proxy failed:', err)
      addToast(t('groups.update_failed'), 'error')
    }
  }

  // Drag-and-drop: optimistic reorder
  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return

    const oldIndex = groups.findIndex((g) => g.uuid === active.id)
    const newIndex = groups.findIndex((g) => g.uuid === over.id)

    // Optimistic update
    const newGroups = arrayMove(groups, oldIndex, newIndex)
    setGroups(newGroups)

    // Save to backend
    try {
      await groupsApi.reorder(newGroups.map((g) => g.uuid))
    } catch (err) {
      console.error('Reorder failed:', err)
      addToast(t('groups.reorder_failed'), 'error')
      // Revert
      await loadGroups()
    }
  }

  return (
    <div className="max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <h1
          className="text-xl font-semibold"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('groups.title')}
        </h1>
        <motion.button
          onClick={handleAdd}
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
          {t('groups.add')}
        </motion.button>
      </div>

      {/* Group List with Drag-and-Drop */}
      <div className="space-y-2">
        {groups.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-20"
          >
            <FolderOpen
              size={32}
              className="mx-auto mb-3"
              style={{ color: 'var(--color-text-muted)' }}
            />
            <p
              className="text-sm"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('common.no_data')}
            </p>
          </motion.div>
        ) : (
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <SortableContext
              items={groups.map((g) => g.uuid)}
              strategy={verticalListSortingStrategy}
            >
              {groups.map((group) => (
                <SortableGroupCard
                  key={group.uuid}
                  group={group}
                  isEditing={editId === group.ID}
                  onEdit={() => setEditId(editId === group.ID ? null : group.ID)}
                  onDelete={() => handleDelete(group.ID)}
                  onSave={(data) => handleSave(group.ID, data)}
                  onCancel={() => setEditId(null)}
                  onRefresh={() => handleRefresh(group.ID)}
                  onRefreshProxy={() => handleRefreshProxy(group.ID)}
                  canDelete={groups.length > 1}
                  t={t}
                />
              ))}
            </SortableContext>
          </DndContext>
        )}
      </div>
    </div>
  )
}