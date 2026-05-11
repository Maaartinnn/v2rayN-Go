<p align="center">
  <h1 align="center">v2rayN-Go</h1>
  <p align="center">A lightweight, blazing-fast proxy control center with a modern web interface</p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white" alt="React">
  <img src="https://img.shields.io/badge/License-GPLv3-blue.svg" alt="License">
</p>

> English | **[中文](README.md)**

---

## ✨ Features

- 🚀 **Single Binary** — One executable with embedded frontend, zero dependencies, just run
- 🌐 **Modern Web UI** — Anthropic-style warm beige theme, React 19 + Tailwind CSS v4 minimalist control panel
- 🔌 **Multi-Core Support** — Compatible with Xray-core, Sing-box, Mihomo, with full lifecycle management (start/stop/status monitoring/log collection)
- 📡 **Multi-Protocol Parsing** — VMess, VLESS, Trojan, Shadowsocks, ShadowsocksR, Hysteria2, Hysteria, TUIC, AnyTLS, WireGuard
- 📋 **Subscription Management** — Concurrent fetching, scheduled auto-update, custom User-Agent, alias regex filtering, one-click link import, transactional atomic updates
- 📦 **Group Management** — Multi-level node groups with drag-and-drop reorder, filter by group, assign subscriptions to groups
- 🖼️ **QR Code Import** — Dedicated import page with drag/paste/upload image or URL for QR code node parsing
- ⚡ **Latency Testing** — TCP Ping and HTTP Ping dual modes, batch concurrent testing, visual latency indicators, auto-save to database
- 🔄 **Node Deduplication** — One-click removal of duplicate nodes
- ⛓️ **Chain Proxy** — DialerProxy support for multi-hop forwarding chains
- 🧩 **Routing Rules** — Visual management of direct/proxy/block routing rules with domain/IP/port conditions
- ♻️ **Strategy Groups** — Visual management of Selector, URLTest, Fallback, and LoadBalance proxy groups
- 🔧 **Type-Safe Config** — Generate Xray / Sing-box kernel configs via Go Structs, no JSON syntax errors
- 📝 **Kernel Config Generation** — Auto-generate Xray and Sing-box configuration files from node data
- 🖥️ **System Service** — Register as Windows Service or systemd daemon with install/start/stop/restart/uninstall/daemon commands
- 🌍 **Multi-Language** — Chinese / English with standalone locale files
- 🌙 **Dark Mode** — Follows system preference or manual override, full Anthropic style adaptation
- ⚙️ **External Config** — CLI flags, config.json, and Web settings page with three-tier priority, auto-persist on Web changes
- 🎯 **System Proxy** — One-click toggle for system-wide proxy
- 📦 **Multi-Source Kernel Download** — GitHub direct, mirror, custom URL, local binary/archive upload (tar.gz/zip auto-extract), automatic platform matching
- 📡 **WebSocket Real-Time Logs** — Live kernel log streaming with level and source filtering
- 🧩 **Code Splitting** — Lazy-loaded non-critical components for faster initial load
- 📦 **Dual Distribution** — Lite (~15MB) and Full (with kernels) editions

---

## 📸 Preview

| Home Dashboard | Node List | Core Manager |
|:---:|:---:|:---:|
| Centered control card + traffic stats | Card-based nodes + protocol badges + group filter | Three-core status + multi-source download/upload |

| Import Nodes | Group Management | Strategy Groups |
|:---:|:---:|:---:|
| Links / QR code / manual add | Drag-and-drop sort + subscription group config | Four strategy group types with visual config |

| Routing Rules | Log Terminal | Settings |
|:---:|:---:|:---:|
| Direct/proxy/block rules | Real-time logs + level filtering | Language/theme/network/system proxy |

---

## 🏗️ Project Structure

