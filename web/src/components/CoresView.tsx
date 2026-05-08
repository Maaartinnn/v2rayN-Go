import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { Download, Upload, Loader2, HardDrive } from 'lucide-react'
import { coresApi } from '../lib/api'
import { useT } from '../lib/i18n'

interface CoreInfo {
  name: string
  display_name: string
  repo: string
  version: string
  latest_version: string
  binary_name: string
}

export function CoresView() {
  const [cores, setCores] = useState<CoreInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState<string>('')
  const t = useT()

  useEffect(() => {
    loadCores()
  }, [])

  const loadCores = async () => {
    setLoading(true)
    try {
      const res = await coresApi.list()
      setCores(res.data || [])
    } catch (err) {
      console.error('Failed to load cores:', err)
    }
    setLoading(false)
  }

  const handleDownload = async (coreName: string) => {
    setDownloading(coreName)
    try {
      await coresApi.download(coreName)
      // Poll for completion
      setTimeout(async () => {
        await loadCores()
        setDownloading('')
      }, 5000)
    } catch (err) {
      console.error('Download failed:', err)
      setDownloading('')
    }
  }

  const handleUpload = (coreName: string) => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.zip,.tar.gz,.tgz,.exe,*'
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0]
      if (!file) return
      const formData = new FormData()
      formData.append('core_name', coreName)
      formData.append('binary', file)
      try {
        await coresApi.upload(formData)
        await loadCores()
      } catch (err) {
        console.error('Upload failed:', err)
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
          {cores.map((core, index) => (
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
                {/* Core info */}
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
                      {core.version ? (
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
                      ) : (
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
                      )}
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

                {/* Actions */}
                <div className="flex items-center gap-2">
                  {/* Upload button */}
                  <motion.button
                    onClick={() => handleUpload(core.name)}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors cursor-pointer"
                    style={{
                      backgroundColor: 'var(--color-muted)',
                      borderColor: 'var(--color-border)',
                      color: 'var(--color-muted-foreground)',
                      fontFamily: 'var(--font-heading)',
                    }}
                    whileTap={{ scale: 0.95 }}
                    title={t('cores.download_upload')}
                  >
                    <Upload size={13} />
                    {t('cores.download_upload')}
                  </motion.button>

                  {/* Download button */}
                  <motion.button
                    onClick={() => handleDownload(core.name)}
                    disabled={downloading === core.name}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer"
                    style={{
                      backgroundColor: downloading === core.name ? 'var(--color-muted)' : 'var(--color-primary)',
                      color: downloading === core.name ? 'var(--color-muted-foreground)' : 'var(--color-primary-foreground)',
                      boxShadow: downloading === core.name ? 'none' : 'var(--shadow-btn)',
                      fontFamily: 'var(--font-heading)',
                    }}
                    whileTap={{ scale: 0.95 }}
                  >
                    {downloading === core.name ? (
                      <Loader2 size={13} className="animate-spin" />
                    ) : (
                      <Download size={13} />
                    )}
                    {downloading === core.name ? t('cores.downloading') : t('cores.download_github')}
                  </motion.button>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      )}
    </div>
  )
}