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
import { render, screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { ModelProviderIcon } from '@/components/model-provider-icon'
import { resolveModelProvider } from '@/lib/model-provider'

vi.mock('@/lib/lobe-icon', () => ({
  getLobeIcon: (iconName: string) => <span data-lobe-icon={iconName} />,
}))

describe('resolveModelProvider', () => {
  test.each([
    ['gpt-4o', 'OpenAI', 'OpenAI'],
    ['o3-mini', 'OpenAI', 'OpenAI'],
    ['claude-sonnet-4-5', 'Claude', 'Claude.Color'],
    ['gemini-2.5-pro', 'Gemini', 'Gemini.Color'],
    ['deepseek-r1', 'DeepSeek', 'DeepSeek.Color'],
    ['grok-3', 'Grok', 'Grok'],
    ['qwen-max', 'Qwen', 'Qwen.Color'],
  ])(
    'maps %s to %s (%s)',
    (modelName, expectedLabel, expectedIcon) => {
      const provider = resolveModelProvider(modelName)

      expect(provider).not.toBeNull()
      expect(provider?.label).toBe(expectedLabel)
      expect(provider?.icon).toBe(expectedIcon)
    }
  )

  test('returns null for blank or unrecognized custom model names', () => {
    expect(resolveModelProvider('')).toBeNull()
    expect(resolveModelProvider('   ')).toBeNull()
    expect(resolveModelProvider('custom-internal-v1')).toBeNull()
  })

  test('renders dynamic provider icon and falls back cleanly for unknown models', () => {
    const { rerender, container } = render(
      <ModelProviderIcon modelName='gpt-4o' />
    )
    expect(
      container.querySelector('[data-lobe-icon="OpenAI"]')
    ).toBeInTheDocument()

    rerender(<ModelProviderIcon modelName='claude-sonnet-4-5' />)
    expect(
      container.querySelector('[data-lobe-icon="Claude.Color"]')
    ).toBeInTheDocument()

    rerender(<ModelProviderIcon modelName='custom-model' />)
    expect(
      container.querySelector('[data-provider="Generic"]')
    ).toBeInTheDocument()

    expect(screen.queryByRole('img')).not.toBeInTheDocument()
  })
})
