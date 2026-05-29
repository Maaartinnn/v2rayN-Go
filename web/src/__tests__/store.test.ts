import { describe, it, expect, beforeEach } from 'vitest'
import { useStore, type ProfileListItem, type CoreStatus } from '../store'

describe('useStore', () => {
  beforeEach(() => {
    // Reset store to initial state
    useStore.setState({
      isConnected: false,
      coreStatuses: [],
      profileList: [],
      activeProfileUUID: null,
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

  // ==================== Profile List ====================

  describe('setProfileList', () => {
    it('should set profileList array directly', () => {
      const items: ProfileListItem[] = [
        {
           uuid: 'test-uuid', name: 'TestNode',
           proxy_protocol: 'vless', address: 'host.com:443',
           core_type: '', test_result: '', is_active: false, group_uuid: 'group-1',
          protocol_color: { bg: '#fff', text: '#000' },
          core_color: { bg: '#eee', text: '#333' },
          latency_color: 'var(--color-error)',
        },
      ]
      useStore.getState().setProfileList(items)
      expect(useStore.getState().profileList).toHaveLength(1)
      expect(useStore.getState().profileList[0].name).toBe('TestNode')
    })

    it('should accept function updater', () => {
      useStore.getState().setProfileList([])
      useStore.getState().setProfileList((prev) => [
        ...prev,
        {
           uuid: 'test', name: 'New',
           proxy_protocol: 'vless', address: '0.0.0.0:0',
           core_type: '', test_result: '', is_active: false, group_uuid: '',
          protocol_color: { bg: '#fff', text: '#000' },
          core_color: { bg: '#eee', text: '#333' },
          latency_color: 'var(--color-error)',
        },
      ])
      expect(useStore.getState().profileList).toHaveLength(1)
    })
  })

  describe('setActiveProfileUUID', () => {
    it('should set active profile UUID', () => {
      useStore.getState().setActiveProfileUUID('test-uuid')
      expect(useStore.getState().activeProfileUUID).toBe('test-uuid')
    })

    it('should clear active profile UUID', () => {
      useStore.getState().setActiveProfileUUID('test-uuid')
      useStore.getState().setActiveProfileUUID(null)
      expect(useStore.getState().activeProfileUUID).toBeNull()
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
    it('should add toast with type and return id', () => {
      const id = useStore.getState().addToast('Hello', 'success')
      expect(useStore.getState().toasts).toHaveLength(1)
      expect(useStore.getState().toasts[0].message).toBe('Hello')
      expect(useStore.getState().toasts[0].type).toBe('success')
      expect(useStore.getState().toasts[0].id).toBe(id)
    })

    it('should default to info type when no type provided', () => {
      useStore.getState().addToast('Default')
      expect(useStore.getState().toasts[0].type).toBe('info')
    })

    it('should add toast with custom color', () => {
      useStore.getState().addToast('Custom', 'info', {
        color: { bg: '#f0f', text: '#000' },
      })
      expect(useStore.getState().toasts[0].color).toEqual({ bg: '#f0f', text: '#000' })
    })

    it('should add toast with action', () => {
      const onClick = () => {}
      useStore.getState().addToast('With action', 'warning', {
        action: { label: 'Update', onClick },
      })
      expect(useStore.getState().toasts[0].action?.label).toBe('Update')
      expect(useStore.getState().toasts[0].action?.onClick).toBe(onClick)
    })

    it('should add toast with duration', () => {
      useStore.getState().addToast('Timed', 'success', { duration: 2000 })
      expect(useStore.getState().toasts[0].duration).toBe(2000)
    })

    it('should not auto-remove (no setTimeout in store)', () => {
      useStore.getState().addToast('Persistent', 'error')
      // Toast should still be present after a tick (no auto-removal in store)
      expect(useStore.getState().toasts).toHaveLength(1)
    })

    it('should remove toast by id', () => {
      const id = useStore.getState().addToast('Remove me')
      expect(useStore.getState().toasts).toHaveLength(1)
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