```
v2rayN-Go/
├── web/                           # Frontend (React 19 + Vite + Tailwind CSS v4)
│   └── src/
│       ├── components/            # UI Components
│       │   ├── Sidebar.tsx        # Collapsible navigation sidebar
│       │   ├── HomeView.tsx       # Dashboard control panel (traffic stats + quick actions)
│       │   ├── NodesView.tsx      # Node management (search/group/dedup/latency test/activate)
│       │   ├── ImportView.tsx     # Import page (links/QR code/manual add)
│       │   ├── GroupsView.tsx     # Group management (CRUD / drag-and-drop / subscription config)
│       │   ├── NodeEditForm.tsx   # Node edit/create form
│       │   ├── CoresView.tsx      # Core management (multi-source download/upload/start/stop)
│       │   ├── RoutingView.tsx    # Routing rule management
│       │   ├── StrategyGroupView.tsx  # Strategy group management
│       │   ├── SettingsView.tsx   # Settings (language/theme/network/system proxy)
│       │   ├── LogConsole.tsx     # Log terminal (level/source filtering)
│       │   ├── ErrorBoundary.tsx  # Error boundary
│       │   └── ui/               # Atomic UI components
│       ├── lib/
│       │   ├── api.ts             # API client (Axios)
│       │   ├── i18n.ts            # i18n + theme management
│       │   ├── useWebSocket.ts    # WebSocket hook
│       │   └── coreMap.ts         # Core name mapping utility
│       ├── locales/               # Standalone locale files
│       │   ├── zh-CN.ts           # Chinese
│       │   └── en-US.ts           # English
│       ├── store.ts               # Zustand global state management
│       ├── App.tsx                # Root component (routing + layout)
│       ├── main.tsx               # Entry point
│       └── index.css              # Global styles (Tailwind CSS v4)
└── src/                           # Backend (Go)
    ├── main.go                    # Entry point: init config → load config → execute CLI
    ├── cmd/cli.go                 # CLI commands + flag parsing
    │                              # Commands: install/uninstall/start/stop/restart/daemon/help
    ├── config/                    # App config (three-tier priority)
    │   └── config.go              # AppConfig definition, JSON loading, CLI parsing, settings persistence
    ├── database/                  # SQLite database (pure Go, no CGO required)
    │   ├── db.go                  # DB initialization / AutoMigrate / default group creation
    │   ├── models.go              # Profile, NodeGroup, RoutingRule, StrategyGroup, AppSetting
    │   └── utils.go               # UUID generation
    ├── parser/                    # Multi-protocol parsers + QR code decoder
    │   ├── parser.go              # Parse entry dispatch + batch parsing + subscription content parsing
    │   ├── vmess.go / vless.go / trojan.go
    │   ├── shadowsocks.go / ssr.go
    │   ├── hysteria2.go / hysteria.go / tuic.go
    │   ├── anytls.go              # AnyTLS protocol parser
    │   ├── wireguard.go           # WireGuard protocol parser
    │   ├── qrcode.go              # QR code decoding
    │   └── utils.go               # Base64 decode / URL parse / name extraction utilities
    ├── subscription/              # Subscription management service
    │   ├── subscription.go        # Subscription fetch/parse/filter/update, auto-update scheduler
    │   └── ping.go                # TCP Ping + HTTP Ping batch concurrent latency testing
    ├── configbuilder/             # Type-safe kernel config generator
    │   ├── xray.go                # Xray config struct definitions and generation
    │   ├── singbox.go             # Sing-box config struct definitions and generation
    │   └── utils.go               # Common utility functions
    ├── core/                      # Kernel process manager
    │   └── admin.go               # Three-core lifecycle management (start/stop/status/log/graceful shutdown)
    ├── updater/                   # Kernel online download & update
    │   └── updater.go             # Xray / Sing-box / Mihomo multi-source download (tar.gz/zip auto-extract)
    ├── service/                   # System service integration
    │   └── service.go             # Windows service / systemd registration & management / foreground direct run
    └── web/                       # Web server
        ├── embed.go               # Frontend static asset embedding (go:embed dist/*)
        └── server.go              # HTTP routes / RESTful API / WebSocket / static files / kernel config generation
```

---

## 🚀 Quick Start

### Download

