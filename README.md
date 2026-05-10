<p align="center">
  <h1 align="center">v2rayN-Go</h1>
  <p align="center">轻量、极速、基于 Web 界面管理的通用代理控制中心</p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white" alt="React">
  <img src="https://img.shields.io/badge/License-GPLv3-blue.svg" alt="License">
</p>

> **[English](README-en.md)** | 中文

---

## ✨ 特性

- 🚀 **单文件部署** — 一个二进制文件包含前后端，零依赖，解压即用
- 🌐 **现代化 Web UI** — Anthropic 风格暖米色主题，React 19 + Tailwind CSS v4 打造的极简控制面板
- 🔌 **多内核支持** — 兼容 Xray-core、Sing-box、Mihomo 等主流代理内核
- 📡 **多协议解析** — 支持 VMess、VLESS、Trojan、Shadowsocks、ShadowsocksR、Hysteria2、Hysteria、TUIC 等协议
- 📋 **订阅管理** — 并发拉取订阅、自动更新、自定义 User-Agent、别名正则过滤、一键导入分享链接
- 📦 **分组管理** — 多层级节点分组，支持拖拽排序、按分组筛选、分配订阅到分组
- 🖼️ **二维码导入** — 独立的导入页面，支持拖拽/粘贴/上传图片或输入 URL 解析二维码中的节点
- ⚡ **延迟测速** — TCP Ping 批量并发测速，可视化延迟状态
- 🔄 **节点去重** — 一键去除重复节点
- 🧩 **路由规则** — 可视化管理直连/代理/拦截路由规则
- ♻️ **策略组** — 可视化管理手动切换、自动测速、故障转移、负载均衡四种策略组
- 🔧 **强类型配置** — 通过 Go Struct 生成内核配置，杜绝 JSON 语法错误
- 🖥️ **系统服务** — 支持注册为 Windows 服务 / systemd 守护进程
- 🌍 **多语言** — 支持中文 / English，独立语言包文件
- 🌙 **深色模式** — 跟随系统自动切换 / 手动设置，完整 Anthropic 风格适配
- ⚙️ **配置外置** — 支持 CLI 参数、config.json、Web 设置页三级配置管理
- 🎯 **系统代理** — 一键开启/关闭系统代理
- 📦 **多源内核下载** — GitHub 直连、镜像源、自定义 URL、本地上传二进制/压缩包
- 📝 **实时日志** — 带过滤功能的实时日志终端，支持按级别和来源筛选
- 🧩 **代码分割** — 非首屏组件懒加载，首屏加载更快速
- 📦 **双轨发行** — Lite 版（~15MB）与 Full 版（含内核）供选择

---

## 📸 界面预览

| 首页控制面板 | 节点列表 | 内核管理 |
|:---:|:---:|:---:|
| 居中控制卡片 + 流量统计 | 卡片式节点 + 协议徽章 + 分组筛选 | 三核状态 + 多源下载/上传 |

| 导入节点 | 分组管理 | 策略组 |
|:---:|:---:|:---:|
| 链接/二维码/手动添加 | 拖拽排序 + 订阅分组配置 | 四种策略组可视化配置 |

| 路由规则 | 日志终端 | 设置 |
|:---:|:---:|:---:|
| 直连/代理/拦截规则 | 实时日志 + 级别过滤 | 语言/主题/网络/系统代理 |

---

## 🏗️ 项目架构

