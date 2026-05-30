# System Patterns

## Architecture

### 分层架构
```
Web Handler → Service → Database (GORM/SQLite)
     ↓
  Core Service → CoreAdminManager (xray/sing-box/mihomo)
```

### Go 后端结构
```
src/
├── main.go              # 入口
├── cmd/cli.go           # CLI 命令行
├── config/              # 配置管理
├── configbuilder/       # 代理配置生成器（singbox/xray）
├── core/                # 内核进程管理
├── coredef/             # 核心常量定义
├── database/            # 数据库层（GORM + SQLite）
│   ├── models.go        # 全量数据模型（Profile/NodeGroup/RoutingRule/...）
│   ├── profile_summary.go  # 精简 DTO（ProfileListItem/ColorPair）
│   └── profile_colors.go   # 颜色映射纯函数（不入库）
├── httpclient/          # HTTP 客户端
├── parser/              # 协议解析器
├── ping/                # 测速服务
├── service/             # 业务逻辑层
├── subscription/        # 订阅管理
├── sysmgr/              # 系统管理（系统代理等）
├── updater/             # 内核更新器
└── web/                 # Web 服务器 + Handler
```

### 前端结构
```
web/src/
├── main.tsx             # React 入口
├── App.tsx              # 根组件（路由 + ToastContainer 挂载）
├── store.ts             # Zustand 全局状态（ProfileListItem/Toast/...）
├── components/
│   ├── NodesView.tsx    # 节点列表（使用精简数据 + uuid 标识）
│   ├── HomeView.tsx     # 首页（从 profileList 查找激活节点）
│   ├── NodeEditForm.tsx # 节点编辑表单（接收完整 Profile）
│   └── ui/
│       ├── ToastContainer.tsx  # 通用 Toast 通知（ToastItem + ToastContainer）
│       └── ...                 # 其他原子组件
├── lib/
│   ├── api.ts           # API 调用层
│   ├── coreMap.ts       # 协议-内核兼容映射
│   ├── i18n.ts          # 多语言 + 主题管理
│   └── useWebSocket.ts  # WebSocket Hook
├── locales/             # 国际化字典
│   ├── zh-CN.ts
│   └── en-US.ts
└── __tests__/           # 前端测试
```

## Key Design Patterns

### 1. 依赖注入 (DI)
- Web Server 是纯 DI 容器，显式注入所有 Service 和 Handler
- Handler 通过构造函数接收 Service，不依赖全局状态
- 唯一的全局状态：`database.DB`（GORM 连接）

### 2. 数据传输优化（DTO 模式）
- **精简 DTO**：`ProfileListItem` 仅含列表展示所需字段（12 个），不传 `raw_link`/`proxy_credential` 等大字段
- **按需加载**：列表返回精简数据，编辑时通过 `GET /api/profiles/{uuid}` 获取完整 Profile
- **颜色后端驱动**：颜色映射纯函数在后端计算，随 DTO 一起返回 JSON，前端直接使用

### 3. 前端标识统一（uuid）
- 前端列表 `key`、多选 `Set<string>`、激活比较、删除/编辑操作全部使用 `uuid`
- Store 中 `activeProfileUUID: string | null` 替代原来的 `activeProfile: Profile | null`
- HomeView 和 App.tsx 从 `profileList` 中 `.find()` 查找激活节点

### 4. Toast 通知模式（数据与行为分离）
- **Store 只负责数据**：`addToast` 不含 `setTimeout`，纯数据操作
- **组件管理生命周期**：`ToastItem` 的 `useEffect` 管理定时器，卸载时清理，避免竞态条件
- **类型安全**：`Toast` 接口定义 `type/color/action/duration`

### 5. 排序系统
- 所有序列表使用 `sort_order` 字段，步长 10
- `SortBetween` 插值、`Rebalance` 重排、`SortInsertSafe` 冲突检测
- 整数溢出保护：`safeAdd`/`safeSub` 检测溢出返回 0

### 6. 错误处理
- Service 层定义三种业务错误：`ErrNotFound`(404)、`ErrValidation`(400)、`ErrConflict`(409)
- `mapServiceError` 统一映射为 HTTP 状态码
- 500 错误内部细节写 slog，前端仅收到泛化提示
- 所有 `fmt.Errorf` 使用 `%w` 保留错误链，支持 `errors.Is`/`errors.As`
- 统一 JSON 错误响应格式：`{"error": "msg", "code": status_code}`

### 7. 日志系统
- 统一使用 `log/slog`，在 `main.go` 入口处通过 `coredef.InitLogger()` 初始化
- 结构化日志：`slog.Info("msg", "key", value)` 替代 `log.Printf`
- 日志级别：debug < info < warn < error，可在初始化时配置
- CLI 用户输出使用 `fmt.Println`，不混入日志流

### 8. 实时通信
- WebSocket 广播核心状态、日志、流量指标
- `WSHandler` 实现 `StatusBroadcaster` 接口

### 9. 前端状态管理
- Zustand 单一 store，通过 `setState` 精确更新
- 组件通过 selector 避免不必要的重渲染

### 10. 协议解析
- `ParseLink` 根据协议前缀分发到对应解析器
- 每种协议一个独立文件，返回统一的 `database.Profile` 结构

### 11. 断电安全防护
- **原子写入**：`config.AtomicWriteFile` 导出函数，写临时文件 → `f.Sync()` 强制刷盘 → `os.Rename` 原子替换
- **仅用于命脉配置**：`config.json` 的保存和 `.bak` 恢复使用原子写入；Xray/Sing-box 配置文件是派生数据，使用普通 `os.WriteFile`
- **`.bak` 容灾回滚**：`loadJSONConfig` 在文件损坏（0KB / JSON 解析失败）时自动从 `config.json.bak` 恢复；`BackupConfig` 仅在应用完整启动后调用，避免脏数据污染
- **SQLite WAL 模式**：连接参数 `_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)`，断电最多丢失最近一个事务，数据库结构不会损坏
- **sync.Once 保证**：`BackupConfig` 使用 `sync.Once` 防止重复备份

### 12. 无文件落地（Fileless Execution）
- **stdin 模式**：默认通过 `cmd.Stdin` 管道将 JSON 配置注入内核进程，全程不触碰物理磁盘
- **内核 stdin 支持**：Xray `-config stdin:` / Sing-box `-c stdin:` / Mihomo `-d . -f -`
- **Functional Options**：`StartCore(coreType, "", core.WithStdin(data))` 零侵入扩展，向后兼容文件模式
- **`BuildBytes` 接口**：`ConfigBuilder.BuildBytes()` 仅返回 JSON 字节不写文件，`Build()` 保留给调试模式
- **调试开关**：`CoreConfigDebug` 配置字段，开启后写入文件并使用传统文件模式启动
- **跨平台进程安全**：`process_unix.go`（Setpgid + kill -pid）/ `process_windows.go`（HideWindow + Process.Kill）
- **工作目录**：stdin 模式下 `cmd.Dir` 设为内核二进制所在目录（Mihomo 需要加载 geoip.db 等资源）
