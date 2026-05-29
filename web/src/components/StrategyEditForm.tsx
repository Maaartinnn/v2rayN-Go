import { useState, useEffect } from 'react'
import { profileApi } from '../lib/api'
import { useT } from '../lib/i18n'
import type { ProfileListItem } from '../store'
import { EditFormCard } from './ui/EditFormCard'
import { FormField } from './ui/FormField'
import { FormActions } from './ui/FormActions'
import { inputStyle } from './ui/formStyles'

interface StrategyEditFormProps {
  onClose: () => void
  onSaved: () => void
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

export function StrategyEditForm({ onClose, onSaved, groupUUID }: StrategyEditFormProps) {
  const t = useT()

  const [name, setName] = useState('')
  const [strategyType, setStrategyType] = useState('selector')
  const [memberUUIDs, setMemberUUIDs] = useState<string[]>([])
  const [testURL, setTestURL] = useState('https://www.gstatic.com/generate_204')
  const [testInterval, setTestInterval] = useState('300')
  const [balanceStrategy, setBalanceStrategy] = useState('random')
  // Available proxy profiles for member selection
  const [allProfiles, setAllProfiles] = useState<ProfileListItem[]>([])

  useEffect(() => {
    profileApi.list(groupUUID).then((res) => {
      // Filter out strategy group nodes (only show proxy nodes)
      const proxyNodes = (res.data || []).filter(
        (p: ProfileListItem) => !['selector', 'urltest', 'fallback', 'loadbalance'].includes(p.proxy_protocol)
      ) as ProfileListItem[]
      setAllProfiles(proxyNodes)
    }).catch(() => {})
  }, [groupUUID])

  const toggleMember = (uuid: string) => {
    setMemberUUIDs((prev) =>
      prev.includes(uuid) ? prev.filter((u) => u !== uuid) : [...prev, uuid]
    )
  }

  const handleSubmit = async () => {
    if (!name.trim()) return

    const payload = {
      group_uuid: groupUUID || '',
      name: name.trim(),
      proxy_protocol: strategyType,
      proxy_address: '', // strategy nodes have no address
      proxy_port: 0,
      strategy_member_uuids: JSON.stringify(memberUUIDs),
      strategy_test_url: testURL,
      strategy_test_interval: parseInt(testInterval) || 300,
      strategy_type: balanceStrategy,
      is_active: false,
      sort_order: 0,
    }

    try {
      await profileApi.create(payload)
      onSaved()
      onClose()
    } catch (err) {
      console.error('Failed to create strategy group:', err)
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
        submitLabel={t('nodes.confirm')}
        submitDisabled={!name.trim() || !groupUUID}
      />
    </EditFormCard>
  )
}