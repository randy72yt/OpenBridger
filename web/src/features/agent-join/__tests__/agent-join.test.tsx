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
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { ThemeProvider } from '@/context/theme-provider'
import { useSystemConfigStore } from '@/stores/system-config-store'

import { AgentJoinPage } from '../index'

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

const originalConfig = useSystemConfigStore.getState()

beforeEach(() => {
  localStorage.clear()
  useSystemConfigStore
    .getState()
    .setConfig({ systemName: 'OpenBridger', footerHtml: '' })
})

afterEach(() => {
  useSystemConfigStore.setState(originalConfig)
  localStorage.clear()
  vi.restoreAllMocks()
})

function renderAgentJoinWithStatus(statusPayload: Record<string, unknown>) {
  const root = createRootRoute()
  const agentJoin = createRoute({
    getParentRoute: () => root,
    path: '/agent-join',
    component: AgentJoinPage,
  })
  const router = createRouter({
    routeTree: root.addChildren([agentJoin]),
    history: createMemoryHistory({ initialEntries: ['/agent-join'] }),
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  client.setQueryData(['status'], statusPayload)

  const utils = render(
    <ThemeProvider>
      <QueryClientProvider client={client}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </ThemeProvider>
  )

  return { ...utils, client }
}

describe('AgentJoinPage', { timeout: 15000 }, () => {
  it('renders default support@openbridger.com and QR placeholder when WeChat is unconfigured', async () => {
    const { client } = renderAgentJoinWithStatus({
      docs_link: '/docs',
      server_address: 'https://openbridger.com',
    })

    expect(
      await screen.findByText('agentJoin.hero.badge')
    ).toBeInTheDocument()
    expect(document.querySelector('.ob-landing-theme')).toBeInTheDocument()
    expect(
      screen.getAllByText('support@openbridger.com').length
    ).toBeGreaterThanOrEqual(1)
    expect(
      screen.getByTestId('partner-wechat-qr-placeholder')
    ).toBeInTheDocument()

    for (const stepNum of ['01', '02', '03', '04']) {
      expect(screen.getByText(stepNum)).toBeInTheDocument()
    }

    expect(
      screen.queryByText(/快跑|Kokolo779|一镜到底/)
    ).not.toBeInTheDocument()

    client.clear()
  })

  it('renders hot-configured WeChat ID and uploaded WeChat QR image from /api/status', async () => {
    const qrDataUrl =
      'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='
    const { client } = renderAgentJoinWithStatus({
      contact_email: 'support@openbridger.com',
      contact_wechat_id: 'OpenBridger_BD',
      contact_wechat_qrcode: qrDataUrl,
    })

    const wechatMatches = await screen.findAllByText('OpenBridger_BD')
    expect(wechatMatches.length).toBeGreaterThanOrEqual(1)
    const qrBox = screen.getByTestId('partner-wechat-qr')
    const qrImg = qrBox.querySelector('img')
    expect(qrImg).toHaveAttribute('src', qrDataUrl)

    client.clear()
  })
})
