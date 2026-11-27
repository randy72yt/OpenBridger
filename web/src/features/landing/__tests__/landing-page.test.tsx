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
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { ThemeProvider } from '@/context/theme-provider'
import { useAuthStore } from '@/stores/auth-store'
import { useSystemConfigStore } from '@/stores/system-config-store'

import { LandingPage } from '../index'
import { deriveLandingShowcase } from '../lib/live-models'

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

vi.mock('@/features/pricing/hooks/use-pricing-data', () => ({
  usePricingData: () => ({
    models: [],
    vendors: [],
    groupRatio: {},
    usableGroup: {},
    supportedEndpoint: {},
    autoGroups: [],
    isLoading: false,
    error: null,
  }),
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
  useSystemConfigStore
    .getState()
    .setConfig({ systemName: 'OpenBridger', footerHtml: '' })
})

afterEach(() => {
  useAuthStore.getState().auth.setUser(null)
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

  it.each([
    { signedIn: false, destination: '/sign-up' },
    { signedIn: true, destination: '/keys' },
  ])(
    'routes both API key actions to $destination',
    async ({ signedIn, destination }) => {
      mockViewportAndMotion({ isDesktop: false, reducedMotion: true })
      if (signedIn) {
        useAuthStore
          .getState()
          .auth.setUser({ id: 1, username: 'tester', role: 1 })
      }
      const { client } = renderLandingPage()
      const actions = await screen.findAllByRole('button', {
        name: 'landing.hero.ctaPrimary',
      })
      expect(actions).toHaveLength(2)
      for (const action of actions) {
        expect(action).toHaveAttribute('href', destination)
      }
      client.clear()
    }
  )

  it('renders API relay sections, supports code language switching, and preserves mandatory New API attribution', async () => {
    mockViewportAndMotion({ isDesktop: false, reducedMotion: true })
    const { client } = renderLandingPage()

    // 1. Multi-model hub section + latest frontier model names + HOT partner badge
    expect(
      await screen.findByRole('heading', { name: 'OpenAI · Codex' })
    ).toBeVisible()
    expect(
      screen.getByRole('heading', { name: 'Anthropic Claude' })
    ).toBeVisible()
    expect(screen.getByRole('heading', { name: 'DeepSeek' })).toBeVisible()
    expect(screen.getAllByText('HOT').length).toBeGreaterThanOrEqual(1)
    expect(
      screen.getAllByText('gpt-6.2-astra').length
    ).toBeGreaterThanOrEqual(1)
    expect(
      screen.getAllByText('claude-opus-5-5').length
    ).toBeGreaterThanOrEqual(1)
    expect(
      screen.getAllByText('gemini-4-neon').length
    ).toBeGreaterThanOrEqual(1)
    expect(
      screen.getAllByText('deepseek-v4.1').length
    ).toBeGreaterThanOrEqual(1)

    // Scoped warm-orange theme wrapper is mounted
    expect(document.querySelector('.ob-landing-theme')).toBeInTheDocument()

    // 2. Pricing preview section
    expect(screen.getByText('No Models Found')).toBeVisible()

    // 3. Real-time model status section
    expect(screen.getByText('99.94%')).toBeVisible()

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
      screen.queryByText(/快跑影创|快跑画布|一镜到底/)
    ).not.toBeInTheDocument()

    client.clear()
  })

  it('automatically picks newest model versions from live channel models and honors admin tagline override', () => {
    const showcase = deriveLandingShowcase([
      { model_name: 'gpt-4o' },
      { model_name: 'gpt-6.2-sol' },
      { model_name: 'gpt-7.1-astra', enable_groups: ['vip-codex'] },
      { model_name: 'gpt-7.1-astra-2027-02-01' },
      { model_name: 'text-embedding-3-large' },
      { model_name: 'claude-3-5-sonnet-20241022' },
      { model_name: 'claude-opus-6-2' },
      { model_name: 'claude-opus-6-2-high' },
      { model_name: 'gemini-2.5-pro' },
      { model_name: 'gemini-5-carbon' },
      { model_name: 'deepseek-v3' },
      { model_name: 'deepseek-v5.2-flash' },
      { model_name: 'qwen3-235b-a22b' },
      { model_name: 'qwen5-max' },
    ])

    const openaiCard = showcase.providers.find((p) => p.id === 'openai')
    const claudeCard = showcase.providers.find((p) => p.id === 'anthropic')
    const qwenCard = showcase.providers.find((p) => p.id === 'qwen')
    expect(openaiCard?.models[0]).toBe('gpt-7.1-astra')
    expect(openaiCard?.models).not.toContain('text-embedding-3-large')
    // Deduplicates dated snapshot and effort variants of the same model
    expect(openaiCard?.models).not.toContain('gpt-7.1-astra-2027-02-01')
    expect(claudeCard?.models).toEqual([
      'claude-opus-6-2',
      'claude-3-5-sonnet-20241022',
    ])
    // 235b parameter count does not outrank Qwen 5 generation
    expect(qwenCard?.models[0]).toBe('qwen5-max')

    expect(showcase.heroModelsTagline).toBe(
      'GPT-7.1 · Codex · Claude 6.2 · Gemini 5 · DeepSeek V5.2'
    )
    expect(showcase.terminalModels.chatLabel).toBe('GPT-7.1')
    expect(showcase.terminalModels.claudeLabel).toBe('Claude 6.2')
    expect(showcase.healthItems[0]).toMatchObject({
      modelName: 'gpt-7.1-astra',
      groupName: 'vip-codex',
    })
    expect(
      showcase.getSnippets('https://api.openbridger.com')[0]?.code
    ).toContain('model="gpt-7.1-astra"')

    const overridden = deriveLandingShowcase(
      [{ model_name: 'gpt-7.1-astra' }],
      '  GPT-7.1 满血直连 · 校园专享  '
    )
    expect(overridden.heroModelsTagline).toBe('GPT-7.1 满血直连 · 校园专享')
  })
})
