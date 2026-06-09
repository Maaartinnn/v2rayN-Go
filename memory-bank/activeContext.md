# Active Context

## Current Work Focus
custom_base_path 前端感知修复完成：html/template 注入 + 哈希路由 + 3xx 重定向前缀补回 + 测试用例。

## Recent Changes

### custom_base_path 前端感知修复（2026-06-09）

#### 问题根因
设置 custom_base_path 后，前端无法正常访问页面。API 请求发到 `/api/...`（无前缀），被 `withBasePath` 返回 404。
根因是前端所有代码硬编码了无前缀路径，不知道 custom_base_path 的存在。

#### 修复方案：运行时注入 + 哈希路由
- **`web/index.html`**：占位符改为 Go 模板语法 `{{ .BasePath }}`
- **`src/web/server.go`**：`html/template` + `Option("missingkey=error")` 安全注入，启动时一次性渲染并缓存
- **`web/src/lib/basePath.ts`**（新建）：公共模块，统一判断 `{{ .BasePath }}` 占位符（本地 dev 兼容）
- **`web/src/lib/api.ts`**：axios baseURL 改为 `${basePath}/api`
- **`web/src/lib/useWebSocket.ts`**：WebSocket URL 拼接 basePath
- **`web/src/components/AccountView.tsx`**：退出登录重定向带 basePath
- **`web/src/App.tsx`**：wouter 切换为哈希路由模式（`useHashLocation`），消除 SPA 路由匹配问题
- **`web/vite.config.ts`**：顶级 `base: './'`，Vite 构建输出相对路径
- **`web/src/vite-env.d.ts`**：`Window.__BASE_PATH__` 类型声明 + Go 模板注释

#### Go 1.22+ 精确路由 + redirectWriter
- **精确路由 `{$}`**：`mux.HandleFunc("GET /{$}", indexHandler)` + `mux.HandleFunc("GET /my-secret/{$}", indexHandler)`
- **`redirectWriter`**：包装 ResponseWriter，拦截所有 3xx 重定向，对相对路径（以 `/` 开头）补回前缀，绝对 URL 不动

#### 测试用例（`handler_test.go`）
- `TestWithBasePath_EmptyBasePath`：空 basePath 直接透传
- `TestWithBasePath_StripsPrefix`：前缀剥离（API/根路径/精确匹配/404）
- `TestWithBasePath_RedirectRoot`：根路径 302 重定向到 /my-secret/
- `TestRedirectWriter_PrefixesRedirectLocation`：相对路径补前缀 + 绝对 URL 不动
- `TestRedirectWriter_3xxRange`：所有 3xx 状态码都被拦截

#### 文档更新
- README.md 新增 🛰️ 动态网络防御特性描述

### 安全改造计划（2026-06-02）

#### 阶段一：底层拓荒与数据模型
- `database/models.go`：新增 User 模型（UUID/Username/PasswordHash/JWTSecret/TOTPSecret/TOTPEnabled/Role）
- `database/db.go`：AutoMigrate 挂载 User + SeedDefaults()（app_settings 默认 KV + admin 用户高亮密码打印）
- `cmd/cli.go`：admin 子命令（`admin set <pwd>` / `admin random`），在 flag.Parse() 前拦截
- `main.go`：admin 命令前置拦截 + 非 admin 路径的 Init 移到 sysmgr 中调用
- `sysmgr/os_service.go`：RunDirect 和 App.run 中加入 SeedDefaults 调用

#### 阶段二：后端鉴权与 API 防线
- `service/auth_service.go`：AuthService（Login/JWT 签发验证/ChangePassword/RotateJWTSecret/EnableTOTP/VerifyAndActivateTOTP/DisableTOTP）
- `web/handler_auth.go`：AuthHandler（8 个 API 端点：login/me/change-password/totp/enable|verify|disable/sessions/revoke-all）
- `web/middleware_auth.go`：JWT 认证中间件 + 白名单放行
- `web/server.go`：AuthHandler 注册 + AuthMiddleware 包装 + withBasePath + getSettingFromDB + force_https HTTPS 启动
- `web/cert.go`：ECDSA P-256 自签名证书自动生成，10 年有效期，SAN 覆盖 localhost/127.0.0.1/::1，复用 config.AtomicWriteFile

#### 阶段三：前端登录与拦截
- `web/src/lib/api.ts`：Axios 请求拦截器（自动注入 Token）+ 响应拦截器（401 跳转）+ authApi
- `web/src/components/LoginView.tsx`：极简居中卡片登录页（用户名+密码+TOTP 动态码）
- `web/src/components/AccountView.tsx`：修改密码卡片 + TOTP 两步验证卡片（QR 码渲染 qrcode.react）+ 会话管理卡片
- `web/src/App.tsx`：路由守卫（authState 三态：loading/authenticated/unauthenticated）+ AuthenticatedApp 子组件
- `web/src/components/Sidebar.tsx`：底部新增账户入口（UserCircle 图标 + /account 路由）
- `web/src/locales/zh-CN.ts` + `en-US.ts`：新增 auth/account/settings.server 相关键值

#### 阶段四：动态网络纵深防御
- `web/src/components/SettingsView.tsx`：新增服务器设置卡片（HTTPS Toggle + basePath Input + JWT 过期时间 + 重启提示 Toast）

