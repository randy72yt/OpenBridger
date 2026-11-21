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
import { describe, expect, it, vi } from 'vitest'

import { ModelDetailsApi } from '../components/model-details-api'
import { ModelDetailsApps } from '../components/model-details-apps'
import type { PricingModel } from '../types'

vi.mock('@/lib/api', () => ({
  getStatus: vi.fn().mockResolvedValue(null),
}))

const model: PricingModel = {
  id: 1,
  model_name: 'catalog-model',
  quota_type: 0,
  model_ratio: 1,
  completion_ratio: 1,
  enable_groups: ['default', 'vip'],
}

function renderDetails() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={client}>
      <ModelDetailsApps model={model} />
      <ModelDetailsApi model={model} endpointMap={{}} />
    </QueryClientProvider>
  )
}

describe('model detail sources without fabricated operations data', () => {
  it('does not invent app rankings or rate limits', () => {
    renderDetails()

    expect(screen.getByText('catalog-model')).toBeVisible()
    expect(
      screen.getByText('No app usage data available for this model.')
    ).toBeVisible()
    expect(screen.getByText('No rate-limit data yet.')).toBeVisible()
    expect(screen.queryByText('Cline')).not.toBeInTheDocument()
    expect(screen.queryByText('RPM')).not.toBeInTheDocument()
    expect(screen.queryByText('99.94%')).not.toBeInTheDocument()
  })
})