Download the appropriate package from the [Releases](https://github.com/Maaartinnn/v2rayN-Go/releases) page:

| Version | Description | Size |
|---------|-------------|------|
| `v2rayN-Go-windows-amd64-lite.zip` | Windows 64-bit Lite | ~15MB |
| `v2rayN-Go-linux-amd64-lite.tar.gz` | Linux 64-bit Lite | ~15MB |
| `v2rayN-Go-darwin-amd64-lite.tar.gz` | macOS 64-bit Lite | ~15MB |

### Run

```bash
# Run directly (foreground mode, recommended for development)
./v2rayN-Go

# Override config with CLI flags
./v2rayN-Go --listen-ip 0.0.0.0 --port 8080 --socks-port 10808

# Install as system service (auto-start on boot)
./v2rayN-Go install
./v2rayN-Go start

# Other commands
./v2rayN-Go stop       # Stop the service
./v2rayN-Go restart    # Restart the service
./v2rayN-Go uninstall  # Uninstall the service
./v2rayN-Go help       # Show help
```

After starting, open your browser and navigate to **http://127.0.0.1:2017**

### Obtain Kernels

On first run with the Lite edition, go to the **Core Manager** page which supports multiple ways to obtain kernels:
- **GitHub Direct** — Auto-fetch latest release from GitHub Releases with automatic OS/architecture detection
- **GitHub Mirror** — Download via configured mirror URL
- **Custom URL** — Manually enter download URL
- **Local Upload** — Upload binary files or tar.gz/zip archives (auto-extract)

Kernels are stored in the local `bin/` directory, organized by type (`bin/xray/`, `bin/sing_box/`, `bin/mihomo/`).

---

## ⚙️ Configuration

v2rayN-Go supports three-tier configuration priority (highest to lowest):

### 1. CLI Flags (Highest Priority)

```bash
./v2rayN-Go --listen-ip 0.0.0.0 --port 8080 --socks-port 10808 --http-port 10809
```

| Flag | Description | Default |
|------|-------------|---------|
| `--listen-ip` | Listen IP address | `127.0.0.1` |
| `--port` | Web UI port | `2017` |
| `--socks-port` | SOCKS5 proxy port | `10808` |
| `--http-port` | HTTP proxy port | `10809` |
| `--outbound-ip` | Outbound bind IP | `0.0.0.0` |
| `--github-mirror` | GitHub mirror URL | empty (direct) |

### 2. config.json File

Create a `config.json` in the same directory as the executable:

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

### 3. Web Settings Page

After startup, use the **Settings** page to visually modify network parameters, GitHub mirror, system proxy, etc. Changes are automatically persisted to `config.json`.

---

## 🔨 Build from Source

### Prerequisites

- **Go** 1.22+
- **Node.js** 20+
- **npm** 9+

### Build Steps

```bash
# 1. Build frontend
cd web
npm install
npm run build
cd ..

# 2. Build backend (frontend assets are embedded via go:embed)
cd src
go build -ldflags="-s -w" -o v2rayN-Go.exe .
```

### Cross-Compilation

```bash
# Linux ARM64
cd src
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o v2rayN-Go-linux-arm64 .

# macOS ARM64 (Apple Silicon)
cd src
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o v2rayN-Go-darwin-arm64 .
```

---

## 📡 Supported Protocols

| Protocol | Parsing | Config Gen (Xray) | Config Gen (Sing-box) |
|----------|:-------:|:-----------------:|:---------------------:|
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

## 🧩 Supported Kernels

| Kernel | GitHub Repository | One-Click Download | Local Upload | Mirror/Custom URL | Log Collection |
|--------|-------------------|:------------------:|:------------:|:-----------------:|:--------------:|
| Xray-core | [XTLS/Xray-core](https://github.com/XTLS/Xray-core) | ✅ | ✅ | ✅ | ✅ |
| Sing-box | [SagerNet/sing-box](https://github.com/SagerNet/sing-box) | ✅ | ✅ | ✅ | ✅ |
| Mihomo | [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) | ✅ | ✅ | ✅ | ✅ |

---

## ⚙️ Tech Stack

### Backend
- **Language**: Go 1.22+
- **Web Framework**: Standard library `net/http`
- **Database**: SQLite (`glebarez/sqlite`, pure Go, no CGO required)
- **ORM**: GORM
- **WebSocket**: gorilla/websocket
- **System Service**: kardianos/service
- **QR Code Decoder**: gozxing

### Frontend
- **Framework**: React 19 + TypeScript
- **Build Tool**: Vite 8
- **Styling**: Tailwind CSS v4 + Anthropic-style design system
- **Routing**: wouter (lightweight React router)
- **State Management**: Zustand 5
- **Drag & Drop**: @dnd-kit
- **Animations**: Framer Motion 12
- **Command Palette**: cmdk
- **Icons**: Lucide Icons
- **HTTP Client**: Axios

---

## ⭐ Star History

<a href="https://www.star-history.com/?repos=Maaartinnn%2Fv2rayN-Go&type=date&logscale=&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=Maaartinnn/v2rayN-Go&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=Maaartinnn/v2rayN-Go&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=Maaartinnn/v2rayN-Go&type=date&legend=top-left" />
 </picture>
</a>

## 📄 License

[GNU General Public License v3.0](LICENSE)