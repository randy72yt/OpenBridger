import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { useSystemConfigStore } from '@/stores/system-config-store'

import { Footer } from '../components/footer'

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
})

describe('public footer', () => {
  it('shows pending contacts with local policy links and the upstream attribution', async () => {
    const root = createRootRoute({ component: Footer })
    const router = createRouter({
      routeTree: root,
      history: createMemoryHistory({ initialEntries: ['/'] }),
    })
    render(<RouterProvider router={router} />)
    expect(await screen.findByText(/Contact email/)).toHaveTextContent(
      'To be announced'
    )
    expect(screen.getByText(/Community/)).toHaveTextContent('To be announced')
    expect(
      screen.getByRole('link', { name: 'User Agreement' })
    ).toHaveAttribute('href', '/user-agreement')
    expect(
      screen.getByRole('link', { name: 'Privacy Policy' })
    ).toHaveAttribute('href', '/privacy-policy')
    expect(screen.getByRole('link', { name: 'New API' })).toHaveAttribute(
      'href',
      'https://github.com/QuantumNous/new-api'
    )
    expect(
      screen.queryByRole('link', { name: /cun_ai|wintokenai/ })
    ).not.toBeInTheDocument()
    expect(screen.getAllByRole('link')).toHaveLength(3)
    const footer = screen.getByRole('contentinfo')
    const layout = footer.querySelector('[data-slot=card-content]')
    expect(layout).toHaveClass('flex-col', 'xl:flex-row')
    expect(screen.getByRole('navigation')).toContainElement(
      screen.getByRole('link', { name: 'Privacy Policy' })
    )
  })
})
