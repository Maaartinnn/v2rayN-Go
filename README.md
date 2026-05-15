<p align="center">
  <h1 align="center">v2rayN-Go</h1>
  <p align="center">轻量、极速、基于 Web 界面管理的通用代理控制中心</p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white" alt="React">
  <img src="https://img.shields.io/badge/License-GPLv3-blue.svg" alt="License">
</p>

> **[English](README-en.md)** | 中文

---

## ✨ 特性

- 🚀 **单文件部署** — 一个二进制文件包含前后端，零依赖，解压即用
- 🌐 **现代化 Web UI** — Anthropic 风格暖米色主题，React 19 + Tailwind CSS v4 打造的极简控制面板
- 🔌 **多内核支持** — 兼容 Xray-core、Sing-box、Mihomo 等主流代理内核，支持完整生命周期管理（启动/停止/状态监控/日志收集/优雅退出）
- 📡 **多协议解析** — 支持 VMess、VLESS、Trojan、Shadowsocks、ShadowsocksR、Hysteria2、Hysteria、TUIC、AnyTLS、WireGuard 等协议
- 📋 **订阅管理** — 并发拉取订阅、定期自动更新、自定义 User-Agent、别名正则过滤、一键导入分享链接，支持事务式原子更新
- 📦 **分组管理** — 多层级节点分组，支持拖拽排序、按分组筛选、分配合并到分组
- 🖼️ **二维码导入** — 独立的导入页面，支持拖拽/粘贴/上传图片或输入 URL 解析二维码中的节点
- ⚡ **延迟测速** — TCP Ping 与 HTTP Ping 两种测速方式，批量并发测速，可视化延迟状态，自动写入数据库
- 🔄 **节点去重** — 一键去除重复节点
- ⛓️ **链式代理** — 支持前置代理串联（DialerProxy），灵活构建多跳转发链路
- 🧩 **路由规则** — 可视化管理直连/代理/拦截路由规则，支持域名/IP/端口条件
- ♻️ **策略组** — 可视化管理 Selector（手动切换）、URLTest（自动测速）、Fallback（故障转移）、LoadBalance（负载均衡）四种策略组
- 🔧 **强类型配置** — 通过 Go Struct 生成 Xray / Sing-box 内核配置，杜绝 JSON 语法错误
- 📝 **内核配置生成** — 基于节点数据自动生成 Xray 与 Sing-box 内核配置文件
- 🖥️ **系统服务** — 支持注册为 Windows 服务 / systemd 守护进程，支持前台运行、后台守护及 install/start/stop/restart/uninstall 命令
- 🌍 **多语言** — 支持中文 / English，独立语言包文件
- 🌙 **深色模式** — 跟随系统自动切换 / 手动设置，完整 Anthropic 风格适配
- ⚙️ **配置外置** — 支持 CLI 参数、config.json、Web 设置页三级配置管理，Web 修改自动持久化
- 🎯 **系统代理** — 一键开启/关闭系统代理
- 📦 **多源内核下载** — GitHub 直连、镜像源、自定义 URL、本地上传二进制/压缩包（支持 tar.gz/zip 自动解压），自动匹配平台架构
- 🔍 **CPU 微架构检测** — mihomo amd64 自动检测 CPU 级别（v1/v2/v3/v4），匹配最优内核版本
- ⏱️ **缓存加速** — GitHub Release API 自动缓存（5 分钟 TTL），减少请求次数
- 📡 **WebSocket 实时日志** — 内核运行日志实时推送，支持按级别和来源筛选
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
│       │   ├── HomeView.tsx       # 首页控制面板（流量统计 + 快捷操作）
│       │   ├── NodesView.tsx      # 节点管理（搜索/分组/去重/测速/激活切换）
│       │   ├── ImportView.tsx     # 导入页面（链接/二维码/手动添加）
│       │   ├── GroupsView.tsx     # 分组管理（CRUD / 拖拽排序 / 订阅配置）
│       │   ├── NodeEditForm.tsx   # 节点编辑/新建表单
│       │   ├── CoresView.tsx      # 内核管理（多源下载/上传/启动/停止）
│       │   ├── RoutingView.tsx    # 路由规则管理
│       │   ├── StrategyGroupView.tsx  # 策略组管理
│       │   ├── SettingsView.tsx   # 设置（语言/主题/网络/系统代理）
│       │   ├── LogConsole.tsx     # 日志终端（级别/来源过滤）
│       │   ├── ErrorBoundary.tsx  # 错误边界
│       │   └── ui/               # 通用 UI 原子组件（含 DeleteConfirmBanner 等）
│       ├── lib/
│       │   ├── api.ts             # API 客户端 (Axios)
│       │   ├── i18n.ts            # 多语言 + 主题管理
│       │   ├── useWebSocket.ts    # WebSocket Hook
│       │   └── coreMap.ts         # 内核名称映射工具
│       ├── locales/               # 独立语言包
│       │   ├── zh-CN.ts           # 中文
│       │   └── en-US.ts           # English
│       ├── store.ts               # Zustand 全局状态管理
│       ├── App.tsx                # 根组件（路由 + 布局）
│       ├── main.tsx               # 入口文件
│       └── index.css              # 全局样式 (Tailwind CSS v4)
└── src/                           # 后端 (Go)
    ├── main.go                    # 程序入口：初始化配置 → 加载配置 → 执行 CLI
    ├── cmd/
    │   └── cli.go                 # CLI 命令 + 参数解析
    │                              # 支持：前台运行 install/uninstall/start/stop/restart/daemon/help
    ├── config/
    │   └── config.go              # AppConfig 定义、JSON 加载、CLI 标志解析、三级优先级加载、设置持久化
    ├── coredef/
    │   └── coredef.go             # 内核类型与元数据注册表（唯一事实来源），支持 Xray/Sing-box/Mihomo
    ├── database/                  # SQLite 数据库（纯 Go，无需 CGO）
    │   ├── db.go                  # 数据库初始化 / AutoMigrate / 软删除清理 / 排序重平衡 / 默认分组创建
    │   ├── models.go              # Profile, NodeGroup, RoutingRule, StrategyGroup, AppSetting
    │   └── utils.go               # UUID 生成
    ├── parser/                    # 多协议解析器 + 二维码解码
    │   ├── parser.go              # 解析入口分发 + 批量解析 + 订阅内容解析
    │   ├── vmess.go / vless.go / trojan.go
    │   ├── shadowsocks.go / ssr.go
    │   ├── hysteria2.go / hysteria.go / tuic.go
    │   ├── anytls.go              # AnyTLS 协议解析
    │   ├── wireguard.go           # WireGuard 协议解析
    │   ├── qrcode.go              # 二维码解码
    │   └── utils.go               # Base64 解码 / URL 解析 / 名称提取等工具函数
    ├── subscription/              # 订阅管理服务
    │   ├── subscription.go        # 订阅拉取/解析/过滤/更新，自动更新调度器
    │   └── ping.go                # TCP Ping + HTTP Ping 批量并发测速服务
    ├── configbuilder/             # 强类型内核配置生成器
    │   ├── xray.go                # Xray 内核配置结构体定义与生成（含策略组 balancer 支持）
    │   ├── singbox.go             # Sing-box 内核配置结构体定义与生成
    │   └── utils.go               # 通用工具函数
    ├── core/                      # 内核进程管理
    │   └── admin.go               # 三核生命周期管理（启动/停止/状态监控/日志收集/优雅退出/超时强制终止）
    ├── httpclient/                # 统一 HTTP 客户端
    │   └── httpclient.go          # 自动注入 User-Agent，支持普通/代理两种模式，基于 go 标准库 transport 克隆
    ├── updater/                   # 内核在线下载更新
    │   └── updater.go             # Xray / Sing-box / Mihomo 多源下载（镜像降级 + mihomo CPU 级别降级），tar.gz/zip 自动解压
    ├── sysmgr/                    # 系统服务管理
    │   └── sysmgr.go              # 前台运行 / 后台守护 / Windows 服务 + systemd 注册与生命周期管理
    └── web/                       # Web 服务器
        ├── embed.go               # 前端静态资源嵌入（go:embed dist/*）
        └── server.go              # HTTP 路由 / RESTful API / WebSocket / 静态文件 / 内核配置生成
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
- **GitHub 直连** — 从 GitHub Releases 自动拉取最新版，自动匹配操作系统与架构
- **GitHub 镜像** — 通过配置的镜像源下载
- **自定义 URL** — 手动输入下载地址
- **本地上传** — 支持上传二进制文件或 tar.gz/zip 压缩包（自动解压）

