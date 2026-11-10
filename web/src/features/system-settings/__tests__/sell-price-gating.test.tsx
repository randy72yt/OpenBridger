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
import { useState, type ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { pricingOptions } from '@/features/model-pricing/pricing'
import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { SettingsPageProvider } from '../components/settings-page-context'
import { RatioSettingsCard } from '../models/ratio-settings-card'

const modelDefaults = {
  ...pricingOptions({}),
  ModelPrice: '{}',
  ModelRatio: '{}',
  CacheRatio: '{}',
  CreateCacheRatio: '{}',
  CompletionRatio: '{}',
  ImageRatio: '{}',
  AudioRatio: '{}',
  AudioCompletionRatio: '{}',
  ExposeRatioEnabled: false,
  BillingMode: '{}',
  BillingExpr: '{}',
}

const groupDefaults = {
  GroupRatio: '{}',
  TopupGroupRatio: '{}',
  UserUsableGroups: '{}',
  GroupGroupRatio: '{}',
  AutoGroups: '[]',
  MaxTokenAutoGroups: 5,
  DefaultUseAutoGroup: false,
  GroupSpecialUsableGroup: '{}',
}

afterEach(() => {
  useAuthStore.getState().auth.reset()
  window.localStorage.clear()
  vi.restoreAllMocks()
})

function TabHost(props: { children: ReactNode }) {
  const [title, setTitle] = useState<HTMLSpanElement | null>(null)
  return (
    <>
      <span ref={setTitle} />
      <SettingsPageProvider
        actionsContainer={null}
        titleStatusContainer={title}
      >
        {title ? props.children : null}
      </SettingsPageProvider>
    </>
  )
}

async function renderCard(
  pricingControlEnabled: boolean,
  visibleTabs: Array<
    'models' | 'unset-models' | 'groups' | 'tool-prices' | 'upstream-sync'
  >
) {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'root', role: 100 })
  vi.spyOn(api, 'get').mockImplementation(async (url) => {
    if (url === '/api/status') {
      return {
        data: { data: { pricing_control_enabled: pricingControlEnabled } },
      }
    }
    if (url === '/api/option/model_pricing') {
      return {
        data: {
          success: true,
          data: {
            entries: [],
            options: pricingOptions({}),
            empty_version: 'empty',
          },
        },
      }
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
          <TabHost>
            <RatioSettingsCard
              modelDefaults={modelDefaults}
              groupDefaults={groupDefaults}
              toolPricesDefault='{}'
              visibleTabs={visibleTabs}
            />
          </TabHost>
        </QueryClientProvider>
      ),
    }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  await router.load()
  return render(<RouterProvider router={router} />)
}

describe('billing sell-price gating', () => {
  it('hides upstream price sync and model price saves when dynamic pricing control is on', async () => {
    await renderCard(true, [
      'models',
      'unset-models',
      'tool-prices',
      'upstream-sync',
    ])

    expect(
      await screen.findByText('Sell prices are published by dynamic pricing')
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('tab', { name: 'Upstream price sync' })
    ).not.toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Tool prices' })).toBeInTheDocument()
    expect(
      screen.getByText('Sell prices are published by dynamic pricing')
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Reset prices' })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Save model prices' })
    ).not.toBeInTheDocument()
  })

  it('keeps upstream price sync when dynamic pricing control is off', async () => {
    await renderCard(false, [
      'models',
      'unset-models',
      'tool-prices',
      'upstream-sync',
    ])

    expect(
      await screen.findByRole('button', { name: 'Reset prices' })
    ).toBeEnabled()
    expect(
      screen.getByRole('tab', { name: 'Upstream price sync' })
    ).toBeInTheDocument()
    expect(
      screen.queryByText('Sell prices are published by dynamic pricing')
    ).not.toBeInTheDocument()
  })

  it('labels group ratio as a second multiplier on the published sell price', async () => {
    await renderCard(true, ['groups'])

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Group ratio is a second multiplier'
    )
    expect(
      screen.getByText(
        'It scales the published model sell price for a user group and does not replace that sell price.'
      )
    ).toBeInTheDocument()
  })
})
