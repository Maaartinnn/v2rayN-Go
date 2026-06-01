# Active Context

## Current Work Focus
协议内核智能选择 + 能力矩阵 + Mihomo ConfigBuilder 已完成，可继续策略组重构或测试扩展。

## Recent Changes

### 1. 内核配置调试输出统一到 binConfig (2026-06-01)
- SaveXrayConfig / SaveSingboxConfig / SaveMihomoConfig 输出路径统一到 `{AppDir}/binConfig/`

### 2. Mihomo ConfigBuilder 实现 (2026-06-01)
- 新建 `configbuilder/mihomo.go`：完整 YAML 配置结构体（MihomoConfig / MihomoProxy / MihomoProxyGroup）
- 基础字段强类型定义 + `Extra map[string]any` + `yaml:",inline"` 处理协议专属参数
- 利用 Go 1.26 `new(expr)` 语法创建 `*bool` 指针字面量
- 支持 8 种协议：vmess, vless, trojan, ss, hysteria2, tuic, socks5, http
- TLS/Reality/传输层(ws/h2/grpc/tcp) 全部覆盖
- `configbuilder/mihomo_builder.go`：实现 ConfigBuilder 接口，`BuildBytes()` 返回 YAML 字节

### 3. 内核选择后端化 (2026-06-01)
- 新建 `coredef/protocol_cores.go`：`ProtocolCoreMap` 协议→内核兼容映射表（唯一事实来源）
- `CoreService.GetCompatibleInstalledCores()`：交叉查询映射表 + 本地 bin/ 目录
- `CoreService.GetInstalledCoreMatrix()`：一次性计算所有协议的能力矩阵
- `CoreService.Start()` 复用 `GetCompatibleInstalledCores` 选择最佳默认内核（替换硬编码 xray）
- `CoreService.Stop()` 核心类型为空时调用 `StopAll()`

### 4. API 能力矩阵 + 前端 NodeEditForm 重构 (2026-06-01)
- 新增 `GET /api/profiles/core-matrix` 端点：返回完整能力矩阵 `{"vmess": ["xray", ...]}`
- `GET /api/profiles/{uuid}` 保持返回原始 profile（职责单一）
- `ProfileHandler` 注入 `CoreService`，`NewProfileHandler` 新增 `coreSvc` 参数
- 前端 NodeEditForm：删除 `coresApi.list()` / `coreMap.ts` 过滤逻辑，改用后端能力矩阵字典查询
- 协议切换时零延迟查字典，无需网络请求

### 5. HomeView Bug 修复 (2026-06-01)
- 修复前端硬编码 `'xray'` 问题：HomeView 传空字符串，后端自动判断内核

### 6. 无文件落地 Fileless Execution (2026-05-31)
- stdin 模式 + Functional Options + BuildBytes + 跨平台进程安全

### 7. 断电安全防护改造 (2026-05-31)
- AtomicWriteFile + .bak backup rollback + SQLite WAL mode

## Next Steps
- 策略组重构：StrategyGroup 表 → Profile 虚拟节点（组合模式）
- 扩展测试覆盖
- coreMap.ts 清理（PROTOCOLS / NETWORKS / TLS_OPTIONS / SECURITY_METHODS 仍被 NodeEditForm 使用）

## Important Patterns
- **能力矩阵**：后端一次性下发所有协议的可用内核矩阵，前端字典查询
- **协议→内核映射**：`coredef.ProtocolCoreMap` 是唯一事实来源
- **Mihomo YAML**：基础字段强类型 + `Extra map[string]any` + `yaml:",inline"` 处理可选协议参数
- **API 职责单一**：`GET /api/profiles/{uuid}` 返回原始 profile，`GET /api/profiles/core-matrix` 返回能力矩阵
- Data transfer optimization (DTO), Toast notification, testing, logging, constants patterns