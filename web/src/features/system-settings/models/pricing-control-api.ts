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
import { isAxiosError } from 'axios'

import { api } from '@/lib/api'

import type {
  ModelPricePolicy,
  PricingApiResponse,
  PricingCalibrationPage,
  PricingCalibrationResult,
  PricingOfferPage,
  PricingOfferSyncResult,
  PricingProposalPage,
  PricingRisk,
  PricingSyncRunsResponse,
  UpstreamModelOffer,
} from './pricing-control-types'

export type PricingOfferListParams = {
  model?: string
  keyword?: string
  sourceType?: string
  includeExpired?: boolean
  limit?: number
  offset?: number
}

export async function getPricingOffers(params: PricingOfferListParams = {}) {
  const response = await api.get<PricingApiResponse<PricingOfferPage>>(
    '/api/pricing-control/offers',
    {
      params: {
        model: params.model || undefined,
        q: params.keyword || undefined,
        source_type: params.sourceType || undefined,
        include_expired: params.includeExpired ? 'true' : undefined,
        limit: params.limit ?? 50,
        offset: params.offset ?? 0,
      },
    }
  )
  return response.data.data
}

export async function getPricingPolicies() {
  const response = await api.get<PricingApiResponse<ModelPricePolicy[]>>(
    '/api/pricing-control/policies'
  )
  return response.data.data
}

export async function getPricingRisks() {
  const response = await api.get<PricingApiResponse<PricingRisk[]>>(
    '/api/pricing-control/risk'
  )
  return response.data.data
}

export async function getPricingProposals(
  params: { status?: string; limit?: number; offset?: number } = {}
) {
  const response = await api.get<PricingApiResponse<PricingProposalPage>>(
    '/api/pricing-control/proposals',
    {
      params: {
        status: params.status || undefined,
        limit: params.limit ?? 50,
        offset: params.offset ?? 0,
      },
    }
  )
  return response.data.data
}

export async function getPricingSyncRuns(limit = 10) {
  const response = await api.get<PricingApiResponse<PricingSyncRunsResponse>>(
    '/api/pricing-control/sync-runs',
    { params: { limit } }
  )
  return response.data.data
}

export async function importPricingOffers(offers: UpstreamModelOffer[]) {
  const response = await api.post('/api/pricing-control/offers/import', {
    offers,
  })
  return response.data
}

export async function syncPricingOffers() {
  try {
    const response = await api.post<PricingApiResponse<PricingOfferSyncResult>>(
      '/api/pricing-control/offers/sync',
      {},
      { skipErrorHandler: true }
    )
    return response.data
  } catch (error) {
    if (isAxiosError<PricingApiResponse<PricingOfferSyncResult>>(error)) {
      const payload = error.response?.data
      if (payload?.data?.channels) return payload
    }
    throw error
  }
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

export async function getPricingCalibrations(
  params: { keyword?: string; limit?: number; offset?: number } = {}
) {
  const response = await api.get<PricingApiResponse<PricingCalibrationPage>>(
    '/api/pricing-control/calibrations',
    {
      params: {
        q: params.keyword || undefined,
        limit: params.limit ?? 50,
        offset: params.offset ?? 0,
      },
    }
  )
  return response.data.data
}

export async function syncPricingCalibrations() {
  const response = await api.post<PricingApiResponse<PricingCalibrationResult>>(
    '/api/pricing-control/calibrations/sync',
    {},
    { skipErrorHandler: true }
  )
  return response.data
}

export async function pinPricingCalibration(payload: {
  channel_id: number
  public_model: string
  effective_ratio: number
  reason?: string
}) {
  const response = await api.put('/api/pricing-control/calibrations', payload)
  return response.data
}
