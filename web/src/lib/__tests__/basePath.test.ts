import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'

/**
 * basePath 模块测试
 *
 * 测试范围：lib/basePath.ts
 *   - 生产环境：window.__BASE_PATH__ 为实际路径（如 "/my-secret"）
 *   - 本地开发：window.__BASE_PATH__ 为字面量 '{{ .BasePath }}'
 *   - 空值保护：window.__BASE_PATH__ 未设置或为空字符串
 *
 * 使用 Vitest 原生 vi.stubGlobal / vi.unstubAllGlobals：
 *   - stubGlobal(name, value) 模拟全局变量，等价于 window.name = value
 *   - unstubAllGlobals() 自动恢复所有被 stub 的全局变量，无需手动备份还原
 *   - 结合 vi.resetModules() 确保每次 import 都重新执行模块代码
 */

describe('basePath', () => {
  beforeEach(() => {
    // 清除模块缓存，确保每次 import 都是全新的模块执行
    // 因为 import 的变量是模块级常量，不重置的话首次 import 后就缓存了
    vi.resetModules()
  })

  afterEach(() => {
    // 自动恢复所有被 vi.stubGlobal 修改的全局变量
    // 无需手动备份原始值（如 const ORIGINAL = window.__BASE_PATH__）
    vi.unstubAllGlobals()
  })

  it('应在本地开发模式（Go 模板字面量）时返回空字符串', async () => {
    // 模拟 Go html/template 未渲染时的占位符 Vite 不解析 Go 模板
    vi.stubGlobal('__BASE_PATH__', '{{ .BasePath }}')
    const { basePath } = await import('../basePath')
    expect(basePath).toBe('')
  })

  it('应在生产环境返回实际的 custom_base_path 值', async () => {
    // 模拟 Go 后端渲染后的实际路径（如 "/my-secret"）
    vi.stubGlobal('__BASE_PATH__', '/my-secret')
    const { basePath } = await import('../basePath')
    expect(basePath).toBe('/my-secret')
  })

  it('应在 __BASE_PATH__ 未定义时返回空字符串（容错保护）', async () => {
    // 不调用 stubGlobal，模拟 window.__BASE_PATH__ 完全未定义的情况
    const { basePath } = await import('../basePath')
    expect(basePath).toBe('')
  })

  it('应在 __BASE_PATH__ 为空字符串时返回空字符串', async () => {
    // 模拟未设置 custom_base_path 时后端注入的空字符串
    vi.stubGlobal('__BASE_PATH__', '')
    const { basePath } = await import('../basePath')
    expect(basePath).toBe('')
  })
})