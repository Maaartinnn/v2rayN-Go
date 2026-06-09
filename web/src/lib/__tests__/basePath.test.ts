import { describe, it, expect, beforeEach, afterAll, vi } from 'vitest'

/**
 * basePath 模块测试
 *
 * 测试范围：lib/basePath.ts
 *   - 生产环境：window.__BASE_PATH__ 为实际路径（如 "/my-secret"）
 *   - 本地开发：window.__BASE_PATH__ 为字面量 '{{ .BasePath }}'
 *   - 空值保护：window.__BASE_PATH__ 未设置或为空字符串
 */

// 备份原始值，避免测试间互相污染
const ORIGINAL = window.__BASE_PATH__

beforeEach(() => {
  // 每个测试前删除 __BASE_PATH__，并清除模块缓存
  delete (window as any).__BASE_PATH__
  vi.resetModules()
})

afterAll(() => {
  // 恢复原始值
  ;(window as any).__BASE_PATH__ = ORIGINAL
})

describe('basePath', () => {
  it('should return empty string when __BASE_PATH__ is the Go template literal (dev mode)', async () => {
    ;(window as any).__BASE_PATH__ = '{{ .BasePath }}'
    // 使用 vi.resetModules() 确保每次动态 import 都重新执行模块代码
    const { basePath } = await import('../basePath')
    expect(basePath).toBe('')
  })

  it('should return the actual path when __BASE_PATH__ is set (production)', async () => {
    ;(window as any).__BASE_PATH__ = '/my-secret'
    const { basePath } = await import('../basePath')
    expect(basePath).toBe('/my-secret')
  })

  it('should return empty string when __BASE_PATH__ is not set', async () => {
    const { basePath } = await import('../basePath')
    expect(basePath).toBe('')
  })

  it('should return empty string when __BASE_PATH__ is empty string', async () => {
    ;(window as any).__BASE_PATH__ = ''
    const { basePath } = await import('../basePath')
    expect(basePath).toBe('')
  })
})
