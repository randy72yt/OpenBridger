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
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { ThemeProvider } from '@/context/theme-provider'
import { useSystemConfigStore } from '@/stores/system-config-store'

import { LandingPage } from '../index'

vi.mock('@lobehub/icons', () => {
  const IconStub = () => null
  const WithColor = Object.assign(IconStub, { Color: IconStub })
  return {
    OpenAI: WithColor,
    Claude: WithColor,
    Gemini: WithColor,
    DeepSeek: WithColor,
    Qwen: WithColor,
    Grok: WithColor,
    Zhipu: WithColor,
    Minimax: WithColor,
    Moonshot: WithColor,
    Doubao: WithColor,
    Kling: WithColor,
    Jimeng: WithColor,
    ByteDance: WithColor,
  }
})

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

vi.mock('@/features/home/hooks', () => ({
  useHomePageContent: () => ({
    content: '',
    isLoaded: true,
    isUrl: false,
  }),
}))

const pricingData = vi.hoisted(() => ({
  models: [] as Array<Record<string, unknown>>,
  vendors: [] as unknown[],
  groupRatio: {} as Record<string, number>,
  usableGroup: {} as Record<string, string>,
  supportedEndpoint: {} as Record<string, unknown>,
  autoGroups: [] as string[],
  isLoading: false,
  error: null as Error | null,
}))

vi.mock('@/features/pricing/hooks/use-pricing-data', () => ({
  usePricingData: () => pricingData,
}))

const originalConfig = useSystemConfigStore.getState()

function mockViewportAndMotion(options: {
  isDesktop: boolean
  reducedMotion: boolean
}) {
  const originalMatchMedia = window.matchMedia
  vi.spyOn(window, 'matchMedia').mockImplementation((query: string) => {
    const base = originalMatchMedia(query)
    if (query === '(min-width: 768px)') {
      return { ...base, matches: options.isDesktop }
    }
    if (query === '(prefers-reduced-motion: reduce)') {
      return { ...base, matches: options.reducedMotion }
    }
    return base
  })
}

