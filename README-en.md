<p align="center">
  <h1 align="center">v2rayN-Go</h1>
  <p align="center">A lightweight, blazing-fast proxy control center with a modern web interface</p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white" alt="React">
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License">
</p>

> English | **[中文](README.md)**

---

## ✨ Features

- 🚀 **Single Binary** — One executable with embedded frontend, zero dependencies, just run
- 🌐 **Modern Web UI** — Claude-inspired minimalist control panel built with React + Tailwind CSS
- 🔌 **Multi-Core Support** — Compatible with Xray-core, Sing-box, and other proxy engines
- 📡 **Multi-Protocol Parsing** — VMess, VLESS, Trojan, Shadowsocks, ShadowsocksR, Hysteria2, TUIC
- 📋 **Subscription Management** — Concurrent fetching, auto-update, one-click link import
- ⚡ **Latency Testing** — Batch TCP Ping with concurrent workers, visual latency indicators
- 🔧 **Type-Safe Config** — Generate kernel configs via Go Structs, no JSON syntax errors
- 🖥️ **System Service** — Register as Windows Service or systemd daemon
- 🌍 **Multi-Language** — Chinese / English with auto-detection
- 🌙 **Dark Mode** — Follows system preference, log terminal adapts automatically
- 📦 **Dual Distribution** — Lite (~15MB) and Full (with kernels) editions

---

## 📸 Preview

| Home Dashboard | Node List | Log Terminal |
|:---:|:---:|:---:|
| Centered control card + traffic stats | Card-based nodes + protocol badges | macOS-style terminal + syntax highlighting |

---

## 🏗️ Project Structure

```
v2rayN-Go/
├── web/                           # Frontend (React + Vite + Tailwind CSS)
│   └── src/
│       ├── components/            # UI Components
│       │   ├── Sidebar.tsx        # Navigation sidebar
│       │   ├── HomeView.tsx       # Dashboard control panel
│       │   ├── NodesView.tsx      # Node management
│       │   └── LogConsole.tsx     # Log terminal
│       ├── lib/
│       │   ├── api.ts             # API client
│       │   ├── i18n.ts            # Internationalization
│       │   ├── useWebSocket.ts    # WebSocket hook
│       │   └── useDarkMode.ts     # Dark mode hook
│       └── store.ts               # Zustand global state
└── src/                           # Backend (Go)
    ├── cmd/cli.go                 # CLI commands
    ├── config/                    # App configuration
    ├── database/                  # SQLite database (pure Go)
    ├── parser/                    # Multi-protocol parsers
    ├── subscription/              # Subscription service + latency testing
    ├── configbuilder/             # Type-safe config generator
    ├── core/                      # Kernel process manager
    ├── updater/                   # Kernel online updater
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

On first run with the Lite edition, go to the **Cores** page and click the download button to automatically fetch Xray-core or Sing-box from GitHub Releases into the local `bin/` directory.

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

## ⚙️ Tech Stack

### Backend
- **Language**: Go 1.22+
- **Web Framework**: Standard library `net/http`
- **Database**: SQLite (`glebarez/sqlite`, pure Go, no CGO required)
- **ORM**: GORM
- **WebSocket**: gorilla/websocket
- **System Service**: kardianos/service

### Frontend
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite 8
- **Styling**: Tailwind CSS (Zinc neutral color scheme)
- **State Management**: Zustand
- **Animations**: Framer Motion
- **Icons**: Lucide Icons
- **HTTP Client**: Axios

---

## 📄 License

[MIT License](LICENSE)