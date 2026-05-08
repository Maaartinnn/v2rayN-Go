// 协议-内核兼容映射表
// 定义每种协议支持哪些内核，以及推荐的最优内核

export interface CoreMapping {
  name: string
  displayName: string
  subDir: string
  binaryName: string
}

// 支持的内核列表
export const CORES: CoreMapping[] = [
  { name: 'xray', displayName: 'Xray', subDir: 'xray', binaryName: 'xray' },
  { name: 'sing-box', displayName: 'Sing-box', subDir: 'sing_box', binaryName: 'sing-box' },
  { name: 'mihomo', displayName: 'Mihomo', subDir: 'mihomo', binaryName: 'mihomo' },
]

// 协议 → 支持的内核列表（按推荐优先级排序）
export const PROTOCOL_CORE_MAP: Record<string, string[]> = {
  vmess:        ['xray', 'sing-box', 'mihomo'],
  vless:        ['xray', 'sing-box'],
  trojan:       ['xray', 'sing-box', 'mihomo'],
  shadowsocks:  ['xray', 'sing-box', 'mihomo'],
  shadowsocksr: ['mihomo'],
  hysteria2:    ['sing-box', 'mihomo'],
  hysteria:     ['sing-box', 'mihomo'],
  tuic:         ['sing-box', 'mihomo'],
  wireguard:    ['sing-box', 'mihomo'],
  anytls:       ['sing-box'],
  socks:        ['xray', 'sing-box'],
  http:         ['xray', 'sing-box'],
}

// 获取协议支持的内核列表
export function getSupportedCores(protocol: string): string[] {
  return PROTOCOL_CORE_MAP[protocol] || []
}

// 获取协议的最优推荐内核（列表中第一个）
export function getRecommendedCore(protocol: string): string | null {
  const cores = getSupportedCores(protocol)
  return cores.length > 0 ? cores[0] : null
}

// 根据已安装的内核列表，获取协议的最优已安装内核
export function getBestInstalledCore(protocol: string, installedCores: string[]): string | null {
  const supported = getSupportedCores(protocol)
  for (const core of supported) {
    if (installedCores.includes(core)) {
      return core
    }
  }
  return null
}

// 所有支持的协议列表
export const PROTOCOLS = [
  { value: 'vmess', label: 'VMess' },
  { value: 'vless', label: 'VLESS' },
  { value: 'trojan', label: 'Trojan' },
  { value: 'shadowsocks', label: 'Shadowsocks' },
  { value: 'hysteria2', label: 'Hysteria2' },
  { value: 'wireguard', label: 'WireGuard' },
  { value: 'tuic', label: 'TUIC' },
  { value: 'socks', label: 'SOCKS' },
  { value: 'http', label: 'HTTP' },
] as const

// 传输协议列表
export const NETWORKS = [
  { value: 'tcp', label: 'TCP' },
  { value: 'ws', label: 'WebSocket' },
  { value: 'grpc', label: 'gRPC' },
  { value: 'h2', label: 'HTTP/2' },
  { value: 'quic', label: 'QUIC' },
] as const

// TLS 选项
export const TLS_OPTIONS = [
  { value: '', label: 'None' },
  { value: 'tls', label: 'TLS' },
  { value: 'reality', label: 'Reality' },
] as const

// 加密方式
export const SECURITY_METHODS = [
  'auto', 'aes-128-gcm', 'chacha20-poly1305', 'none', 'zero',
] as const