```
v2rayN-Go/
├── web/                           # 前端 (React 19 + Vite + Tailwind CSS v4)
│   └── src/
│       ├── components/            # UI 组件
│       │   ├── Sidebar.tsx        # 可折叠侧边导航栏
│       │   ├── HomeView.tsx       # 首页控制面板
│       │   ├── NodesView.tsx      # 节点管理（搜索/分组/去重/测速）
│       │   ├── ImportView.tsx     # 导入页面（链接/二维码/手动添加）
│       │   ├── GroupsView.tsx     # 分组管理（CRUD / 拖拽排序 / 订阅配置）
│       │   ├── NodeEditForm.tsx   # 节点编辑/新建表单
│       │   ├── CoresView.tsx      # 内核管理（多源下载/上传）
│       │   ├── RoutingView.tsx    # 路由规则管理
│       │   ├── StrategyGroupView.tsx  # 策略组管理
│       │   ├── SettingsView.tsx   # 设置（语言/主题/网络/系统代理）
│       │   ├── LogConsole.tsx     # 日志终端（级别/来源过滤）
│       │   └── ErrorBoundary.tsx  # 错误边界
│       ├── locales/               # 独立语言包
│       │   ├── zh-CN.ts           # 中文
│       │   └── en-US.ts           # English
│       ├── lib/
│       │   ├── api.ts             # API 客户端
│       │   ├── i18n.ts            # 多语言 + 主题管理
│       │   └── useWebSocket.ts    # WebSocket Hook
│       └── store.ts               # Zustand 全局状态
└── src/                           # 后端 (Go)
    ├── main.go                    # 程序入口
    ├── cmd/cli.go                 # CLI 命令行 + 参数解析
    ├── config/                    # 应用配置（三级优先级）
    │   └── config.go              # AppConfig 定义与加载
    ├── database/                  # SQLite 数据库 (纯 Go)
    │   ├── db.go                  # 数据库初始化 / AutoMigrate
    │   ├── models.go              # Profile, Subscription, NodeGroup, RoutingRule, StrategyGroup
    │   └── utils.go               # UUID 生成
    ├── parser/                    # 多协议解析器 + QR码解码
    │   ├── parser.go              # 解析入口分发
    │   ├── vmess.go / vless.go / trojan.go
    │   ├── shadowsocks.go / ssr.go
    │   ├── hysteria2.go / hysteria.go / tuic.go
    │   └── qrcode.go              # 二维码解码
    ├── subscription/              # 订阅服务 + 延迟测速
    │   └── subscription.go        # 订阅拉取 / 节点解析 / TCPing 测速
    ├── configbuilder/             # 强类型配置生成器
    │   └── builder.go             # Xray / Sing-box 内核配置生成
    ├── core/                      # 内核进程管理
    │   └── admin.go               # 内核启动 / 停止 / 状态监控
    ├── updater/                   # 内核在线更新
    │   └── updater.go             # Xray / Sing-box / Mihomo 多源下载
    ├── service/                   # 系统服务集成
    │   └── service.go             # Windows 服务 / systemd
    └── web/                       # Web 服务器 + go:embed
        └── server.go              # HTTP 路由 / API 端点 / WebSocket / 静态文件
```

---

## 🚀 快速开始

### 下载

