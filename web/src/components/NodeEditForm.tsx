import { useState, useEffect } from 'react'
import { Zap } from 'lucide-react'
import { profileApi, coresApi } from '../lib/api'
import { useT } from '../lib/i18n'
import type { Profile } from '../store'
import {
  PROTOCOLS, NETWORKS, TLS_OPTIONS, SECURITY_METHODS,
  getBestInstalledCore, getSupportedCores,
} from '../lib/coreMap'
import { EditFormCard } from './ui/EditFormCard'
import { FormField } from './ui/FormField'
import { FormActions } from './ui/FormActions'
import { inputStyle } from './ui/formStyles'

interface NodeEditFormProps {
  onClose: () => void
  onSaved: (updatedProfile?: Profile) => void
  groupId?: number
  editData?: Profile
}

export function NodeEditForm({ onClose, onSaved, groupId, editData }: NodeEditFormProps) {
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

  // Smart core selection
  const [recommendedCore, setRecommendedCore] = useState<string | null>(null)
  const [installedCores, setInstalledCores] = useState<string[]>([])

  // Pre-fill form when editing
  useEffect(() => {
    if (editData) {
      setName(editData.name || '')
      setProtocol(editData.protocol || 'vmess')
      setAddress(editData.address || '')
      setPort(String(editData.port || 443))
      setUuid(editData.uuid || '')
      setSecurity(editData.security || 'auto')
      setNetwork(editData.network || 'tcp')
      setHost(editData.host || '')
      setPath(editData.path || '')
      setTls(editData.tls || '')
      setSni(editData.sni || '')
      setFingerprint(editData.fingerprint || '')
      setAllowInsecure(editData.allow_insecure || false)
      setPublicKey(editData.public_key || '')
      setShortId(editData.short_id || '')
      setFlow(editData.flow || '')
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
      group_id: groupId || editData?.group_id || 0,
      name: name.trim(),
      address: address.trim(),
      port: parseInt(port) || 443,
      protocol,
      uuid: uuid.trim(),
      security,
      network,
      host: host.trim(),
      path: path.trim(),
      tls,
      sni: sni.trim(),
      fingerprint: fingerprint.trim(),
      allow_insecure: allowInsecure,
      public_key: publicKey.trim(),
      short_id: shortId.trim(),
      flow: flow.trim(),
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
        {/* Row 1: Protocol + Name */}
        <div className="grid grid-cols-2 gap-3">
          <FormField label={t('routing.type')} cols="1/2">
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
          <FormField label={t('strategy.name')} cols="1/2">
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Node"
              className="w-full px-3 py-2 text-sm rounded-lg border"
              style={inputStyle}
            />
          </FormField>
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

        {/* Smart Core Recommendation */}
        {supportedCores.length > 0 && (
          <div
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs"
            style={{
              backgroundColor: recommendedCore ? 'var(--color-success-dim)' : 'var(--color-warning-dim)',
              color: recommendedCore ? 'var(--color-success)' : 'var(--color-warning)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            <Zap size={13} />
            {recommendedCore ? (
              <span>
                推荐内核: <strong>{recommendedCore}</strong>
                {supportedCores.length > 1 && (
                  <span style={{ opacity: 0.7 }}>
                    {' '}(也支持: {supportedCores.filter(c => c !== recommendedCore).join(', ')})
                  </span>
                )}
              </span>
            ) : (
              <span>
                无已安装内核支持此协议 (需要: {supportedCores.join(', ')})
              </span>
            )}
          </div>
        )}
      </div>

      <FormActions
        onCancel={onClose}
        onSubmit={handleSubmit}
        cancelLabel={t('nodes.cancel')}
        submitLabel={isEditing ? t('nodes.save') : t('nodes.confirm')}
        submitDisabled={!name.trim() || !address.trim() || !port}
      />
    </EditFormCard>
  )
}