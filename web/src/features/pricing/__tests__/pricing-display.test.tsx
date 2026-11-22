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
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ThemeProvider } from '@/context/theme-provider'
import { routeTree } from '@/routeTree.gen'
import {
  DEFAULT_CURRENCY_CONFIG,
  useSystemConfigStore,
} from '@/stores/system-config-store'

import { formatPrice } from '../lib/price'
import type { PricingData, PricingModel } from '../types'

const displayState = vi.hoisted(() => ({
  mode: 'catalog' as 'catalog' | 'empty' | 'error',
}))

const fixtureModel: PricingModel = {
  id: 1,
  model_name: 'fixture-model',
  description: 'Local fixture, not a published price',
  quota_type: 0,
  model_ratio: 1,
  completion_ratio: 2,
  cache_ratio: 0.5,
  enable_groups: ['default', 'premium'],
  vendor_id: 1,
  supported_endpoint_types: ['openai'],
  tags: 'chat',
}

const otherModel: PricingModel = {
  id: 2,
  model_name: 'other-model',
  quota_type: 0,
  model_ratio: 4,
  completion_ratio: 1,
  enable_groups: ['default'],
  vendor_id: 1,
}

const groupRatio = { default: 1, premium: 3 }

const catalog: PricingData = {
  success: true,
  data: [fixtureModel, otherModel],
  vendors: [{ id: 1, name: 'Fixture Vendor' }],
  group_ratio: groupRatio,
  usable_group: {
    default: { desc: 'default', ratio: 1 },
    premium: { desc: 'premium', ratio: 3 },
  },
  supported_endpoint: {},
  auto_groups: [],
}

const emptyCatalog: PricingData = {
  success: true,
  data: [],
  vendors: [],
  group_ratio: {},
  usable_group: {},
  supported_endpoint: {},
  auto_groups: [],
}

vi.mock('@visactor/react-vchart', () => ({
  VChart: () => null,
}))

vi.mock('@/lib/vchart', () => ({
  VCHART_OPTION: {},
}))

vi.mock('@/hooks/use-notifications', () => ({
  useNotifications: () => ({
    popoverOpen: false,
    setPopoverOpen: vi.fn(),
    unreadCount: 0,
    activeTab: 'announcements',
    setActiveTab: vi.fn(),
    notice: '',
    announcements: [],
    loading: false,
  }),
}))

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api')>('@/lib/api')
  return {
    ...actual,
    getStatus: vi.fn(async () => ({
      price: 2,
      usd_exchange_rate: 7,
      quota_display_type: 'USD',
      system_name: 'OpenBridger',
    })),
    api: new Proxy(actual.api, {
      get(target, prop, receiver) {
        if (prop !== 'get') return Reflect.get(target, prop, receiver)
        return async (url: string) => {
          if (url === '/api/pricing') {
            if (displayState.mode === 'error') {
              throw new Error('pricing unavailable')
            }
            return {
              data: displayState.mode === 'empty' ? emptyCatalog : catalog,
            }
          }
          if (url === '/api/perf-metrics/summary') {
            return { data: { success: true, data: { models: [] } } }
          }
          return { data: { success: true, data: null } }
        }
      },
    }),
  }
})

function pricedModel(model: PricingModel): PricingModel {
  return { ...model, group_ratio: groupRatio }
}

function renderSquare(search = '') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({
      initialEntries: [`/pricing${search}`],
    }),
  })
  return render(
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </ThemeProvider>
  )
}

describe('pricing display skin', () => {
  beforeEach(() => {
    displayState.mode = 'catalog'
    localStorage.clear()
    useSystemConfigStore.getState().setConfig({
      currency: { ...DEFAULT_CURRENCY_CONFIG, quotaDisplayType: 'USD' },
    })
  })

  it('keeps URL filters, group prices, and card or table values on the same model', async () => {
    const user = userEvent.setup()
    renderSquare('?group=premium&tokenUnit=K&view=table&rechargePrice=true')

    expect(await screen.findByText('fixture-model')).toBeVisible()
    expect(screen.queryByText('other-model')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Recharge' })).toHaveAttribute(
      'aria-pressed',
      'true'
    )
    expect(screen.getByRole('button', { name: '/1K' })).toHaveAttribute(
      'aria-pressed',
      'true'
    )
    expect(screen.getByRole('button', { name: 'Table view' })).toHaveAttribute(
      'aria-pressed',
      'true'
    )

    const premiumRecharge = {
      input: formatPrice(pricedModel(fixtureModel), 'input', 'K', true, 2, 7, 'premium'),
      output: formatPrice(pricedModel(fixtureModel), 'output', 'K', true, 2, 7, 'premium'),
      cache: formatPrice(pricedModel(fixtureModel), 'cache', 'K', true, 2, 7, 'premium'),
    }
    expect(screen.getByRole('table')).toHaveTextContent(premiumRecharge.input)
    expect(screen.getByRole('table')).toHaveTextContent(premiumRecharge.output)
    expect(screen.getByRole('table')).toHaveTextContent(premiumRecharge.cache)

    await user.click(screen.getByRole('button', { name: 'Card view' }))
    const cardPrices = screen.getByRole('group', { name: 'Pricing' })
    expect(cardPrices).toHaveTextContent(premiumRecharge.input)
    expect(cardPrices).toHaveTextContent(premiumRecharge.output)
    expect(cardPrices).toHaveTextContent(premiumRecharge.cache)
    expect(cardPrices).toHaveTextContent('Input')
    expect(cardPrices).toHaveTextContent('Output')
    expect(cardPrices).toHaveTextContent('Cached')

    const performance = await screen.findByLabelText(
      'Performance metrics for the last 24 hours'
    )
    expect(within(performance).getAllByText(/^—/)).toHaveLength(3)
    expect(performance).not.toHaveTextContent('0.0%')

    await user.click(screen.getByRole('button', { name: '/1M' }))
    await user.click(screen.getByRole('button', { name: 'Standard' }))
    const standardMillion = {
      input: formatPrice(pricedModel(fixtureModel), 'input', 'M', false, 2, 7, 'premium'),
      output: formatPrice(pricedModel(fixtureModel), 'output', 'M', false, 2, 7, 'premium'),
      cache: formatPrice(pricedModel(fixtureModel), 'cache', 'M', false, 2, 7, 'premium'),
    }
    await waitFor(() => {
      expect(cardPrices).toHaveTextContent(standardMillion.input)
    })
    expect(cardPrices).toHaveTextContent(standardMillion.output)
    expect(cardPrices).toHaveTextContent(standardMillion.cache)
    expect(cardPrices).not.toHaveTextContent(premiumRecharge.input)

    await user.click(screen.getByRole('button', { name: /^default/ }))
    expect(await screen.findByText('other-model')).toBeVisible()
    const defaultGroup = formatPrice(
      pricedModel(fixtureModel),
      'input',
      'M',
      false,
      2,
      7,
      'default'
    )
    expect(screen.getAllByRole('group', { name: 'Pricing' })[0]).toHaveTextContent(
      defaultGroup
    )
    expect(defaultGroup).not.toBe(standardMillion.input)
  })

  it('shows an empty catalog separately from a failed pricing request', async () => {
    displayState.mode = 'empty'
    const empty = renderSquare()
    expect(await screen.findByText('No models found')).toBeVisible()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    empty.unmount()

    displayState.mode = 'error'
    renderSquare()
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Failed to load model pricing'
    )
    expect(screen.queryByText('No models found')).not.toBeInTheDocument()
  })
})
