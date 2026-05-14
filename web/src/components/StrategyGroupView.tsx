import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Plus, Trash2, Edit3, Shuffle, Zap, ArrowDownUp, BarChart3 } from 'lucide-react'
import { strategyGroupsApi, profileApi } from '../lib/api'
import { useT } from '../lib/i18n'
import type { Profile } from '../store'
import { EditFormCard } from './ui/EditFormCard'
import { FormField } from './ui/FormField'
import { FormActions } from './ui/FormActions'
import { inputStyle, inputHeadingStyle } from './ui/formStyles'

interface StrategyGroup {
  ID: number
  name: string
  type: string
  description: string
  profile_ids: string
  test_url: string
  test_interval: number
  strategy: string
  sort_order: number
  enabled: boolean
}

const GROUP_TYPES = [
  { value: 'selector', labelKey: 'strategy.selector' as const, icon: Shuffle },
  { value: 'urltest', labelKey: 'strategy.urltest' as const, icon: Zap },
  { value: 'fallback', labelKey: 'strategy.fallback' as const, icon: ArrowDownUp },
  { value: 'loadbalance', labelKey: 'strategy.loadbalance' as const, icon: BarChart3 },
]

// ========== Strategy Group Edit Form (独立悬浮卡片) ==========
function StrategyGroupEditForm({
  group,
  profiles,
  onSave,
  onCancel,
  t,
}: {
  group?: StrategyGroup
  profiles: Profile[]
  onSave: (data: Partial<StrategyGroup>) => void
  onCancel: () => void
  t: (key: any, params?: Record<string, any>) => string
}) {
  const isEditing = !!group
  const [formName, setFormName] = useState(group?.name || '')
  const [formType, setFormType] = useState(group?.type || 'selector')
  const [formTestURL, setFormTestURL] = useState(group?.test_url || 'https://www.gstatic.com/generate_204')
  const [formTestInterval, setFormTestInterval] = useState(String(group?.test_interval || 300))
  const [formStrategy, setFormStrategy] = useState(group?.strategy || 'round-robin')
  const [formProfileIDs, setFormProfileIDs] = useState<number[]>(() => {
    if (!group) return []
    try {
      return JSON.parse(group.profile_ids || '[]')
    } catch {
      return []
    }
  })

  const toggleProfileID = (id: number) => {
    setFormProfileIDs(prev =>
      prev.includes(id) ? prev.filter(p => p !== id) : [...prev, id]
    )
  }

  const handleSubmit = () => {
    if (!formName.trim()) return
    onSave({
      name: formName,
      type: formType,
      description: group?.description || '',
      test_url: formTestURL,
      test_interval: parseInt(formTestInterval) || 300,
      strategy: formStrategy,
      profile_ids: JSON.stringify(formProfileIDs),
    })
  }

  return (
    <EditFormCard>
      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <FormField label={t('strategy.name')} cols="1/2">
            <input
              type="text"
              value={formName}
              onChange={(e) => setFormName(e.target.value)}
              placeholder={t('strategy.group_name_placeholder')}
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={inputHeadingStyle}
            />
          </FormField>
          <FormField label={t('strategy.type')} cols="1/2">
            <select
              value={formType}
              onChange={(e) => setFormType(e.target.value)}
              className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
              style={inputHeadingStyle}
            >
              <option value="selector">{t('strategy.selector')}</option>
              <option value="urltest">{t('strategy.urltest')}</option>
              <option value="fallback">{t('strategy.fallback')}</option>
              <option value="loadbalance">{t('strategy.loadbalance')}</option>
            </select>
          </FormField>
        </div>

        {formType === 'loadbalance' && (
          <FormField label={t('strategy.strategy_label')}>
            <select
              value={formStrategy}
              onChange={(e) => setFormStrategy(e.target.value)}
              className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
              style={inputHeadingStyle}
            >
              <option value="round-robin">{t('strategy.round_robin')}</option>
              <option value="least-load">{t('strategy.least_load')}</option>
              <option value="random">{t('strategy.random')}</option>
            </select>
          </FormField>
        )}

        {(formType === 'urltest' || formType === 'fallback') && (
          <div className="grid grid-cols-2 gap-3">
            <FormField label={t('strategy.test_url')} cols="1/2">
              <input
                type="text"
                value={formTestURL}
                onChange={(e) => setFormTestURL(e.target.value)}
                placeholder="https://www.gstatic.com/generate_204"
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </FormField>
            <FormField label={t('strategy.test_interval')} cols="1/2">
              <input
                type="text"
                value={formTestInterval}
                onChange={(e) => setFormTestInterval(e.target.value)}
                placeholder="300"
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </FormField>
          </div>
        )}

        {/* Member nodes selection */}
        <FormField label={`${t('strategy.members')} (${t('strategy.members_selected', { count: formProfileIDs.length })})`}>
          <div
            className="max-h-40 overflow-y-auto rounded-lg border p-2 space-y-1"
            style={{ backgroundColor: 'var(--color-overlay)', borderColor: 'var(--color-border)' }}
          >
            {profiles.length === 0 ? (
              <p className="text-xs text-center py-4" style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}>
                {t('nodes.no_nodes')}
              </p>
            ) : (
              profiles.map(p => (
                <label
                  key={p.ID}
                  className="flex items-center gap-2 px-2 py-1 rounded cursor-pointer hover:bg-(--color-muted) transition-colors"
                >
                  <input
                    type="checkbox"
                    checked={formProfileIDs.includes(p.ID)}
                    onChange={() => toggleProfileID(p.ID)}
                    className="rounded"
                  />
                  <span className="text-xs truncate" style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}>
                    {p.name}
                  </span>
                  <span className="text-[10px] px-1 py-0.5 rounded" style={{ backgroundColor: 'var(--color-muted)', color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}>
                    {p.protocol}
                  </span>
                </label>
              ))
            )}
          </div>
        </FormField>
      </div>

      <FormActions
        onCancel={onCancel}
        onSubmit={handleSubmit}
        cancelLabel={t('nodes.cancel')}
        submitLabel={isEditing ? t('nodes.save') : t('nodes.confirm')}
        submitDisabled={!formName.trim()}
      />
    </EditFormCard>
  )
}

