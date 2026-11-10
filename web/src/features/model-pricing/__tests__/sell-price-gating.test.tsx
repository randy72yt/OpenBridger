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
  createRootRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { ModelPricingPanel } from '../model-pricing-panel'
import { pricingOptions } from '../pricing'

const pricingPayload = {
  success: true,
  data: {
    entries: [
      {
        model_name: 'example-model',
        version: 'v1',
        configured: { ModelPrice: 1.5 },
        effective: { ModelPrice: 1.5 },
      },
    ],
    options: pricingOptions({ ModelPrice: '{"example-model":1.5}' }),
    empty_version: 'empty',
  },
}

afterEach(() => {
  useAuthStore.getState().auth.reset()
  window.localStorage.clear()
  vi.restoreAllMocks()
})

async function renderPanel(pricingControlEnabled: boolean) {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'root', role: 100 })
  vi.spyOn(api, 'get').mockImplementation(async (url) => {
    if (url === '/api/status') {
      return {
        data: { data: { pricing_control_enabled: pricingControlEnabled } },
      }
    }
    if (url === '/api/option/model_pricing') {
      return { data: pricingPayload }
    }
    return { data: { success: true, data: [], vendors: [] } }
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <QueryClientProvider client={client}>
          <ModelPricingPanel modelName='example-model' />
        </QueryClientProvider>
      ),
    }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  await router.load()
  return render(<RouterProvider router={router} />)
}

describe('model pricing editor sell-price gating', () => {
  it('makes the per-model editor read-only when dynamic pricing control is on', async () => {
    await renderPanel(true)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Sell prices are published by dynamic pricing'
    )
    expect(
      screen.getByRole('link', { name: 'Open dynamic pricing' })
    ).toHaveAttribute('href', '/system-settings/models/pricing-control')
    expect(
      screen.queryByRole('button', { name: 'Save model prices' })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Restore default pricing' })
    ).not.toBeInTheDocument()
    expect(await screen.findByPlaceholderText('0.01')).toBeDisabled()
  })

  it('keeps save available when dynamic pricing control is off', async () => {
    await renderPanel(false)

    expect(
      await screen.findByRole('button', { name: 'Save model prices' })
    ).toBeEnabled()
    expect(
      screen.getByRole('button', { name: 'Restore default pricing' })
    ).toBeEnabled()
    expect(
      screen.queryByText('Sell prices are published by dynamic pricing')
    ).not.toBeInTheDocument()
  })
})
