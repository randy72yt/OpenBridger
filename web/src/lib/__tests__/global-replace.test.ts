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

import { getPreviewText } from '@/features/dashboard/lib/text'
import { parseHttpStatusCodeRules } from '@/lib/http-status-code-rules'
import { sanitizeCssVariableName } from '@/lib/utils'

describe('global replaceAll equivalents', () => {
  test('sanitizeCssVariableName keeps the previous character mapping', () => {
    expect(sanitizeCssVariableName('gpt-3.5 turbo/x')).toBe('gpt-3-5-turbo-x')
    expect(sanitizeCssVariableName('a+b')).toBe('ab')
  })

  test('getPreviewText strips tags and markdown symbols before truncating', () => {
    expect(getPreviewText('<b>Hello</b> *world*', 100)).toBe('Hello world')
    expect(getPreviewText('# Title', 4)).toBe('Titl...')
  })

  test('parseHttpStatusCodeRules still accepts the fullwidth comma and inner spaces', () => {
    const parsed = parseHttpStatusCodeRules('200， 201 - 204')
    expect(parsed.ok).toBe(true)
    expect(parsed.normalized).toBe('200-204')
  })
})
