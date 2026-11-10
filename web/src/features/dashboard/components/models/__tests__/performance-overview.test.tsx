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
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'

import { PerformanceOverview } from '../performance-overview'

function adminModel(overrides: Record<string, unknown> = {}) {
  return {
    model_name: 'example-model',
    real: { requests: 3, success: 3, avg_latency_ms: 900, avg_ttft_ms: 300, avg_tps: 40 },
    probe: { requests: 2, success: 2, avg_latency_ms: 800, avg_ttft_ms: 250, avg_tps: 30 },
    internal: { requests: 1, success: 1, avg_latency_ms: 0, avg_ttft_ms: 0, avg_tps: 0 },
    client_error: { requests: 4, success: 0, avg_latency_ms: 0, avg_ttft_ms: 0, avg_tps: 0 },
    public_v1: { model_name: 'example-model', avg_latency_ms: 900, success_rate: 100, avg_tps: 40 },
    public_v2: {
      model_name: 'example-model',
      avg_latency_ms: 860,
      success_rate: 100,
      avg_tps: 36,
      availability_24h: 100,
      sample_count: 5,
      insufficient: false,
    },
    availability_24h: 100,
    availability_7d: 99.5,
    error_codes: { 'channel:invalid_key': 2 },
    series: [],
    last_probe_at: 0,
    probe_tier: 'cheap',
    probe_token_param: 'max_tokens',
    ...overrides,
  }
}

let queryClient: QueryClient

beforeEach(() => {
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  })
})

afterEach(() => {
  queryClient.clear()
  vi.restoreAllMocks()
})

describe('PerformanceOverview (admin)', () => {
  it('renders the admin table with source split, availability and probe info', async () => {
    vi.spyOn(api, 'get').mockResolvedValue({
      data: {
        success: true,
        data: {
          models: [adminModel()],
          probe_budget_used_quota: 100000,
          probe_budget_quota: 500000,
          generated_at: 1794300000,
        },
      },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <PerformanceOverview />
      </QueryClientProvider>
    )

    await waitFor(() =>
      expect(screen.getByText('example-model')).toBeVisible()
    )
    expect(screen.getByText('3/2/1/4')).toBeVisible()
    expect(screen.getAllByText('100.00%').length).toBeGreaterThan(0)
    expect(screen.getByText(/99\.50%/)).toBeVisible()
    expect(screen.getByText('channel:invalid_key ×2')).toBeVisible()
    expect(screen.getByText(/cheap · —/)).toBeVisible()
    expect(screen.getByText(/Probe budget/)).toBeVisible()
  })

  it('keeps the empty state when the admin summary has no models', async () => {
    vi.spyOn(api, 'get').mockResolvedValue({
      data: {
        success: true,
        data: {
          models: [],
          probe_budget_used_quota: 0,
          probe_budget_quota: 500000,
          generated_at: 1794300000,
        },
      },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <PerformanceOverview />
      </QueryClientProvider>
    )

    await waitFor(() =>
      expect(screen.getByText('No performance data available')).toBeVisible()
    )
  })
})
