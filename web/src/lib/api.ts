import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// ========== Core API ==========
export const coreApi = {
  start: (coreType: string, configPath: string) =>
    api.post('/core/start', { core_type: coreType, config_path: configPath }),
  stop: (coreType: string) =>
    api.post('/core/stop', { core_type: coreType }),
  status: () =>
    api.get('/core/status'),
}

// ========== Profile API ==========
export const profileApi = {
  list: () => api.get('/profiles'),
  get: (id: number) => api.get(`/profiles/${id}`),
  create: (data: any) => api.post('/profiles', data),
  update: (id: number, data: any) => api.put(`/profiles/${id}`, data),
  delete: (id: number) => api.delete(`/profiles/${id}`),
  select: (id: number) => api.post(`/profiles/${id}/select`),
  ping: (id: number) => api.post(`/profiles/${id}/ping`),
  pingAll: () => api.post('/profiles/ping-all'),
  importLinks: (links: string) => api.post('/profiles/import', { links }),
}

// ========== Subscription API ==========
export const subscriptionApi = {
  list: () => api.get('/subscriptions'),
  create: (data: any) => api.post('/subscriptions', data),
  update: (id: number, data: any) => api.put(`/subscriptions/${id}`, data),
  delete: (id: number) => api.delete(`/subscriptions/${id}`),
  refresh: (id: number) => api.post(`/subscriptions/${id}/refresh`),
  refreshAll: () => api.post('/subscriptions/refresh-all'),
}

// ========== Groups API ==========
export const groupsApi = {
  list: () => api.get('/groups'),
  create: (data: { name: string; description?: string; color?: string }) =>
    api.post('/groups', data),
  update: (data: { ID: number; name: string; description?: string; color?: string }) =>
    api.put('/groups', data),
  delete: (id: number) => api.delete('/groups', { data: { id } }),
}

// ========== Profile Enhancements ==========
export const profileEnhancedApi = {
  dedup: () => api.post('/profiles/dedup'),
  importImage: (formData: FormData) =>
    api.post('/profiles/import-image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 30000,
    }),
}

// ========== Strategy Groups API ==========
export const strategyGroupsApi = {
  list: () => api.get('/strategy-groups'),
  get: (id: number) => api.get(`/strategy-groups/${id}`),
  create: (data: any) => api.post('/strategy-groups', data),
  update: (id: number, data: any) => api.put(`/strategy-groups/${id}`, data),
  delete: (id: number) => api.delete(`/strategy-groups/${id}`),
}

// ========== Routing API ==========
export const routingApi = {
  list: () => api.get('/routing-rules'),
  create: (data: any) => api.post('/routing-rules', data),
  update: (id: number, data: any) => api.put(`/routing-rules/${id}`, data),
  delete: (id: number) => api.delete(`/routing-rules/${id}`),
}

// ========== Core Hub API ==========
export const coresApi = {
  list: () => api.get('/cores'),
  checkUpdates: () => api.get('/cores/check-updates'),
  detectVersions: () => api.get('/cores/detect-versions'),
  download: (coreName: string) => api.post('/cores/download', { core_name: coreName }),
  downloadUrl: (coreName: string, downloadUrl: string) =>
    api.post('/cores/download-url', { core_name: coreName, download_url: downloadUrl }),
  upload: (formData: FormData) =>
    api.post('/cores/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 300000,
    }),
}

// ========== Updater API (legacy) ==========
export const updaterApi = {
  check: () => api.get('/updater/check'),
  download: (coreName: string) => api.post(`/updater/download/${coreName}`),
}

// ========== Settings API ==========
export const settingsApi = {
  get: () => api.get('/settings'),
  save: (data: {
    listen_ip?: string
    socks_port?: number
    http_port?: number
    outbound_ip?: string
    github_mirror?: string
  }) => api.post('/settings', data),
}

// ========== System Proxy API ==========
export const proxyApi = {
  setSystemProxy: (enabled: boolean, port: number) =>
    api.post('/proxy/system', { enabled, port }),
  getSystemProxy: () => api.get('/proxy/system'),
}

export default api
