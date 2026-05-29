# Tech Context

## Technologies

### Backend (Go)
- **Go 1.26** with modules
- **GORM** (via `gorm.io/gorm`) — ORM for SQLite
- **SQLite** (via `github.com/glebarez/sqlite` — pure Go, no CGO)
- **Google UUID** — 生成唯一标识
- **Gorilla WebSocket** — 实时状态推送
- **gozxing** — 二维码解析
- **kardianos/service** — 系统服务管理

### Frontend (React)
- **React 19** + TypeScript 6
- **Vite 8** — 构建工具
- **Zustand 5** — 状态管理
- **TailwindCSS 4** — 样式（@tailwindcss/vite 插件）
- **Framer Motion** — 动画
- **wouter** — 轻量路由
- **Axios** — HTTP 客户端
- **@dnd-kit** — 拖拽排序
- **Lucide React** — 图标
- **Vitest + @testing-library/react + jsdom** — 测试

### CI/CD
- GitHub Actions
- 三个工作流：`test.yml`（独立测试）、`build-on-push.yml`（构建）、`build-and-release.yml`（发布）

## Development Setup
```bash
# 后端
cd src && go test ./...

# 前端
cd web && npm test        # 单次运行
cd web && npm run test:watch  # 监听模式
cd web && npm run build   # TypeScript 编译 + Vite 构建
```

## Technical Constraints
- SQLite 不支持并发写入（通过 busy_timeout 缓解）
- CGO_ENABLED=0 纯 Go 编译，跨平台无依赖
- 前端嵌入后端（go:embed），单一二进制分发

## API 设计模式
- **列表精简传输**：`GET /api/profiles/` 返回 `ProfileListItem[]`（12 字段），不传大字段
- **按需完整数据**：`GET /api/profiles/{uuid}` 返回完整 `Profile`（30+ 字段）
- **颜色后端计算**：协议/内核/延迟颜色由后端纯函数计算，随 DTO 返回
- **WebSocket 广播**：核心状态、日志、流量指标实时推送