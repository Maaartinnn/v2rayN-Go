import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Plus, RefreshCw, Trash2, Edit3, Check, X, Link, Clock } from 'lucide-react'
import { subscriptionApi } from '../lib/api'
import { useT } from '../lib/i18n'

interface Subscription {
  ID: number
  name: string
  url: string
  enabled: boolean
  auto_update: boolean
  update_interval: number
  user_agent: string
  last_update_time: string
  node_count: number
  group_id: number
}

export function SubscriptionsView() {
  const [subs, setSubs] = useState<Subscription[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [formName, setFormName] = useState('')
  const [formUrl, setFormUrl] = useState('')
  const [formUA, setFormUA] = useState('')
  const [formInterval, setFormInterval] = useState('24')
  const [refreshing, setRefreshing] = useState(false)
  const t = useT()

  useEffect(() => {
    loadSubs()
  }, [])

  const loadSubs = async () => {
    try {
      const res = await subscriptionApi.list()
      setSubs(res.data || [])
    } catch {
      setSubs([])
    }
  }

  const handleAdd = async () => {
    if (!formName.trim() || !formUrl.trim()) return
    try {
      await subscriptionApi.create({
        name: formName,
        url: formUrl,
        user_agent: formUA,
        update_interval: parseInt(formInterval) * 3600,
        auto_update: true,
        enabled: true,
      })
      resetForm()
      setShowAdd(false)
      await loadSubs()
    } catch (err) {
      console.error('Add subscription failed:', err)
    }
  }

  const handleUpdate = async () => {
    if (editId === null || !formName.trim() || !formUrl.trim()) return
    try {
      await subscriptionApi.update(editId, {
        ID: editId,
        name: formName,
        url: formUrl,
        user_agent: formUA,
        update_interval: parseInt(formInterval) * 3600,
      })
      resetForm()
      setEditId(null)
      await loadSubs()
    } catch (err) {
      console.error('Update subscription failed:', err)
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await subscriptionApi.delete(id)
      await loadSubs()
    } catch (err) {
      console.error('Delete subscription failed:', err)
    }
  }

  const handleRefreshAll = async () => {
    setRefreshing(true)
    try {
      await subscriptionApi.refreshAll()
      setTimeout(() => {
        setRefreshing(false)
        loadSubs()
      }, 3000)
    } catch (err) {
      console.error('Refresh failed:', err)
      setRefreshing(false)
    }
  }

  const startEdit = (sub: Subscription) => {
    setEditId(sub.ID)
    setFormName(sub.name)
    setFormUrl(sub.url)
    setFormUA(sub.user_agent || '')
    setFormInterval(String(Math.floor((sub.update_interval || 86400) / 3600)))
    setShowAdd(false)
  }

  const resetForm = () => {
    setFormName('')
    setFormUrl('')
    setFormUA('')
    setFormInterval('24')
  }

  const cancelForm = () => {
    resetForm()
    setShowAdd(false)
    setEditId(null)
  }

  const formatDate = (dateStr: string) => {
    if (!dateStr || dateStr === '0001-01-01T00:00:00Z') return '-'
    return new Date(dateStr).toLocaleString()
  }

  const renderForm = () => (
    <motion.div
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: 'auto' }}
      exit={{ opacity: 0, height: 0 }}
      className="mb-4 overflow-hidden"
    >
      <div
        className="rounded-xl border p-5"
        style={{
          backgroundColor: 'var(--color-card)',
          borderColor: 'var(--color-border)',
          boxShadow: 'var(--shadow-card)',
        }}
      >
        <div className="space-y-3">
          <div>
            <label
              className="text-xs font-medium block mb-1"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('subs.name')}
            </label>
            <input
              type="text"
              value={formName}
              onChange={(e) => setFormName(e.target.value)}
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
              {t('subs.url')}
            </label>
            <input
              type="text"
              value={formUrl}
              onChange={(e) => setFormUrl(e.target.value)}
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={{
                backgroundColor: 'var(--color-overlay)',
                borderColor: 'var(--color-border)',
                color: 'var(--color-foreground)',
                fontFamily: 'var(--font-mono)',
              }}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                className="text-xs font-medium block mb-1"
                style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
              >
                {t('subs.user_agent')}
              </label>
              <input
                type="text"
                value={formUA}
                onChange={(e) => setFormUA(e.target.value)}
                placeholder="ClashForAndroid/2.5.12"
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
                {t('subs.interval')}
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
            </div>
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-4">
          <button
            onClick={cancelForm}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
            style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            <X size={13} />
            {t('nodes.cancel')}
          </button>
          <button
            onClick={editId !== null ? handleUpdate : handleAdd}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
            style={{
              backgroundColor: 'var(--color-primary)',
              color: 'var(--color-primary-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            <Check size={13} />
            {editId !== null ? t('nodes.save') : t('nodes.confirm')}
          </button>
        </div>
      </div>
    </motion.div>
  )

  return (
    <div className="max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <h1
          className="text-xl font-semibold"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('subs.title')}
        </h1>
        <div className="flex gap-2">
          <motion.button
            onClick={handleRefreshAll}
            disabled={refreshing}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors cursor-pointer"
            style={{
              backgroundColor: 'var(--color-muted)',
              borderColor: 'var(--color-border)',
              color: 'var(--color-muted-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
            whileTap={{ scale: 0.95 }}
          >
            <RefreshCw size={13} className={refreshing ? 'animate-spin' : ''} />
            {t('subs.refresh_all')}
          </motion.button>
          <motion.button
            onClick={() => { cancelForm(); setShowAdd(!showAdd) }}
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
            {t('subs.add')}
          </motion.button>
        </div>
      </div>

      {/* Add/Edit Form */}
      <AnimatePresence>
        {(showAdd || editId !== null) && renderForm()}
      </AnimatePresence>

      {/* Subscription List */}
      <div className="space-y-1.5">
        {subs.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-20"
          >
            <Link
              size={32}
              className="mx-auto mb-3"
              style={{ color: 'var(--color-text-muted)' }}
            />
            <p
              className="text-sm"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('subs.no_subscriptions')}
            </p>
            <p
              className="text-xs mt-1"
              style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
            >
              {t('subs.add_hint')}
            </p>
          </motion.div>
        ) : (
          subs.map((sub, index) => (
            <motion.div
              key={sub.ID}
              layout
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.2, delay: index * 0.02 }}
              className="rounded-xl border px-4 py-3"
              style={{
                backgroundColor: 'var(--color-card)',
                borderColor: 'var(--color-border)',
                boxShadow: 'var(--shadow-card)',
              }}
            >
              <div className="flex items-center justify-between">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span
                      className="text-sm font-medium truncate"
                      style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                    >
                      {sub.name}
                    </span>
                    {sub.node_count > 0 && (
                      <span
                        className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                        style={{
                          backgroundColor: 'var(--color-accent-dim)',
                          color: 'var(--color-accent-warm)',
                          fontFamily: 'var(--font-heading)',
                        }}
                      >
                        {sub.node_count} {t('subs.node_count')}
                      </span>
                    )}
                  </div>
                  <p
                    className="text-xs mt-0.5 truncate"
                    style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                  >
                    {sub.url}
                  </p>
                  <div className="flex items-center gap-3 mt-1">
                    <span
                      className="text-[10px] flex items-center gap-1"
                      style={{ color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)' }}
                    >
                      <Clock size={10} />
                      {formatDate(sub.last_update_time)}
                    </span>
                    {sub.auto_update && (
                      <span
                        className="text-[10px]"
                        style={{ color: 'var(--color-success)', fontFamily: 'var(--font-heading)' }}
                      >
                        {t('subs.auto_update')} · {Math.floor((sub.update_interval || 86400) / 3600)}h
                      </span>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-1 flex-shrink-0 ml-3">
                  <button
                    onClick={() => startEdit(sub)}
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
                    onClick={() => handleDelete(sub.ID)}
                    className="p-1.5 rounded-md transition-colors cursor-pointer"
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
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            </motion.div>
          ))
        )}
      </div>
    </div>
  )
}