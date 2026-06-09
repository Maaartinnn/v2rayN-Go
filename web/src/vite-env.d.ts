/// <reference types="vite/client" />

/**
 * Go 后端通过 html/template 在启动时注入的 custom_base_path 值。
 * - 生产环境：Go 模板引擎将 {{ .BasePath }} 渲染为实际路径（如 "/my-secret"）
 * - 本地开发（npm run dev）：Vite 不解析 Go 模板，值保持字面量 '{{ .BasePath }}'，
 *   前端代码（lib/basePath.ts）会将其视为空字符串
 */
interface Window {
  __BASE_PATH__: string;
}
