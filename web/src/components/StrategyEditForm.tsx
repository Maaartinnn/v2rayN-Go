import { useState, useEffect } from 'react'
import { profileApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { isStrategyGroup } from '../lib/constants'
import type { Profile, ProfileListItem } from '../store'
import { EditFormCard } from './ui/EditFormCard'
import { FormField } from './ui/FormField'
import { FormActions } from './ui/FormActions'
import { inputStyle } from './ui/formStyles'

interface StrategyEditFormProps {
  editData?: Profile            // 编辑模式：传入完整 Profile 数据
  onClose: () => void
  onSaved: (updatedProfile?: Profile) => void
  groupUUID?: string
}

const STRATEGY_TYPES = [
  { value: 'selector', labelKey: 'strategy.selector' },
  { value: 'urltest', labelKey: 'strategy.urltest' },
  { value: 'fallback', labelKey: 'strategy.fallback' },
  { value: 'loadbalance', labelKey: 'strategy.loadbalance' },
]

const BALANCE_STRATEGIES = [
  { value: 'round-robin', labelKey: 'strategy.round_robin' },
  { value: 'least-load', labelKey: 'strategy.least_load' },
  { value: 'random', labelKey: 'strategy.random' },
]

// parseMemberUUIDs 从 JSON 字符串解析成员 UUID 列表，容错返回空数组
function parseMemberUUIDs(json: string | undefined): string[] {
  if (!json) return []
  try {
    const arr = JSON.parse(json)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

export function StrategyEditForm({ editData, onClose, onSaved, groupUUID }: StrategyEditFormProps) {
  const t = useT()
  const isEditMode = !!editData

  const [name, setName] = useState('')
  const [strategyType, setStrategyType] = useState('selector')
  const [memberUUIDs, setMemberUUIDs] = useState<string[]>([])
  const [testURL, setTestURL] = useState('https://www.gstatic.com/generate_204')
  const [testInterval, setTestInterval] = useState('300')
  const [balanceStrategy, setBalanceStrategy] = useState('random')

  // Available proxy profiles for member selection
  const [allProfiles, setAllProfiles] = useState<ProfileListItem[]>([])

  // 表单回填：监听 editData 变化，确保 Drawer 复用时表单状态正确重置
  useEffect(() => {
    if (editData) {
      setName(editData.name)
      setStrategyType(editData.proxy_protocol)
      setMemberUUIDs(parseMemberUUIDs(editData.strategy_member_uuids))
      setTestURL(editData.strategy_test_url || 'https://www.gstatic.com/generate_204')
      setTestInterval(String(editData.strategy_test_interval || 300))
      setBalanceStrategy(editData.strategy_type || 'random')
    } else {
      // 新建模式：重置为默认值
      setName('')
      setStrategyType('selector')
      setMemberUUIDs([])
      setTestURL('https://www.gstatic.com/generate_204')
      setTestInterval('300')
      setBalanceStrategy('random')
    }
  }, [editData])

  // 加载可选的代理节点列表（排除策略组节点）
  useEffect(() => {
    profileApi.list(groupUUID || editData?.group_uuid).then((res) => {
      const proxyNodes = (res.data || []).filter(
        (p: ProfileListItem) => !isStrategyGroup(p.proxy_protocol)
      ) as ProfileListItem[]
      setAllProfiles(proxyNodes)
    }).catch(() => {})
  }, [groupUUID, editData?.group_uuid])

  const toggleMember = (uuid: string) => {
    setMemberUUIDs((prev) =>
      prev.includes(uuid) ? prev.filter((u) => u !== uuid) : [...prev, uuid]
    )
  }

  const handleSubmit = async () => {
    if (!name.trim()) return

    const payload: Record<string, unknown> = {
      name: name.trim(),
      proxy_protocol: strategyType,
      proxy_address: '',
      proxy_port: 0,
      strategy_member_uuids: JSON.stringify(memberUUIDs),
      strategy_test_url: testURL,
      strategy_test_interval: parseInt(testInterval) || 300,
      strategy_type: balanceStrategy,
    }

    try {
      if (isEditMode && editData) {
        // 编辑模式：通过 update API 更新
        const res = await profileApi.update(editData.uuid, payload)
        onSaved(res.data)
      } else {
        // 新建模式
        payload.group_uuid = groupUUID || ''
        payload.is_active = false
        payload.sort_order = 0
        await profileApi.create(payload)
        onSaved()
      }
      onClose()
    } catch (err) {
      console.error('Failed to save strategy group:', err)
    }
  }

  const showTestURL = strategyType === 'urltest' || strategyType === 'fallback'
  const showBalance = strategyType === 'loadbalance'

  return (
    <EditFormCard>
      <div className="space-y-4">
        {/* Name */}
        <FormField label={t('strategy.name')}>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t('strategy.group_name_placeholder')}
            className="w-full px-3 py-2 text-sm rounded-lg border"
            style={inputStyle}
          />
        </FormField>

        {/* Strategy Type */}
        <FormField label={t('strategy.type')}>
          <select
            value={strategyType}
            onChange={(e) => setStrategyType(e.target.value)}
            className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
            style={inputStyle}
          >
            {STRATEGY_TYPES.map((st) => (
              <option key={st.value} value={st.value}>{t(st.labelKey as Parameters<typeof t>[0])}</option>
            ))}
          </select>
        </FormField>

        {/* Test URL (for urltest/fallback) */}
        {showTestURL && (
          <div className="grid grid-cols-2 gap-3">
            <FormField label={t('strategy.test_url')} cols="1/2">
              <input
                type="text"
                value={testURL}
                onChange={(e) => setTestURL(e.target.value)}
                placeholder="https://www.gstatic.com/generate_204"
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </FormField>
            <FormField label={t('strategy.test_interval')} cols="1/2">
              <input
                type="number"
                value={testInterval}
                onChange={(e) => setTestInterval(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </FormField>
          </div>
        )}

        {/* Balance Strategy (for loadbalance) */}
        {showBalance && (
          <FormField label={t('strategy.strategy_label')}>
            <select
              value={balanceStrategy}
              onChange={(e) => setBalanceStrategy(e.target.value)}
              className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
              style={inputStyle}
            >
              {BALANCE_STRATEGIES.map((bs) => (
                <option key={bs.value} value={bs.value}>{t(bs.labelKey as Parameters<typeof t>[0])}</option>
              ))}
            </select>
          </FormField>
        )}

        {/* Member Selection */}
        <FormField label={t('strategy.members')}>
          <div className="text-xs mb-2" style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}>
            {t('strategy.members_selected', { count: memberUUIDs.length })}
          </div>
          <div
            className="max-h-48 overflow-y-auto rounded-lg border p-2 space-y-1"
            style={{
              backgroundColor: 'var(--color-muted)',
              borderColor: 'var(--color-border-subtle)',
            }}
          >
            {allProfiles.length === 0 ? (
              <div className="text-center py-4 text-xs" style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}>
                {t('common.no_data')}
              </div>
            ) : (
              allProfiles.map((p) => (
                <label
                  key={p.uuid}
                  className="flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer transition-colors"
                  style={{
                    backgroundColor: memberUUIDs.includes(p.uuid) ? 'var(--color-accent-dim)' : 'transparent',
                  }}
                >
                  <input
                    type="checkbox"
                    checked={memberUUIDs.includes(p.uuid)}
                    onChange={() => toggleMember(p.uuid)}
                  />
                  <span className="text-xs truncate" style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}>
                    {p.name}
                  </span>
                  <span className="text-[10px] shrink-0 ml-auto" style={{ color: 'var(--color-muted-foreground)' }}>
                    {p.proxy_protocol}
                  </span>
                </label>
              ))
            )}
          </div>
        </FormField>

      </div>

      <FormActions
        onCancel={onClose}
        onSubmit={handleSubmit}
        cancelLabel={t('nodes.cancel')}
        submitLabel={isEditMode ? t('nodes.save') || '保存' : t('nodes.confirm')}
        submitDisabled={!name.trim() || (!isEditMode && !groupUUID)}
      />
    </EditFormCard>
  )
}