// AccountView.tsx — 账户管理页面
//
// 功能区块：
//   1. 修改密码（旧密码 + 新密码 + 确认，独立提交按钮，成功后自动 RotateJWTSecret）
//   2. 两步验证 TOTP（开关 + 二维码 + 动态码验证）
//   3. 会话管理（退出登录 / 注销所有设备）

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { QRCodeSVG } from 'qrcode.react'
import { Lock, Shield, LogOut } from 'lucide-react'
import { authApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { useStore } from '../store'

// 读取 Go 后端注入的 custom_base_path，用于退出登录后重定向到正确的路径
const basePath =
  window.__BASE_PATH__ === '__INJECT_BASE_PATH__' ? '' : (window.__BASE_PATH__ || '')

export function AccountView() {
  const t = useT()
  const addToast = useStore(s => s.addToast)

  // ── 用户信息 ──────────────────────────────────────────────────────
  const [user, setUser] = useState<{ uuid: string; username: string; role: number; totp_enabled: boolean } | null>(null)

  useEffect(() => {
    authApi.me().then(res => setUser(res.data)).catch(() => {})
  }, [])

  // ── 卡片样式（与 SettingsView 一致）────────────────────────────────
  const cardStyle = {
    backgroundColor: 'var(--color-card)',
    borderColor: 'var(--color-border)',
    boxShadow: 'var(--shadow-card)',
  }
  const labelStyle: React.CSSProperties = {
    color: 'var(--color-foreground)',
    fontFamily: 'var(--font-heading)',
  }
  const inputStyle: React.CSSProperties = {
    backgroundColor: 'var(--color-background)',
    color: 'var(--color-foreground)',
    border: '1px solid var(--color-border)',
    fontFamily: 'var(--font-heading)',
  }
  const focusBlur = {
    onFocus: (e: React.FocusEvent<HTMLInputElement>) => (e.target.style.borderColor = 'var(--color-primary)'),
    onBlur: (e: React.FocusEvent<HTMLInputElement>) => (e.target.style.borderColor = 'var(--color-border)'),
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1
        className="text-xl font-semibold mb-6"
        style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
      >
        {t('account.title')}
      </h1>

      {/* ── 卡片 1：修改密码 ──────────────────────────────────────── */}
      <ChangePasswordCard t={t} addToast={addToast} cardStyle={cardStyle} labelStyle={labelStyle} inputStyle={inputStyle} focusBlur={focusBlur} />

      {/* ── 卡片 2：两步验证 (TOTP) ───────────────────────────────── */}
      <TOTPCard t={t} addToast={addToast} user={user} setUser={setUser} cardStyle={cardStyle} labelStyle={labelStyle} inputStyle={inputStyle} focusBlur={focusBlur} />

      {/* ── 卡片 3：会话管理 ──────────────────────────────────────── */}
      <SessionCard t={t} addToast={addToast} cardStyle={cardStyle} />
    </div>
  )
}

// ═══════════════════════════════════════════════════════════════════════════
// 子组件：修改密码卡片
// ═══════════════════════════════════════════════════════════════════════════

function ChangePasswordCard({ t, addToast, cardStyle, labelStyle, inputStyle, focusBlur }: any) {
  const [oldPwd, setOldPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')
  const [confirmPwd, setConfirmPwd] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async () => {
    // 1. 三个字段都不能为空
    if (!oldPwd || !newPwd || !confirmPwd) return

    // 2. 新旧密码不能相同
    if (newPwd === oldPwd) {
      addToast(t('account.password_same_as_old'), 'error', { duration: 5000 })
      return
    }

    // 3. 两次新密码必须一致
    if (newPwd !== confirmPwd) {
      addToast(t('account.password_mismatch'), 'error', { duration: 5000 })
      return
    }

    // 4. 最少 6 位
    if (newPwd.length < 6) {
      addToast(t('account.password_too_short'), 'error', { duration: 5000 })
      return
    }

    setLoading(true)
    try {
      const res = await authApi.changePassword({ old_password: oldPwd, new_password: newPwd })
      // 后端返回新 JWT，更新 localStorage 使当前设备无缝续用
      const newToken = res.data.token
      if (newToken) {
        localStorage.setItem('auth_token', newToken)
      }
      addToast(t('account.password_changed'), 'success', { duration: 5000 })
      setOldPwd('')
      setNewPwd('')
      setConfirmPwd('')
    } catch (err: any) {
      const msg = err.response?.data?.error || 'Error'
      addToast(msg, 'error', { duration: 5000 })
    } finally {
      setLoading(false)
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: 0.05 }}
      className="rounded-xl border p-6 mb-4"
      style={cardStyle}
    >
      <div className="flex items-center gap-2.5 mb-5">
        <Lock size={16} style={{ color: 'var(--color-muted-foreground)' }} />
        <h3 className="text-sm font-semibold" style={labelStyle}>
          {t('account.change_password')}
        </h3>
      </div>

      <div className="space-y-4">
        <div className="space-y-1.5">
          <label className="text-xs font-medium" style={labelStyle}>{t('account.old_password')}</label>
          <input
            type="password"
            value={oldPwd}
            onChange={e => setOldPwd(e.target.value)}
            className="w-full px-3 py-2 text-sm rounded-lg outline-none transition-colors"
            style={inputStyle}
            {...focusBlur}
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-xs font-medium" style={labelStyle}>{t('account.new_password')}</label>
          <input
            type="password"
            value={newPwd}
            onChange={e => setNewPwd(e.target.value)}
            className="w-full px-3 py-2 text-sm rounded-lg outline-none transition-colors"
            style={inputStyle}
            {...focusBlur}
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-xs font-medium" style={labelStyle}>{t('account.confirm_password')}</label>
          <input
            type="password"
            value={confirmPwd}
            onChange={e => setConfirmPwd(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleSubmit() }}
            className="w-full px-3 py-2 text-sm rounded-lg outline-none transition-colors"
            style={inputStyle}
            {...focusBlur}
          />
        </div>
        <button
          onClick={handleSubmit}
          disabled={loading}
          className="btn-primary px-4 py-2 text-sm"
        >
          {loading ? '...' : t('account.submit_password')}
        </button>
      </div>
    </motion.div>
  )
}

// ═══════════════════════════════════════════════════════════════════════════
// 子组件：TOTP 两步验证卡片
// ═══════════════════════════════════════════════════════════════════════════

function TOTPCard({ t, addToast, user, setUser, cardStyle, labelStyle, inputStyle, focusBlur }: any) {
  const [totpSetup, setTotpSetup] = useState<{ secret: string; otpauth_url: string } | null>(null)
  const [verifyCode, setVerifyCode] = useState('')
  const [disablePassword, setDisablePassword] = useState('')
  const [loading, setLoading] = useState(false)

  // 开启 TOTP 流程
  const handleEnable = async () => {
    setLoading(true)
    try {
      const res = await authApi.enableTOTP()
      setTotpSetup(res.data)
    } catch (err: any) {
      addToast(err.response?.data?.error || 'Error', 'error')
    } finally {
      setLoading(false)
    }
  }

  // 验证并激活 TOTP
  const handleVerify = async () => {
    if (verifyCode.length !== 6) return
    setLoading(true)
    try {
      await authApi.verifyTOTP(verifyCode)
      addToast(t('account.totp_enabled_ok'), 'success')
      setUser((u: any) => u ? { ...u, totp_enabled: true } : u)
      setTotpSetup(null)
      setVerifyCode('')
    } catch (err: any) {
      addToast(err.response?.data?.error || 'Error', 'error')
    } finally {
      setLoading(false)
    }
  }

  // 关闭 TOTP
  const handleDisable = async () => {
    if (!disablePassword) return
    setLoading(true)
    try {
      await authApi.disableTOTP(disablePassword)
      addToast(t('account.totp_disabled_ok'), 'success')
      setUser((u: any) => u ? { ...u, totp_enabled: false } : u)
      setDisablePassword('')
    } catch (err: any) {
      addToast(err.response?.data?.error || 'Error', 'error')
    } finally {
      setLoading(false)
    }
  }

  const isEnabled = user?.totp_enabled

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: 0.1 }}
      className="rounded-xl border p-6 mb-4"
      style={cardStyle}
    >
      <div className="flex items-center justify-between mb-5">
        <div className="flex items-center gap-2.5">
          <Shield size={16} style={{ color: 'var(--color-muted-foreground)' }} />
          <h3 className="text-sm font-semibold" style={labelStyle}>
            {t('account.totp_title')}
          </h3>
        </div>
        <span
          className="text-xs px-2 py-0.5 rounded-full"
          style={{
            backgroundColor: isEnabled ? 'var(--color-success-light, #dcfce7)' : 'var(--color-muted)',
            color: isEnabled ? 'var(--color-success)' : 'var(--color-muted-foreground)',
            fontFamily: 'var(--font-heading)',
          }}
        >
          {isEnabled ? t('account.totp_enabled') : t('account.totp_disabled')}
        </span>
      </div>

      {/* 已启用 + 未在配置中 → 显示关闭入口 */}
      {isEnabled && !totpSetup && (
        <div className="space-y-3">
          <p className="text-xs" style={{ color: 'var(--color-muted-foreground)' }}>
            {t('account.totp_disable_confirm')}
          </p>
          <div className="flex items-center gap-2">
            <input
              type="password"
              value={disablePassword}
              onChange={e => setDisablePassword(e.target.value)}
              placeholder={t('account.old_password')}
              className="flex-1 px-3 py-2 text-sm rounded-lg outline-none transition-colors"
              style={inputStyle}
              {...focusBlur}
            />
            <button
              onClick={handleDisable}
              disabled={loading || !disablePassword}
              className="btn-danger px-4 py-2 text-sm"
            >
              {t('common.disabled')}
            </button>
          </div>
        </div>
      )}

      {/* 未启用 + 未在配置中 → 显示开启按钮 */}
      {!isEnabled && !totpSetup && (
        <button
          onClick={handleEnable}
          disabled={loading}
          className="btn-primary px-4 py-2 text-sm"
        >
          {loading ? '...' : t('common.enabled')}
        </button>
      )}

      {/* 配置中（已调用 enable，等待验证）→ 显示二维码 + 动态码输入 */}
      {totpSetup && (
        <div className="space-y-4">
          <p className="text-xs" style={{ color: 'var(--color-muted-foreground)' }}>
            {t('account.totp_scan')}
          </p>

          {/* 二维码 */}
          <div className="flex justify-center">
            <div className="p-3 rounded-lg" style={{ backgroundColor: '#fff' }}>
              <QRCodeSVG value={totpSetup.otpauth_url} size={160} />
            </div>
          </div>

          {/* 密钥（可手动输入） */}
          <div className="space-y-1">
            <label className="text-xs font-medium" style={labelStyle}>{t('account.totp_secret_label')}</label>
            <code
              className="block w-full px-3 py-2 text-xs rounded-lg select-all"
              style={{
                backgroundColor: 'var(--color-muted)',
                color: 'var(--color-foreground)',
                fontFamily: 'monospace',
              }}
            >
              {totpSetup.secret}
            </code>
          </div>

          {/* 动态码输入 + 确认按钮 */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium" style={labelStyle}>{t('account.totp_enter_code')}</label>
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={verifyCode}
                onChange={e => setVerifyCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                maxLength={6}
                inputMode="numeric"
                className="flex-1 px-3 py-2 text-sm rounded-lg outline-none transition-colors tracking-[0.5em]"
                style={inputStyle}
                {...focusBlur}
              />
              <button
                onClick={handleVerify}
                disabled={loading || verifyCode.length !== 6}
                className="btn-primary px-4 py-2 text-sm"
              >
                {loading ? '...' : t('account.totp_confirm_enable')}
              </button>
            </div>
          </div>

          {/* 取消配置 */}
          <button
            onClick={() => { setTotpSetup(null); setVerifyCode('') }}
            className="btn-ghost text-xs"
          >
            {t('common.back')}
          </button>
        </div>
      )}
    </motion.div>
  )
}

