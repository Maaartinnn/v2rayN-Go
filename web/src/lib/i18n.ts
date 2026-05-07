import { create } from 'zustand'

type Lang = 'en' | 'zh'

interface I18nState {
  lang: Lang
  setLang: (lang: Lang) => void
}

function detectLang(): Lang {
  const stored = localStorage.getItem('lang')
  if (stored === 'en' || stored === 'zh') return stored
  const nav = navigator.language.toLowerCase()
  if (nav.startsWith('zh')) return 'zh'
  return 'en'
}

export const useI18n = create<I18nState>((set) => ({
  lang: detectLang(),
  setLang: (lang) => {
    localStorage.setItem('lang', lang)
    set({ lang })
  },
}))

// ========== Translation Dictionary ==========

const translations = {
  en: {
    // Sidebar
    'nav.home': 'Home',
    'nav.nodes': 'Nodes',
    'nav.logs': 'Logs',
    'nav.settings': 'Settings',
    'nav.cores': 'Cores',

    // Home
    'home.connected': 'Connected',
    'home.disconnected': 'Disconnected',
    'home.no_node': 'No node selected',
    'home.upload': 'Upload',
    'home.download': 'Download',
    'home.latency': 'Latency',

    // Nodes
    'nodes.title': 'Nodes',
    'nodes.test_all': 'Test All',
    'nodes.import': 'Import',
    'nodes.no_nodes': 'No nodes yet',
    'nodes.import_hint': 'Import share links or add a subscription',
    'nodes.cancel': 'Cancel',
    'nodes.import_placeholder': 'Paste share links here (vmess://, vless://, trojan://, ss://, ...)\nOne link per line',

    // Logs
    'logs.title': 'Logs',
    'logs.clear': 'Clear',
    'logs.no_logs': 'No logs yet',
    'logs.start_hint': 'Start the core to see logs',

    // Settings
    'settings.title': 'Settings',
    'settings.coming_soon': 'Settings panel coming soon...',

    // Cores
    'cores.title': 'Core Manager',
    'cores.coming_soon': 'Core download & update panel coming soon...',

    // Common
    'common.error': 'Error',
    'common.success': 'Success',
  },
  zh: {
    // 侧边栏
    'nav.home': '首页',
    'nav.nodes': '节点',
    'nav.logs': '日志',
    'nav.settings': '设置',
    'nav.cores': '内核',

    // 首页
    'home.connected': '已连接',
    'home.disconnected': '未连接',
    'home.no_node': '未选择节点',
    'home.upload': '上传',
    'home.download': '下载',
    'home.latency': '延迟',

    // 节点
    'nodes.title': '节点',
    'nodes.test_all': '全部测速',
    'nodes.import': '导入',
    'nodes.no_nodes': '暂无节点',
    'nodes.import_hint': '导入分享链接或添加订阅',
    'nodes.cancel': '取消',
    'nodes.import_placeholder': '在此粘贴分享链接 (vmess://, vless://, trojan://, ss://, ...)\n每行一个链接',

    // 日志
    'logs.title': '日志',
    'logs.clear': '清空',
    'logs.no_logs': '暂无日志',
    'logs.start_hint': '启动内核后即可查看日志',

    // 设置
    'settings.title': '设置',
    'settings.coming_soon': '设置面板即将推出...',

    // 内核
    'cores.title': '内核管理',
    'cores.coming_soon': '内核下载与更新面板即将推出...',

    // 通用
    'common.error': '错误',
    'common.success': '成功',
  },
} as const

type TranslationKey = keyof typeof translations.en

export function t(key: TranslationKey): string {
  const lang = useI18n.getState().lang
  return translations[lang][key] ?? key
}

// React hook for reactive translations
export function useT() {
  const lang = useI18n((s) => s.lang)
  return (key: TranslationKey): string => {
    return translations[lang][key] ?? key
  }
}