# Active Context

## Current Work Focus

前端架构优化：节点列表精简传输 + 通用 Toast 通知系统。

## Recent Changes

### 节点列表精简传输重构（2026-05-29）

**1. 后端新增 DTO 和颜色映射：**
- 新建 `src/database/profile_summary.go` — `ColorPair` + `ProfileListItem` DTO（9 个展示字段 + 3 个颜色字段）
- 新建 `src/database/profile_colors.go` — 协议/内核/延迟颜色映射纯函数（常量查表，不入库）
- 修改 `src/service/profile_service.go` — 新增 `ListSummary` 方法，GORM `Select()` 只查 9 个字段
- 修改 `src/web/handler_profile.go` — `handleList` 改调 `ListSummary`
- 修改 `src/web/handler_test.go` — List 测试改为断言 `ProfileListItem` 结构

**2. 前端统一使用精简数据 + uuid 标识：**
- 修改 `web/src/store.ts` — 新增 `ProfileListItem`/`ColorPair` 类型；`profiles` → `profileList`；`activeProfile` → `activeProfileUUID`
- 修改 `web/src/components/NodesView.tsx` — 使用精简数据 + 后端返回颜色；多选统一用 `uuid`；编辑时通过 `profileApi.get(uuid)` 按需获取完整数据；新增内核徽标显示
- 修改 `web/src/components/HomeView.tsx` — 适配新 store
- 修改 `web/src/App.tsx` — 适配新 store
- 修改 `web/src/__tests__/store.test.ts` — 测试适配新 store

### 通用 Toast 通知系统（2026-05-29）

**1. Store 层升级：**
- 新增 `Toast` + `ToastAction` 接口，`addToast` 支持 `color`/`action`/`duration` 选项
- 移除 Store 中的 `setTimeout`，定时器生命周期下放到组件层

**2. 组件层：**
- 新建 `web/src/components/ui/ToastContainer.tsx` — `ToastItem`（定时器管理）+ `ToastContainer`（aria-live + 响应式）
- 修改 `web/src/App.tsx` — 挂载 `<ToastContainer />`
- 修改 `web/src/lib/useWebSocket.ts` — `addToast` 调用添加 `duration: 5000`

**3. 编辑按钮 Toast 错误通知：**
- 修改 `web/src/components/NodesView.tsx` — `handleEditClick` catch 中调用 `addToast`
- 修改 `web/src/locales/zh-CN.ts` + `web/src/locales/en-US.ts` — 新增 `nodes.edit_load_failed`

### 后端代码质量提升（2026-05-28）

- 提取硬编码常量至 `coredef/constants.go`
- 统一日志系统（log → slog），新建 `coredef/logger.go`
- 错误链规范化审计（全部 112 处 `fmt.Errorf` 已正确使用 `%w`）

### 测试编写（2026-05-28）

- Go 后端新增 19 个测试文件，覆盖 parser/database/service/config/web
- 前端新增 3 个测试文件，Vitest 框架搭建
- CI 3 个 GitHub Actions 工作流配置

## Next Steps

- 可继续扩展前端交互组件测试
- 可添加 E2E 测试
- 可添加 configbuilder 配置构建器测试

## Important Patterns

### 数据传输优化模式
- **精简 DTO**：`ProfileListItem` 仅含列表展示字段，`GET /api/profiles/{uuid}` 按需获取完整数据
- **颜色后端驱动**：颜色映射纯函数在后端计算，前端直接使用，改颜色只改后端常量
- **uuid 统一标识**：前端列表 key、多选、激活比较全部使用 `uuid`，不依赖 `ID`

### Toast 通知模式
- **Store 只负责数据**：`addToast` 不含 `setTimeout`，定时器由 `ToastItem` 组件的 `useEffect` 管理
- **类型安全**：`Toast` 接口定义完整配置（type/color/action/duration）
- **i18n 适配**：Toast 消息使用 `t()` 翻译函数

### 测试模式
- **Go 测试**: `setupTestDB(t)` + `t.Cleanup(CleanupTestDB)` 内存 SQLite 隔离
- **前端测试**: Zustand `setState` 重置 + `getState()` 直接调用 actions
- **HTTP 测试**: `httptest.NewRecorder` + `httptest.NewRequest` + `mux.ServeHTTP`

### 日志模式
- **统一使用 `log/slog`**：`slog.Info/Warn/Error` 替代 `log.Printf/Println`
- **初始化**：`coredef.InitLogger("info", os.Stderr)` 在 `main.go` 入口处调用
- **CLI 输出**：`fmt.Println` 用于用户交互信息（非日志）

### 常量模式
- **全局默认值**：`coredef.DefaultWebPort` 等常量在 `coredef/constants.go`
- **业务限制**：`coredef.MultipartMaxMemoryDefault` 等
- **内部常量**：WebSocket 参数保持在 `web/handler_ws.go` 内