// ═══════════════════════════════════════════════════════════════════════════
// 子组件：会话管理卡片
// ═══════════════════════════════════════════════════════════════════════════

function SessionCard({ t, addToast, cardStyle }: any) {
  const [showConfirm, setShowConfirm] = useState(false)

  // 退出登录（仅清除当前设备 token）
  // 重定向时需带上 basePath 前缀，确保落在正确的路由前缀下
  const handleLogout = () => {
    localStorage.removeItem('auth_token')
    window.location.href = basePath || '/'
  }

  // 注销所有设备（后端刷新 JWTSecret）
  const handleRevokeAll = async () => {
    try {
      await authApi.revokeAllSessions()
      addToast(t('account.sessions_revoked'), 'success')
      // 刷新后本设备也失效了，需要重新登录
      // 重定向时需带上 basePath 前缀，确保落在正确的路由前缀下
      localStorage.removeItem('auth_token')
      window.location.href = basePath || '/'
    } catch (err: any) {
      addToast(err.response?.data?.error || 'Error', 'error')
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: 0.15 }}
      className="rounded-xl border p-6 mb-4"
      style={cardStyle}
    >
      <div className="flex items-center gap-2.5 mb-5">
        <LogOut size={16} style={{ color: 'var(--color-muted-foreground)' }} />
        <h3
          className="text-sm font-semibold"
          style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
        >
          {t('account.sessions')}
        </h3>
      </div>

      <div className="space-y-3">
        {/* 退出登录 */}
        <button
          onClick={handleLogout}
          className="btn-secondary w-full px-4 py-2.5 text-sm text-left"
        >
          {t('account.logout')}
        </button>

        {/* 注销所有设备 */}
        {!showConfirm ? (
            <button
              onClick={() => setShowConfirm(true)}
              className="btn-danger w-full px-4 py-2.5 text-sm text-left"
            >
              {t('account.revoke_all')}
            </button>
        ) : (
          /* 确认横幅（红色警告条风格） */
          <div
            className="rounded-lg p-4 space-y-3"
            style={{
              backgroundColor: 'var(--color-destructive-light, #fef2f2)',
              border: '1px solid var(--color-destructive, #ef4444)',
            }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--color-destructive, #ef4444)' }}>
              {t('account.revoke_all_confirm')}
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={handleRevokeAll}
                className="btn-danger px-4 py-1.5 text-xs"
              >
                {t('account.revoke_all')}
              </button>
              <button
                onClick={() => setShowConfirm(false)}
                className="btn-ghost px-4 py-1.5 text-xs"
              >
                {t('nodes.cancel')}
              </button>
            </div>
          </div>
        )}
      </div>
    </motion.div>
  )
}