import { describe, it, expect } from 'vitest'
import { AVAILABLE_LANGUAGES } from '../i18n'

describe('AVAILABLE_LANGUAGES', () => {
  it('should contain zh and en', () => {
    expect(AVAILABLE_LANGUAGES).toHaveLength(2)
    const codes = AVAILABLE_LANGUAGES.map((l) => l.code)
    expect(codes).toContain('zh')
    expect(codes).toContain('en')
  })

  it('each language should have code and label', () => {
    for (const lang of AVAILABLE_LANGUAGES) {
      expect(lang.code).toBeTruthy()
      expect(lang.label).toBeTruthy()
    }
  })
})