内核下载后存放在本地 `bin/` 目录，按内核类型分目录存放（`bin/xray/`、`bin/sing_box/`、`bin/mihomo/`）。

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

启动后在 **设置** 页面中可视化修改网络参数、GitHub 镜像、系统代理等，保存后自动持久化到 `config.json`。

---

## 🔨 从源码构建

### 环境要求

- **Go** 1.26+
- **Node.js** 20+
- **npm** 9+

### 构建步骤

```bash
# 1. 构建前端
cd web
npm install
npm run build
cd ..

# 2. 构建后端（前端产物通过 go:embed 嵌入到二进制文件中）
cd src
go build -ldflags="-s -w" -o v2rayN-Go.exe .
```

### 交叉编译

```bash
# Linux ARM64
cd src
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o v2rayN-Go-linux-arm64 .

# macOS ARM64 (Apple Silicon)
cd src
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o v2rayN-Go-darwin-arm64 .
```

### 一键构建脚本

Windows 下可使用项目根目录的 `src/dev-build.cmd` 一键完成前端构建 + 后端编译：

```cmd
cd src
dev-build.cmd
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
| AnyTLS | ✅ | — | — |
| WireGuard | ✅ | — | — |

---

## 🧩 支持的内核

| 内核 | GitHub 仓库 | 一键下载 | 本地上传 | 镜像/自定义 URL | 日志收集 |
|------|-------------|:--------:|:--------:|:--------------:|:--------:|
| Xray-core | [XTLS/Xray-core](https://github.com/XTLS/Xray-core) | ✅ | ✅ | ✅ | ✅ |
| Sing-box | [SagerNet/sing-box](https://github.com/SagerNet/sing-box) | ✅ | ✅ | ✅ | ✅ |
| Mihomo | [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) | ✅ | ✅ | ✅ | ✅ |

---

## ⚙️ 技术栈

### 后端
- **语言**: Go 1.26+
- **Web 框架**: 标准库 `net/http`
- **数据库**: SQLite（`glebarez/sqlite`，纯 Go，无需 CGO）
- **ORM**: GORM
- **WebSocket**: gorilla/websocket
- **系统服务**: kardianos/service
- **QR 码解码**: gozxing
- **UUID**: google/uuid
- **CPU 检测**: golang.org/x/sys（用于 mihomo CPU 微架构级别检测）

### 前端
- **框架**: React 19 + TypeScript
- **构建工具**: Vite 8
- **样式**: Tailwind CSS v4 + Anthropic 风格设计系统
- **路由**: wouter（轻量级 React 路由）
- **状态管理**: Zustand 5
- **拖拽排序**: @dnd-kit
- **动画**: Framer Motion 12
- **命令面板**: cmdk
- **图标**: Lucide Icons
- **HTTP 客户端**: Axios

---

## ⭐ Star 历史

<a href="https://www.star-history.com/?repos=Maaartinnn%2Fv2rayN-Go&type=date&logscale=&legend=top-left">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=Maaartinnn/v2rayN-Go&type=date&theme=dark&legend=top-left" />
    <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=Maaartinnn/v2rayN-Go&type=date&legend=top-left" />
    <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=Maaartinnn/v2rayN-Go&type=date&legend=top-left" />
  </picture>
</a>

## 📄 License

[GNU General Public License v3.0](LICENSE)