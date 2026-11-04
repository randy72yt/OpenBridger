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
import { render, screen, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { PricingControlSection } from '../pricing-control-section'

vi.mock('../pricing-control-api', () => ({
  getPricingPolicies: vi.fn(async () => [
    {
      id: 1,
      public_model: '*',
      service_tier: 'default',
      scope: 'global',
      primary_channel_id: 0,
      backup_channel_id: 0,
      target_margin_bps: 3500,
      minimum_margin_bps: 2000,
      overhead_bps: 500,
      fallback_probability_bps: 1000,
      auto_publish_change_bps: 0,
      price_locked: false,
      enabled: true,
    },
  ]),
  getPricingRisks: vi.fn(async () => [
    {
      public_model: 'gpt-5.4',
      service_tier: 'default',
      code: 'primary_offer_missing_or_expired',
      channel_id: 7,
      channel_role: 'primary',
      expires_at: 1,
      suggested_action: 'resync_offers',
    },
  ]),
  getPricingSyncRuns: vi.fn(async () => ({ items: [], latest: null })),
  getPricingOffers: vi.fn(async () => ({
    items: [
      {
        id: 1,
        channel_id: 7,
        upstream_model: 'gpt-5.4',
        public_model: 'gpt-5.4',
        input_cost: 1,
        output_cost: 2,
        cache_read_cost: 0,
        upstream_group_ratio: 1.8,
        upstream_group: 'claude-max',
        currency: 'USD',
        source_type: 'new-api-pricing',
        source_version: 'v1',
        collected_at: 1,
        expires_at: 4102444800,
        enabled: true,
      },
    ],
    total: 1,
    limit: 50,
    offset: 0,
  })),
  getPricingProposals: vi.fn(async () => ({
    items: [],
    total: 0,
    limit: 50,
    offset: 0,
  })),
  importPricingOffers: vi.fn(),
  syncPricingOffers: vi.fn(),
  savePricingPolicy: vi.fn(),
  recalculatePricing: vi.fn(),
  updatePricingProposal: vi.fn(),
}))

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('pricing control section', () => {
  it('renders the summary cards, the global policy, and the inventory', async () => {
    render(<PricingControlSection />, { wrapper })

    await waitFor(() =>
      expect(screen.getByText('Global policy')).toBeInTheDocument()
    )

    expect(screen.getByText('Current upstream offers')).toBeInTheDocument()
    expect(screen.getByText('Active pricing policies')).toBeInTheDocument()
    expect(screen.getByText('Pending proposals')).toBeInTheDocument()
    expect(screen.getByText('Pricing risks')).toBeInTheDocument()

    // The inventory shows base cost, the settlement ratio, and the actual cost.
    expect(screen.getByText('$1.8')).toBeInTheDocument()
    expect(screen.getByText('claude-max')).toBeInTheDocument()
    expect(screen.getByText('No valid upstream offer')).toBeInTheDocument()
  })
})
