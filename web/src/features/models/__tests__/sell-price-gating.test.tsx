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
import { afterEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { ModelsPrimaryButtons } from '../components/models-primary-buttons'
import { ModelsProvider } from '../components/models-provider'

afterEach(() => {
  useAuthStore.getState().auth.reset()
  window.localStorage.clear()
  vi.restoreAllMocks()
})

function renderButtons() {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'root', role: 100 })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={client}>
      <ModelsProvider>
        <ModelsPrimaryButtons />
      </ModelsProvider>
    </QueryClientProvider>
  )
}

describe('models page sell-price gating', () => {
  it('hides Sync pricing and keeps Sync metadata when dynamic pricing control is on', async () => {
    vi.spyOn(api, 'get').mockResolvedValue({
      data: { data: { pricing_control_enabled: true } },
    })

    renderButtons()

    expect(
      await screen.findByRole('button', { name: 'Sync metadata' })
    ).toBeEnabled()
    await waitFor(() => {
      expect(
        screen.queryByRole('button', { name: 'Sync pricing' })
      ).not.toBeInTheDocument()
    })
  })

  it('shows Sync pricing when dynamic pricing control is off', async () => {
    vi.spyOn(api, 'get').mockResolvedValue({
      data: { data: { pricing_control_enabled: false } },
    })

    renderButtons()

    expect(
      await screen.findByRole('button', { name: 'Sync pricing' })
    ).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Sync metadata' })).toBeEnabled()
  })
})
