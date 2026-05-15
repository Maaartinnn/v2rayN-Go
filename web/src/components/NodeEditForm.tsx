import { useState, useEffect } from 'react'
import { Zap } from 'lucide-react'
import { profileApi, coresApi, groupsApi } from '../lib/api'
import { useT } from '../lib/i18n'
import type { Profile } from '../store'
import {
  PROTOCOLS, NETWORKS, TLS_OPTIONS, SECURITY_METHODS,
  getBestInstalledCore, getSupportedCores,
} from '../lib/coreMap'
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

  // Smart core selection
  const [recommendedCore, setRecommendedCore] = useState<string | null>(null)
  const [installedCores, setInstalledCores] = useState<string[]>([])
  const [kernelMode, setKernelMode] = useState<'auto' | 'manual'>('auto')
  const [manualCore, setManualCore] = useState('')

  // Load groups on mount
  useEffect(() => {
    groupsApi.list().then((res) => {
      setGroups(res.data || [])
    }).catch(() => {})
  }, [])

  // Pre-fill form when editing
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
    } else {
      // 新增模式默认自动
      setKernelMode('auto')
      setManualCore('')
    }
  }, [editData])

  // Load installed cores on mount
  useEffect(() => {
    coresApi.list().then((res) => {
      const cores = (res.data || [])
        .filter((c: { version: string }) => c.version === 'installed')
        .map((c: { name: string }) => c.name)
      setInstalledCores(cores)
    }).catch(() => {})
  }, [])

  // Update recommended core when protocol changes
  useEffect(() => {
    const best = getBestInstalledCore(protocol, installedCores)
    setRecommendedCore(best)
  }, [protocol, installedCores])

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
        const res = await profileApi.update(editData.ID, payload)
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

  const supportedCores = getSupportedCores(protocol)

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
          <FormField label="Address" cols="2/3">
            <input
              type="text"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="example.com"
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={inputStyle}
            />
          </FormField>
          <FormField label="Port" cols="1/3">
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
          <FormField label={showUUID ? 'UUID / ID' : 'Password'}>
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
          <FormField label="Security">
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
          <FormField label="Flow">
            <select
              value={flow}
              onChange={(e) => setFlow(e.target.value)}
              className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
              style={inputStyle}
            >
              <option value="">None</option>
              <option value="xtls-rprx-vision">xtls-rprx-vision</option>
            </select>
          </FormField>
        )}

        {/* Transport */}
        {showTransport && (
          <>
            <div className="grid grid-cols-2 gap-3">
              <FormField label="Transport" cols="1/2">
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
              <FormField label="TLS" cols="1/2">
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
                <FormField label="Host" cols="1/2">
                  <input
                    type="text"
                    value={host}
                    onChange={(e) => setHost(e.target.value)}
                    placeholder="example.com"
                    className="w-full px-3 py-2 text-sm rounded-lg border"
                    style={inputStyle}
                  />
                </FormField>
                <FormField label="Path" cols="1/2">
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
                <FormField label="SNI" cols="1/2">
                  <input
                    type="text"
                    value={sni}
                    onChange={(e) => setSni(e.target.value)}
                    placeholder="example.com"
                    className="w-full px-3 py-2 text-sm rounded-lg border"
                    style={inputStyle}
                  />
                </FormField>
                <FormField label="Fingerprint" cols="1/2">
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
                  <FormField label="SNI (ServerName)" cols="1/2">
                    <input
                      type="text"
                      value={sni}
                      onChange={(e) => setSni(e.target.value)}
                      placeholder="example.com"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </FormField>
                  <FormField label="Fingerprint" cols="1/2">
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
                  <FormField label="Public Key" cols="1/2">
                    <input
                      type="text"
                      value={publicKey}
                      onChange={(e) => setPublicKey(e.target.value)}
                      placeholder="Public key"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </FormField>
                  <FormField label="Short ID" cols="1/2">
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
                  Allow Insecure (skip certificate verification)
                </span>
              </label>
            )}
          </>
        )}

        {/* 内核设置 */}
        {supportedCores.length > 0 && (
          <>
            <div className="grid grid-cols-2 gap-3">
              <FormField label="内核选择" cols="1/2">
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
                  {kernelMode === 'auto' ? '自动' : '手动'}
                </button>
              </FormField>
              <FormField label={kernelMode === 'auto' ? '推荐内核' : '手动内核'} cols="1/2">
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
                      <span>无可用内核</span>
                    )}
                  </div>
                ) : (
                  <select
                    value={manualCore}
                    onChange={(e) => setManualCore(e.target.value)}
                    className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
                    style={inputStyle}
                  >
                    {supportedCores.map((c) => {
                      const isInstalled = installedCores.includes(c)
                      return (
                        <option key={c} value={c}>
                          {c}{isInstalled ? ' ✓' : ' (未安装)'}
                        </option>
                      )
                    })}
                  </select>
                )}
              </FormField>
            </div>
            {/* 提示信息 */}
            {kernelMode === 'manual' && supportedCores.length > 1 && (
              <p
                className="text-xs"
                style={{ color: 'var(--color-muted-foreground)', opacity: 0.6, fontFamily: 'var(--font-heading)' }}
              >
                此协议也支持: {supportedCores.filter(c => c !== manualCore).join(', ')}
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