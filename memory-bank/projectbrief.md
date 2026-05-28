# Project Brief

## v2rayN-Go

v2rayN-Go 是一个用 Go 语言重写的 v2rayN 代理管理工具，提供 Web UI 界面管理代理节点、分组、路由规则和策略组。

## 核心目标

- 提供跨平台（Windows/Linux/macOS）的代理管理后端
- 支持多种代理协议：VMess、VLESS、Trojan、Shadowsocks、ShadowsocksR、Hysteria/Hysteria2、TUIC、WireGuard、AnyTLS
- 支持多种内核：Xray、Sing-box、Mihomo
- 提供 React Web UI 前端进行可视化管理
- 支持订阅管理、节点测速、批量导入/去重
- 支持系统代理设置、路由规则配置

## 技术架构

- **后端**: Go (Gin-style HTTP handlers + GORM + SQLite)
- **前端**: React 19 + TypeScript + Vite + Zustand + TailwindCSS 4
- **通信**: REST API + WebSocket（实时状态推送）