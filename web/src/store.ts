import { create } from 'zustand'

export interface Profile {
  ID: number
  name: string
  address: string
  port: number
  protocol: string
  test_result: string
  is_active: boolean
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