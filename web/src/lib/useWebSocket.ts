import { useEffect, useRef, useCallback } from 'react'
import { useStore } from '../store'

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const { addLog, setMetrics, setCoreStatuses, setDownloadProgress, clearDownloadProgress, addToast, setCoreVersions } = useStore()

  const connect = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/ws`

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      console.log('WebSocket connected')
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        switch (data.type) {
          case 'log':
            addLog(data.payload)
            break
          case 'metrics':
            setMetrics(data.payload)
            break
          case 'status':
            setCoreStatuses(data.payload)
            break
          case 'download_progress':
            setDownloadProgress(data.payload.core_name, data.payload)
            break
          case 'download_complete':
            if (data.payload.success) {
              addToast(`${data.payload.core_name} 下载完成`, 'success')
            } else {
              addToast(`${data.payload.core_name} 下载失败: ${data.payload.error}`, 'error')
            }
            clearDownloadProgress(data.payload.core_name)
            break
          case 'core_versions':
            setCoreVersions(data.payload)
            break
        }
      } catch {
        // ignore parse errors
      }
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected, reconnecting...')
      reconnectTimer.current = setTimeout(connect, 3000)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [addLog, setMetrics, setCoreStatuses, setDownloadProgress, clearDownloadProgress, addToast, setCoreVersions])

  useEffect(() => {
    connect()
    return () => {
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  return wsRef
}