### QR 识别前端化迁移（2026-06-02）
- **问题**：`/api/profiles/import-image` 存在 SSRF 死代码（URL 下载）、无文件大小限制、无 MIME 校验、先全量读入再验证等安全问题
- **方案**：彻底删除后端 QR 识别能力，迁移到前端浏览器端完成
- **后端清理**：
  - `handler_profile.go`：删除 `handleImportImage` + `importParsedLinks` + 路由注册，清理 `io`/`parser` 导入
  - `profile_service.go`：删除无调用方的 `ImportParsedLinks` 方法
  - `parser/qrcode.go`：整个文件删除（gozxing + image.Decode）
  - `go.mod`/`go.sum`：`go mod tidy` 移除 `gozxing` 依赖
- **前端新增/修改**：
  - 安装 `jsqr`（纯 JS QR 解码库，~50KB）
  - `components/tools/QrScanner.tsx`：独立 QR 解码组件，通过 `React.lazy()` 按需加载，集成 `useT()` i18n + `addToast()` 通知，大图自动等比缩放（≤1000px）防 OOM，StrictMode 双重执行保护
  - `ImportView.tsx`：图片选择后设置 `qrFile` 状态，渲染 `QrScanner` → 解码成功调用已有 `POST /api/profiles/import`
  - `api.ts`：删除 `profileEnhancedApi.importImage`
  - `locales/zh-CN.ts` + `en-US.ts`：新增 `qr.*` 翻译 key（6 条）
- **安全收益**：消除 SSRF、消除内存溢出风险、图片零网络传输、移除 Go 图片解码依赖

### 安全改造后续 Bug 修复（2026-06-02）
- **密码确认逻辑**：AccountView 三字段空值检查 + 新旧密码相同判断 + 密码长度独立错误消息
- **改密后无缝续用**：ChangePassword 返回新 JWT Token，前端更新 localStorage
- **TOTP 防攻击**：未开 TOTP 但输入动态码时拒绝登录
- **Toast 自动消失**：LoginView/SettingsView/AccountView 所有土司加 5 秒 duration
- **按钮样式统一**：LoginView + AccountView 改用 btn-primary/btn-secondary/btn-danger/btn-ghost CSS class
- **ToastContainer 全局挂载**：移至 App 顶层，确保登录页也能看到通知

### app_settings 表读写（2026-06-02）
- `SettingsService.UpdateSettings` 新增 `ForceHTTPS`、`CustomBasePath`、`JwtExpireHours` 字段
- GORM 原生 Upsert（`clause.OnConflict`）一条 SQL 完成插入或更新
- `GetSettings` 合并 config.json + app_settings 返回完整配置快照

### 路由前缀规范化（2026-06-02）
- 存储规范：纯路径名（无斜杠），空字符串表示无前缀
- 后端正则校验 `^[a-zA-Z0-9_-]+$`，拒绝含 `/` 的输入
- `withBasePath` 兼容无斜杠纯路径名，自动加 `/` 前缀比对
- 前端 `onBlur` 时 `trim()` 后直接保存，placeholder 改为 `my-path`

### JWT 过期时间可配置（2026-06-02）
- `app_settings` 存储 `jwt_expire_hours`，默认 24 小时
- SettingsView Server Section 新增 number Input，失焦保存
- 后端正整数校验 1-8760，空值回退默认 24

### 局部更新 + 失焦保存（2026-06-01）
- `settings_service.go`：dirty flag + 三步校验
- `SettingsView.tsx`：移除全局保存按钮，失焦自动保存

### 内核配置调试输出统一到 binConfig (2026-06-01)
### Mihomo ConfigBuilder 实现 (2026-06-01)
### 内核选择后端化 + API 能力矩阵 (2026-06-01)
### HomeView Bug 修复 (2026-06-01)
### 无文件落地 Fileless Execution (2026-05-31)
### 断电安全防护改造 (2026-05-31)

## Next Steps
- 策略组重构：StrategyGroup 表 → Profile 虚拟节点（组合模式）
- 扩展测试覆盖
- 移动端响应式布局优化

## Important Patterns
- JWT 使用用户专属 Secret（HS256），RotateJWTSecret 使旧 Token 失效
- TOTP 使用 pquerna/otp 库，默认时间窗口 ±30 秒
- CLI admin 命令在 flag.Parse() 之前拦截（避免 flag 冲突）
- Auth guard 拆分为 App（检测 token）+ AuthenticatedApp（业务逻辑）两层
- ToastContainer 必须在 App 顶层渲染，不能放在 AuthenticatedApp 内部
- 自签名证书使用 ECDSA P-256，复用 config.AtomicWriteFile 断电安全写入
- withBasePath 动态路由前缀包装，存储纯路径名（无斜杠），运行时自动加 `/`
- **html/template 注入**：`template.New().Option("missingkey=error").ParseFS()` 安全渲染，防止拼写错误
- **redirectWriter**：拦截 3xx 重定向补回前缀，只处理相对路径，绝对 URL 不动
- **哈希路由**：wouter `useHashLocation`，URL 格式 `/my-secret/#/nodes`，浏览器只请求根路径
- **前端公共模块**：`lib/basePath.ts` 统一导出 `basePath`，消除重复判断
- **app_settings Upsert**：GORM `clause.OnConflict` 原生 SQL，一条语句完成插入或更新
- **能力矩阵**：后端一次性下发所有协议的可用内核矩阵，前端字典查询
- **Mihomo YAML**：基础字段强类型 + `Extra map[string]any` + `yaml:",inline"`
- **按钮规范**：所有按钮使用 `.btn-primary`/`.btn-secondary`/`.btn-ghost`/`.btn-danger` CSS class
