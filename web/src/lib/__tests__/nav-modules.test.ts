import { describe, expect, it } from 'vitest'

import { parseHeaderNavModules as parseAdminNav } from '@/features/system-settings/maintenance/config'

import { parseHeaderNavModules } from '../nav-modules'

describe('rankings navigation defaults', () => {
  it('hides rankings in public navigation and administration defaults', () => {
    expect(parseHeaderNavModules(undefined).rankings.enabled).toBe(false)
    expect(parseAdminNav(undefined).rankings.enabled).toBe(false)
    expect(parseHeaderNavModules(undefined).pricing.enabled).toBe(true)
  })

  it('retains an explicit administrator choice to enable rankings', () => {
    const configuration = JSON.stringify({
      rankings: { enabled: true, requireAuth: true },
    })
    expect(parseHeaderNavModules(configuration).rankings).toEqual({
      enabled: true,
      requireAuth: true,
    })
    expect(parseAdminNav(configuration).rankings).toEqual({
      enabled: true,
      requireAuth: true,
    })
  })
})
