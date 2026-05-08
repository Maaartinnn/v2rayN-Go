import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Plus, Trash2, Edit3, Check, X, Route, Shield, Zap, Globe } from 'lucide-react'
import { routingApi } from '../lib/api'
import { useT } from '../lib/i18n'

interface RoutingRule {
  ID: number
  name: string
  type: string
  domain: string
  ip: string
  port: string
  enabled: boolean
  sort_order: number
}

export function RoutingView() {
  const [rules, setRules] = useState<RoutingRule[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [formName, setFormName] = useState('')
  const [formType, setFormType] = useState('direct')
  const [formDomain, setFormDomain] = useState('')
  const [formIP, setFormIP] = useState('')
  const t = useT()

  useEffect(() => {
    loadRules()
  }, [])

  const loadRules = async () => {
    try {
      const res = await routingApi.list()
      setRules(res.data || [])
    } catch {
      setRules([])
    }
  }

  const handleAdd = async () => {
    if (!formName.trim()) return
    try {
      await routingApi.create({
        name: formName,
        type: formType,
        domain: formDomain,
        ip: formIP,
        enabled: true,
        sort_order: rules.length,
      })
      resetForm()
      setShowAdd(false)
      await loadRules()
    } catch (err) {
      console.error('Add rule failed:', err)
    }
  }

  const handleUpdate = async () => {
    if (editId === null || !formName.trim()) return
    try {
      await routingApi.update(editId, {
        ID: editId,
        name: formName,
        type: formType,
        domain: formDomain,
        ip: formIP,
      })
      resetForm()
      setEditId(null)
      await loadRules()
    } catch (err) {
      console.error('Update rule failed:', err)
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await routingApi.delete(id)
      await loadRules()
    } catch (err) {
      console.error('Delete rule failed:', err)
    }
  }

  const startEdit = (rule: RoutingRule) => {
    setEditId(rule.ID)
    setFormName(rule.name)
    setFormType(rule.type)
    setFormDomain(rule.domain)
    setFormIP(rule.ip)
    setShowAdd(false)
  }

  const resetForm = () => {
    setFormName('')
    setFormType('direct')
    setFormDomain('')
    setFormIP('')
  }

  const cancelForm = () => {
    resetForm()
    setShowAdd(false)
    setEditId(null)
  }

  const getTypeColor = (type: string) => {
    const colors: Record<string, { bg: string; text: string; icon: typeof Route }> = {
      direct: { bg: 'var(--color-success-dim)', text: 'var(--color-success)', icon: Zap },
      proxy: { bg: 'var(--color-accent-dim)', text: 'var(--color-accent-warm)', icon: Globe },
      block: { bg: 'var(--color-error-dim)', text: 'var(--color-error)', icon: Shield },
    }
    return colors[type] || { bg: 'var(--color-muted)', text: 'var(--color-muted-foreground)', icon: Route }
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
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                className="text-xs font-medium block mb-1"
                style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
              >
                {t('routing.type')}
              </label>
              <select
                value={formType}
                onChange={(e) => setFormType(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
                style={{
                  backgroundColor: 'var(--color-overlay)',
                  borderColor: 'var(--color-border)',
                  color: 'var(--color-foreground)',
                  fontFamily: 'var(--font-heading)',
                }}
              >
                <option value="direct">{t('routing.direct')}</option>
                <option value="proxy">{t('routing.proxy')}</option>
                <option value="block">{t('routing.block')}</option>
              </select>
            </div>
            <div>
              <label
                className="text-xs font-medium block mb-1"
                style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
              >
                {t('common.search')}
              </label>
              <input
                type="text"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                placeholder="Rule name"
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={{
                  backgroundColor: 'var(--color-overlay)',
                  borderColor: 'var(--color-border)',
                  color: 'var(--color-foreground)',
                  fontFamily: 'var(--font-heading)',
                }}
              />
            </div>
          </div>
          <div>
            <label
              className="text-xs font-medium block mb-1"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('routing.domain')}
            </label>
            <input
              type="text"
              value={formDomain}
              onChange={(e) => setFormDomain(e.target.value)}
              placeholder="google.com, github.com, geosite:cn"
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
              {t('routing.ip')}
            </label>
            <input
              type="text"
              value={formIP}
              onChange={(e) => setFormIP(e.target.value)}
              placeholder="192.168.0.0/16, geoip:cn"
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
      <div className="flex items-center justify-between mb-5">
        <h1
          className="text-xl font-semibold"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('routing.title')}
        </h1>
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
          {t('routing.add')}
        </motion.button>
      </div>

      <AnimatePresence>
        {(showAdd || editId !== null) && renderForm()}
      </AnimatePresence>

      <div className="space-y-1.5">
        {rules.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-20"
          >
            <Route
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
          rules.map((rule, index) => {
            const typeInfo = getTypeColor(rule.type)
            const TypeIcon = typeInfo.icon
            return (
              <motion.div
                key={rule.ID}
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
                  <div className="flex items-center gap-3 min-w-0">
                    <div
                      className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
                      style={{ backgroundColor: typeInfo.bg }}
                    >
                      <TypeIcon size={14} style={{ color: typeInfo.text }} />
                    </div>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span
                          className="text-sm font-medium truncate"
                          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                        >
                          {rule.name}
                        </span>
                        <span
                          className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                          style={{
                            backgroundColor: typeInfo.bg,
                            color: typeInfo.text,
                            fontFamily: 'var(--font-heading)',
                          }}
                        >
                          {rule.type}
                        </span>
                      </div>
                      <p
                        className="text-xs mt-0.5 truncate"
                        style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                      >
                        {rule.domain && `🌐 ${rule.domain}`}
                        {rule.domain && rule.ip && ' · '}
                        {rule.ip && `📡 ${rule.ip}`}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-1 flex-shrink-0">
                    <button
                      onClick={() => startEdit(rule)}
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
                      onClick={() => handleDelete(rule.ID)}
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
            )
          })
        )}
      </div>
    </div>
  )
}