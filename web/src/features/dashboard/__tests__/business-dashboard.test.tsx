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
import { render, screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { ThemeProvider } from '@/context/theme-provider'
import { getAdminBusinessDashboard } from '@/features/dashboard/api'

import { BusinessDashboard } from '../components/business/business-dashboard'

vi.mock('@/features/dashboard/api', () => ({
  getAdminBusinessDashboard: vi.fn(),
}))

vi.mock('@visactor/react-vchart', () => ({
  VChart: () => <div data-testid='mock-vchart' />,
}))

vi.mock('@/lib/use-chart-theme', () => ({
  useChartTheme: () => ({ resolvedTheme: 'light', themeReady: true }),
}))

vi.mock('@/lib/lobe-icon', () => ({
  getLobeIcon: (iconName: string) => <span data-lobe-icon={iconName} />,
}))

describe('BusinessDashboard', () => {
  test('renders executive KPIs, payment channel breakdown, and model gross margins', async () => {
    vi.mocked(getAdminBusinessDashboard).mockResolvedValueOnce({
      success: true,
      data: {
        summary: {
          cash_revenue_usd: 145.5,
          topup_revenue_usd: 115.5,
          subscription_revenue_usd: 30,
          paid_order_count: 8,
          paying_users: 5,
          arppu_usd: 29.1,
          redemption_usd: 20,
          redemption_count: 2,
          consumption_usd: 60,
          costed_consumption_usd: 50,
          uncosted_consumption_usd: 10,
          estimated_cost_usd: 22.5,
          estimated_gross_profit_usd: 27.5,
          estimated_gross_margin_pct: 55,
          cost_coverage_pct: 83.33,
          total_requests: 1200,
          total_tokens: 4500000,
          total_users: 48,
          new_users: 12,
          active_users: 19,
          wallet_liability_usd: 310.25,
          channel_balance_usd: 890,
        },
        daily_trend: [
          {
            date: '2026-11-25',
            timestamp: 1764028800,
            cash_revenue_usd: 145.5,
            redemption_usd: 20,
            consumption_usd: 60,
            estimated_cost_usd: 22.5,
            new_users: 12,
            paid_orders: 8,
          },
        ],
        payment_channels: [
          {
            channel: 'stripe',
            provider: 'stripe',
            order_type: 'topup',
            order_count: 5,
            original_currency: 'USD',
            original_amount: 100,
            revenue_usd: 100,
          },
          {
            channel: 'epay:alipay',
            provider: 'epay',
            order_type: 'topup',
            order_count: 3,
            original_currency: 'CNY',
            original_amount: 111.6,
            revenue_usd: 15.5,
          },
        ],
        model_margins: [
          {
            model_name: 'gpt-4o',
            channel_id: 1,
            channel_name: 'OpenAI Direct',
            requests: 900,
            prompt_tokens: 2000000,
            completion_tokens: 1000000,
            consumption_usd: 50,
            estimated_cost_usd: 22.5,
            gross_profit_usd: 27.5,
            gross_margin_pct: 55,
            cost_covered: true,
          },
          {
            model_name: 'custom-v1',
            channel_id: 2,
            channel_name: 'Relay Pool',
            requests: 300,
            prompt_tokens: 1000000,
            completion_tokens: 500000,
            consumption_usd: 10,
            estimated_cost_usd: null,
            gross_profit_usd: null,
            gross_margin_pct: null,
            cost_covered: false,
          },
        ],
      },
    })

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })

    render(
      <ThemeProvider>
        <QueryClientProvider client={client}>
          <BusinessDashboard />
        </QueryClientProvider>
      </ThemeProvider>
    )

    expect(
      await screen.findByTestId('business-kpi-grid')
    ).toBeInTheDocument()
    expect(screen.getByText('$145.50')).toBeInTheDocument()
    expect(screen.getByText('$310.25')).toBeInTheDocument()
    expect(screen.getByText('epay:alipay')).toBeInTheDocument()
    expect(screen.getByText('gpt-4o')).toBeInTheDocument()
    expect(screen.getAllByText('55.0%').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('business.uncosted')).toBeInTheDocument()

    client.clear()
  })
})
