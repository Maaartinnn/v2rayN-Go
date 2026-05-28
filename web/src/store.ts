import { create } from 'zustand'

// 颜色对（背景色 + 文字色），由后端计算返回
export interface ColorPair {
  bg: string
  text: string
}

// ProfileListItem 节点列表精简数据，由 GET /api/profiles 返回。
// 仅包含前端列表展示和操作所需字段，编辑时通过 GET /api/profiles/{uuid} 获取完整数据。
export interface ProfileListItem {
  uuid: string
  name: string
  proxy_protocol: string
  proxy_address: string
  proxy_port: number
  core_type: string
  test_result: string
  is_active: boolean
  group_uuid: string
  protocol_color: ColorPair   // 后端计算的协议徽标颜色
  core_color: ColorPair       // 后端计算的内核徽标颜色
  latency_color: string       // 后端计算的延迟指示灯颜色（CSS 变量）
}

// Profile 完整节点数据，由 GET /api/profiles/{uuid} 返回。
// 仅在编辑节点时按需获取，列表页面不使用此类型。
export interface Profile {
  ID: number
  uuid: string
  name: string
  proxy_address: string
  proxy_port: number
  proxy_protocol: string
  proxy_credential: string
  proxy_alter_id: number
  proxy_security: string
  proxy_network: string
  proxy_tls: string
  proxy_sni: string
  proxy_fingerprint: string
  proxy_allow_insecure: boolean
  proxy_host: string
  proxy_path: string
  proxy_seed: string
  proxy_flow: string
  proxy_public_key: string
  proxy_short_id: string
  proxy_sider_sni: string
  proxy_dialer_proxy: string
  raw_link: string
  test_result: string
  last_test_time: string
  is_active: boolean
  sort_order: number
  core_type: string
  group_uuid: string
}

export interface CoreStatus {
  type: string
  status: string
  pid: number
  start_time: string
  error_msg: string
}

export interface LogEntry {
  time: string
  level: string
  content: string
  source: string // "system", "xray", "sing-box", "mihomo"
}

export interface Metrics {
  upload_speed: number
  download_speed: number
  upload_total: number
  download_total: number
}

export interface DownloadProgress {
  core_name: string
  downloaded: number
  total: number
  percentage: number
  status: string // "downloading", "complete", "error"
  error?: string
}

interface AppState {
  // Core
  isConnected: boolean
  coreStatuses: CoreStatus[]
  setConnected: (v: boolean) => void
  setCoreStatuses: (s: CoreStatus[]) => void

  // Profile List（精简数据，用于列表展示）
  profileList: ProfileListItem[]
  setProfileList: (p: ProfileListItem[] | ((prev: ProfileListItem[]) => ProfileListItem[])) => void

  // Active Profile（仅存 uuid，避免存储完整 Profile 数据）
  activeProfileUUID: string | null
  setActiveProfileUUID: (uuid: string | null) => void

  // Metrics
  metrics: Metrics
  setMetrics: (m: Partial<Metrics>) => void

  // Core versions (async via WebSocket)
  coreVersions: Record<string, string>
  setCoreVersions: (versions: Record<string, string>) => void

  // Download progress
  downloadProgress: Record<string, DownloadProgress>
  setDownloadProgress: (coreName: string, progress: DownloadProgress) => void
  clearDownloadProgress: (coreName: string) => void

  // Toast notifications
  toasts: Array<{ id: number; message: string; type: 'success' | 'error' | 'info' }>
  addToast: (message: string, type: 'success' | 'error' | 'info') => void
  removeToast: (id: number) => void

  // Logs
  logs: LogEntry[]
  addLog: (entry: LogEntry) => void
  clearLogs: () => void

  // UI
}

export const useStore = create<AppState>((set) => ({
  // Core
  isConnected: false,
  coreStatuses: [],
  setConnected: (v) => set({ isConnected: v }),
  setCoreStatuses: (s) => set({
    coreStatuses: s,
    isConnected: s.some(c => c.status === 'running'),
  }),

  // Profile List（精简数据，用于列表展示）
  profileList: [],
  setProfileList: (p) => set((state) => ({
    profileList: typeof p === 'function' ? (p as (prev: ProfileListItem[]) => ProfileListItem[])(state.profileList) : p
  })),

  // Active Profile（仅存 uuid）
  activeProfileUUID: null,
  setActiveProfileUUID: (uuid) => set({ activeProfileUUID: uuid }),

  // Metrics
  metrics: { upload_speed: 0, download_speed: 0, upload_total: 0, download_total: 0 },
  setMetrics: (m) => set((state) => ({ metrics: { ...state.metrics, ...m } })),

  // Core versions (async via WebSocket)
  coreVersions: {},
  setCoreVersions: (versions) => set({ coreVersions: versions }),

  // Download progress
  downloadProgress: {},
  setDownloadProgress: (coreName, progress) => set((state) => ({
    downloadProgress: { ...state.downloadProgress, [coreName]: progress },
  })),
  clearDownloadProgress: (coreName) => set((state) => {
    const { [coreName]: _, ...rest } = state.downloadProgress
    return { downloadProgress: rest }
  }),

  // Toast notifications
  toasts: [],
  addToast: (message, type) => {
    const id = Date.now()
    set((state) => ({
      toasts: [...state.toasts, { id, message, type }],
    }))
    // Auto remove after 5 seconds
    setTimeout(() => {
      set((state) => ({
        toasts: state.toasts.filter((t) => t.id !== id),
      }))
    }, 5000)
  },
  removeToast: (id) => set((state) => ({
    toasts: state.toasts.filter((t) => t.id !== id),
  })),

  // Logs
  logs: [],
  addLog: (entry) => set((state) => ({
    logs: [...state.logs.slice(-499), entry],
  })),
  clearLogs: () => set({ logs: [] }),

}))