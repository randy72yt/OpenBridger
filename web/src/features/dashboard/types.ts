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
import type { TimeGranularity } from '@/lib/time'

// ============================================================================
// Quota & Usage Data Types
// ============================================================================

export interface QuotaDataItem {
  id?: number
  user_id?: number
  username?: string
  model_name?: string
  created_at: number
  token_used?: number
  count?: number
  quota?: number
}

export interface FlowQuotaDataItem {
  user_id?: number
  username?: string
  node_name?: string
  use_group?: string
  token_id?: number
  token_name?: string
  channel_id?: number
  channel_name?: string
  model_name?: string
  token_used?: number
  count?: number
  quota?: number
}

export type FlowMetric = 'quota' | 'tokens' | 'requests'

export type FlowOverflowMode = 'aggregate' | 'hide'

export type FlowRole = 'user' | 'admin' | 'root'

export type FlowNodeKind =
  | 'user'
  | 'node'
  | 'token'
  | 'group'
  | 'model'
  | 'channel'

export interface FlowNodeFilter {
  kind: FlowNodeKind
  id: string
}

export interface FlowLinkSelection {
  source: string
  target: string
}

export interface FlowBuildOptions {
  role?: FlowRole
  selectedUsers?: string[]
  selectedNodes?: FlowNodeFilter[]
  activeNode?: FlowNodeFilter
  activeLink?: FlowLinkSelection
  colorPalette?: readonly string[]
  visibleStages?: FlowNodeKind[]
  topNodeLimit?: number
  overflowMode?: FlowOverflowMode
  // When true, sensitive node labels (users, tokens, nodes, groups, channels)
  // are partially masked in the rendered graph while keeping node identity so
  // the Sankey shape stays intact.
  maskSensitive?: boolean
  // Resolves the label for a token whose record no longer exists (deleted).
  // Lets the caller inject a localized string such as "Deleted (123)".
  deletedTokenLabel?: (tokenId: number) => string
  otherNodeLabel?: (kind: FlowNodeKind) => string
}

export interface DashboardFlowNode {
  id: string
  label: string
  kind: FlowNodeKind
  value: number
  requests: number
  quota: number
  tokens: number
  color: string
  colorKey: string
  highlighted?: boolean
  dimmed?: boolean
}

export interface DashboardFlowLink {
  source: string
  target: string
  value: number
  requests: number
  quota: number
  tokens: number
  sourceLabel: string
  targetLabel: string
  color: string
  linkColor: string
  linkAlpha: number
  hoverColor: string
  colorKey: string
  share: number
  highlighted?: boolean
  dimmed?: boolean
}

export interface DashboardFlowGraph {
  nodes: DashboardFlowNode[]
  links: DashboardFlowLink[]
}

export interface FlowUserFilterOption {
  value: string
  label: string
  valueLabel: string
  valueRaw: number
  color: string
}

export interface FlowNodeFilterOption {
  kind: FlowNodeKind
  value: string
  label: string
  valueLabel: string
  valueRaw: number
  color: string
}

export interface FlowFilterOptions {
  users: FlowUserFilterOption[]
  nodes: FlowNodeFilterOption[]
}

export interface FlowSummary {
  quota: number
  tokens: number
  requests: number
}

export interface ProcessedFlowData {
  summary: FlowSummary
  flow: DashboardFlowGraph
  filterOptions: FlowFilterOptions
}

// ============================================================================
// Uptime Monitoring Types
// ============================================================================

export interface UptimeMonitor {
  name: string
  uptime: number
  status: number
  group?: string
}

export interface UptimeGroupResult {
  categoryName: string
  monitors: UptimeMonitor[]
}

// ============================================================================
// Admin Business Dashboard Types
// ============================================================================

export interface BusinessDashboardSummary {
  cash_revenue_usd: number
  topup_revenue_usd: number
  subscription_revenue_usd: number
  paid_order_count: number
  paying_users: number
  arppu_usd: number
  redemption_usd: number
  redemption_count: number
  consumption_usd: number
  costed_consumption_usd: number
  uncosted_consumption_usd: number
  estimated_cost_usd: number
  estimated_gross_profit_usd: number
  estimated_gross_margin_pct: number | null
  cost_coverage_pct: number
  total_requests: number
  total_tokens: number
  total_users: number
  new_users: number
  active_users: number
  wallet_liability_usd: number
  channel_balance_usd: number
}

export interface BusinessDailyPoint {
  date: string
  timestamp: number
  cash_revenue_usd: number
  redemption_usd: number
  consumption_usd: number
  estimated_cost_usd: number
  new_users: number
  paid_orders: number
}

export interface PaymentChannelStat {
  channel: string
  provider: string
  order_type: 'topup' | 'subscription' | string
  order_count: number
  original_currency: string
  original_amount: number
  revenue_usd: number
}

export interface ModelMarginStat {
  model_name: string
  channel_id: number
  channel_name: string
  requests: number
  prompt_tokens: number
  completion_tokens: number
  consumption_usd: number
  estimated_cost_usd: number | null
  gross_profit_usd: number | null
  gross_margin_pct: number | null
  cost_covered: boolean
}

export interface BusinessDashboardData {
  summary: BusinessDashboardSummary
  daily_trend: BusinessDailyPoint[]
  payment_channels: PaymentChannelStat[]
  model_margins: ModelMarginStat[]
}

// ============================================================================
// Dashboard Filter Types
// ============================================================================

export interface DashboardFilters {
  start_timestamp?: Date
  end_timestamp?: Date
  time_granularity?: TimeGranularity
  username?: string
}

export type ConsumptionDistributionChartType = 'bar' | 'area'

export type ModelAnalyticsChartTab = 'trend' | 'proportion' | 'top'

export interface DashboardChartPreferences {
  consumptionDistributionChart: ConsumptionDistributionChartType
  modelAnalyticsChart: ModelAnalyticsChartTab
  defaultTimeRangeDays: number
  defaultTimeGranularity: TimeGranularity
}

// User analytics selections are held by the dashboard parent so they survive
// switching between dashboard sub-sections, matching the model/flow filters.
export interface UserChartsFilters {
  timeGranularity: TimeGranularity
  selectedRange: number
  topUserLimit: number
}

// ============================================================================
// API Info Types
// ============================================================================

export interface ApiInfoItem {
  url: string
  route: string
  description: string
  color: string
}

export interface PingStatus {
  latency: number | null
  testing: boolean
  error: boolean
}

export type PingStatusMap = Record<string, PingStatus>

// ============================================================================
// Chart Types
// ============================================================================

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type VChartSpec = Record<string, any>

export interface ProcessedChartData {
  spec_pie: VChartSpec
  spec_line: VChartSpec
  spec_area: VChartSpec
  spec_model_line: VChartSpec
  spec_rank_bar: VChartSpec
  totalQuotaDisplay: string
  totalCountDisplay: string
}

export interface ProcessedUserChartData {
  spec_user_rank: VChartSpec
  spec_user_trend: VChartSpec
}

// ============================================================================
// Announcement Types
// ============================================================================

export interface AnnouncementItem {
  id?: number
  content: string
  publishDate?: string
  type?: 'default' | 'ongoing' | 'success' | 'warning' | 'error'
  extra?: string
}

// ============================================================================
// FAQ Types
// ============================================================================

export interface FAQItem {
  id?: number
  question: string
  answer: string
}
