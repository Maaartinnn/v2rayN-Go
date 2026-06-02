# Progress

## What Works
All core features are implemented and tested.

## What's Built
- Web UI (React 19 + TypeScript + TailwindCSS 4 + Framer Motion)
- 前端暗色/亮色主题、中英双语、Toast 通知
- 节点管理：列表/新增/编辑/拖拽排序/删除/批量测速/拖拽导入
- 节点列表精简传输（ProfileListItem DTO）+ 颜色后端驱动
- 前端 uuid 标识 + HomeView/App.tsx 从 profileList 查找激活节点
- 节点编辑表单内核选择：后端能力矩阵（core-matrix）一次性下发，协议切换零延迟
- 策略组管理（创建/编辑/排序）— 基础功能
- 通用 Toast 通知系统（Store + 组件分离）
- 订阅管理：解析、分组、多协议支持
- 测速：单节点/批量（并发 20），延迟结果缓存
- 路由规则管理（CRUD/排序）
- 设置管理（端口/监听 IP/GitHub 镜像/核心配置调试开关）
- 系统代理管理（Windows/macOS）
- 内核管理：上传/下载/版本检测（异步更新进度条）
- **三内核配置构建器**：Xray (JSON) + Sing-box (JSON) + Mihomo (YAML)
- **Mihomo ConfigBuilder**：完整 YAML 生成，支持 8 种协议 + TLS/Reality/传输层
- **无文件落地（Fileless Execution）**：stdin 模式 + Functional Options + 跨平台进程安全
- **断电安全防护**：AtomicWriteFile + .bak 容灾 + SQLite WAL
- **协议→内核智能选择**：ProtocolCoreMap 映射表 + GetCompatibleInstalledCores + GetInstalledCoreMatrix
- **Mihomo 配置调试输出**：binConfig/mihomo_config.yaml
- **局部更新 + 失焦保存**：SettingsService dirty flag + 三步校验 + SettingsView Blur 自动保存
- **安全改造计划（2026-06-02）**：
  - User 模型（UUID/PasswordHash/JWTSecret/TOTP/Role）
  - SeedDefaults（app_settings 默认 KV + admin 高亮密码打印）
  - CLI admin 子命令（`admin set <pwd>` / `admin random`）
  - AuthService（Login/JWT/ChangePassword/RotateJWTSecret/TOTP 全套）
  - AuthHandler（8 个 API 端点）+ AuthMiddleware（JWT 认证 + 白名单）
  - 自签名证书（ECDSA P-256，10 年有效期，SAN 覆盖 localhost）
  - 动态路由前缀（withBasePath）+ force_https HTTPS 启动
  - LoginView 极简登录页 + AccountView 账户管理（改密/TOTP QR 码/会话管理）
  - App.tsx 路由守卫（Auth guard 三态 + AuthenticatedApp 子组件）
  - Sidebar 账户入口 + SettingsView 服务器设置卡片（HTTPS/basePath/JWT 过期时间）
  - Axios 拦截器（自动注入 Token + 401 跳转）+ authApi
  - i18n 中英文新增 auth/account/settings.server 键值
  - ToastContainer 移至 App 顶层（登录页也能看到通知）
  - 按钮样式统一（btn-primary/secondary/ghost/danger）
  - 密码确认逻辑修复 + 新旧密码相同判断 + 独立错误消息
  - ChangePassword 返回新 JWT（当前设备不退出）
  - TOTP 防攻击（未开 TOTP 但输入动态码拒绝登录）
  - Toast 自动消失（5 秒 duration）
  - app_settings 表 GORM Upsert 读写（ForceHTTPS/CustomBasePath/JwtExpireHours）
  - 路由前缀规范化（纯路径名无斜杠 + 正则校验）
  - JWT 过期时间前端可配置（1-8760 小时）
- Go 后端全部测试通过（7 packages）
- 前端 TypeScript 编译无错误

## What's Left To Build
- 策略组重构：StrategyGroup 表 → Profile 虚拟节点（组合模式）
- 扩展前端组件测试

## Current Status
- 安全改造计划四阶段全部完成，go vet + vite build 通过
- 项目稳定运行

## Known Issues
- 策略组功能需要重构为 Profile 虚拟节点模式
- Mihomo 配置生成器暂不支持 WireGuard 协议

## Decisions
- 设置页面失焦保存：前端 handleBlur 单字段 API，后端 dirty flag 避免无效 Sync()
- Go 1.26 + new(expr) 语法（Mihomo 指针字面量）
- Mihomo YAML 结构体：基础字段强类型 + Extra map + yaml:",inline"（平衡覆盖度与可维护性）
- 协议→内核映射表位于 coredef/protocol_cores.go（唯一事实来源）
- API 职责单一：GET /api/profiles/{uuid} 返回原始 profile，GET /api/profiles/core-matrix 返回能力矩阵
- 内核配置调试输出统一到 binConfig/ 目录
- Mihomo stdin 模式使用 `-d . -f -`（工作目录为内核二进制目录）
- JWT 使用用户专属 Secret（HS256），RotateJWTSecret 使旧 Token 失效
- TOTP 使用 pquerna/otp 库，默认时间窗口 ±30 秒
- CLI admin 命令在 flag.Parse() 之前拦截（避免 flag 冲突）
- Auth guard 拆分为 App（检测 token）+ AuthenticatedApp（业务逻辑）两层
- 自签名证书使用 ECDSA P-256，复用 config.AtomicWriteFile 断电安全写入
- withBasePath 动态路由前缀包装，根路径重定向到 basePath