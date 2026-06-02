import { useEffect, useRef, useState } from 'react'
import jsQR from 'jsqr'
import { useT } from '../../lib/i18n'
import { useStore } from '../../store'

/**
 * QrScanner — 二维码图片解析组件
 *
 * 职责单一：接收用户上传的图片 File，在浏览器端解码二维码并返回链接数组。
 * 通过 React.lazy() 按需加载，避免主包体积膨胀。
 *
 * 安全特性：
 * - 图片完全在浏览器内存中处理，零网络传输
 * - 大图自动等比缩放（最大 1000px），防止 OOM 和 UI 卡顿
 */

/** Canvas 最大边长，超过此尺寸的图片将等比缩放 */
const CANVAS_MAX_SIZE = 1000

interface QrScannerProps {
  /** 用户选择的图片文件 */
  file: File
  /** 解码成功回调，返回解析到的链接数组 */
  onResult: (links: string[]) => void
  /** 解码失败回调（可选，组件内部已集成 toast 通知） */
  onError?: () => void
}

export default function QrScanner({ file, onResult, onError }: QrScannerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [status, setStatus] = useState<'decoding' | 'done'>('decoding')
  const t = useT()
  const { addToast } = useStore()

  // 使用 ref 防止 React 18 StrictMode 下的重复执行
  const processedRef = useRef(false)

  useEffect(() => {
    // StrictMode 双重执行保护
    if (processedRef.current) return
    processedRef.current = true

    const img = new Image()
    const url = URL.createObjectURL(file)

    img.onload = () => {
      try {
        const canvas = canvasRef.current!
        const ctx = canvas.getContext('2d', { willReadFrequently: true })!

        // 大图等比缩放：二维码识别不需要高分辨率，
        // 缩放后大幅降低内存占用，jsQR 解码速度也成倍提升
        let width = img.width
        let height = img.height

        if (width > CANVAS_MAX_SIZE || height > CANVAS_MAX_SIZE) {
          const ratio = Math.min(CANVAS_MAX_SIZE / width, CANVAS_MAX_SIZE / height)
          width = Math.round(width * ratio)
          height = Math.round(height * ratio)
        }

        canvas.width = width
        canvas.height = height
        ctx.drawImage(img, 0, 0, width, height)

        const imageData = ctx.getImageData(0, 0, width, height)

        // jsQR 解码
        const code = jsQR(imageData.data, imageData.width, imageData.height)
        URL.revokeObjectURL(url)

        if (!code) {
          addToast(t('qr.no_qr_found'), 'error', { duration: 5000 })
          onError?.()
          return
        }

        const text = code.data.trim()
        if (!text) {
          addToast(t('qr.empty_qr'), 'error', { duration: 5000 })
          onError?.()
          return
        }

        // 按换行分割（与原后端逻辑一致，单个二维码可能包含多行链接）
        const links = text.split('\n').map(l => l.trim()).filter(Boolean)
        if (links.length === 0) {
          addToast(t('qr.no_valid_links'), 'error', { duration: 5000 })
          onError?.()
          return
        }

        setStatus('done')
        onResult(links)
      } catch {
        URL.revokeObjectURL(url)
        addToast(t('qr.decode_failed'), 'error', { duration: 5000 })
        onError?.()
      }
    }

    img.onerror = () => {
      URL.revokeObjectURL(url)
      addToast(t('qr.load_failed'), 'error', { duration: 5000 })
      onError?.()
    }

    img.src = url

    return () => {
      URL.revokeObjectURL(url)
    }
  }, [file, onResult, onError, t, addToast])

  return (
    <div className="flex items-center gap-2 text-xs" style={{ color: 'var(--color-muted-foreground)' }}>
      {/* 隐藏的 canvas 用于图像像素数据提取 */}
      <canvas ref={canvasRef} className="hidden" />
      {status === 'decoding' && <span>{t('qr.decoding')}</span>}
    </div>
  )
}