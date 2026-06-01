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
- Go 后端全部测试通过（7 packages）
- 前端 TypeScript 编译无错误

## What's Left To Build
- 策略组重构：StrategyGroup 表 → Profile 虚拟节点（组合模式）
- 扩展前端组件测试

## Current Status
- 断电安全防护 + 无文件落地 + Mihomo ConfigBuilder + 协议→内核能力矩阵 已完成
- Go 全量测试 7/7 PASS，TypeScript 编译无错误
- 项目稳定运行

## Known Issues
- 策略组功能需要重构为 Profile 虚拟节点模式
- Mihomo 配置生成器暂不支持 WireGuard 协议

## Decisions
- Go 1.26 + new(expr) 语法（Mihomo 指针字面量）
- Mihomo YAML 结构体：基础字段强类型 + Extra map + yaml:",inline"（平衡覆盖度与可维护性）
- 协议→内核映射表位于 coredef/protocol_cores.go（唯一事实来源）
- API 职责单一：GET /api/profiles/{uuid} 返回原始 profile，GET /api/profiles/core-matrix 返回能力矩阵
- 内核配置调试输出统一到 binConfig/ 目录
- Mihomo stdin 模式使用 `-d . -f -`（工作目录为内核二进制目录）