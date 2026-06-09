/**
 * basePath — custom_base_path 的前端感知模块
 *
 * Go 后端通过 html/template 在启动时将 custom_base_path 注入到 index.html 中：
 *   window.__BASE_PATH__ = '{{ .BasePath }}';
 *
 * 生产环境下 Go 模板引擎会将其渲染为实际值（如 "/my-secret"），
 * 本地开发（npm run dev）时 Vite 不解析 Go 模板，值保持字面量 '{{ .BasePath }}'。
 *
 * 所有需要感知 base path 的模块统一从这里导入，避免重复判断。
 */
export const basePath =
  window.__BASE_PATH__ === '{{ .BasePath }}' ? '' : (window.__BASE_PATH__ || '')