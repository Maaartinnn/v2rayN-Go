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
├── configbuilder/       # 代理配置生成器（xray/singbox/mihomo）
├── core/                # 内核进程管理
├── coredef/             # 核心常量定义 + 协议→内核映射表
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
│   ├── NodeEditForm.tsx # 节点编辑表单（使用后端能力矩阵）
│   └── ui/
│       ├── ToastContainer.tsx  # 通用 Toast 通知（ToastItem + ToastContainer）
│       └── ...                 # 其他原子组件
├── lib/
│   ├── api.ts           # API 调用层（含 coreMatrix 端点）
│   ├── coreMap.ts       # 协议/网络/TLS 常量定义（已被 NodeEditForm 引用）
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
- **stdin 模式**：默认通过 `cmd.Stdin` 管道将 JSON/YAML 配置注入内核进程，全程不触碰物理磁盘
- **内核 stdin 支持**：Xray `-config stdin:` / Sing-box `-c stdin:` / Mihomo `-d . -f -`
- **Functional Options**：`StartCore(coreType, "", core.WithStdin(data))` 零侵入扩展，向后兼容文件模式
- **`BuildBytes` 接口**：`ConfigBuilder.BuildBytes()` 仅返回 JSON/YAML 字节不写文件，`Build()` 保留给调试模式
- **调试开关**：`CoreConfigDebug` 配置字段，开启后写入 `binConfig/` 目录并使用传统文件模式启动
- **跨平台进程安全**：`process_unix.go`（Setpgid + kill -pid）/ `process_windows.go`（HideWindow + Process.Kill）
- **工作目录**：stdin 模式下 `cmd.Dir` 设为内核二进制所在目录（Mihomo 需要加载 geoip.db 等资源）

### 13. Mihomo ConfigBuilder (2026-06-01)
- **YAML 配置**：Mihomo 使用 YAML 格式（非 JSON），ConfigBuilder 接口统一
- **结构体设计**：基础字段强类型（name/type/server/port） + `Extra map[string]any` + `yaml:",inline"` 处理协议专属参数（ws-opts, grpc-opts, reality-opts 等）
- **Go 1.26 特性**：利用 `new(expr)` 创建 `*bool` 指针字面量，配合 `omitempty` 避免零值误判
- **路由规则格式**：Mihomo 规则为纯字符串 `"TYPE,PARAM,POLICY"`（DOMAIN-SUFFIX/IP-CIDR/DST-PORT/MATCH）
- **VLESS 支持**：Mihomo 支持 VLESS 协议（从 Clash.Meta 内核起支持）

### 14. 设置局部更新 + 失焦保存 (2026-06-01)
- **Dirty Flag**：`UpdateSettings` 中 `changed := false`，只有字段值真正改变时才标记 `changed = true`，末尾只有 `changed` 时才调用 `SaveJSONConfig()`
- **三步拦截**（判空→判变→判合法）：端口 1-65535，IP 通过 `net.ParseIP` 校验，非法数据永不触碰内存和磁盘
- **值未变零 I/O**：如果所有字段都没变，直接返回 nil，跳过 AtomicWriteFile + Sync() 开销
- **失焦保存**：前端 `handleBlur(field, value)` 组装单字段 JSON 发送，后端校验失败时 `loadSettings()` 回滚脏输入
- **回车即保存**：输入框 `onKeyDown` Enter 触发 `e.currentTarget.blur()`
- **Toggle 立即保存**：无失焦概念，`onClick` 时先 setState 再调用 `handleBlur`

### 15. 协议→内核能力矩阵 (2026-06-01)
- **ProtocolCoreMap**：`coredef/protocol_cores.go` 定义协议→内核映射表（唯一事实来源），每种协议的内核列表按推荐优先级排序
- **GetCompatibleInstalledCores**：交叉查询映射表 + `updater.GetLocalCores()` 检查 bin/ 下二进制文件，返回协议兼容且已安装的内核列表（返回空数组而非 nil，避免 JSON 序列化为 null）
- **GetInstalledCoreMatrix**：遍历所有协议返回 `map[string][]string`，前端一次性加载，协议切换时零延迟查字典
- **API 设计**：`GET /api/profiles/core-matrix` 返回能力矩阵（独立端点，职责单一），`GET /api/profiles/{uuid}` 保持返回原始 profile
- **前端使用**：NodeEditForm 使用 `coreMatrix[protocol]` 获取当前协议可用内核，删除了前端 coreMap.ts 过滤逻辑

### 16. 按钮样式规范 (2026-06-02)
项目 `index.css` 中定义了 4 种全局按钮 CSS class，所有按钮必须使用这些 class，禁止内联 style 定义按钮外观：

| Class | 用途 | 视觉特征 |
|-------|------|----------|
| `.btn-primary` | 主要操作（提交、确认、登录） | 深色背景 + 白色文字 + 橙色阴影 |
| `.btn-secondary` | 次要操作（退出登录、普通操作） | 卡片背景 + 前景色文字 + 边框 |
| `.btn-ghost` | 辅助操作（取消、返回） | 无背景无边框 + 灰色文字，hover 出现背景 |
| `.btn-danger` | 危险操作（删除、注销、关闭） | 红色背景 + 白色文字 |

用法示例：`className="btn-primary px-4 py-2 text-sm"`（尺寸/间距仍用 Tailwind 控制）

### 17. ToastContainer 全局挂载 (2026-06-02)
- `ToastContainer` 必须渲染在 `App` 组件顶层（auth state 判断之前），确保登录页和已认证页面都能看到 toast 通知
- 如果 `ToastContainer` 放在 `AuthenticatedApp` 内部，登录页调用 `addToast()` 时 Toast 组件不在 DOM 树中，通知无法显示
- `AuthenticatedApp` 内部不需要再重复渲染 `ToastContainer`
