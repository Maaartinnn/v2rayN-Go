// SettingsView.tsx — 设置页面
//
// 设计要点：
//   - 移除全局保存按钮，改为失焦自动保存（Blur → API）
//   - 输入框：onBlur 触发保存，onKeyDown Enter 触发 blur
//   - Toggle/Select：onChange 时立即触发保存（无失焦概念）
//   - 后端返回 400 时，回滚脏输入并重新加载

import { motion } from 'framer-motion'
import { Globe, Monitor, Sun, Moon } from 'lucide-react'
import { AVAILABLE_LANGUAGES, useI18n, useT } from '../lib/i18n'
import { useState, useEffect } from 'react'
import { settingsApi } from '../lib/api'
import { useStore } from '../store'

export function SettingsView() {
  const t = useT()
  const { lang, setLang, theme, setTheme } = useI18n()
  const [listenIP, setListenIP] = useState('127.0.0.1')
  const [socksPort, setSocksPort] = useState('10808')
  const [httpPort, setHttpPort] = useState('10809')
  const [outboundIP, setOutboundIP] = useState('0.0.0.0')
  const [githubMirror, setGithubMirror] = useState('')
  const [coreConfigDebug, setCoreConfigDebug] = useState(false)
  const [forceHttps, setForceHttps] = useState(false)
  const [basePath, setBasePath] = useState('')
  const [jwtExpireHours, setJwtExpireHours] = useState('24')
  const addToast = useStore(s => s.addToast)

  useEffect(() => {
    loadSettings()
  }, [])

  // loadSettings 从后端加载当前配置，用于初始化和错误回滚
  const loadSettings = async () => {
    try {
      const res = await settingsApi.get()
      const data = res.data
      if (data.listen_ip) setListenIP(data.listen_ip)
      if (data.socks_port) setSocksPort(String(data.socks_port))
      if (data.http_port) setHttpPort(String(data.http_port))
      if (data.outbound_ip) setOutboundIP(data.outbound_ip)
      if (data.github_mirror !== undefined) setGithubMirror(data.github_mirror || '')
      setCoreConfigDebug(!!data.core_config_debug)
      // 服务器设置（从 app_settings 表读取）
      if (data.force_https !== undefined) setForceHttps(data.force_https === 'true')
      // custom_base_path 存储格式：纯路径名（无斜杠），空字符串表示无前缀
      if (data.custom_base_path !== undefined) {
        setBasePath(data.custom_base_path || '')
      }
      // JWT 过期时间（小时），默认 24
      if (data.jwt_expire_hours !== undefined) {
        setJwtExpireHours(data.jwt_expire_hours || '24')
      }
    } catch (err) {
      console.error('Failed to load settings:', err)
    }
  }

  // handleBlur 通用失焦保存函数
  //
  // 输入框 onBlur 时调用，组装单字段 JSON 发送到后端。
  // 如果后端校验失败（HTTP 400），回滚脏输入：重新从后端加载配置覆盖前端状态。
  //
  // 参数：
  //   - field: 配置字段名（与 JSON key 一致）
  //   - value: 字段值
  const handleBlur = async (field: string, value: string | number | boolean) => {
    try {
      await settingsApi.save({ [field]: value })
    } catch (err: any) {
      const msg = err?.response?.data?.error || 'Save failed'
      console.error(`Settings save failed (${field}):`, msg)
      // 校验失败时回滚：重新从后端加载配置，覆盖用户的非法输入
      loadSettings()
    }
  }

  // handleKeyDown 回车键触发失焦（blur），间接触发保存
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.currentTarget.blur()
    }
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
          <select
            value={lang}
            onChange={(e) => setLang(e.target.value as any)}
            className="px-3 py-1.5 text-xs font-medium rounded-lg border cursor-pointer"
            style={{
              backgroundColor: 'var(--color-muted)',
              borderColor: 'var(--color-border)',
              color: 'var(--color-foreground)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            {AVAILABLE_LANGUAGES.map((l) => (
              <option key={l.code} value={l.code}>
                {l.label}
              </option>
            ))}
          </select>
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
              value={listenIP}
              onChange={(e) => setListenIP(e.target.value)}
              onBlur={() => handleBlur('listen_ip', listenIP)}
              onKeyDown={handleKeyDown}
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
              value={socksPort}
              onChange={(e) => setSocksPort(e.target.value)}
              onBlur={() => handleBlur('socks_port', parseInt(socksPort) || 0)}
              onKeyDown={handleKeyDown}
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
              value={httpPort}
              onChange={(e) => setHttpPort(e.target.value)}
              onBlur={() => handleBlur('http_port', parseInt(httpPort) || 0)}
              onKeyDown={handleKeyDown}
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
              value={outboundIP}
              onChange={(e) => setOutboundIP(e.target.value)}
              onBlur={() => handleBlur('outbound_ip', outboundIP)}
              onKeyDown={handleKeyDown}
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
          value={githubMirror}
          onChange={(e) => setGithubMirror(e.target.value)}
          onBlur={() => handleBlur('github_mirror', githubMirror)}
          onKeyDown={handleKeyDown}
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

      {/* Advanced Section */}
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, delay: 0.24, ease: [0.16, 1, 0.3, 1] }}
        className="rounded-xl border p-6 mb-6"
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
          {t('settings.advanced') ?? 'Advanced'}
        </h3>

        {/* Core Config Debug Toggle */}
        <div className="flex items-center justify-between">
          <div className="flex-1 mr-4">
            <span
              className="text-sm block"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.core_config_debug')}
            </span>
            <span
              className="text-xs block mt-1"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.core_config_debug_hint')}
            </span>
          </div>
          <button
            onClick={() => {
              const newVal = !coreConfigDebug
              setCoreConfigDebug(newVal)
              // Toggle 无失焦概念，onChange 时立即触发保存
              handleBlur('core_config_debug', newVal)
            }}
            className="relative inline-flex h-6 w-11 items-center rounded-full transition-colors cursor-pointer shrink-0"
            style={{
              backgroundColor: coreConfigDebug ? 'var(--color-primary)' : 'var(--color-muted)',
            }}
          >
            <span
              className="inline-block h-4 w-4 transform rounded-full transition-transform"
              style={{
                backgroundColor: coreConfigDebug ? 'var(--color-primary-foreground)' : 'var(--color-muted-foreground)',
                transform: coreConfigDebug ? 'translateX(24px)' : 'translateX(4px)',
              }}
            />
          </button>
        </div>
      </motion.div>

      {/* ── Server Section（服务器级设置，修改后需重启） ─────────────── */}
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, delay: 0.3 }}
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
          {t('settings.server')}
        </h3>

        {/* Force HTTPS Toggle */}
        <div className="flex items-center justify-between mb-5">
          <div className="flex-1 mr-4">
            <span
              className="text-sm block"
              style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.force_https')}
            </span>
            <span
              className="text-xs block mt-1"
              style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
            >
              {t('settings.force_https_hint')}
            </span>
          </div>
          <button
            onClick={() => {
              const newVal = !forceHttps
              setForceHttps(newVal)
              // 保存到 app_settings 并提醒需重启
              settingsApi.save({ force_https: String(newVal) } as any).then(() => {
                addToast(t('settings.restart_required'), 'warning', { duration: 5000 })
              }).catch(() => loadSettings())
            }}
            className="relative inline-flex h-6 w-11 items-center rounded-full transition-colors cursor-pointer shrink-0"
            style={{
              backgroundColor: forceHttps ? 'var(--color-primary)' : 'var(--color-muted)',
            }}
          >
            <span
              className="inline-block h-4 w-4 transform rounded-full transition-transform"
              style={{
                backgroundColor: forceHttps ? 'var(--color-primary-foreground)' : 'var(--color-muted-foreground)',
                transform: forceHttps ? 'translateX(24px)' : 'translateX(4px)',
              }}
            />
          </button>
        </div>

        {/* Custom Base Path */}
        <div className="space-y-1.5">
          <label
            className="text-xs font-medium"
            style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {t('settings.base_path')}
          </label>
          <span
            className="text-xs block mb-1.5"
            style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {t('settings.base_path_hint')}
          </span>
            <input
              type="text"
              value={basePath}
              onChange={e => setBasePath(e.target.value)}
              onKeyDown={handleKeyDown}
              onBlur={e => {
                // 路由前缀存储规范：纯路径名（无斜杠），空字符串表示无前缀
                // 前端 trim 后直接保存，斜杠由后端正则兜底拒绝
                const val = e.target.value.trim()
                setBasePath(val)
                settingsApi.save({ custom_base_path: val } as any).then(() => {
                  addToast(t('settings.restart_required'), 'warning', { duration: 5000 })
                }).catch(() => loadSettings())
              }}
              placeholder="my-path"
              className="w-full px-3 py-2 text-sm rounded-lg outline-none transition-colors"
              style={{
                backgroundColor: 'var(--color-background)',
                color: 'var(--color-foreground)',
                border: '1px solid var(--color-border)',
                fontFamily: 'var(--font-heading)',
              }}
              onFocus={e => (e.target.style.borderColor = 'var(--color-primary)')}
              onBlurCapture={e => (e.target.style.borderColor = 'var(--color-border)')}
            />
        </div>

        {/* JWT 过期时间 */}
        <div className="space-y-1.5 mt-5">
          <label
            className="text-xs font-medium"
            style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {t('settings.jwt_expire_hours')}
          </label>
          <span
            className="text-xs block mb-1.5"
            style={{ color: 'var(--color-muted-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {t('settings.jwt_expire_hours_hint')}
          </span>
          <input
            type="number"
            min="1"
            max="8760"
            value={jwtExpireHours}
            onChange={e => setJwtExpireHours(e.target.value)}
            onKeyDown={handleKeyDown}
            // 失焦保存：后端校验正整数范围 1-8760，非法值自动回滚
            onBlur={() => handleBlur('jwt_expire_hours', jwtExpireHours)}
            className="w-40 px-3 py-2 text-sm rounded-lg outline-none transition-colors"
            style={{
              backgroundColor: 'var(--color-background)',
              color: 'var(--color-foreground)',
              border: '1px solid var(--color-border)',
              fontFamily: 'var(--font-heading)',
            }}
            onFocus={e => (e.target.style.borderColor = 'var(--color-primary)')}
            onBlurCapture={e => (e.target.style.borderColor = 'var(--color-border)')}
          />
        </div>
      </motion.div>
    </div>
  )
}
