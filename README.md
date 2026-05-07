<p align="center">
  <h1 align="center">v2rayN-Go</h1>
  <p align="center">轻量、极速、基于 Web 界面管理的通用代理控制中心</p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white" alt="React">
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License">
</p>

> **[English](README-en.md)** | 中文

---

## ✨ 特性

- 🚀 **单文件部署** — 一个二进制文件包含前后端，零依赖，解压即用
- 🌐 **现代化 Web UI** — React + Tailwind CSS 打造的 Claude 风格极简控制面板
- 🔌 **多内核支持** — 兼容 Xray-core、Sing-box 等主流代理内核
- 📡 **多协议解析** — 支持 VMess、VLESS、Trojan、Shadowsocks、ShadowsocksR、Hysteria2、TUIC 等协议
- 📋 **订阅管理** — 并发拉取订阅、自动更新、一键导入分享链接
- ⚡ **延迟测速** — TCP Ping 批量并发测速，可视化延迟状态
- 🔧 **强类型配置** — 通过 Go Struct 生成内核配置，杜绝 JSON 语法错误
- 🖥️ **系统服务** — 支持注册为 Windows 服务 / systemd 守护进程
- 🌍 **多语言** — 支持中文 / English，自动检测系统语言
- 🌙 **深色模式** — 跟随系统自动切换，日志终端自适应主题
- 📦 **双轨发行** — Lite 版（~15MB）与 Full 版（含内核）供选择

---

## 📸 界面预览

| 首页控制面板 | 节点列表 | 日志终端 |
|:---:|:---:|:---:|
| 居中控制卡片 + 流量统计 | 卡片式节点 + 协议徽章 | 仿 macOS 终端 + 语法高亮 |

---

## 🏗️ 项目架构

```
v2rayN-Go/
├── web/                           # 前端 (React + Vite + Tailwind CSS)
│   └── src/
│       ├── components/            # UI 组件
│       │   ├── Sidebar.tsx        # 侧边导航栏
│       │   ├── HomeView.tsx       # 首页控制面板
│       │   ├── NodesView.tsx      # 节点管理
│       │   └── LogConsole.tsx     # 日志终端
│       ├── lib/
│       │   ├── api.ts             # API 客户端
│       │   ├── i18n.ts            # 多语言模块
│       │   ├── useWebSocket.ts    # WebSocket Hook
│       │   └── useDarkMode.ts     # 深色模式 Hook
│       └── store.ts               # Zustand 全局状态
└── src/                           # 后端 (Go)
    ├── cmd/cli.go                 # CLI 命令行
    ├── config/                    # 应用配置
    ├── database/                  # SQLite 数据库 (纯 Go)
    ├── parser/                    # 多协议解析器
    ├── subscription/              # 订阅服务 + 延迟测速
    ├── configbuilder/             # 强类型配置生成器
    ├── core/                      # 内核进程管理
    ├── updater/                   # 内核在线更新
    ├── service/                   # 系统服务集成
    └── web/                       # Web 服务器 + go:embed
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

### 下载内核

精简版首次运行时，进入 **Cores** 页面点击下载按钮，自动从 GitHub Releases 拉取 Xray-core 或 Sing-box 到本地 `bin/` 目录。

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

## ⚙️ 技术栈

### 后端
- **语言**: Go 1.22+
- **Web 框架**: 标准库 `net/http`
- **数据库**: SQLite (`glebarez/sqlite`，纯 Go，无需 CGO)
- **ORM**: GORM
- **WebSocket**: gorilla/websocket
- **系统服务**: kardianos/service

### 前端
- **框架**: React 18 + TypeScript
- **构建工具**: Vite 8
- **样式**: Tailwind CSS (Zinc 中性色系)
- **状态管理**: Zustand
- **动画**: Framer Motion
- **图标**: Lucide Icons
- **HTTP 客户端**: Axios

---

## 📄 License

[MIT License](LICENSE)