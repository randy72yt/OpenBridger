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
export type UpstreamModelOffer = {
  id: number
  channel_id: number
  upstream_model: string
  public_model: string
  input_cost: number
  output_cost: number
  cache_read_cost: number
  upstream_group_ratio?: number | null
  currency: string
  source_type: string
  source_version: string
  collected_at: number
  expires_at: number
  enabled: boolean
}

export type ModelPricePolicy = {
  id?: number
  public_model: string
  service_tier: string
  primary_channel_id: number
  backup_channel_id: number
  target_margin_bps: number
  minimum_margin_bps: number
  overhead_bps: number
  fallback_probability_bps: number
  auto_publish_change_bps: number
  price_locked: boolean
  enabled: boolean
}

export type ProposalStatus = 'pending' | 'approved' | 'published' | 'rejected'

export type ModelPriceProposal = {
  id: number
  public_model: string
  service_tier: string
  expected_input_cost: number
  expected_output_cost: number
  proposed_input_price: number
  proposed_output_price: number
  stress_input_cost: number
  stress_output_cost: number
  expected_margin_bps: number
  worst_margin_bps: number
  primary_channel_name: string
  backup_channel_name: string
  primary_collected_at: number
  primary_expires_at: number
  backup_collected_at: number
  backup_expires_at: number
  source_type: string
  source_version: string
  current_input_price: number | null
  current_output_price: number | null
  input_price_change_bps: number | null
  output_price_change_bps: number | null
  pricing_changed: boolean
  status: ProposalStatus
  created_at: number
}

export type PricingRisk = {
  public_model: string
  service_tier: string
  code: string
}

export type PricingApiResponse<T> = {
  success: boolean
  message: string
  data: T
}

export type PricingOfferSyncResult = {
  imported: number
  skipped: number
  created_proposals: number
  channels: {
    channel_id: number
    imported: number
    skipped: number
    warnings?: string[]
    error?: string
  }[]
}