function renderLandingPage() {
  const root = createRootRoute()
  const home = createRoute({
    getParentRoute: () => root,
    path: '/',
    component: LandingPage,
  })
  const pricing = createRoute({
    getParentRoute: () => root,
    path: '/pricing',
    component: () => <h1>Model Square Page</h1>,
  })
  const signUp = createRoute({
    getParentRoute: () => root,
    path: '/sign-up',
    component: () => <h1>Sign Up Page</h1>,
  })
  const signIn = createRoute({
    getParentRoute: () => root,
    path: '/sign-in',
    component: () => <h1>Sign In Page</h1>,
  })
  const docs = createRoute({
    getParentRoute: () => root,
    path: '/docs',
    component: () => <h1>Docs Page</h1>,
  })
  const about = createRoute({
    getParentRoute: () => root,
    path: '/about',
    component: () => <h1>About Page</h1>,
  })
  const userAgreement = createRoute({
    getParentRoute: () => root,
    path: '/user-agreement',
    component: () => <h1>User Agreement Page</h1>,
  })
  const privacyPolicy = createRoute({
    getParentRoute: () => root,
    path: '/privacy-policy',
    component: () => <h1>Privacy Policy Page</h1>,
  })
  const dashboard = createRoute({
    getParentRoute: () => root,
    path: '/dashboard',
    component: () => <h1>Dashboard Page</h1>,
  })

  const router = createRouter({
    routeTree: root.addChildren([
      home,
      pricing,
      signUp,
      signIn,
      docs,
      about,
      userAgreement,
      privacyPolicy,
      dashboard,
    ]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })

  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  client.setQueryData(['status'], {
    docs_link: '/docs',
    server_address: 'https://api.openbridger.com',
  })

  return {
    client,
    ...render(
      <ThemeProvider>
        <QueryClientProvider client={client}>
          <RouterProvider router={router} />
        </QueryClientProvider>
      </ThemeProvider>
    ),
  }
}

beforeEach(() => {
  localStorage.clear()
  pricingData.models = []
  pricingData.groupRatio = {}
  pricingData.isLoading = false
  pricingData.error = null
  useSystemConfigStore
    .getState()
    .setConfig({ systemName: 'OpenBridger', footerHtml: '' })
})

afterEach(() => {
  useSystemConfigStore.setState(originalConfig)
  localStorage.clear()
  vi.restoreAllMocks()
})

describe('LandingPage (P0)', { timeout: 15000 }, () => {
  it('uses CSS fallback without mounting 3D scene on mobile viewports', async () => {
    mockViewportAndMotion({ isDesktop: false, reducedMotion: false })
    const { client } = renderLandingPage()

    expect(
      await screen.findByTestId('landing-hero-css-fallback')
    ).toBeInTheDocument()
    expect(
      screen.queryByTestId('landing-hero-3d-canvas')
    ).not.toBeInTheDocument()
    client.clear()
  })

  it('uses CSS fallback without mounting 3D scene when prefers-reduced-motion is enabled on desktop', async () => {
    mockViewportAndMotion({ isDesktop: true, reducedMotion: true })
    const { client } = renderLandingPage()

    expect(
      await screen.findByTestId('landing-hero-css-fallback')
    ).toBeInTheDocument()
    expect(
      screen.queryByTestId('landing-hero-3d-canvas')
    ).not.toBeInTheDocument()
    client.clear()
  })

  it('renders API relay sections, supports code language switching, and preserves mandatory New API attribution', async () => {
    mockViewportAndMotion({ isDesktop: false, reducedMotion: true })
    const { client } = renderLandingPage()

    // 1. Multi-model hub section
    expect(
      await screen.findByRole('heading', { name: 'Anthropic Claude' })
    ).toBeVisible()
    expect(screen.getByRole('heading', { name: 'DeepSeek' })).toBeVisible()

    // Scoped warm-orange theme wrapper is mounted
    expect(document.querySelector('.ob-landing-theme')).toBeInTheDocument()

    // 2. Pricing preview has no fallback prices when the catalog is empty
    expect(screen.getAllByText('claude-sonnet-4-6').length).toBeGreaterThan(0)
    expect(screen.getByText('Prices are unavailable right now.')).toBeVisible()
    expect(screen.queryByText('$3.00')).not.toBeInTheDocument()
    expect(screen.queryByText('$15.00')).not.toBeInTheDocument()

    // 3. Model health has no fabricated sample
    expect(screen.getByText('No sample data yet.')).toBeVisible()
    expect(screen.queryByText('99.94%')).not.toBeInTheDocument()
    expect(screen.queryByText('420ms')).not.toBeInTheDocument()
    expect(screen.queryByText('200+')).not.toBeInTheDocument()
    expect(screen.getByText('Example')).toBeVisible()
    expect(screen.queryByText('landing.terminal.cost')).not.toBeInTheDocument()

    // 4. 3-line code integration tabs
    const tsTab = screen.getByRole('tab', { name: 'TypeScript / Node.js' })
    await userEvent.click(tsTab)
    expect(
      screen.getByText(
        (_, element) =>
          element?.tagName.toLowerCase() === 'code' &&
          element.textContent?.includes("import OpenAI from 'openai'") === true
      )
    ).toBeVisible()

    // 5. Mandatory NOTICE attribution in footer
    expect(
      screen.getByText(
        /Frontend design and development by New API contributors\./
      )
    ).toBeVisible()
    const newApiLink = screen.getByRole('link', { name: 'New API' })
    expect(newApiLink).toHaveAttribute(
      'href',
      'https://github.com/QuantumNous/new-api'
    )

    // 6. Excludes video-creation / lottery / invoice copy
    expect(
      screen.queryByText(
        /快跑|快跑影创|快跑画布|一镜到底|短剧|抽奖|开票|返现加盟/
      )
    ).not.toBeInTheDocument()
    const homeLink = screen.getAllByRole('link', { name: 'Home' })[0]
    expect(homeLink?.className).toContain('rounded-full')
    expect(homeLink?.className).toContain('bg-[var(--ob-primary-soft)]')
    expect(homeLink?.className).not.toContain('underline')

    client.clear()
  })

  it('shows catalog prices from the pricing API and hides them when that request fails', async () => {
    mockViewportAndMotion({ isDesktop: false, reducedMotion: true })
    pricingData.groupRatio = { default: 1 }
    pricingData.models = [
      {
        model_name: 'catalog-model',
        vendor_name: 'Catalog',
        quota_type: 0,
        model_ratio: 1.11,
        completion_ratio: 2,
        cache_ratio: 0.5,
        enable_groups: ['default'],
        tags: 'Tools',
      },
    ]
    const loaded = renderLandingPage()
    expect(await screen.findByText('catalog-model')).toBeVisible()
    expect(screen.getByText(/\$2\.22/)).toBeVisible()
    expect(screen.queryByText('Prices are unavailable right now.')).not.toBeInTheDocument()
    expect(screen.queryByText('$3.00')).not.toBeInTheDocument()
    loaded.client.clear()
    cleanup()

    pricingData.models = []
    pricingData.error = new Error('pricing down')
    const failed = renderLandingPage()
    expect(
      await screen.findByText('Prices are unavailable right now.')
    ).toBeVisible()
    expect(screen.queryByText('catalog-model')).not.toBeInTheDocument()
    expect(screen.queryByText(/\$2\.22/)).not.toBeInTheDocument()
    expect(screen.queryByText('$3.00')).not.toBeInTheDocument()
    failed.client.clear()
  })
})
