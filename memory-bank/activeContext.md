# Active Context

## Current Work Focus
安全改造计划已全部实施完毕，后端+前端+CLI 全部编译通过。

## Recent Changes

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
- `web/src/components/SettingsView.tsx`：新增服务器设置卡片（HTTPS Toggle + basePath Input + 重启提示 Toast）

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
- 自签名证书使用 ECDSA P-256，复用 config.AtomicWriteFile 断电安全写入
- withBasePath 动态路由前缀包装，根路径重定向到 basePath
- **能力矩阵**：后端一次性下发所有协议的可用内核矩阵，前端字典查询
- **Mihomo YAML**：基础字段强类型 + `Extra map[string]any` + `yaml:",inline"`