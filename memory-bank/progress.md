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

### Frontend (React)
- **Web UI 全功能实现**
- 节点管理（CRUD + 批量导入 + 去重 + 测速）
- 分组管理（订阅分组 + 普通分组）
- 路由规则管理、策略组管理
- 内核管理（下载、启动/停止、版本检测）
- 实时日志 + 流量监控（WebSocket）
- 国际化（中/英文）+ 主题切换

### Testing
- **Go 后端测试**: parser(~70), database(~35), service(~50), config(~6), web(~18) — 全部通过
- **前端测试**: coreMap(25), store(16), i18n(2) — 43 个测试全部通过
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

1. **2026-05**: 完成项目测试编写
   - Go 后端：parser → database → service → config → web handler 逐层覆盖
   - 前端：搭建 Vitest 测试框架，覆盖纯函数 + store
   - CI：创建独立 test.yml + 在 build 工作流中添加测试门控