import { useEffect, useState, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Download, Upload, Loader2, HardDrive, Link, X, ChevronDown, ExternalLink, GitBranch } from 'lucide-react'
import { coresApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { useStore } from '../store'

interface CoreInfo {
  name: string
  display_name: string
  repo: string
  version: string
  latest_version: string
  binary_name: string
  sub_dir: string
}

export function CoresView() {
  const [cores, setCores] = useState<CoreInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState<Record<string, boolean>>({})
  const [customUrlCore, setCustomUrlCore] = useState<string>('')
  const [customUrl, setCustomUrl] = useState('')
  const [menuOpen, setMenuOpen] = useState<string>('')
  const menuRef = useRef<HTMLDivElement>(null)
  const t = useT()
  const { downloadProgress, addToast, coreVersions } = useStore()

  useEffect(() => {
    loadCores()
    // Trigger async version detection via API
    coresApi.detectVersions().catch(() => {})
  }, [])

  // Close menu on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen('')
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const loadCores = async () => {
    setLoading(true)
    try {
      // Fast: local info only, no network
      const res = await coresApi.list()
      setCores(res.data || [])
    } catch (err) {
      console.error('Failed to load cores:', err)
    }
    setLoading(false)

    // Async: check latest versions from GitHub (non-blocking)
    try {
      const updateRes = await coresApi.checkUpdates()
      const latestVersions = updateRes.data?.latest_versions || {}
      setCores(prev => prev.map(core => ({
        ...core,
        latest_version: latestVersions[core.name] || core.latest_version,
      })))
    } catch (err) {
      console.error('Failed to check updates:', err)
    }
  }

  const handleDownload = async (coreName: string) => {
    setDownloading(prev => ({ ...prev, [coreName]: true }))
    setMenuOpen('')
    try {
      await coresApi.download(coreName)
    } catch (err: any) {
      console.error('Download failed:', err)
      const msg = err?.response?.data?.error || err?.message || 'Unknown error'
      addToast(t('cores.download_failed', { name: coreName, error: msg }), 'error')
      setDownloading(prev => ({ ...prev, [coreName]: false }))
    }
  }

  const handleCustomUrlDownload = async () => {
    if (!customUrlCore || !customUrl.trim()) return
    setDownloading(prev => ({ ...prev, [customUrlCore]: true }))
    setMenuOpen('')
    try {
      await coresApi.downloadUrl(customUrlCore, customUrl.trim())
      setCustomUrlCore('')
      setCustomUrl('')
    } catch (err: any) {
      console.error('Custom URL download failed:', err)
      const msg = err?.response?.data?.error || err?.message || 'Unknown error'
      addToast(t('cores.download_failed', { name: customUrlCore, error: msg }), 'error')
      setDownloading(prev => ({ ...prev, [customUrlCore]: false }))
    }
  }

  const handleUpload = (coreName: string, acceptArchive: boolean) => {
    setMenuOpen('')
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = acceptArchive ? '.zip,.tar.gz,.tgz' : '.exe,*'
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0]
      if (!file) return
      const formData = new FormData()
      formData.append('core_name', coreName)
      formData.append('binary', file)
      try {
        await coresApi.upload(formData)
        addToast(t('cores.upload_success', { name: coreName }), 'success')
        await loadCores()
      } catch (err: any) {
        console.error('Upload failed:', err)
        const msg = err?.response?.data?.error || err?.message || 'Unknown error'
        addToast(t('cores.upload_failed', { name: coreName, error: msg }), 'error')
      }
    }
    input.click()
  }

  const getCoreIcon = (name: string) => {
    const colors: Record<string, string> = {
      xray: '#6A9BCC',
      'sing-box': '#6B8F47',
      mihomo: '#C9943A',
    }
    return colors[name] || 'var(--color-muted-foreground)'
  }

  const getGithubUrl = (repo: string) => `https://github.com/${repo}`

  // Normalize version: strip 'v' prefix for comparison
  const normalizeVersion = (v: string) => v ? v.replace(/^v/i, '') : ''

  const hasUpdate = (core: CoreInfo) => {
    const ver = normalizeVersion(coreVersions[core.name] || core.version)
    const latest = normalizeVersion(core.latest_version)
    return ver && ver !== 'installed' && latest && ver !== latest
  }

  return (
    <div className="max-w-3xl mx-auto">
      <h1
        className="text-xl font-semibold mb-6"
        style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
      >
        {t('cores.title')}
      </h1>

      {loading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 size={24} className="animate-spin" style={{ color: 'var(--color-muted-foreground)' }} />
        </div>
      ) : (
        <div className="space-y-3">
          {cores.map((core, index) => {
            const progress = downloadProgress[core.name]
            const isDownloading = downloading[core.name] || !!progress

            return (
              <motion.div
                key={core.name}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3, delay: index * 0.08, ease: [0.16, 1, 0.3, 1] }}
                className="rounded-xl border p-5"
                style={{
                  backgroundColor: 'var(--color-card)',
                  borderColor: 'var(--color-border)',
                  boxShadow: 'var(--shadow-card)',
                }}
              >
                <div className="flex items-center justify-between">
                  {/* Left: Core info */}
                  <div className="flex items-center gap-3">
                    <div
                      className="w-10 h-10 rounded-lg flex items-center justify-center"
                      style={{ backgroundColor: `${getCoreIcon(core.name)}15` }}
                    >
                      <HardDrive size={18} style={{ color: getCoreIcon(core.name) }} />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span
                          className="text-sm font-semibold"
                          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                        >
                          {core.display_name}
                        </span>
                        {(() => {
                          const ver = coreVersions[core.name] || core.version
                          const isInstalled = !!ver
                          const hasKnownVersion = ver && ver !== 'installed'
                          const update = hasUpdate(core)

                          return (
                            <>
                              {/* 徽标1：状态 */}
                              {!isInstalled ? (
                                <span
                                  className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                                  style={{
                                    backgroundColor: 'var(--color-muted)',
                                    color: 'var(--color-text-muted)',
                                    fontFamily: 'var(--font-heading)',
                                  }}
                                >
                                  {t('cores.not_installed')}
                                </span>
                              ) : update ? (
                                <span
                                  className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                                  style={{
                                    backgroundColor: 'var(--color-warning-dim, #FEF3C7)',
                                    color: 'var(--color-warning, #D97706)',
                                    fontFamily: 'var(--font-heading)',
                                  }}
                                >
                                  {t('cores.has_update')}
                                </span>
                              ) : (
                                <span
                                  className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                                  style={{
                                    backgroundColor: 'var(--color-success-dim)',
                                    color: 'var(--color-success)',
                                    fontFamily: 'var(--font-heading)',
                                  }}
                                >
                                  {t('cores.installed')}
                                </span>
                              )}

                              {/* 徽标2：版本号（未安装时隐藏） */}
                              {isInstalled && (
                                <span
                                  className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                                  style={{
                                    backgroundColor: hasKnownVersion ? 'var(--color-success-dim)' : 'var(--color-warning-dim, #FEF3C7)',
                                    color: hasKnownVersion ? 'var(--color-success)' : 'var(--color-warning, #D97706)',
                                    fontFamily: 'var(--font-mono)',
                                  }}
                                >
                                  {hasKnownVersion ? ver : t('cores.unknown_version')}
                                </span>
                              )}
                            </>
                          )
                        })()}
                      </div>
                      <p
                        className="text-xs mt-0.5"
                        style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                      >
                        {core.repo}
                        {core.latest_version && (
                          <span style={{ fontFamily: 'var(--font-heading)' }}> · {t('cores.latest')}: {core.latest_version}</span>
                        )}
                      </p>
                    </div>
                  </div>

                  {/* Right: Actions */}
                  <div className="flex items-center gap-2">
                    {/* Download progress indicator */}
                    {progress && (
                      <div className="flex items-center gap-2 mr-2">
                        <div
                          className="w-24 h-1.5 rounded-full overflow-hidden"
                          style={{ backgroundColor: 'var(--color-muted)' }}
                        >
                          <motion.div
                            className="h-full rounded-full"
                            style={{ backgroundColor: 'var(--color-primary)' }}
                            initial={{ width: 0 }}
                            animate={{ width: `${progress.percentage}%` }}
                            transition={{ duration: 0.3 }}
                          />
                        </div>
                        <span
                          className="text-[10px] font-medium tabular-nums"
                          style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-mono)' }}
                        >
                          {progress.percentage}%
                        </span>
                      </div>
                    )}

                    {/* GitHub repo link */}
                    <motion.a
                      href={getGithubUrl(core.repo)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors"
                      style={{
                        backgroundColor: 'var(--color-muted)',
                        borderColor: 'var(--color-border)',
                        color: 'var(--color-muted-foreground)',
                        fontFamily: 'var(--font-heading)',
                      }}
                      whileTap={{ scale: 0.95 }}
                      title={core.repo}
                    >
                      <ExternalLink size={13} />
                    </motion.a>

                    {/* Unified download button with submenu */}
                    <div className="relative" ref={menuOpen === core.name ? menuRef : undefined}>
                      <motion.button
                        onClick={() => {
                          if (isDownloading) return
                          setMenuOpen(menuOpen === core.name ? '' : core.name)
                        }}
                        disabled={isDownloading}
                        className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium cursor-pointer ${isDownloading ? '' : 'btn-primary'}`}
                        style={{
                          fontFamily: 'var(--font-heading)',
                        }}
                        whileTap={{ scale: 0.95 }}
                      >
                        {isDownloading ? (
                          <Loader2 size={15} className="animate-spin" />
                        ) : (
                          <Download size={15} />
                        )}
                        {isDownloading ? t('cores.downloading') : t('cores.download')}
                        {!isDownloading && <ChevronDown size={12} />}
                      </motion.button>

                      {/* Submenu */}
                      <AnimatePresence>
                        {menuOpen === core.name && (
                          <motion.div
                            initial={{ opacity: 0, y: -4, scale: 0.95 }}
                            animate={{ opacity: 1, y: 0, scale: 1 }}
                            exit={{ opacity: 0, y: -4, scale: 0.95 }}
                            transition={{ duration: 0.15 }}
                            className="absolute right-0 top-full mt-1 z-50 w-52 rounded-lg border py-1"
                            style={{
                              backgroundColor: 'var(--color-card)',
                              borderColor: 'var(--color-border)',
                              boxShadow: '0 8px 24px rgba(0,0,0,0.12)',
                            }}
                          >
                            <button
                              onClick={() => {
                                setCustomUrlCore(customUrlCore === core.name ? '' : core.name)
                                setCustomUrl('')
                                setMenuOpen('')
                              }}
                              className="w-full flex items-center gap-2 px-3 py-2 text-xs hover:opacity-80 transition-opacity cursor-pointer"
                              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                            >
                              <Link size={13} />
                              {t('cores.download_custom')}
                            </button>
                            <button
                              onClick={() => handleUpload(core.name, false)}
                              className="w-full flex items-center gap-2 px-3 py-2 text-xs hover:opacity-80 transition-opacity cursor-pointer"
                              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                            >
                              <Upload size={13} />
                              {t('cores.upload_binary')}
                            </button>
                            <button
                              onClick={() => handleUpload(core.name, true)}
                              className="w-full flex items-center gap-2 px-3 py-2 text-xs hover:opacity-80 transition-opacity cursor-pointer"
                              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                            >
                              <Upload size={13} />
                              {t('cores.upload_archive')}
                            </button>
                            <div className="mx-3 my-1 border-t" style={{ borderColor: 'var(--color-border-subtle)' }} />
                            <button
                              onClick={() => handleDownload(core.name)}
                              className="w-full flex items-center gap-2 px-3 py-2 text-xs hover:opacity-80 transition-opacity cursor-pointer"
                              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
                            >
                              <GitBranch size={13} />
                              {t('cores.download_github')}
                            </button>
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  </div>
                </div>

                {/* Custom URL input */}
                <AnimatePresence>
                  {customUrlCore === core.name && (
                    <motion.div
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: 'auto' }}
                      exit={{ opacity: 0, height: 0 }}
                      className="overflow-hidden"
                    >
                      <div className="flex items-center gap-2 mt-4 pt-4 border-t" style={{ borderColor: 'var(--color-border-subtle)' }}>
                        <input
                          type="text"
                          value={customUrl}
                          onChange={(e) => setCustomUrl(e.target.value)}
                          placeholder="https://mirror.example.com/releases/download/v1.0.0/xray-windows-64.zip"
                          className="flex-1 px-3 py-2 text-xs rounded-lg border"
                          style={{
                            backgroundColor: 'var(--color-overlay)',
                            borderColor: 'var(--color-border)',
                            color: 'var(--color-foreground)',
                            fontFamily: 'var(--font-mono)',
                          }}
                        />
                        <motion.button
                          onClick={handleCustomUrlDownload}
                          disabled={!customUrl.trim() || isDownloading}
                          className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded-lg transition-colors cursor-pointer disabled:opacity-50"
                          style={{
                            backgroundColor: 'var(--color-primary)',
                            color: 'var(--color-primary-foreground)',
                            fontFamily: 'var(--font-heading)',
                          }}
                          whileTap={{ scale: 0.95 }}
                        >
                          {isDownloading ? (
                            <Loader2 size={13} className="animate-spin" />
                          ) : (
                            <Download size={13} />
                          )}
                          {t('cores.download_custom')}
                        </motion.button>
                        <button
                          onClick={() => { setCustomUrlCore(''); setCustomUrl('') }}
                          className="p-2 rounded-md transition-colors cursor-pointer"
                          style={{ color: 'var(--color-text-muted)' }}
                        >
                          <X size={14} />
                        </button>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </motion.div>
            )
          })}
        </div>
      )}
    </div>
  )
}