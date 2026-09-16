import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { Features } from '../features'
import { Hero } from '../hero'
import { HowItWorks } from '../how-it-works'
import { Stats } from '../stats'

beforeEach(() => {
  vi.spyOn(window, 'scrollTo').mockImplementation(() => undefined)
  const matchMedia = window.matchMedia
  vi.spyOn(window, 'matchMedia').mockImplementation((query) => ({
    ...matchMedia(query),
    matches: query === '(prefers-reduced-motion: reduce)',
  }))
})

describe('marketing homepage', () => {
  it('shows agent tools without outbound links and retains model discovery', async () => {
    const root = createRootRoute()
    const home = createRoute({
      getParentRoute: () => root,
      path: '/',
      component: Hero,
    })
    const pricing = createRoute({
      getParentRoute: () => root,
      path: '/pricing',
      component: () => <h1>Model catalog</h1>,
    })
    const router = createRouter({
      routeTree: root.addChildren([home, pricing]),
      history: createMemoryHistory({ initialEntries: ['/'] }),
    })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    client.setQueryData(['status'], { docs_link: '/docs' })
    render(
      <QueryClientProvider client={client}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    )
    const link = await screen.findByRole('button', {
      name: 'View Pricing',
    })
    expect(link).toHaveAttribute('href', '/pricing')
    const tools = screen.getByRole('list', { name: 'Supported Applications' })
    for (const name of ['Codex', 'Claude Code', 'DeepSeek Harness']) {
      expect(within(tools).getByText(name, { selector: 'span' })).toBeVisible()
    }
    expect(within(tools).queryByRole('link')).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Practical tutorials' })
    ).toHaveAttribute('href', '/docs')

    expect(screen.queryByText('Cherry Studio')).not.toBeInTheDocument()
    expect(screen.queryByText('CC Switch')).not.toBeInTheDocument()
    await userEvent.click(link)
    expect(
      await screen.findByRole('heading', { name: 'Model catalog' })
    ).toBeVisible()
    client.clear()
  })

  it('describes learning and development use cases', () => {
    render(<Features />)
    expect(
      screen.getByRole('heading', { name: 'Learning and course projects' })
    ).toBeVisible()
    expect(
      screen.getByRole('heading', {
        name: 'Development and application integration',
      })
    ).toBeVisible()
  })

  it('constrains section separators instead of drawing full-width rules', () => {
    render(
      <>
        <Stats />
        <HowItWorks />
      </>
    )
    const separators = screen.getAllByRole('separator')
    expect(separators).toHaveLength(3)
    for (const separator of separators) {
      expect(separator).toHaveClass('max-w-[min(36rem,80%)]')
    }
  })
})