// ========== Main StrategyGroupView ==========
export function StrategyGroupView() {
  const [groups, setGroups] = useState<StrategyGroup[]>([])
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const t = useT()

  useEffect(() => {
    loadGroups()
    loadProfiles()
  }, [])

  const loadGroups = async () => {
    try {
      const res = await strategyGroupsApi.list()
      setGroups(res.data || [])
    } catch {
      setGroups([])
    }
  }

  const loadProfiles = async () => {
    try {
      const res = await profileApi.list()
      setProfiles(res.data || [])
    } catch {
      setProfiles([])
    }
  }

  const handleAdd = async (data: Partial<StrategyGroup>) => {
    try {
      await strategyGroupsApi.create({
        ...data,
        enabled: true,
      })
      setShowAdd(false)
      await loadGroups()
    } catch (err) {
      console.error('Add strategy group failed:', err)
    }
  }

  const handleUpdate = async (id: number, data: Partial<StrategyGroup>) => {
    try {
      await strategyGroupsApi.update(id, { ID: id, ...data })
      setEditId(null)
      await loadGroups()
    } catch (err) {
      console.error('Update strategy group failed:', err)
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await strategyGroupsApi.delete(id)
      await loadGroups()
    } catch (err) {
      console.error('Delete strategy group failed:', err)
    }
  }

  const getTypeInfo = (type: string) => {
    const info = GROUP_TYPES.find(gt => gt.value === type)
    return info || GROUP_TYPES[0]
  }

  const getProfileNames = (idsStr: string) => {
    try {
      const ids: number[] = JSON.parse(idsStr || '[]')
      return ids.map(id => {
        const p = profiles.find(pr => pr.ID === id)
        return p ? p.name : `#${id}`
      })
    } catch {
      return []
    }
  }

  return (
    <div className="max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-5">
        <h1
          className="text-xl font-semibold"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('strategy.title')}
        </h1>
        <motion.button
          onClick={() => {
            setEditId(null)
            setShowAdd(!showAdd)
          }}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium cursor-pointer btn-primary"
          style={{ fontFamily: 'var(--font-heading)' }}
          whileTap={{ scale: 0.95 }}
        >
          <Plus size={13} />
          {t('strategy.add')}
        </motion.button>
      </div>

      {/* Add form at top */}
      <AnimatePresence>
        {showAdd && (
          <StrategyGroupEditForm
            key="add-form"
            profiles={profiles}
            onSave={handleAdd}
            onCancel={() => setShowAdd(false)}
            t={t}
          />
        )}
      </AnimatePresence>

      <div className="space-y-1.5">
        {groups.length === 0 ? (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-center py-20">
            <Shuffle size={32} className="mx-auto mb-3" style={{ color: 'var(--color-text-muted)' }} />
            <p className="text-sm" style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}>
              {t('common.no_data')}
            </p>
          </motion.div>
        ) : (
          groups.map((group, index) => {
            const typeInfo = getTypeInfo(group.type)
            const TypeIcon = typeInfo.icon
            const memberNames = getProfileNames(group.profile_ids)
            return (
              <div key={group.ID}>
                <motion.div
                  layout
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.2, delay: index * 0.02 }}
                  className="rounded-xl border px-4 py-3"
                  style={{ backgroundColor: 'var(--color-card)', borderColor: 'var(--color-border)', boxShadow: 'var(--shadow-card)' }}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3 min-w-0">
                      <div
                        className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0"
                        style={{ backgroundColor: 'var(--color-accent-dim)' }}
                      >
                        <TypeIcon size={14} style={{ color: 'var(--color-accent-warm)' }} />
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium truncate" style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}>
                            {group.name}
                          </span>
                          <span
                            className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                            style={{ backgroundColor: 'var(--color-accent-dim)', color: 'var(--color-accent-warm)', fontFamily: 'var(--font-heading)' }}
                          >
                            {group.type === 'selector' ? t('strategy.selector') :
                             group.type === 'urltest' ? t('strategy.urltest') :
                             group.type === 'fallback' ? t('strategy.fallback') :
                             group.type === 'loadbalance' ? t('strategy.loadbalance') :
                             group.type}
                          </span>
                        </div>
                        <p className="text-xs mt-0.5 truncate" style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}>
                          {memberNames.length > 0 ? memberNames.join(', ') : t('strategy.members') + ': 0'}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-1 shrink-0">
                      <button
                        onClick={() => {
                          setShowAdd(false)
                          setEditId(editId === group.ID ? null : group.ID)
                        }}
                        className="p-1.5 rounded-md transition-colors cursor-pointer"
                        style={{ color: 'var(--color-text-muted)' }}
                        onMouseEnter={(e) => { e.currentTarget.style.color = 'var(--color-accent-warm)'; e.currentTarget.style.backgroundColor = 'var(--color-accent-dim)' }}
                        onMouseLeave={(e) => { e.currentTarget.style.color = 'var(--color-text-muted)'; e.currentTarget.style.backgroundColor = 'transparent' }}
                      >
                        <Edit3 size={13} />
                      </button>
                      <button
                        onClick={() => handleDelete(group.ID)}
                        className="p-1.5 rounded-md transition-colors cursor-pointer"
                        style={{ color: 'var(--color-text-muted)' }}
                        onMouseEnter={(e) => { e.currentTarget.style.color = 'var(--color-error)'; e.currentTarget.style.backgroundColor = 'var(--color-error-dim)' }}
                        onMouseLeave={(e) => { e.currentTarget.style.color = 'var(--color-text-muted)'; e.currentTarget.style.backgroundColor = 'transparent' }}
                      >
                        <Trash2 size={13} />
                      </button>
                    </div>
                  </div>
                </motion.div>

                {/* Edit form below the group card */}
                <AnimatePresence>
                  {editId === group.ID && (
                    <StrategyGroupEditForm
                      key={`edit-${group.ID}`}
                      group={group}
                      profiles={profiles}
                      onSave={(data) => handleUpdate(group.ID, data)}
                      onCancel={() => setEditId(null)}
                      t={t}
                    />
                  )}
                </AnimatePresence>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}