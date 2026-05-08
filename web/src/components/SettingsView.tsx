import { motion } from 'framer-motion'
import { Globe, Monitor, Sun, Moon, Save } from 'lucide-react'
import { useI18n, useT } from '../lib/i18n'
import { useState } from 'react'

export function SettingsView() {
  const t = useT()
  const { lang, setLang, theme, setTheme } = useI18n()
  const [saved, setSaved] = useState(false)

  const handleSave = () => {
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const themeOptions = [
    { value: 'light' as const, icon: Sun, label: t('settings.theme_light') },
    { value: 'dark' as const, icon: Moon, label: t('settings.theme_dark') },
    { value: 'system' as const, icon: Monitor, label: t('settings.theme_system') },
  ]

  return (
    <div className="max-w-2xl mx-auto">
      <h1
        className="text-xl font-semibold mb-6"
        style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
      >
        {t('settings.title')}
      </h1>

      {/* Appearance Section */}
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }}
        className="rounded-xl border p-6 mb-4"
        style={{
          backgroundColor: 'var(--color-card)',
          borderColor: 'var(--color-border)',
          boxShadow: 'var(--shadow-card)',
        }}
      >
        <h3
          className="text-sm font-semibold mb-5"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('settings.appearance')}
        </h3>

        {/* Language */}
        <div className="flex items-center justify-between mb-5">
          <div className="flex items-center gap-2.5">
            <Globe size={16} style={{ color: 'var(--color-muted-foreground)' }} />
            <span
              className="text-sm"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.language')}
            </span>
          </div>
          <div
            className="flex rounded-lg overflow-hidden border"
            style={{ borderColor: 'var(--color-border)' }}
          >
            <button
              onClick={() => setLang('zh')}
              className="px-3 py-1.5 text-xs font-medium transition-colors cursor-pointer"
              style={{
                backgroundColor: lang === 'zh' ? 'var(--color-primary)' : 'var(--color-muted)',
                color: lang === 'zh' ? 'var(--color-primary-foreground)' : 'var(--color-muted-foreground)',
                fontFamily: 'var(--font-heading)',
              }}
            >
              中文
            </button>
            <button
              onClick={() => setLang('en')}
              className="px-3 py-1.5 text-xs font-medium transition-colors cursor-pointer"
              style={{
                backgroundColor: lang === 'en' ? 'var(--color-primary)' : 'var(--color-muted)',
                color: lang === 'en' ? 'var(--color-primary-foreground)' : 'var(--color-muted-foreground)',
                fontFamily: 'var(--font-heading)',
              }}
            >
              English
            </button>
          </div>
        </div>

        {/* Theme */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <Monitor size={16} style={{ color: 'var(--color-muted-foreground)' }} />
            <span
              className="text-sm"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.theme')}
            </span>
          </div>
          <div
            className="flex rounded-lg overflow-hidden border"
            style={{ borderColor: 'var(--color-border)' }}
          >
            {themeOptions.map((opt) => {
              const Icon = opt.icon
              const isActive = theme === opt.value
              return (
                <button
                  key={opt.value}
                  onClick={() => setTheme(opt.value)}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium transition-colors cursor-pointer"
                  style={{
                    backgroundColor: isActive ? 'var(--color-primary)' : 'var(--color-muted)',
                    color: isActive ? 'var(--color-primary-foreground)' : 'var(--color-muted-foreground)',
                    fontFamily: 'var(--font-heading)',
                  }}
                >
                  <Icon size={12} />
                  {opt.label}
                </button>
              )
            })}
          </div>
        </div>
      </motion.div>

      {/* Network Section */}
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, delay: 0.08, ease: [0.16, 1, 0.3, 1] }}
        className="rounded-xl border p-6 mb-4"
        style={{
          backgroundColor: 'var(--color-card)',
          borderColor: 'var(--color-border)',
          boxShadow: 'var(--shadow-card)',
        }}
      >
        <h3
          className="text-sm font-semibold mb-5"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('settings.network')}
        </h3>

        <div className="space-y-4">
          {/* Listen IP */}
          <div className="flex items-center justify-between">
            <span
              className="text-sm"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.listen_ip')}
            </span>
            <input
              type="text"
              defaultValue="127.0.0.1"
              className="w-40 px-3 py-1.5 text-sm rounded-lg border text-right"
              style={{
                backgroundColor: 'var(--color-overlay)',
                borderColor: 'var(--color-border)',
                color: 'var(--color-foreground)',
                fontFamily: 'var(--font-mono)',
              }}
            />
          </div>

          {/* SOCKS Port */}
          <div className="flex items-center justify-between">
            <span
              className="text-sm"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.socks_port')}
            </span>
            <input
              type="number"
              defaultValue="10808"
              className="w-40 px-3 py-1.5 text-sm rounded-lg border text-right"
              style={{
                backgroundColor: 'var(--color-overlay)',
                borderColor: 'var(--color-border)',
                color: 'var(--color-foreground)',
                fontFamily: 'var(--font-mono)',
              }}
            />
          </div>

          {/* HTTP Port */}
          <div className="flex items-center justify-between">
            <span
              className="text-sm"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.http_port')}
            </span>
            <input
              type="number"
              defaultValue="10809"
              className="w-40 px-3 py-1.5 text-sm rounded-lg border text-right"
              style={{
                backgroundColor: 'var(--color-overlay)',
                borderColor: 'var(--color-border)',
                color: 'var(--color-foreground)',
                fontFamily: 'var(--font-mono)',
              }}
            />
          </div>

          {/* Outbound IP */}
          <div className="flex items-center justify-between">
            <span
              className="text-sm"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.outbound_ip')}
            </span>
            <input
              type="text"
              defaultValue="0.0.0.0"
              className="w-40 px-3 py-1.5 text-sm rounded-lg border text-right"
              style={{
                backgroundColor: 'var(--color-overlay)',
                borderColor: 'var(--color-border)',
                color: 'var(--color-foreground)',
                fontFamily: 'var(--font-mono)',
              }}
            />
          </div>
        </div>
      </motion.div>

      {/* GitHub Mirror Section */}
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, delay: 0.16, ease: [0.16, 1, 0.3, 1] }}
        className="rounded-xl border p-6 mb-6"
        style={{
          backgroundColor: 'var(--color-card)',
          borderColor: 'var(--color-border)',
          boxShadow: 'var(--shadow-card)',
        }}
      >
        <h3
          className="text-sm font-semibold mb-2"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('settings.github_mirror')}
        </h3>
        <p
          className="text-xs mb-4"
          style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('settings.github_mirror_hint')}
        </p>
        <input
          type="text"
          placeholder="https://mirror.example.com"
          className="w-full px-3 py-2 text-sm rounded-lg border"
          style={{
            backgroundColor: 'var(--color-overlay)',
            borderColor: 'var(--color-border)',
            color: 'var(--color-foreground)',
            fontFamily: 'var(--font-mono)',
          }}
        />
      </motion.div>

      {/* Save Button */}
      <motion.button
        onClick={handleSave}
        className="flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium transition-all cursor-pointer"
        style={{
          backgroundColor: saved ? 'var(--color-success)' : 'var(--color-primary)',
          color: 'var(--color-primary-foreground)',
          boxShadow: 'var(--shadow-btn)',
          fontFamily: 'var(--font-heading)',
        }}
        whileHover={{ scale: 1.02 }}
        whileTap={{ scale: 0.98 }}
      >
        <Save size={14} />
        {saved ? t('settings.saved') : t('settings.save')}
      </motion.button>
    </div>
  )
}