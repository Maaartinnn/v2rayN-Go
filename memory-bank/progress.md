# Progress

## What Works

### Backend (Go)
- **全部核心功能实现并测试通过**
- 10 种协议解析器（vmess/vless/trojan/ss/ssr/hysteria/hysteria2/tuic/wireguard/anytls）
- 数据库 CRUD（Profile/NodeGroup/RoutingRule/StrategyGroup/AppSetting）
- 排序系统（SortBetween/SortInsert/Rebalance/ReorderEntity，含整数溢出保护）
- Service 层（ProfileService/GroupService/RoutingRuleService/StrategyGroupService/PingService）
- Web handler 层（REST API + WebSocket + 错误映射）
- 配置管理（AppConfig + JSON 持久化 + CLI 参数优先级）
- 内核管理（CoreService + CoreAdminManager）
- **节点列表精简传输**：`ListSummary` 返回 `ProfileListItem` DTO + 后端计算颜色

### Frontend (React)
- **Web UI 全功能实现**
- 节点管理（CRUD + 批量导入 + 去重 + 测速 + 编辑按需加载完整数据）
- 分组管理（订阅分组 + 普通分组）
- 路由规则管理、策略组管理
- 内核管理（下载、启动/停止、版本检测）
- 实时日志 + 流量监控（WebSocket）
- 国际化（中/英文）+ 主题切换
- **通用 Toast 通知系统**（自定义颜色/操作按钮/可选自动消失/响应式/无障碍）
- **节点列表精简数据**：使用 `ProfileListItem` + 后端驱动颜色 + uuid 统一标识

### Testing
- **Go 后端测试**: parser(~70), database(~35), service(~50), config(~6), web(~18) — 全部通过
- **前端测试**: coreMap(25), store(22), i18n(2) — 49 个测试全部通过
- **CI/CD**: 三个 GitHub Actions 工作流已配置

## What's Left to Build

- 短期内无已知功能缺口
- 可扩展方向：
  - 更多前端组件测试（Dialog、Table、DnD 等交互组件）
  - E2E 测试
  - 性能/负载测试

## Known Issues

- SQLite 并发写入需依赖 busy_timeout
- 排序整数溢出时触发 Rebalance（设计如此）

## Evolution of Project Decisions

1. **2026-05-29**: 节点列表精简传输 + 通用 Toast 通知系统
   - 后端新增 `ProfileListItem` DTO + 颜色映射纯函数（不入库）
   - 前端统一用 `uuid` 作为唯一标识，移除对 `ID` 的依赖
   - 通用 Toast 组件：Store 只负责数据，定时器生命周期由组件管理
   - 编辑节点时按需请求完整 Profile（`GET /api/profiles/{uuid}`）

2. **2026-05-28**: 完成项目测试编写 + 后端代码质量提升
   - Go 后端：parser → database → service → config → web handler 逐层覆盖
   - 前端：搭建 Vitest 测试框架，覆盖纯函数 + store
   - CI：创建独立 test.yml + 在 build 工作流中添加测试门控
   - 常量提取、日志统一（log → slog）、错误链审计