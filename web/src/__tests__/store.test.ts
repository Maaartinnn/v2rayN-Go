import { describe, it, expect, beforeEach } from 'vitest'
import { useStore, type Profile, type CoreStatus } from '../store'

describe('useStore', () => {
  beforeEach(() => {
    // Reset store to initial state
    useStore.setState({
      isConnected: false,
      coreStatuses: [],
      profiles: [],
      activeProfile: null,
      metrics: { upload_speed: 0, download_speed: 0, upload_total: 0, download_total: 0 },
      coreVersions: {},
      downloadProgress: {},
      toasts: [],
      logs: [],
    })
  })

  // ==================== Core Status ====================

  describe('setConnected', () => {
    it('should set isConnected', () => {
      useStore.getState().setConnected(true)
      expect(useStore.getState().isConnected).toBe(true)
    })

    it('should toggle isConnected', () => {
      useStore.getState().setConnected(true)
      expect(useStore.getState().isConnected).toBe(true)
      useStore.getState().setConnected(false)
      expect(useStore.getState().isConnected).toBe(false)
    })
  })

  describe('setCoreStatuses', () => {
    it('should set core statuses', () => {
      const statuses: CoreStatus[] = [
        { type: 'xray', status: 'running', pid: 1234, start_time: '', error_msg: '' },
      ]
      useStore.getState().setCoreStatuses(statuses)
      expect(useStore.getState().coreStatuses).toEqual(statuses)
    })

    it('should set isConnected=true when any core is running', () => {
      const statuses: CoreStatus[] = [
        { type: 'xray', status: 'running', pid: 1234, start_time: '', error_msg: '' },
      ]
      useStore.getState().setCoreStatuses(statuses)
      expect(useStore.getState().isConnected).toBe(true)
    })

    it('should set isConnected=false when no core is running', () => {
      const statuses: CoreStatus[] = [
        { type: 'xray', status: 'stopped', pid: 0, start_time: '', error_msg: '' },
      ]
      useStore.getState().setCoreStatuses(statuses)
      expect(useStore.getState().isConnected).toBe(false)
    })
  })

  // ==================== Profiles ====================

  describe('setProfiles', () => {
    it('should set profiles array directly', () => {
      const profiles: Profile[] = [
        {
          ID: 1, uuid: 'test-uuid', name: 'TestNode',
          proxy_address: 'host.com', proxy_port: 443, proxy_protocol: 'vless',
          proxy_credential: '', proxy_alter_id: 0, proxy_security: '', proxy_network: 'tcp',
          proxy_tls: '', proxy_sni: '', proxy_fingerprint: '', proxy_allow_insecure: false,
          proxy_host: '', proxy_path: '', proxy_seed: '', proxy_flow: '',
          proxy_public_key: '', proxy_short_id: '', proxy_sider_sni: '',
          proxy_dialer_proxy: '', raw_link: '', test_result: '', last_test_time: '',
          is_active: false, sort_order: 10, core_type: '', group_uuid: 'group-1',
        },
      ]
      useStore.getState().setProfiles(profiles)
      expect(useStore.getState().profiles).toHaveLength(1)
      expect(useStore.getState().profiles[0].name).toBe('TestNode')
    })

    it('should accept function updater', () => {
      useStore.getState().setProfiles([])
      useStore.getState().setProfiles((prev) => [
        ...prev,
        {
          ID: 1, uuid: 'test', name: 'New',
          proxy_address: '', proxy_port: 0, proxy_protocol: 'vless',
          proxy_credential: '', proxy_alter_id: 0, proxy_security: '', proxy_network: '',
          proxy_tls: '', proxy_sni: '', proxy_fingerprint: '', proxy_allow_insecure: false,
          proxy_host: '', proxy_path: '', proxy_seed: '', proxy_flow: '',
          proxy_public_key: '', proxy_short_id: '', proxy_sider_sni: '',
          proxy_dialer_proxy: '', raw_link: '', test_result: '', last_test_time: '',
          is_active: false, sort_order: 10, core_type: '', group_uuid: '',
        },
      ])
      expect(useStore.getState().profiles).toHaveLength(1)
    })
  })

  describe('setActiveProfile', () => {
    it('should set active profile', () => {
      const profile: Profile = {
        ID: 1, uuid: 'test', name: 'Active',
        proxy_address: '', proxy_port: 0, proxy_protocol: 'vless',
        proxy_credential: '', proxy_alter_id: 0, proxy_security: '', proxy_network: '',
        proxy_tls: '', proxy_sni: '', proxy_fingerprint: '', proxy_allow_insecure: false,
        proxy_host: '', proxy_path: '', proxy_seed: '', proxy_flow: '',
        proxy_public_key: '', proxy_short_id: '', proxy_sider_sni: '',
        proxy_dialer_proxy: '', raw_link: '', test_result: '', last_test_time: '',
        is_active: true, sort_order: 10, core_type: '', group_uuid: '',
      }
      useStore.getState().setActiveProfile(profile)
      expect(useStore.getState().activeProfile?.name).toBe('Active')
    })

    it('should clear active profile', () => {
      useStore.getState().setActiveProfile(null)
      expect(useStore.getState().activeProfile).toBeNull()
    })
  })

  // ==================== Metrics ====================

  describe('setMetrics', () => {
    it('should merge metrics', () => {
      useStore.getState().setMetrics({ download_speed: 100 })
      expect(useStore.getState().metrics.download_speed).toBe(100)
      expect(useStore.getState().metrics.upload_speed).toBe(0)
    })
  })

  // ==================== Core Versions ====================

  describe('setCoreVersions', () => {
    it('should set versions', () => {
      useStore.getState().setCoreVersions({ xray: 'v1.8.0' })
      expect(useStore.getState().coreVersions.xray).toBe('v1.8.0')
    })
  })

  // ==================== Download Progress ====================

  describe('download progress', () => {
    it('should set and clear download progress', () => {
      useStore.getState().setDownloadProgress('xray', {
        core_name: 'xray', downloaded: 50, total: 100, percentage: 50, status: 'downloading',
      })
      expect(useStore.getState().downloadProgress.xray?.percentage).toBe(50)

      useStore.getState().clearDownloadProgress('xray')
      expect(useStore.getState().downloadProgress.xray).toBeUndefined()
    })
  })

  // ==================== Toasts ====================

  describe('toasts', () => {
    it('should add and remove toasts', () => {
      useStore.getState().addToast('Hello', 'success')
      expect(useStore.getState().toasts).toHaveLength(1)
      expect(useStore.getState().toasts[0].message).toBe('Hello')
      expect(useStore.getState().toasts[0].type).toBe('success')

      const id = useStore.getState().toasts[0].id
      useStore.getState().removeToast(id)
      expect(useStore.getState().toasts).toHaveLength(0)
    })
  })

  // ==================== Logs ====================

  describe('logs', () => {
    it('should add log entry', () => {
      useStore.getState().addLog({ time: '12:00', level: 'info', content: 'test', source: 'system' })
      expect(useStore.getState().logs).toHaveLength(1)
      expect(useStore.getState().logs[0].content).toBe('test')
    })

    it('should clear logs', () => {
      useStore.getState().addLog({ time: '12:00', level: 'info', content: 'test', source: 'system' })
      useStore.getState().clearLogs()
      expect(useStore.getState().logs).toHaveLength(0)
    })

    it('should cap logs at 500 entries', () => {
      for (let i = 0; i < 510; i++) {
        useStore.getState().addLog({ time: '12:00', level: 'info', content: `log-${i}`, source: 'system' })
      }
      expect(useStore.getState().logs.length).toBeLessThanOrEqual(500)
    })
  })
})