import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Check, X, Zap } from 'lucide-react'
import { profileApi, coresApi } from '../lib/api'
import { useT } from '../lib/i18n'
import {
  PROTOCOLS, NETWORKS, TLS_OPTIONS, SECURITY_METHODS,
  getBestInstalledCore, getSupportedCores,
} from '../lib/coreMap'

interface NodeEditFormProps {
  onClose: () => void
  onSaved: () => void
  groupId?: number
}

export function NodeEditForm({ onClose, onSaved, groupId }: NodeEditFormProps) {
  const t = useT()

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

    try {
      await profileApi.create({
        group_id: groupId || 0,
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
        is_active: false,
        sort_order: 0,
      })
      onSaved()
      onClose()
    } catch (err) {
      console.error('Failed to create node:', err)
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

  const inputStyle = {
    backgroundColor: 'var(--color-overlay)',
    borderColor: 'var(--color-border)',
    color: 'var(--color-foreground)',
    fontFamily: 'var(--font-mono)',
  }

  const labelStyle = {
    color: 'var(--color-muted-foreground)',
    fontFamily: 'var(--font-heading)' as const,
  }

  return (
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
        <div className="space-y-4">
          {/* Row 1: Protocol + Name */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs font-medium block mb-1" style={labelStyle}>
                {t('routing.type')}
              </label>
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
            </div>
            <div>
              <label className="text-xs font-medium block mb-1" style={labelStyle}>
                {t('strategy.name')}
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="My Node"
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </div>
          </div>

          {/* Row 2: Address + Port */}
          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <label className="text-xs font-medium block mb-1" style={labelStyle}>
                Address
              </label>
              <input
                type="text"
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                placeholder="example.com"
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </div>
            <div>
              <label className="text-xs font-medium block mb-1" style={labelStyle}>
                Port
              </label>
              <input
                type="text"
                value={port}
                onChange={(e) => setPort(e.target.value)}
                placeholder="443"
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </div>
          </div>

          {/* UUID / Password */}
          {(showUUID || showPassword) && (
            <div>
              <label className="text-xs font-medium block mb-1" style={labelStyle}>
                {showUUID ? 'UUID / ID' : 'Password'}
              </label>
              <input
                type="text"
                value={uuid}
                onChange={(e) => setUuid(e.target.value)}
                placeholder={showUUID ? '00000000-0000-0000-0000-000000000000' : 'password'}
                className="w-full px-3 py-2 text-sm rounded-lg border"
                style={inputStyle}
              />
            </div>
          )}

          {/* VMess Security */}
          {showSecurity && (
            <div>
              <label className="text-xs font-medium block mb-1" style={labelStyle}>
                Security
              </label>
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
            </div>
          )}

          {/* VLESS Flow */}
          {showFlow && (
            <div>
              <label className="text-xs font-medium block mb-1" style={labelStyle}>
                Flow
              </label>
              <select
                value={flow}
                onChange={(e) => setFlow(e.target.value)}
                className="w-full px-3 py-2 text-sm rounded-lg border cursor-pointer"
                style={inputStyle}
              >
                <option value="">None</option>
                <option value="xtls-rprx-vision">xtls-rprx-vision</option>
              </select>
            </div>
          )}

          {/* Transport */}
          {showTransport && (
            <>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium block mb-1" style={labelStyle}>
                    Transport
                  </label>
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
                </div>
                <div>
                  <label className="text-xs font-medium block mb-1" style={labelStyle}>
                    TLS
                  </label>
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
                </div>
              </div>

              {/* Host & Path for ws/h2/grpc */}
              {network !== 'tcp' && (
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-xs font-medium block mb-1" style={labelStyle}>
                      Host
                    </label>
                    <input
                      type="text"
                      value={host}
                      onChange={(e) => setHost(e.target.value)}
                      placeholder="example.com"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </div>
                  <div>
                    <label className="text-xs font-medium block mb-1" style={labelStyle}>
                      Path
                    </label>
                    <input
                      type="text"
                      value={path}
                      onChange={(e) => setPath(e.target.value)}
                      placeholder={network === 'ws' ? '/ws' : network === 'grpc' ? 'grpc-service' : '/path'}
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </div>
                </div>
              )}

              {/* TLS fields */}
              {tls === 'tls' && (
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-xs font-medium block mb-1" style={labelStyle}>
                      SNI
                    </label>
                    <input
                      type="text"
                      value={sni}
                      onChange={(e) => setSni(e.target.value)}
                      placeholder="example.com"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </div>
                  <div>
                    <label className="text-xs font-medium block mb-1" style={labelStyle}>
                      Fingerprint
                    </label>
                    <input
                      type="text"
                      value={fingerprint}
                      onChange={(e) => setFingerprint(e.target.value)}
                      placeholder="chrome / firefox / random"
                      className="w-full px-3 py-2 text-sm rounded-lg border"
                      style={inputStyle}
                    />
                  </div>
                </div>
              )}

              {/* Reality fields */}
              {showReality && (
                <div className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="text-xs font-medium block mb-1" style={labelStyle}>
                        SNI (ServerName)
                      </label>
                      <input
                        type="text"
                        value={sni}
                        onChange={(e) => setSni(e.target.value)}
                        placeholder="example.com"
                        className="w-full px-3 py-2 text-sm rounded-lg border"
                        style={inputStyle}
                      />
                    </div>
                    <div>
                      <label className="text-xs font-medium block mb-1" style={labelStyle}>
                        Fingerprint
                      </label>
                      <input
                        type="text"
                        value={fingerprint}
                        onChange={(e) => setFingerprint(e.target.value)}
                        placeholder="chrome"
                        className="w-full px-3 py-2 text-sm rounded-lg border"
                        style={inputStyle}
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="text-xs font-medium block mb-1" style={labelStyle}>
                        Public Key
                      </label>
                      <input
                        type="text"
                        value={publicKey}
                        onChange={(e) => setPublicKey(e.target.value)}
                        placeholder="Public key"
                        className="w-full px-3 py-2 text-sm rounded-lg border"
                        style={inputStyle}
                      />
                    </div>
                    <div>
                      <label className="text-xs font-medium block mb-1" style={labelStyle}>
                        Short ID
                      </label>
                      <input
                        type="text"
                        value={shortId}
                        onChange={(e) => setShortId(e.target.value)}
                        placeholder="Short ID"
                        className="w-full px-3 py-2 text-sm rounded-lg border"
                        style={inputStyle}
                      />
                    </div>
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
                  <span className="text-xs" style={labelStyle}>
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

        {/* Action buttons */}
        <div className="flex justify-end gap-2 mt-5">
          <button
            onClick={onClose}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
            style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            <X size={13} />
            {t('nodes.cancel')}
          </button>
          <button
            onClick={handleSubmit}
            disabled={!name.trim() || !address.trim() || !port}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer disabled:opacity-50"
            style={{
              backgroundColor: 'var(--color-primary)',
              color: 'var(--color-primary-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            <Check size={13} />
            {t('nodes.confirm')}
          </button>
        </div>
      </div>
    </motion.div>
  )
}