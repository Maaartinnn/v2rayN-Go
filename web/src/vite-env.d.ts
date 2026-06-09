/// <reference types="vite/client" />

/**
 * Go 后端在运行时注入的 custom_base_path 值。
 * - 生产环境：Go 替换占位符后注入实际路径（如 "/my-secret"）
 * - 本地开发：值为字面量 '__INJECT_BASE_PATH__'，前端代码会将其视为空字符串
 */
interface Window {
  __BASE_PATH__: string;
}