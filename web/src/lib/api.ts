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
  // list 获取节点列表，支持按分组和关键词筛选（服务端过滤）。
  list: (groupUuid?: string, q?: string) => {
    const params = new URLSearchParams()
    if (groupUuid) params.set('group_uuid', groupUuid)
    if (q) params.set('q', q)
    const qs = params.toString()
    return api.get(`/profiles${qs ? '?' + qs : ''}`)
  },
  get: (uuid: string) => api.get(`/profiles/${uuid}`),
  create: (data: any) => api.post('/profiles', data),
  update: (uuid: string, data: any) => api.put(`/profiles/${uuid}`, data),
  delete: (uuid: string) => api.delete(`/profiles/${uuid}`),
  coreList: (protocol: string) => api.get(`/profiles/core-list?protocol=${protocol}`),
  select: (uuid: string) => api.post(`/profiles/${uuid}/select`),
  ping: (uuid: string) => api.post(`/profiles/${uuid}/ping`),
  pingAll: () => api.post('/profiles/ping-all'),
  importLinks: (links: string, groupUuid?: string) =>
    api.post('/profiles/import', { links, group_uuid: groupUuid || '' }),
  importToGroup: (links: string, groupUuid: string) =>
    api.post('/profiles/import', { links, group_uuid: groupUuid }),
}

// ========== Groups API (unified: normal + subscription groups) ==========
export const groupsApi = {
  list: () => api.get('/groups'),
  get: (uuid: string) => api.get(`/groups/${uuid}`),
  create: (data: any) => api.post('/groups', data),
  update: (uuid: string, data: any) => api.put(`/groups/${uuid}`, data),
  delete: (uuid: string) => api.delete(`/groups/${uuid}`),
  reorder: (uuid: string, beforeUuid: string | null, afterUuid: string | null) =>
    api.put('/groups/reorder', { uuid, before_uuid: beforeUuid || '', after_uuid: afterUuid || '' }),
  refresh: (uuid: string) => api.post(`/groups/${uuid}/refresh`),
  refreshProxy: (uuid: string) => api.post(`/groups/${uuid}/refresh-proxy`),
}

// ========== Profile Enhancements ==========
export const profileEnhancedApi = {
  dedup: (groupUuid?: string) => api.post('/profiles/dedup', { group_uuid: groupUuid || '' }),
  importImage: (formData: FormData) =>
    api.post('/profiles/import-image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 30000,
    }),
}

// ========== Routing API ==========
export const routingApi = {
  list: () => api.get('/routing-rules'),
  create: (data: any) => api.post('/routing-rules', data),
  update: (uuid: string, data: any) => api.put(`/routing-rules/${uuid}`, data),
  delete: (uuid: string) => api.delete(`/routing-rules/${uuid}`),
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

// ========== Settings API ==========
export const settingsApi = {
  get: () => api.get('/settings'),
  save: (data: {
    listen_ip?: string
    socks_port?: number
    http_port?: number
    outbound_ip?: string
    github_mirror?: string
    core_config_debug?: boolean
  }) => api.post('/settings', data),
}

// ========== System Proxy API ==========
export const proxyApi = {
  setSystemProxy: (enabled: boolean, port: number) =>
    api.post('/proxy/system', { enabled, port }),
  getSystemProxy: () => api.get('/proxy/system'),
}

export default api