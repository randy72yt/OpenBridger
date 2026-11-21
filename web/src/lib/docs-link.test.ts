/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { describe, expect, test } from 'vitest'

import { resolveDocsHref } from './docs-link'

describe('resolveDocsHref', () => {
  test('uses the language home for the default docs link, empty config, and /docs', () => {
    for (const input of [
      undefined,
      null,
      '',
      '   ',
      '/docs',
      '/docs/',
      'https://docs.openbridger.com',
      'https://docs.openbridger.com/',
    ]) {
      expect(resolveDocsHref(input, 'zhCN')).toBe(
        'https://docs.openbridger.com/zh/'
      )
      expect(resolveDocsHref(input, 'en')).toBe(
        'https://docs.openbridger.com/en/'
      )
      expect(resolveDocsHref(input, 'zhTW')).toBe(
        'https://docs.openbridger.com/zh/'
      )
      expect(resolveDocsHref(input, 'zh-TW')).toBe(
        'https://docs.openbridger.com/zh/'
      )
    }
  })

  test('does not invent a Traditional Chinese docs path', () => {
    expect(resolveDocsHref(undefined, 'zhTW')).not.toContain('zh-tw')
    expect(resolveDocsHref(undefined, 'zh-Hant')).toBe(
      'https://docs.openbridger.com/zh/'
    )
  })

  test('keeps a custom URL, including path, query, and hash', () => {
    const custom = 'https://help.example.com/guide?lang=en#start'
    expect(resolveDocsHref(custom, 'zhCN')).toBe(custom)
    expect(resolveDocsHref('https://docs.example.com/', 'en')).toBe(
      'https://docs.example.com/'
    )
    expect(
      resolveDocsHref(
        'https://docs.openbridger.com/zh/guide/quickstart/?q=1#call',
        'en'
      )
    ).toBe('https://docs.openbridger.com/zh/guide/quickstart/?q=1#call')
    expect(
      resolveDocsHref('https://docs.openbridger.com/en/#top', 'zhCN')
    ).toBe('https://docs.openbridger.com/en/#top')
    expect(resolveDocsHref('https://docs.openbridger.com/?utm=1', 'zhCN')).toBe(
      'https://docs.openbridger.com/?utm=1'
    )
  })

  test('keeps a custom relative path and does not send /docs back to itself', () => {
    expect(resolveDocsHref('/help/start', 'zhCN')).toBe('/help/start')
    for (const language of ['en', 'zhCN', 'zhTW']) {
      const href = resolveDocsHref('/docs', language)
      expect(href.startsWith('https://docs.openbridger.com/')).toBe(true)
      expect(href).not.toBe('/docs')
      expect(href).not.toContain('/zh/zh')
      expect(href).not.toContain('/en/en')
    }
  })

  test('does not follow a non-http docs link', () => {
    expect(resolveDocsHref('javascript:alert(1)', 'en')).toBe(
      'https://docs.openbridger.com/en/'
    )
    expect(resolveDocsHref('//docs.example.com', 'zhCN')).toBe(
      'https://docs.openbridger.com/zh/'
    )
  })
})