从 [Releases](https://github.com/Maaartinnn/v2rayN-Go/releases) 页面下载对应平台的压缩包：

| 版本 | 说明 | 体积 |
|------|------|------|
| `v2rayN-Go-windows-amd64-lite.zip` | Windows 64位 精简版 | ~15MB |
| `v2rayN-Go-linux-amd64-lite.tar.gz` | Linux 64位 精简版 | ~15MB |
| `v2rayN-Go-darwin-amd64-lite.tar.gz` | macOS 64位 精简版 | ~15MB |

### 运行

```bash
# 直接运行（前台模式，推荐开发调试）
./v2rayN-Go

# 使用 CLI 参数覆盖配置
./v2rayN-Go --listen-ip 0.0.0.0 --port 8080 --socks-port 10808

# 安装为系统服务（开机自启）
./v2rayN-Go install
./v2rayN-Go start

# 其他命令
./v2rayN-Go stop       # 停止服务
./v2rayN-Go restart    # 重启服务
./v2rayN-Go uninstall  # 卸载服务
./v2rayN-Go help       # 查看帮助
```

启动后打开浏览器访问 **http://127.0.0.1:2017**

### 获取内核

精简版首次运行时，进入 **内核管理** 页面，支持多种方式获取内核：
- **GitHub 直连** — 从 GitHub Releases 自动拉取最新版
- **GitHub 镜像** — 通过配置的镜像源下载
- **自定义 URL** — 手动输入下载地址
- **本地上传** — 支持上传二进制文件或压缩包

内核下载后存放在本地 `bin/` 目录。

---

## ⚙️ 配置管理

v2rayN-Go 支持三级配置优先级（从高到低）：

### 1. CLI 启动参数（最高优先级）

```bash
./v2rayN-Go --listen-ip 0.0.0.0 --port 8080 --socks-port 10808 --http-port 10809
```

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--listen-ip` | 监听 IP 地址 | `127.0.0.1` |
| `--port` | Web UI 端口 | `2017` |
| `--socks-port` | SOCKS5 代理端口 | `10808` |
| `--http-port` | HTTP 代理端口 | `10809` |
| `--outbound-ip` | 出站绑定 IP | `0.0.0.0` |
| `--github-mirror` | GitHub 镜像 URL | 空（直连） |

### 2. config.json 配置文件

在可执行文件同目录下创建 `config.json`：

```json
{
  "listen_ip": "0.0.0.0",
  "web_port": 2017,
  "socks_port": 10808,
  "http_port": 10809,
  "outbound_ip": "0.0.0.0",
  "github_mirror": "https://mirror.example.com"
}
```

### 3. Web 设置页面

启动后在 **设置** 页面中可视化修改网络参数、GitHub 镜像、系统代理等，保存后自动写入 `config.json`。

---

## 🔨 从源码构建

### 环境要求

- **Go** 1.22+
- **Node.js** 20+
- **npm** 9+

### 构建步骤

```bash
# 1. 构建前端
cd web
npm install
npm run build
cd ..

# 2. 构建后端（前端产物会通过 go:embed 嵌入）
cd src
go build -ldflags="-s -w" -o v2rayN-Go.exe .
```

### 交叉编译

```bash
# Linux ARM64
cd src
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o v2rayN-Go-linux-arm64 .

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o v2rayN-Go-darwin-arm64 .
```

---

## 📡 支持的协议

| 协议 | 解析 | 配置生成 (Xray) | 配置生成 (Sing-box) |
|------|:----:|:---------------:|:-------------------:|
| VMess | ✅ | ✅ | ✅ |
| VLESS | ✅ | ✅ | ✅ |
| Trojan | ✅ | ✅ | ✅ |
| Shadowsocks | ✅ | ✅ | ✅ |
| ShadowsocksR | ✅ | — | — |
| Hysteria2 | ✅ | — | ✅ |
| Hysteria | ✅ | — | ✅ |
| TUIC | ✅ | — | ✅ |

---

## 🧩 支持的内核

| 内核 | GitHub 仓库 | 一键下载 | 本地上传 | 镜像/自定义 URL |
|------|-------------|:--------:|:--------:|:--------------:|
| Xray-core | [XTLS/Xray-core](https://github.com/XTLS/Xray-core) | ✅ | ✅ | ✅ |
| Sing-box | [SagerNet/sing-box](https://github.com/SagerNet/sing-box) | ✅ | ✅ | ✅ |
| Mihomo | [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) | ✅ | ✅ | ✅ |

---

## ⚙️ 技术栈

### 后端
- **语言**: Go 1.22+
- **Web 框架**: 标准库 `net/http`
- **数据库**: SQLite (`glebarez/sqlite`，纯 Go，无需 CGO)
- **ORM**: GORM
- **WebSocket**: gorilla/websocket
- **系统服务**: kardianos/service
- **QR 码解码**: gozxing

### 前端
- **框架**: React 19 + TypeScript
- **构建工具**: Vite
- **样式**: Tailwind CSS v4 + Anthropic 风格设计系统
- **状态管理**: Zustand
- **动画**: Framer Motion
- **图标**: Lucide Icons
- **HTTP 客户端**: Axios

---

## 📄 License

[GNU General Public License v3.0](LICENSE)