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
import { api } from '@/lib/api'

import type {
  ModelPricePolicy,
  ModelPriceProposal,
  PricingApiResponse,
  PricingRisk,
  UpstreamModelOffer,
} from './pricing-control-types'

export async function getPricingControlData() {
  const [offers, policies, proposals, risks] = await Promise.all([
    api.get<PricingApiResponse<UpstreamModelOffer[]>>(
      '/api/pricing-control/offers'
    ),
    api.get<PricingApiResponse<ModelPricePolicy[]>>(
      '/api/pricing-control/policies'
    ),
    api.get<PricingApiResponse<ModelPriceProposal[]>>(
      '/api/pricing-control/proposals'
    ),
    api.get<PricingApiResponse<PricingRisk[]>>('/api/pricing-control/risk'),
  ])
  return {
    offers: offers.data.data,
    policies: policies.data.data,
    proposals: proposals.data.data,
    risks: risks.data.data,
  }
}

export async function importPricingOffers(offers: UpstreamModelOffer[]) {
  const response = await api.post('/api/pricing-control/offers/import', {
    offers,
  })
  return response.data
}

export async function syncPricingOffers() {
  const response = await api.post('/api/pricing-control/offers/sync', {})
  return response.data
}

export async function savePricingPolicy(policy: ModelPricePolicy) {
  const response = await api.put('/api/pricing-control/policies', policy)
  return response.data
}

export async function recalculatePricing() {
  const response = await api.post('/api/pricing-control/recalculate')
  return response.data
}

export async function updatePricingProposal(
  id: number,
  action: 'approve' | 'reject' | 'publish'
) {
  const response = await api.post(
    `/api/pricing-control/proposals/${id}/${action}`
  )
  return response.data
}
