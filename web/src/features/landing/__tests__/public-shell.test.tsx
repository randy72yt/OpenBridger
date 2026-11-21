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
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { AuthenticatedLayout } from '@/components/layout/components/authenticated-layout'
import { ThemeProvider } from '@/context/theme-provider'

import { PublicShell } from '../components/public-shell'

vi.mock('@/components/layout/components/app-sidebar', () => ({
  AppSidebar: () => <div>Console sidebar</div>,
}))

vi.mock('@/components/layout/components/app-header', () => ({
  AppHeader: () => <div>Console header</div>,
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

function renderInRouter(ui: ReactNode) {
  const root = createRootRoute({
    component: () => ui,
    errorComponent: ({ error }) => <pre>{error.message}</pre>,
  })
  const router = createRouter({
    routeTree: root,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <ThemeProvider>
      <QueryClientProvider client={client}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </ThemeProvider>
  )
}

describe('PublicShell', () => {
  it('renders the shared warm theme, header, and New API attribution', async () => {
    renderInRouter(
      <PublicShell>
        <h1>Public page</h1>
      </PublicShell>
    )

    expect(
      await screen.findByRole('heading', { name: 'Public page' })
    ).toBeVisible()
    expect(document.querySelector('.ob-landing-theme')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'New API' })).toHaveAttribute(
      'href',
      'https://github.com/QuantumNous/new-api'
    )
  })

  it('does not apply the public landing theme to the console layout', async () => {
    renderInRouter(
      <AuthenticatedLayout>
        <h1>Console page</h1>
      </AuthenticatedLayout>
    )

    expect(
      await screen.findByRole('heading', { name: 'Console page' })
    ).toBeVisible()
    expect(screen.getByText('Console header')).toBeVisible()
    expect(document.querySelector('.ob-landing-theme')).not.toBeInTheDocument()
  })
})
