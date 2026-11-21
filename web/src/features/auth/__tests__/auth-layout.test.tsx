import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createRootRoute,
  createRouter,
  createMemoryHistory,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { ThemeProvider } from '@/context/theme-provider'

import { AuthLayout } from '../auth-layout'

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

it('keeps form content available and limits the promotional panel to large screens', async () => {
  const route = createRootRoute({
    component: () => (
      <AuthLayout showcase>
        <button type='button'>Form action</button>
      </AuthLayout>
    ),
  })
  const router = createRouter({
    routeTree: route,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <ThemeProvider>
      <QueryClientProvider client={client}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </ThemeProvider>
  )
  expect(
    await screen.findByRole('button', { name: 'Form action' })
  ).toBeVisible()
  for (const label of [
    'From inspiration to creation, make AI your creative power.',
    'Access multiple AI models with one account for coding assistance, learning and exploration, and application development.',
  ]) {
    expect(screen.getByText(label)).toBeInTheDocument()
  }
  expect(
    screen.getByText('AI application infrastructure').closest('section')
  ).toHaveClass('hidden', 'lg:block')
})
