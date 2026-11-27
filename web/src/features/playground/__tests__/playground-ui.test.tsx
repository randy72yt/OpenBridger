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
import userEvent from '@testing-library/user-event'
import { describe, expect, test, vi } from 'vitest'

import { PlaygroundChat } from '../components/chat/playground-chat'
import {
  PlaygroundHeroGreeting,
  PlaygroundStarterPills,
} from '../components/chat/playground-empty-state'
import { PlaygroundInput } from '../components/input/playground-input'
import { PlaygroundHeader } from '../components/playground-header'
import { DEFAULT_CONFIG, DEFAULT_PARAMETER_ENABLED } from '../constants'
import type { Message } from '../types'

vi.mock('@/lib/lobe-icon', () => ({
  getLobeIcon: (iconName: string) => <span data-lobe-icon={iconName} />,
}))

describe('Playground ChatGPT-style UI', () => {
  test('switches hero emblem dynamically when the selected model changes', () => {
    const { rerender, container } = render(
      <PlaygroundHeroGreeting modelName='gpt-4o' />
    )

    expect(screen.getByText('What can I help with?')).toBeInTheDocument()
    expect(screen.getByText('OpenAI · gpt-4o')).toBeInTheDocument()
    expect(
      container.querySelector('[data-lobe-icon="OpenAI"]')
    ).toBeInTheDocument()

    rerender(<PlaygroundHeroGreeting modelName='claude-sonnet-4-5' />)

    expect(
      screen.getByText('Claude · claude-sonnet-4-5')
    ).toBeInTheDocument()
    expect(
      container.querySelector('[data-lobe-icon="Claude.Color"]')
    ).toBeInTheDocument()
  })

  test('renders model selector inside the bottom-right of the input composer', () => {
    const { container } = render(
      <PlaygroundInput
        config={DEFAULT_CONFIG}
        groups={[{ label: 'Default', value: 'default', ratio: 1 }]}
        groupValue='default'
        models={[
          { label: 'gpt-4o', value: 'gpt-4o' },
          { label: 'claude-sonnet-4-5', value: 'claude-sonnet-4-5' },
        ]}
        modelValue='claude-sonnet-4-5'
        onConfigChange={() => undefined}
        onGroupChange={() => undefined}
        onModelChange={() => undefined}
        onParameterEnabledChange={() => undefined}
        onSubmit={() => undefined}
        parameterEnabled={DEFAULT_PARAMETER_ENABLED}
      />
    )

    const modelSelector = screen.getByRole('combobox')
    expect(modelSelector).toHaveTextContent('claude-sonnet-4-5')
    expect(
      container.querySelector('[data-lobe-icon="Claude.Color"]')
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Send' })).toBeInTheDocument()
  })

  test('submits starter prompt when a quick starter pill is clicked', async () => {
    const user = userEvent.setup()
    const onSelectPrompt = vi.fn()

    render(<PlaygroundStarterPills onSelectPrompt={onSelectPrompt} />)

    await user.click(screen.getByRole('button', { name: 'Summarize text' }))

    expect(onSelectPrompt).toHaveBeenCalledWith('Summarize text')
  })

  test('confirms before clearing conversation when New chat is clicked', async () => {
    const user = userEvent.setup()
    const onClearMessages = vi.fn()

    render(
      <PlaygroundHeader hasMessages onClearMessages={onClearMessages} />
    )

    await user.click(screen.getByRole('button', { name: 'New chat' }))
    expect(screen.getByText('Clear chat history?')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Clear' }))
    expect(onClearMessages).toHaveBeenCalledTimes(1)
  })

  test('renders per-message model avatar and model label in conversation stream', () => {
    const messages: Message[] = [
      {
        key: 'u-1',
        from: 'user',
        versions: [{ id: 'v1', content: 'Compare models' }],
      },
      {
        key: 'a-1',
        from: 'assistant',
        model: 'gemini-2.5-pro',
        status: 'complete',
        versions: [{ id: 'v2', content: 'Hello from Gemini' }],
      },
    ]

    const { container } = render(
      <PlaygroundChat currentModel='gpt-4o' messages={messages} />
    )

    expect(screen.getByTestId('assistant-model-avatar')).toBeInTheDocument()
    expect(
      container.querySelector('[data-lobe-icon="Gemini.Color"]')
    ).toBeInTheDocument()
    expect(screen.getByText('gemini-2.5-pro')).toBeInTheDocument()
  })
})
