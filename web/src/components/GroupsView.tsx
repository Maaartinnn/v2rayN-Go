import { useEffect, useState, useCallback, useMemo } from 'react'
import {
  Plus,
  Trash2,
  Edit3,
  FolderOpen,
  GripVertical,
  Link,
  RefreshCw,
  Globe,
} from 'lucide-react'
import { groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { useStore } from '../store'
import { DeleteConfirmBanner } from './ui/DeleteConfirmBanner'
import { EditFormCard } from './ui/EditFormCard'
import { SmoothCollapse } from './ui/SmoothCollapse'
import { FormField } from './ui/FormField'
import { FormActions } from './ui/FormActions'
import { inputStyle, inputHeadingStyle, textareaStyle } from './ui/formStyles'
import { VirtualSortableList } from './ui/VirtualSortableList'

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

// ========== Group Edit Form ==========
function GroupEditForm({
  group,
  onSave,
  onCancel,
  t,
}: {
  group: NodeGroup
  onSave: (data: Partial<NodeGroup>) => void
  onCancel: () => void
  t: (key: any, params?: Record<string, any>) => string
}) {
  const [formAlias, setFormAlias] = useState(group.alias)
  const [formIsSub, setFormIsSub] = useState(group.is_subscription)
  const [formUrl, setFormUrl] = useState(group.url || '')
  const [formEnableUpdate, setFormEnableUpdate] = useState(group.enable_update)
  const [formInterval, setFormInterval] = useState(String(group.update_interval || 0))
  const [formAliasRegex, setFormAliasRegex] = useState(group.alias_regex || '')
  const [formUserAgent, setUserAgent] = useState(group.user_agent || '')
  const [formNotes, setFormNotes] = useState(group.notes || '')

  const handleSave = () => {
    onSave({
      alias: formAlias,
      is_subscription: formIsSub,
      url: formUrl,
      enabled: group.enabled,
      enable_update: formEnableUpdate,
      update_interval: parseInt(formInterval) || 0,
      alias_regex: formAliasRegex,
      user_agent: formUserAgent,
      notes: formNotes,
    })
  }

  return (
    <EditFormCard>
      <div className="space-y-3">
        {/* Alias + Subscription Toggle */}
        <div className="grid grid-cols-2 gap-3">
          <FormField label={t('groups.alias')} cols="1/2">
            <input
              type="text"
              value={formAlias}
              onChange={(e) => setFormAlias(e.target.value)}
              placeholder={t('groups.default_name')}
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={inputHeadingStyle}
            />
          </FormField>
          <FormField label={t('groups.is_subscription')} cols="1/2">
            <button
              onClick={() => setFormIsSub(!formIsSub)}
              className="w-full px-3 py-2 text-sm rounded-lg border text-left transition-colors cursor-pointer"
              style={{
                ...inputHeadingStyle,
                borderColor: formIsSub ? 'var(--color-primary)' : 'var(--color-border)',
              }}
            >
              {formIsSub ? t('common.yes') : t('common.no')}
            </button>
          </FormField>
        </div>

        {/* Subscription fields (conditional) */}
        <SmoothCollapse isOpen={formIsSub} className="space-y-3">
          {/* URL */}
          <FormField label={t('groups.url')}>
            <input
              type="text"
              value={formUrl}
              onChange={(e) => setFormUrl(e.target.value)}
              placeholder={t('groups.url_placeholder')}
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={inputStyle}
            />
          </FormField>

          {/* Enable Update + Interval */}
          <div className="grid grid-cols-2 gap-3">
            <FormField label={t('groups.enable_update')} cols="1/2">
              <button
                onClick={() => setFormEnableUpdate(!formEnableUpdate)}
                className="w-full px-3 py-2 text-sm rounded-lg border text-left transition-colors cursor-pointer"
                style={{
                  ...inputHeadingStyle,
                  borderColor: formEnableUpdate ? 'var(--color-primary)' : 'var(--color-border)',
                }}
              >
                {formEnableUpdate ? t('common.enabled') : t('common.disabled')}
              </button>
            </FormField>
            <FormField label={t('groups.update_interval')} cols="1/2" hint={t('groups.update_interval_hint')}>
              <input
                type="number"
                value={formInterval}
                onChange={(e) => setFormInterval(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </FormField>
          </div>

          {/* Alias Regex + User Agent */}
          <div className="grid grid-cols-2 gap-3">
            <FormField label={t('groups.alias_regex')} cols="1/2">
              <input
                type="text"
                value={formAliasRegex}
                onChange={(e) => setFormAliasRegex(e.target.value)}
                placeholder={t('groups.alias_regex_placeholder')}
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </FormField>
            <FormField label={t('groups.user_agent')} cols="1/2">
              <input
                type="text"
                value={formUserAgent}
                onChange={(e) => setUserAgent(e.target.value)}
                placeholder={t('groups.user_agent_placeholder')}
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </FormField>
          </div>
        </SmoothCollapse>

        {/* Notes */}
        <FormField label={t('groups.notes')}>
          <textarea
            value={formNotes}
            onChange={(e) => setFormNotes(e.target.value)}
            placeholder={t('groups.notes_placeholder')}
            rows={2}
            className="w-full px-3 py-2 text-sm rounded-lg border resize-none"
            style={textareaStyle}
          />
        </FormField>
      </div>

      <FormActions
        onCancel={onCancel}
        onSubmit={handleSave}
        cancelLabel={t('nodes.cancel')}
        submitLabel={t('nodes.save')}
      />
    </EditFormCard>
  )
}

// ========== Main GroupsView ==========
export function GroupsView() {
  const [groups, setGroups] = useState<NodeGroup[]>([])
  const [editId, setEditId] = useState<number | null>(null)
  const [deleteTargetId, setDeleteTargetId] = useState<number | null>(null)
  const t = useT()
  const { addToast } = useStore()

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
      await groupsApi.update(id, data)
      setEditId(null)
      await loadGroups()
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
    try {
      await groupsApi.delete(id)
      setDeleteTargetId(null)
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

  // 拖拽排序后的持久化
  const handleReorder = useCallback(async (newGroups: NodeGroup[]) => {
    setGroups(newGroups)
    try {
      await groupsApi.reorder(newGroups.map((g) => g.uuid))
    } catch (err) {
      console.error('Reorder failed:', err)
      addToast(t('groups.reorder_failed'), 'error')
      await loadGroups()
    }
  }, [addToast, t, loadGroups])

  // 拖拽开始时关闭编辑/删除
  const handleDragStart = useCallback(() => {
    setEditId(null)
    setDeleteTargetId(null)
  }, [])

  // 渲染额外内容（编辑表单 + 删除确认）
  const renderExtra = useCallback((group: NodeGroup) => {
    return (
      <>
        {editId === group.ID && (
          <div className="pt-2">
            <GroupEditForm
              group={group}
              onSave={(data) => handleSave(group.ID, data)}
              onCancel={() => setEditId(null)}
              t={t}
            />
          </div>
        )}
        <DeleteConfirmBanner
          visible={deleteTargetId === group.ID}
          message={t('groups.delete_confirm', { name: group.alias || t('groups.default_name') })}
          onConfirm={() => handleDelete(group.ID)}
          onCancel={() => setDeleteTargetId(null)}
        />
      </>
    )
  }, [editId, deleteTargetId, t, handleSave, handleDelete])

  // 空状态
  const emptyContent = useMemo(() => (
    <div className="text-center py-20">
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
    </div>
  ), [t])

  // 渲染单个卡片
  const renderItem = useCallback(({ item: group, isDragging, dragListeners, dragAttributes }: {
    item: NodeGroup
    isDragging: boolean
    isOverlay: boolean
    dragListeners: Record<string, any>
    dragAttributes: Record<string, any>
  }) => {
    const displayName = group.alias || t('groups.default_name')

    return (
      <div
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
              {...dragAttributes}
              {...dragListeners}
              className="cursor-grab active:cursor-grabbing p-1 rounded-md shrink-0"
              style={{ color: 'var(--color-text-muted)' }}
            >
              <GripVertical size={14} />
            </div>

            {/* Group Icon */}
            <div
              className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0"
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
                    className="text-[10px] truncate max-w-50"
                    style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
                  >
                    {group.notes}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-1 shrink-0 ml-3">
            {/* Refresh buttons (subscription only) */}
            {group.is_subscription && (
              <>
                <button
                  onClick={() => handleRefresh(group.ID)}
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
                  onClick={() => handleRefreshProxy(group.ID)}
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
              onClick={() => setEditId(editId === group.ID ? null : group.ID)}
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
              onClick={() => {
                if (groups.length <= 1) {
                  addToast(t('groups.cannot_delete'), 'error')
                  return
                }
                setDeleteTargetId(deleteTargetId === group.ID ? null : group.ID)
              }}
              disabled={groups.length <= 1}
              className="p-1.5 rounded-md transition-colors cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
              style={{ color: 'var(--color-text-muted)' }}
              onMouseEnter={(e) => {
                if (groups.length > 1) {
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
      </div>
    )
  }, [t, editId, deleteTargetId, groups, addToast])

  return (
    <div className="max-w-3xl mx-auto flex flex-col" style={{ height: 'calc(100vh - 200px)' }}>
      {/* Header */}
      <div className="flex items-center justify-between mb-5 shrink-0">
        <h1
          className="text-xl font-semibold"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('groups.title')}
        </h1>
        <button
          onClick={handleAdd}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium cursor-pointer btn-primary"
          style={{ fontFamily: 'var(--font-heading)' }}
        >
          <Plus size={13} />
          {t('groups.add')}
        </button>
      </div>

      {/* Virtual Sortable List */}
      <VirtualSortableList
        items={groups}
        onItemsChange={handleReorder}
        renderItem={renderItem}
        renderExtra={renderExtra}
        estimateSize={72}
        overscan={5}
        className="flex-1 min-h-0"
        onDragStart={handleDragStart}
        emptyContent={emptyContent}
      />
    </div>
  )
}