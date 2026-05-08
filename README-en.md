<p align="center">
  <h1 align="center">v2rayN-Go</h1>
  <p align="center">A lightweight, blazing-fast proxy control center with a modern web interface</p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white" alt="React">
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License">
</p>

> English | **[中文](README.md)**

---

## ✨ Features

- 🚀 **Single Binary** — One executable with embedded frontend, zero dependencies, just run
- 🌐 **Modern Web UI** — Anthropic-style warm beige theme, React + Tailwind CSS minimalist control panel
- 🔌 **Multi-Core Support** — Compatible with Xray-core, Sing-box, Mihomo, and other proxy engines
- 📡 **Multi-Protocol Parsing** — VMess, VLESS, Trojan, Shadowsocks, ShadowsocksR, Hysteria2, TUIC
- 📋 **Subscription Management** — Concurrent fetching, auto-update, custom User-Agent, one-click link import
- 📦 **Group Management** — Multi-level node groups, filter by group, assign subscriptions to groups
- 🖼️ **QR Code Import** — Drag/paste/upload images or enter URL to parse QR codes for nodes
- ⚡ **Latency Testing** — Batch TCP Ping with concurrent workers, visual latency indicators
- 🔄 **Node Deduplication** — One-click removal of duplicate nodes
- 🧩 **Routing Rules** — Visual management of direct/proxy/block routing rules
- 🔧 **Type-Safe Config** — Generate kernel configs via Go Structs, no JSON syntax errors
- 🖥️ **System Service** — Register as Windows Service or systemd daemon
- 🌍 **Multi-Language** — Chinese / English with standalone locale files
- 🌙 **Dark Mode** — Follows system preference or manual override, full Anthropic style adaptation
- ⚙️ **External Config** — CLI flags, config.json, and Web settings page with three-tier priority
- 📦 **Dual Distribution** — Lite (~15MB) and Full (with kernels) editions

---

## 📸 Preview

| Home Dashboard | Node List | Core Manager |
|:---:|:---:|:---:|
| Centered control card + traffic stats | Card-based nodes + protocol badges + group filter | Three-core status + one-click download/upload |

| Subscription Manager | Routing Rules | Settings |
|:---:|:---:|:---:|
| Add/edit/refresh subscriptions | Direct/proxy/block rules | Language/theme/network config |

---

## 🏗️ Project Structure

```
v2rayN-Go/
├── web/                           # Frontend (React 19 + Vite + Tailwind CSS)
│   └── src/
│       ├── components/            # UI Components
│       │   ├── Sidebar.tsx        # Collapsible navigation sidebar
│       │   ├── HomeView.tsx       # Dashboard control panel
│       │   ├── NodesView.tsx      # Node management (search/group/dedup/QR)
│       │   ├── SubscriptionsView.tsx  # Subscription management
│       │   ├── CoresView.tsx      # Core Hub management
│       │   ├── RoutingView.tsx    # Routing rule management
│       │   ├── SettingsView.tsx   # Settings (language/theme/network)
│       │   └── LogConsole.tsx     # Log terminal
│       ├── locales/               # Standalone locale files
│       │   ├── zh-CN.ts           # Chinese
│       │   └── en-US.ts           # English
│       ├── lib/
│       │   ├── api.ts             # API client
│       │   ├── i18n.ts            # i18n + theme management
│       │   └── useWebSocket.ts    # WebSocket hook
│       └── store.ts               # Zustand global state
└── src/                           # Backend (Go)
    ├── cmd/cli.go                 # CLI commands + flag parsing
    ├── config/                    # App config (three-tier priority)
    ├── database/                  # SQLite database (pure Go)
    │   └── models.go              # Profile, Subscription, NodeGroup, RoutingRule
    ├── parser/                    # Multi-protocol parsers + QR code decoder
    ├── subscription/              # Subscription service + latency testing
    ├── configbuilder/             # Type-safe config generator
    ├── core/                      # Kernel process manager
    ├── updater/                   # Kernel online updater (Xray/Sing-box/Mihomo)
    ├── service/                   # System service integration
    └── web/                       # Web server + go:embed
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

### Download Kernels

On first run with the Lite edition, go to the **Core Manager** page and click the download button to automatically fetch Xray-core, Sing-box, or Mihomo from GitHub Releases into the local `bin/` directory. You can also upload local binary files.

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

After startup, use the **Settings** page to visually modify network parameters, GitHub mirror, etc. Changes are automatically saved to `config.json`.

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

---

## 🧩 Supported Kernels

| Kernel | GitHub Repository | One-Click Download | Local Upload |
|--------|-------------------|:------------------:|:------------:|
| Xray-core | [XTLS/Xray-core](https://github.com/XTLS/Xray-core) | ✅ | ✅ |
| Sing-box | [SagerNet/sing-box](https://github.com/SagerNet/sing-box) | ✅ | ✅ |
| Mihomo | [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) | ✅ | ✅ |

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
- **State Management**: Zustand
- **Animations**: Framer Motion
- **Icons**: Lucide Icons
- **HTTP Client**: Axios

---

## 📄 License

[GNU General Public License v3.0](LICENSE)