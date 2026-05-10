import { useEffect, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  ScanLine,
  PenLine,
  FolderOpen,
  Link,
  RefreshCw,
  Globe,
  Check,
} from 'lucide-react'
import { groupsApi, profileApi, profileEnhancedApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { useStore } from '../store'
import { NodeEditForm } from './NodeEditForm'

interface NodeGroup {
  ID: number
  uuid: string
  alias: string
  is_subscription: boolean
  node_count: number
  enabled: boolean
}

export function ImportView() {
  const [groups, setGroups] = useState<NodeGroup[]>([])
  const [selectedGroupId, setSelectedGroupId] = useState<number>(0)
  const [importText, setImportText] = useState('')
  const [showManualAdd, setShowManualAdd] = useState(false)
  const [importing, setImporting] = useState(false)
  const [refreshing, setRefreshing] = useState<'direct' | 'proxy' | null>(null)
  const t = useT()
  const { addToast } = useStore()

  const loadGroups = useCallback(async () => {
    try {
      const res = await groupsApi.list()
      const data = res.data || []
      setGroups(data)
      // Auto-select first group if none selected
      if (selectedGroupId === 0 && data.length > 0) {
        setSelectedGroupId(data[0].ID)
      }
    } catch {
      setGroups([])
    }
  }, [selectedGroupId])

  useEffect(() => {
    loadGroups()
  }, [loadGroups])

  const selectedGroup = groups.find((g) => g.ID === selectedGroupId)
  const isSubscriptionGroup = selectedGroup?.is_subscription ?? false

  const handleImport = async () => {
    if (!importText.trim()) {
      addToast(t('import.no_links'), 'error')
      return
    }
    setImporting(true)
    try {
      const res = await profileApi.importToGroup(importText, selectedGroupId)
      const count = res.data?.imported || 0
      addToast(t('import.import_success', { count }), 'success')
      setImportText('')
      await loadGroups()
    } catch (err) {
      console.error('Import failed:', err)
      addToast(t('import.import_failed'), 'error')
    }
    setImporting(false)
  }

  const handleImageImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const formData = new FormData()
    formData.append('image', file)
    if (selectedGroupId > 0) {
      formData.append('group_id', String(selectedGroupId))
    }
    try {
      const res = await profileEnhancedApi.importImage(formData)
      const count = res.data?.imported || 0
      if (count > 0) {
        addToast(t('import.import_success', { count }), 'success')
        await loadGroups()
      }
    } catch (err) {
      console.error('Image import failed:', err)
      addToast(t('import.import_failed'), 'error')
    }
    e.target.value = ''
  }

  const handleRefresh = async (useProxy: boolean) => {
    if (!selectedGroupId || !isSubscriptionGroup) return
    setRefreshing(useProxy ? 'proxy' : 'direct')
    try {
      if (useProxy) {
        await groupsApi.refreshProxy(selectedGroupId)
      } else {
        await groupsApi.refresh(selectedGroupId)
      }
      addToast(t('groups.update_success'), 'success')
    } catch (err) {
      console.error('Refresh failed:', err)
      addToast(t('groups.update_failed'), 'error')
    }
    setRefreshing(null)
  }

  const displayName = (g: NodeGroup) => g.alias || t('groups.default_name')

  return (
    <div className="flex gap-6 max-w-5xl mx-auto">
      {/* Left: Import Area (3/4) */}
      <div className="flex-1 min-w-0">
        <h1
          className="text-xl font-semibold mb-5"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('import.title')}
        </h1>

        {/* Share Links Import */}
        <div
          className="rounded-xl border p-5 mb-4"
          style={{
            backgroundColor: 'var(--color-card)',
            borderColor: 'var(--color-border)',
            boxShadow: 'var(--shadow-card)',
          }}
        >
          <label
            className="text-xs font-medium block mb-2"
            style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {t('import.links')}
          </label>
          <textarea
            value={importText}
            onChange={(e) => setImportText(e.target.value)}
            placeholder={t('import.links_placeholder')}
            className="w-full h-28 rounded-lg p-3 text-sm resize-none border"
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
            {t('import.base64_hint')}
          </p>
          <div className="flex justify-end gap-2 mt-3">
            <motion.button
              onClick={handleImport}
              disabled={importing}
              className="flex items-center gap-1.5 px-4 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
              style={{
                backgroundColor: 'var(--color-primary)',
                color: 'var(--color-primary-foreground)',
                fontFamily: 'var(--font-heading)',
              }}
              whileTap={{ scale: 0.95 }}
            >
              <Check size={13} />
              {t('import.confirm')}
            </motion.button>
          </div>
        </div>

        {/* Action Buttons Row */}
        <div className="flex gap-2 mb-4">
          <motion.button
            onClick={() => setShowManualAdd(!showManualAdd)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors cursor-pointer"
            style={{
              backgroundColor: 'var(--color-muted)',
              borderColor: 'var(--color-border)',
              color: 'var(--color-muted-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
            whileTap={{ scale: 0.95 }}
          >
            <PenLine size={13} />
            {t('import.manual_add')}
          </motion.button>
          <motion.button
            onClick={() => {
              const input = document.createElement('input')
              input.type = 'file'
              input.accept = 'image/*'
              input.onchange = (e) =>
                handleImageImport(e as unknown as React.ChangeEvent<HTMLInputElement>)
              input.click()
            }}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors cursor-pointer"
            style={{
              backgroundColor: 'var(--color-muted)',
              borderColor: 'var(--color-border)',
              color: 'var(--color-muted-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
            whileTap={{ scale: 0.95 }}
          >
            <ScanLine size={13} />
            {t('import.qr_code')}
          </motion.button>
        </div>

        {/* Manual Add Form */}
        <AnimatePresence>
          {showManualAdd && (
            <NodeEditForm
              onClose={() => setShowManualAdd(false)}
              onSaved={loadGroups}
            />
          )}
        </AnimatePresence>
      </div>

      {/* Right: Group Selection Panel (1/4) */}
      <div className="w-64 flex-shrink-0">
        <div className="sticky top-20">
          {/* Update Buttons */}
          <div className="flex gap-2 mb-3">
            <motion.button
              onClick={() => handleRefresh(false)}
              disabled={!isSubscriptionGroup || refreshing !== null}
              className="flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-[10px] font-medium rounded-lg border transition-colors cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
              style={{
                backgroundColor: 'var(--color-muted)',
                borderColor: 'var(--color-border)',
                color: 'var(--color-muted-foreground)',
                fontFamily: 'var(--font-heading)',
              }}
              whileTap={{ scale: 0.95 }}
              title={t('groups.update_no_proxy')}
            >
              <RefreshCw size={11} className={refreshing === 'direct' ? 'animate-spin' : ''} />
              {t('groups.update_no_proxy')}
            </motion.button>
            <motion.button
              onClick={() => handleRefresh(true)}
              disabled={!isSubscriptionGroup || refreshing !== null}
              className="flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-[10px] font-medium rounded-lg border transition-colors cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
              style={{
                backgroundColor: 'var(--color-muted)',
                borderColor: 'var(--color-border)',
                color: 'var(--color-muted-foreground)',
                fontFamily: 'var(--font-heading)',
              }}
              whileTap={{ scale: 0.95 }}
              title={t('groups.update_with_proxy')}
            >
              <Globe size={11} className={refreshing === 'proxy' ? 'animate-spin' : ''} />
              {t('groups.update_with_proxy')}
            </motion.button>
          </div>

          {/* Group List */}
          <div
            className="rounded-xl border overflow-hidden"
            style={{
              backgroundColor: 'var(--color-card)',
              borderColor: 'var(--color-border)',
              boxShadow: 'var(--shadow-card)',
            }}
          >
            <div className="p-2 space-y-1">
              {groups.length === 0 ? (
                <div className="text-center py-8">
                  <FolderOpen
                    size={20}
                    className="mx-auto mb-2"
                    style={{ color: 'var(--color-text-muted)' }}
                  />
                  <p
                    className="text-[10px]"
                    style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
                  >
                    {t('common.no_data')}
                  </p>
                </div>
              ) : (
                groups.map((group) => {
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
                            color: isSelected
                              ? 'var(--color-accent-warm)'
                              : 'var(--color-foreground)',
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
                })
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}