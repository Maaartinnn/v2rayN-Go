import { create } from 'zustand'
import enUS from '../locales/en-US'
import zhCN from '../locales/zh-CN'

export const AVAILABLE_LANGUAGES = [
  { code: 'zh' as const, label: '中文' },
  { code: 'en' as const, label: 'English' },
] as const

export type Lang = (typeof AVAILABLE_LANGUAGES)[number]['code']

interface I18nState {
  lang: Lang
  theme: 'light' | 'dark' | 'system'
  setLang: (lang: Lang) => void
  setTheme: (theme: 'light' | 'dark' | 'system') => void
}

function detectLang(): Lang {
  const stored = localStorage.getItem('lang')
  if (isLang(stored)) return stored
  const nav = navigator.language.toLowerCase()
  if (nav.startsWith('zh')) return 'zh'
  return 'en'
}

function isLang(val: string | null): val is Lang {
  return AVAILABLE_LANGUAGES.some((l) => l.code === val)
}

function detectTheme(): 'light' | 'dark' | 'system' {
  const stored = localStorage.getItem('theme')
  if (stored === 'light' || stored === 'dark' || stored === 'system') return stored
  return 'system'
}

export const useI18n = create<I18nState>((set) => ({
  lang: detectLang(),
  theme: detectTheme(),
  setLang: (lang) => {
    localStorage.setItem('lang', lang)
    set({ lang })
  },
  setTheme: (theme) => {
    localStorage.setItem('theme', theme)
    applyTheme(theme)
    set({ theme })
  },
}))

// ========== Theme Management ==========

export function applyTheme(theme: 'light' | 'dark' | 'system') {
  const root = document.documentElement
  root.classList.remove('light', 'dark')

  if (theme === 'system') {
    // Let CSS media query handle it — remove manual overrides
    root.removeAttribute('data-theme')
  } else {
    root.setAttribute('data-theme', theme)
    root.classList.add(theme)
  }
}

export function initTheme() {
  const theme = useI18n.getState().theme
  applyTheme(theme)
}

// ========== Translation Dictionary ==========

const translations = { en: enUS, zh: zhCN } as const

type TranslationKey = keyof typeof enUS

export function t(key: TranslationKey, params?: Record<string, string | number>): string {
  const lang = useI18n.getState().lang
  let text: string = translations[lang][key] ?? key
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      text = text.replaceAll(`{${k}}`, String(v))
    }
  }
  return text
}

// React hook for reactive translations
export function useT() {
  const lang = useI18n((s) => s.lang)
  return (key: TranslationKey, params?: Record<string, string | number>): string => {
    let text: string = translations[lang][key] ?? key
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        text = text.replaceAll(`{${k}}`, String(v))
      }
    }
    return text
  }
}
