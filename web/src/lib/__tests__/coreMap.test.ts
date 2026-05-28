import { describe, it, expect } from 'vitest'
import {
  CORES,
  PROTOCOL_CORE_MAP,
  PROTOCOLS,
  NETWORKS,
  TLS_OPTIONS,
  SECURITY_METHODS,
  getSupportedCores,
  getRecommendedCore,
  getBestInstalledCore,
} from '../coreMap'

// ==================== Constants ====================

describe('CORES', () => {
  it('should contain 3 cores', () => {
    expect(CORES).toHaveLength(3)
  })

  it('should have xray, sing-box, mihomo', () => {
    const names = CORES.map((c) => c.name)
    expect(names).toEqual(['xray', 'sing-box', 'mihomo'])
  })

  it('each core should have required fields', () => {
    for (const core of CORES) {
      expect(core.name).toBeTruthy()
      expect(core.displayName).toBeTruthy()
      expect(core.subDir).toBeTruthy()
      expect(core.binaryName).toBeTruthy()
    }
  })
})

describe('PROTOCOL_CORE_MAP', () => {
  it('should map vmess to xray first', () => {
    expect(PROTOCOL_CORE_MAP.vmess[0]).toBe('xray')
  })

  it('should map shadowsocksr to mihomo only', () => {
    expect(PROTOCOL_CORE_MAP.shadowsocksr).toEqual(['mihomo'])
  })

  it('should map anytls to sing-box only', () => {
    expect(PROTOCOL_CORE_MAP.anytls).toEqual(['sing-box'])
  })

  it('all protocols should have non-empty core lists', () => {
    for (const [, cores] of Object.entries(PROTOCOL_CORE_MAP)) {
      expect(cores.length).toBeGreaterThan(0)
    }
  })
})

describe('PROTOCOLS', () => {
  it('should contain protocol options', () => {
    expect(PROTOCOLS.length).toBeGreaterThan(0)
  })

  it('each protocol should have value and label', () => {
    for (const p of PROTOCOLS) {
      expect(p.value).toBeTruthy()
      expect(p.label).toBeTruthy()
    }
  })
})

describe('NETWORKS', () => {
  it('should contain tcp', () => {
    const values = NETWORKS.map((n) => n.value)
    expect(values).toContain('tcp')
    expect(values).toContain('ws')
    expect(values).toContain('grpc')
  })
})

describe('TLS_OPTIONS', () => {
  it('should contain none, tls, reality', () => {
    const values = TLS_OPTIONS.map((t) => t.value)
    expect(values).toContain('')
    expect(values).toContain('tls')
    expect(values).toContain('reality')
  })
})

describe('SECURITY_METHODS', () => {
  it('should contain auto', () => {
    expect(SECURITY_METHODS).toContain('auto')
  })

  it('should contain chacha20-poly1305', () => {
    expect(SECURITY_METHODS).toContain('chacha20-poly1305')
  })
})

// ==================== Pure Functions ====================

describe('getSupportedCores', () => {
  it('should return cores for vmess', () => {
    const cores = getSupportedCores('vmess')
    expect(cores).toEqual(['xray', 'sing-box', 'mihomo'])
  })

  it('should return empty array for unknown protocol', () => {
    const cores = getSupportedCores('unknown')
    expect(cores).toEqual([])
  })

  it('should return mihomo for shadowsocksr', () => {
    const cores = getSupportedCores('shadowsocksr')
    expect(cores).toEqual(['mihomo'])
  })
})

describe('getRecommendedCore', () => {
  it('should return xray for vmess', () => {
    expect(getRecommendedCore('vmess')).toBe('xray')
  })

  it('should return mihomo for shadowsocksr', () => {
    expect(getRecommendedCore('shadowsocksr')).toBe('mihomo')
  })

  it('should return null for unknown protocol', () => {
    expect(getRecommendedCore('unknown')).toBeNull()
  })

  it('should return sing-box for hysteria2', () => {
    expect(getRecommendedCore('hysteria2')).toBe('sing-box')
  })
})

describe('getBestInstalledCore', () => {
  it('should return first matching installed core', () => {
    expect(getBestInstalledCore('vmess', ['mihomo', 'xray'])).toBe('xray')
  })

  it('should return null if no core installed', () => {
    expect(getBestInstalledCore('vmess', [])).toBeNull()
  })

  it('should return null if installed cores do not match', () => {
    expect(getBestInstalledCore('shadowsocksr', ['xray', 'sing-box'])).toBeNull()
  })

  it('should return first supported if all installed', () => {
    expect(getBestInstalledCore('trojan', ['sing-box', 'xray', 'mihomo'])).toBe('xray')
  })

  it('should respect priority order', () => {
    // vmess priority: xray > sing-box > mihomo
    expect(getBestInstalledCore('vmess', ['mihomo', 'sing-box'])).toBe('sing-box')
  })
})