import { useState, useEffect } from 'react'
import { Zap } from 'lucide-react'
import { profileApi, groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'
import type { Profile } from '../store'
import { PROTOCOLS, NETWORKS, TLS_OPTIONS, SECURITY_METHODS } from '../lib/coreMap'
import { EditFormCard } from './ui/EditFormCard'
import { FormField } from './ui/FormField'
import { FormActions } from './ui/FormActions'
import { inputStyle, inputHeadingStyle } from './ui/formStyles'

interface NodeEditFormProps {
  onClose: () => void
  onSaved: (updatedProfile?: Profile) => void
  groupUUID?: string
  editData?: Profile
}

export function NodeEditForm({ onClose, onSaved, groupUUID, editData }: NodeEditFormProps) {
  const t = useT()
  const isEditing = !!editData

  // Basic fields
  const [name, setName] = useState('')
  const [protocol, setProtocol] = useState('vmess')
  const [address, setAddress] = useState('')
  const [port, setPort] = useState('443')
  const [uuid, setUuid] = useState('')
  const [security, setSecurity] = useState('auto')

  // Transport
  const [network, setNetwork] = useState('tcp')
  const [host, setHost] = useState('')
  const [path, setPath] = useState('')

  // TLS
  const [tls, setTls] = useState('')
  const [sni, setSni] = useState('')
  const [fingerprint, setFingerprint] = useState('')
  const [allowInsecure, setAllowInsecure] = useState(false)

  // Reality
  const [publicKey, setPublicKey] = useState('')
  const [shortId, setShortId] = useState('')
  const [flow, setFlow] = useState('')

  // Group selection
  const [selectedGroupUUID, setSelectedGroupUUID] = useState<string>(groupUUID || '')
  const [groups, setGroups] = useState<Array<{ ID: number; uuid: string; alias: string; is_subscription: boolean }>>([])

  // Core selection（由后端一次性下发能力矩阵，前端协议切换时零延迟查字典）
  const [coreMatrix, setCoreMatrix] = useState<Record<string, string[]>>({})
  const [kernelMode, setKernelMode] = useState<'auto' | 'manual'>('auto')
  const [manualCore, setManualCore] = useState('')

  // Load groups on mount
  useEffect(() => {
    groupsApi.list().then((res) => {
      setGroups(res.data || [])
    }).catch(() => {})
  }, [])

  // 加载能力矩阵（一次性获取所有协议的可用内核）
  // 编辑/新增均从 GET /api/profiles/core-matrix 获取，保持逻辑统一
  const loadCoreMatrix = () => {
    profileApi.coreMatrix()
      .then(res => setCoreMatrix(res.data?.core_matrix || {}))
      .catch(() => setCoreMatrix({}))
  }

  // Pre-fill form when editing（同时从后端获取能力矩阵）
  useEffect(() => {
    if (editData) {
      setName(editData.name || '')
      setSelectedGroupUUID(editData.group_uuid || '')
      setProtocol(editData.proxy_protocol || 'vmess')
      setAddress(editData.proxy_address || '')
      setPort(String(editData.proxy_port || 443))
      setUuid(editData.proxy_credential || '')
      setSecurity(editData.proxy_security || 'auto')
      setNetwork(editData.proxy_network || 'tcp')
      setHost(editData.proxy_host || '')
      setPath(editData.proxy_path || '')
      setTls(editData.proxy_tls || '')
      setSni(editData.proxy_sni || '')
      setFingerprint(editData.proxy_fingerprint || '')
      setAllowInsecure(editData.proxy_allow_insecure || false)
      setPublicKey(editData.proxy_public_key || '')
      setShortId(editData.proxy_short_id || '')
      setFlow(editData.proxy_flow || '')

      // 内核设置：有 core_type 则手动，否则自动
      if (editData.core_type) {
        setKernelMode('manual')
        setManualCore(editData.core_type)
      } else {
        setKernelMode('auto')
        setManualCore('')
      }

      // 从后端获取完整能力矩阵
      loadCoreMatrix()
    } else {
      // 新增模式默认自动
      setKernelMode('auto')
      setManualCore('')
      // 从后端获取完整能力矩阵
      loadCoreMatrix()
    }
  }, [editData])

  const handleSubmit = async () => {
    if (!name.trim() || !address.trim() || !port) return

    const payload = {
      group_uuid: selectedGroupUUID,
      name: name.trim(),
      proxy_address: address.trim(),
      proxy_port: parseInt(port) || 443,
      proxy_protocol: protocol,
      proxy_credential: uuid.trim(),
      proxy_security: security,
      proxy_network: network,
      proxy_host: host.trim(),
      proxy_path: path.trim(),
      proxy_tls: tls,
      proxy_sni: sni.trim(),
      proxy_fingerprint: fingerprint.trim(),
      proxy_allow_insecure: allowInsecure,
      proxy_public_key: publicKey.trim(),
      proxy_short_id: shortId.trim(),
      proxy_flow: flow.trim(),
      core_type: kernelMode === 'manual' ? manualCore : '',
    }

    try {
      if (isEditing && editData) {
        const res = await profileApi.update(editData.uuid, payload)
        onSaved(res.data)
      } else {
        await profileApi.create({
          ...payload,
          is_active: false,
          sort_order: 0,
        })
        onSaved()
      }
      onClose()
    } catch (err) {
      console.error(`Failed to ${isEditing ? 'update' : 'create'} node:`, err)
    }
  }

  // Determine which fields to show based on protocol
  const showVLESSFields = protocol === 'vless'
  const showUUID = ['vmess', 'vless', 'trojan'].includes(protocol)
  const showPassword = ['trojan', 'shadowsocks', 'hysteria2'].includes(protocol)
  const showSecurity = protocol === 'vmess'
  const showTransport = ['vmess', 'vless', 'trojan'].includes(protocol)
  const showReality = showVLESSFields && tls === 'reality'
  const showFlow = showVLESSFields

  // 当前协议的可用内核 = 从能力矩阵中瞬间查出（零延迟，无需网络请求）
  const currentCores = coreMatrix[protocol] || []
  // 推荐内核 = 当前协议可用内核列表的第一个（后端已按推荐优先级排序）
  const recommendedCore = currentCores.length > 0 ? currentCores[0] : null

  return (
    <EditFormCard>
      <div className="space-y-4">
        {/* Row 0: Name (full width) */}
        <FormField label={t('strategy.name')}>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="My Node"
            className="w-full px-3 py-2 text-sm rounded-lg border"
            style={inputStyle}
          />
        </FormField>

        {/* Row 1: Protocol + Group (group selector only shown when editing) */}
        <div className={isEditing ? "grid grid-cols-2 gap-3" : ""}>
          <FormField label={t('routing.type')} cols={isEditing ? "1/2" : "full"}>
            <select
              value={protocol}
              onChange={(e) => setProtocol(e.target.value)}
              className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
              style={inputStyle}
            >
              {PROTOCOLS.map((p) => (
                <option key={p.value} value={p.value}>{p.label}</option>
              ))}
            </select>
          </FormField>
          {isEditing && (
            <FormField label={t('nodes.group')} cols="1/2">
              <select
                value={selectedGroupUUID}
                onChange={(e) => setSelectedGroupUUID(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
                style={inputStyle}
              >
                {groups.map((g) => (
                  <option key={g.uuid} value={g.uuid}>{g.alias || t('groups.default_name')}</option>
                ))}
              </select>
            </FormField>
          )}
        </div>

        {/* Row 2: Address + Port */}
        <div className="grid grid-cols-3 gap-3">
          <FormField label={t('nodes.address')} cols="2/3">
            <input
              type="text"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="example.com"
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={inputStyle}
            />
          </FormField>
          <FormField label={t('nodes.port')} cols="1/3">
            <input
              type="text"
              value={port}
              onChange={(e) => setPort(e.target.value)}
              placeholder="443"
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={inputStyle}
            />
          </FormField>
        </div>

        {/* UUID / Password */}
        {(showUUID || showPassword) && (
          <FormField label={showUUID ? t('nodes.uuid') : t('nodes.password')}>
            <input
              type="text"
              value={uuid}
              onChange={(e) => setUuid(e.target.value)}
              placeholder={showUUID ? '00000000-0000-0000-0000-000000000000' : 'password'}
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={inputStyle}
            />
          </FormField>
        )}

        {/* VMess Security */}
        {showSecurity && (
          <FormField label={t('nodes.security')}>
            <select
              value={security}
              onChange={(e) => setSecurity(e.target.value)}
              className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
              style={inputStyle}
            >
              {SECURITY_METHODS.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
          </FormField>
        )}

        {/* VLESS Flow */}
        {showFlow && (
          <FormField label={t('nodes.flow')}>
            <select
              value={flow}
              onChange={(e) => setFlow(e.target.value)}
              className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
              style={inputStyle}
            >
              <option value="">{t('nodes.none')}</option>
              <option value="xtls-rprx-vision">xtls-rprx-vision</option>
            </select>
          </FormField>
        )}

        {/* Transport */}
        {showTransport && (
          <>
            <div className="grid grid-cols-2 gap-3">
              <FormField label={t('nodes.transport')} cols="1/2">
                <select
                  value={network}
                  onChange={(e) => setNetwork(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
                  style={inputStyle}
                >
                  {NETWORKS.map((n) => (
                    <option key={n.value} value={n.value}>{n.label}</option>
                  ))}
                </select>
              </FormField>
              <FormField label={t('nodes.tls')} cols="1/2">
                <select
                  value={tls}
                  onChange={(e) => setTls(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
                  style={inputStyle}
                >
                  {TLS_OPTIONS.map((o) => (
                    <option key={o.value} value={o.value}>{o.label}</option>
                  ))}
                </select>
              </FormField>
            </div>

            {/* Host & Path for ws/h2/grpc */}
            {network !== 'tcp' && (
              <div className="grid grid-cols-2 gap-3">
                <FormField label={t('nodes.host')} cols="1/2">
                  <input
                    type="text"
                    value={host}
                    onChange={(e) => setHost(e.target.value)}
                    placeholder="example.com"
                    className="w-full px-3 py-2 text-sm rounded-lg border"
                    style={inputStyle}
                  />
                </FormField>
                <FormField label={t('nodes.path')} cols="1/2">
                  <input
                    type="text"
                    value={path}
                    onChange={(e) => setPath(e.target.value)}
                    placeholder={network === 'ws' ? '/ws' : network === 'grpc' ? 'grpc-service' : '/path'}
                    className="w-full px-3 py-2 text-sm rounded-lg border"
                    style={inputStyle}
                  />
                </FormField>
              </div>
            )}

            {/* TLS fields */}
            {tls === 'tls' && (
              <div className="grid grid-cols-2 gap-3">
                <FormField label={t('nodes.sni')} cols="1/2">
                  <input
                    type="text"
                    value={sni}
                    onChange={(e) => setSni(e.target.value)}
                    placeholder="example.com"
                    className="w-full px-3 py-2 text-sm rounded-lg border"
                    style={inputStyle}
                  />
                </FormField>
                <FormField label={t('nodes.fingerprint')} cols="1/2">
                  <input
                    type="text"
                    value={fingerprint}
                    onChange={(e) => setFingerprint(e.target.value)}
                    placeholder="chrome / firefox / random"
                    className="w-full px-3 py-2 text-sm rounded-lg border"
                    style={inputStyle}
                  />
                </FormField>
              </div>
            )}

            {/* Reality fields */}
            {showReality && (
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <FormField label={t('nodes.sni_servername')} cols="1/2">
                    <input
                      type="text"
                      value={sni}
                      onChange={(e) => setSni(e.target.value)}
                      placeholder="example.com"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </FormField>
                  <FormField label={t('nodes.fingerprint')} cols="1/2">
                    <input
                      type="text"
                      value={fingerprint}
                      onChange={(e) => setFingerprint(e.target.value)}
                      placeholder="chrome"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </FormField>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <FormField label={t('nodes.public_key')} cols="1/2">
                    <input
                      type="text"
                      value={publicKey}
                      onChange={(e) => setPublicKey(e.target.value)}
                      placeholder="Public key"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </FormField>
                  <FormField label={t('nodes.short_id')} cols="1/2">
                    <input
                      type="text"
                      value={shortId}
                      onChange={(e) => setShortId(e.target.value)}
                      placeholder="Short ID"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </FormField>
                </div>
              </div>
            )}

            {/* Allow Insecure */}
            {tls && (
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={allowInsecure}
                  onChange={(e) => setAllowInsecure(e.target.checked)}
                  className="rounded"
                />
                <span className="text-xs" style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}>
                  {t('nodes.allow_insecure')}
                </span>
              </label>
            )}
          </>
        )}

        {/* 内核设置 */}
        {currentCores.length > 0 && (
          <>
            <div className="grid grid-cols-2 gap-3">
              <FormField label={t('nodes.core_selection')} cols="1/2">
                <button
                  type="button"
                  onClick={() => {
                    if (kernelMode === 'auto') {
                      setKernelMode('manual')
                      if (!manualCore && recommendedCore) setManualCore(recommendedCore)
                    } else {
                      setKernelMode('auto')
                      setManualCore('')
                    }
                  }}
                  className="w-full px-3 py-2 text-sm rounded-lg border text-left transition-colors cursor-pointer"
                  style={{
                    ...inputHeadingStyle,
                    borderColor: kernelMode === 'manual' ? 'var(--color-primary)' : 'var(--color-border)',
                  }}
                >
                  {kernelMode === 'auto' ? t('nodes.auto') : t('nodes.manual')}
                </button>
              </FormField>
              <FormField label={kernelMode === 'auto' ? t('nodes.recommended_core') : t('nodes.manual_core')} cols="1/2">
                {kernelMode === 'auto' ? (
                  <div
                    className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm"
                    style={{
                      backgroundColor: recommendedCore ? 'var(--color-success-dim)' : 'var(--color-warning-dim)',
                      color: recommendedCore ? 'var(--color-success)' : 'var(--color-warning)',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    <Zap size={13} />
                    {recommendedCore ? (
                      <span><strong>{recommendedCore}</strong></span>
                    ) : (
                      <span>{t('nodes.no_core_available')}</span>
                    )}
                  </div>
                ) : (
                  <select
                    value={manualCore}
                    onChange={(e) => setManualCore(e.target.value)}
                    className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
                    style={inputStyle}
                  >
                    {currentCores.map((c: string) => (
                      <option key={c} value={c}>{c}</option>
                    ))}
                  </select>
                )}
              </FormField>
            </div>
            {/* 提示信息：手动模式下显示其他兼容内核 */}
            {kernelMode === 'manual' && currentCores.length > 1 && (
              <p
                className="text-xs"
                style={{ color: 'var(--color-muted-foreground)', opacity: 0.6, fontFamily: 'var(--font-heading)' }}
              >
                {t('nodes.also_supported')} {currentCores.filter(c => c !== manualCore).join(', ')}
              </p>
            )}
          </>
        )}
      </div>

      <FormActions
        onCancel={onClose}
        onSubmit={handleSubmit}
        cancelLabel={t('nodes.cancel')}
        submitLabel={isEditing ? t('nodes.save') : t('nodes.confirm')}
        submitDisabled={!name.trim() || !address.trim() || !port || !selectedGroupUUID}
      />
    </EditFormCard>
  )
}