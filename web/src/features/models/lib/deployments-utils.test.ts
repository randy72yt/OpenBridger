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

import { formatRemainingMinutes } from './deployments-utils'

describe('formatRemainingMinutes', () => {
  test('formats a numeric minute count', () => {
    expect(formatRemainingMinutes(90)).toBe('1h 30m')
    expect(formatRemainingMinutes(0)).toBe('0m')
    expect(formatRemainingMinutes(1500)).toBe('1d 1h')
  })

  test('parses numeric strings and rejects other types', () => {
    expect(formatRemainingMinutes('90')).toBe('1h 30m')
    expect(formatRemainingMinutes('nope')).toBeNull()
    expect(formatRemainingMinutes(true)).toBeNull()
    expect(formatRemainingMinutes({})).toBeNull()
    expect(formatRemainingMinutes(null)).toBeNull()
  })
})
