import { useState } from 'react'
import { authApi } from '../lib/api'
import { useT } from '../lib/i18n'
import { useStore } from '../store'

interface LoginViewProps {
  onSuccess: () => void
}

/**
 * LoginView 登录页面
 * 极简居中卡片设计，与项目 Anthropic 暖色调风格一致。
 * 表单：用户名 + 密码 + 可选 TOTP 动态码
 */
export function LoginView({ onSuccess }: LoginViewProps) {
  const t = useT()
  const addToast = useStore(s => s.addToast)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username.trim() || !password.trim()) {
      addToast(t('auth.login_failed'), 'error')
      return
    }

    setLoading(true)
    try {
      const res = await authApi.login({
        username: username.trim(),
        password: password.trim(),
        totp_code: totpCode.trim() || undefined,
      })
      const { token } = res.data
      localStorage.setItem('auth_token', token)
      onSuccess()
    } catch (err: any) {
      const errMsg = err.response?.data?.error || ''
      if (errMsg.includes('totp')) {
        addToast(t('auth.totp_invalid'), 'error')
      } else {
        addToast(t('auth.login_failed'), 'error')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      className="min-h-screen flex items-center justify-center px-4"
      style={{ backgroundColor: 'var(--color-background)' }}
    >
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm rounded-2xl p-8 space-y-6"
        style={{
          backgroundColor: 'var(--color-card)',
          border: '1px solid var(--color-border)',
          boxShadow: '0 4px 24px rgba(0,0,0,0.06)',
        }}
      >
        {/* 品牌标题 */}
        <div className="text-center space-y-1">
          <h1
            className="text-xl font-semibold tracking-tight"
            style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            v2rayN-Go
          </h1>
          <p className="text-xs" style={{ color: 'var(--color-muted-foreground)' }}>
            {t('auth.login')}
          </p>
        </div>

        {/* 用户名 */}
        <div className="space-y-1.5">
          <label
            className="text-xs font-medium"
            style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {t('auth.username')}
          </label>
          <input
            type="text"
            value={username}
            onChange={e => setUsername(e.target.value)}
            autoComplete="username"
            autoFocus
            className="w-full px-3 py-2 text-sm rounded-lg outline-none transition-colors"
            style={{
              backgroundColor: 'var(--color-background)',
              color: 'var(--color-foreground)',
              border: '1px solid var(--color-border)',
              fontFamily: 'var(--font-heading)',
            }}
            onFocus={e => (e.target.style.borderColor = 'var(--color-primary)')}
            onBlur={e => (e.target.style.borderColor = 'var(--color-border)')}
          />
        </div>

        {/* 密码 */}
        <div className="space-y-1.5">
          <label
            className="text-xs font-medium"
            style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {t('auth.password')}
          </label>
          <input
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            autoComplete="current-password"
            className="w-full px-3 py-2 text-sm rounded-lg outline-none transition-colors"
            style={{
              backgroundColor: 'var(--color-background)',
              color: 'var(--color-foreground)',
              border: '1px solid var(--color-border)',
              fontFamily: 'var(--font-heading)',
            }}
            onFocus={e => (e.target.style.borderColor = 'var(--color-primary)')}
            onBlur={e => (e.target.style.borderColor = 'var(--color-border)')}
          />
        </div>

        {/* TOTP 动态码（始终显示，提示可留空） */}
        <div className="space-y-1.5">
          <label
            className="text-xs font-medium"
            style={{ color: 'var(--color-foreground)', fontFamily: 'var(--font-heading)' }}
          >
            {t('auth.totp')}
          </label>
          <input
            type="text"
            value={totpCode}
            onChange={e => setTotpCode(e.target.value)}
            placeholder={t('auth.totp_placeholder')}
            autoComplete="one-time-code"
            inputMode="numeric"
            maxLength={6}
            className="w-full px-3 py-2 text-sm rounded-lg outline-none transition-colors"
            style={{
              backgroundColor: 'var(--color-background)',
              color: 'var(--color-foreground)',
              border: '1px solid var(--color-border)',
              fontFamily: 'var(--font-heading)',
            }}
            onFocus={e => (e.target.style.borderColor = 'var(--color-primary)')}
            onBlur={e => (e.target.style.borderColor = 'var(--color-border)')}
          />
        </div>

        {/* 登录按钮 */}
        <button
          type="submit"
          disabled={loading}
          className="w-full py-2.5 text-sm font-medium rounded-lg transition-opacity hover:opacity-90 disabled:opacity-50"
          style={{
            backgroundColor: 'var(--color-primary)',
            color: 'var(--color-primary-foreground, #fff)',
            fontFamily: 'var(--font-heading)',
          }}
        >
          {loading ? '...' : t('auth.login')}
        </button>
      </form>
    </div>
  )
}