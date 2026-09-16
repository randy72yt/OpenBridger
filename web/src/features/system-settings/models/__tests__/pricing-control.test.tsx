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
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { PricingControlPolicyForm } from '../pricing-control-policy-form'
import type { ModelPriceProposal } from '../pricing-control-types'
import { PricingProposalReview } from '../pricing-proposal-review'

const proposal: ModelPriceProposal = {
  id: 9,
  public_model: 'gpt-5.4-mini',
  service_tier: 'default',
  expected_input_cost: 1.2,
  expected_output_cost: 6,
  proposed_input_price: 1.85,
  proposed_output_price: 9.23,
  stress_input_cost: 1.5,
  stress_output_cost: 7.5,
  expected_margin_bps: 3500,
  worst_margin_bps: 1875,
  primary_channel_name: 'CodeGo',
  backup_channel_name: 'DDD',
  primary_collected_at: 1788912000,
  primary_expires_at: 4102444800,
  backup_collected_at: 1788912000,
  backup_expires_at: 4102444800,
  source_type: 'api',
  source_version: '2026-09-09',
  current_input_price: 2,
  current_output_price: 10,
  input_price_change_bps: -750,
  output_price_change_bps: -770,
  pricing_changed: false,
  status: 'pending',
  created_at: 1788912000,
}

describe('pricing control policy form', () => {
  it('submits the selected model, channels, and default margin safeguards', () => {
    const onSave = vi.fn()
    render(<PricingControlPolicyForm saving={false} onSave={onSave} />)

    fireEvent.change(screen.getByLabelText('Public model'), {
      target: { value: 'gpt-5.4-mini' },
    })
    fireEvent.change(screen.getByLabelText('Primary channel'), {
      target: { value: '11' },
    })
    fireEvent.change(screen.getByLabelText('Backup channel'), {
      target: { value: '12' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'Save pricing policy' }))

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        public_model: 'gpt-5.4-mini',
        primary_channel_id: 11,
        backup_channel_id: 12,
        target_margin_bps: 3500,
        minimum_margin_bps: 2000,
        overhead_bps: 500,
        fallback_probability_bps: 1000,
        enabled: true,
      })
    )
  })
})

describe('pricing proposal review', () => {
  it('shows the live comparison and submits an eligible proposal for approval', () => {
    const onAction = vi.fn()
    render(
      <PricingProposalReview
        proposals={[proposal]}
        recalculating={false}
        updating={false}
        onRecalculate={vi.fn()}
        onAction={onAction}
        onPublish={vi.fn()}
      />
    )

    expect(screen.getByText('CodeGo', { exact: false })).toBeInTheDocument()
    expect(screen.getByText('DDD', { exact: false })).toBeInTheDocument()
    expect(screen.getByText('-7.50%')).toBeInTheDocument()
    expect(screen.getByText('Worst: +18.75%')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Approve' }))
    expect(onAction).toHaveBeenCalledWith(9, 'approve')
  })

  it('blocks approval when the live price changed after proposal calculation', () => {
    render(
      <PricingProposalReview
        proposals={[{ ...proposal, pricing_changed: true }]}
        recalculating={false}
        updating={false}
        onRecalculate={vi.fn()}
        onAction={vi.fn()}
        onPublish={vi.fn()}
      />
    )

    expect(screen.getByText('Stale')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Approve' })).toBeDisabled()
  })
})
