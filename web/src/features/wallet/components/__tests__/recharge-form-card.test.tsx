import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import type { TopupInfo } from '../../types'
import { RechargeFormCard } from '../recharge-form-card'

vi.mock('@/hooks/use-system-config', () => ({
  useSystemConfig: () => ({
    currency: { quotaDisplayType: 'USD', usdExchangeRate: 1 },
  }),
}))

const baseProps = {
  presetAmounts: [{ value: 10 }, { value: 50 }],
  selectedPreset: null,
  onSelectPreset: vi.fn(),
  topupAmount: 0,
  onTopupAmountChange: vi.fn(),
  paymentAmount: 0,
  calculating: false,
  onPaymentMethodSelect: vi.fn(),
  paymentLoading: null,
  redemptionCode: '',
  onRedemptionCodeChange: vi.fn(),
  onRedeem: vi.fn(),
  redeeming: false,
}

const buildTopupInfo = (overrides: Partial<TopupInfo>): TopupInfo =>
  ({
    online_payment_enabled: true,
    enable_online_topup: false,
    enable_stripe_topup: false,
    enable_creem_topup: false,
    enable_redemption: true,
    payment_compliance_confirmed: true,
    pay_methods: [],
    min_topup: 1,
    amount_options: [10, 50],
    ...overrides,
  }) as unknown as TopupInfo

it('degrades to redemption only when the global online-payment switch is off', () => {
  render(
    <RechargeFormCard
      {...baseProps}
      topupInfo={buildTopupInfo({
        online_payment_enabled: false,
        enable_online_topup: true,
        pay_methods: [{ name: 'Alipay', type: 'alipay' }],
      })}
    />
  )

  expect(screen.getByLabelText('Have a Code?')).toBeInTheDocument()
  expect(screen.queryByText('Amount')).not.toBeInTheDocument()
  expect(
    screen.queryByText('Choose an amount and payment method')
  ).not.toBeInTheDocument()
  expect(screen.getByText('Enter your redemption code')).toBeInTheDocument()
})

it('keeps the amount picker when online payment is enabled', () => {
  render(
    <RechargeFormCard
      {...baseProps}
      topupInfo={buildTopupInfo({
        enable_online_topup: true,
        pay_methods: [{ name: 'Alipay', type: 'alipay' }],
      })}
    />
  )

  expect(screen.getByText('Choose an amount and payment method')).toBeInTheDocument()
  expect(screen.getByText('Amount')).toBeInTheDocument()
  expect(screen.getByLabelText('Have a Code?')).toBeInTheDocument()
})
