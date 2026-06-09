import { useState, useEffect, useCallback } from 'react'

/**
 * useBlurSave — 通用失焦自动保存 Hook
 *
 * 实现 draft / committed 双值模式：
 *   - draft:    用户正在编辑的值（显示在输入框中）
 *   - committed: 最后一次成功提交到后端的值（用于回滚）
 *
 * 行为：
 *   - 失焦时比较 draft 与 committed，值变化时才发送请求
 *   - 请求成功 → 更新 committed
 *   - 请求失败 → 将 draft 回滚到 committed（用户体验：输入框值自动还原）
 *   - 初始值（initialValue）变化时同步更新 draft 和 committed
 *
 * 竞态防护：
 *   - 当用户正在编辑（draft !== committed）时，外部 initialValue 的变化不会打断用户输入
 *   - 只有当用户未在编辑时（draft === committed），才同步 initialValue
 *
 * 使用示例：
 *   const { draft, setDraft, saving, handleBlur } = useBlurSave(
 *     'v2rayN-Go',
 *     async (val) => { await api.save({ issuer: val }) },
 *     { validate: (v) => v.length <= 50 }
 *   )
 */
export function useBlurSave<T>(
  initialValue: T,
  save: (value: T) => Promise<void>,
  opts?: { validate?: (v: T) => boolean }
) {
  const [draft, setDraft] = useState<T>(initialValue)
  const [committed, setCommitted] = useState<T>(initialValue)
  const [saving, setSaving] = useState(false)

  // 初始值变化时（如从 API 异步加载完成）同步 draft
  // 竞态防护：仅当用户未在编辑时才覆盖，避免打断用户输入
  useEffect(() => {
    if (draft === committed) {
      // 用户未在编辑，安全同步外部值
      setDraft(initialValue)
      setCommitted(initialValue)
    }
    // 如果 draft !== committed，说明用户正在编辑，不覆盖
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialValue])

  const handleBlur = useCallback(async () => {
    // 值没变，跳过保存
    if (draft === committed) return

    // 前端校验：不通过时回滚到 committed，不发请求
    if (opts?.validate && !opts.validate(draft)) {
      setDraft(committed)
      return
    }

    setSaving(true)
    try {
      await save(draft)
      setCommitted(draft) // 成功：更新 committed
    } catch {
      setDraft(committed) // 失败：回滚 draft 到 committed
    } finally {
      setSaving(false)
    }
  }, [draft, committed, save, opts])

  return { draft, setDraft, saving, handleBlur, committed }
}