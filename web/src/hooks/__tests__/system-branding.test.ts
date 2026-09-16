import { describe, expect, it } from 'vitest'

import { mapStatusDataToConfig } from '../use-system-config'

describe('system branding configuration', () => {
  it('uses OpenBridger identity when the server has no custom brand', () => {
    expect(mapStatusDataToConfig({ system_name: '', logo: '' })).toMatchObject({
      systemName: 'OpenBridger',
      logo: '/openbridger-mark.svg',
    })
  })

  it('preserves an administrator-provided name and logo', () => {
    expect(
      mapStatusDataToConfig({ system_name: 'Team Gateway', logo: '/team.svg' })
    ).toMatchObject({
      systemName: 'Team Gateway',
      logo: '/team.svg',
    })
  })
})
