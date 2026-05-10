import { create } from 'zustand'

export interface Profile {
  ID: number
  name: string
  address: string
  port: number
  protocol: string
  test_result: string
  is_active: boolean
  group_id: number
  group_name: string
  sort_order: number
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

  // Profiles
  profiles: Profile[]
  activeProfile: Profile | null
  setProfiles: (p: Profile[]) => void
  setActiveProfile: (p: Profile | null) => void

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
  currentView: string
  setCurrentView: (v: string) => void
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

  // Profiles
  profiles: [],
  activeProfile: null,
  setProfiles: (p) => set({ profiles: p }),
  setActiveProfile: (p) => set({ activeProfile: p }),

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

  // UI
  currentView: 'home',
  setCurrentView: (v) => set({ currentView: v }),
}))