import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { AffiliateRewardsCard } from '../affiliate-rewards-card'

it('shows a readable referral link and preserves the compliance notice with no rewards', () => {
  render(
    <AffiliateRewardsCard
      user={null}
      affiliateLink='https://example.com/sign-up?aff=sample'
      onTransfer={vi.fn()}
      complianceConfirmed={false}
    />
  )
  expect(
    screen.getByRole('textbox', { name: 'Copy referral link' })
  ).toHaveValue('https://example.com/sign-up?aff=sample')
  expect(screen.getByRole('textbox')).toHaveAttribute('readonly')
  expect(
    screen.queryByRole('button', { name: 'Transfer to Balance' })
  ).not.toBeInTheDocument()
  expect(
    screen.getByText(/Referral reward transfer is disabled/)
  ).toBeInTheDocument()
  expect(screen.getByText(/Earn rewards when users join/)).not.toHaveClass(
    'line-clamp-1'
  )
})
