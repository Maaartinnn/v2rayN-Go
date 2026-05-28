# Active Context

## Current Work Focus

刚完成项目测试编写和 CI 工作流集成。

## Recent Changes

### 测试编写（2026-05-28）

**Go 后端新增测试文件（19 个）：**
- `src/parser/parser_test.go` — ParseLink/ParseLinks 入口 + truncate
- `src/parser/utils_test.go` — base64Decode/parseIntSafe/extractName 等工具函数
- `src/parser/vmess_test.go` — VMess JSON/URI 解析
- `src/parser/vless_test.go` — VLESS 基础/Reality/WS/QUIC
- `src/parser/trojan_test.go` — Trojan 基础/Reality/allowInsecure
- `src/parser/shadowsocks_test.go` — SS URI/全base64/插件 + SSR
- `src/parser/hysteria_test.go` — Hysteria/Hysteria2/TUIC
- `src/parser/wireguard_test.go` — WireGuard 各参数
- `src/parser/anytls_test.go` — AnyTLS 解析
- `src/database/test_helper.go` — InitTestDB/CleanupTestDB 内存 SQLite
- `src/database/sort_utils_test.go` — SortBetween/SortSequence/safeAdd 等纯函数
- `src/database/db_test.go` — 五张表 CRUD + 排序 + UUID 唯一性
- `src/service/errors_test.go` — ErrNotFound/ErrValidation/ErrConflict
- `src/service/profile_service_test.go` — ProfileService 全方法
- `src/service/group_service_test.go` — GroupService 含级联删除
- `src/service/routing_rule_service_test.go` — RoutingRuleService 全方法
- `src/service/strategy_group_service_test.go` — StrategyGroupService 全方法
- `src/config/config_test.go` — AppConfig 默认值/序列化/omitEmpty
- `src/web/handler_test.go` — HTTP handler httptest + 错误映射

**前端新增文件（6 个）：**
- `web/vite.config.ts` — 添加 vitest 测试配置
- `web/src/test/setup.ts` — @testing-library/jest-dom setup
- `web/src/lib/__tests__/coreMap.test.ts` — 纯函数 + 常量测试
- `web/src/lib/__tests__/i18n.test.ts` — 国际化配置测试
- `web/src/__tests__/store.test.ts` — Zustand store 全状态测试
- `web/package.json` — 添加 test/test:watch 脚本

**CI 工作流更新（3 个）：**
- `.github/workflows/test.yml` — 新建独立测试工作流
- `.github/workflows/build-on-push.yml` — 添加 test job 门控
- `.github/workflows/build-and-release.yml` — 添加 test job 门控

## Next Steps

- 可继续扩展前端交互组件测试
- 可添加 E2E 测试
- 可添加 configbuilder 配置构建器测试

## Important Patterns

### 测试模式
- **Go 测试**: `setupTestDB(t)` + `t.Cleanup(CleanupTestDB)` 内存 SQLite 隔离
- **前端测试**: Zustand `setState` 重置 + `getState()` 直接调用 actions
- **HTTP 测试**: `httptest.NewRecorder` + `httptest.NewRequest` + `mux.ServeHTTP`