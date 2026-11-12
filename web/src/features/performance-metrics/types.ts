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
export type PerformanceSeriesPoint = {
  ts: number
  avg_ttft_ms: number
  avg_latency_ms: number
  success_rate: number
  avg_tps: number
}

export type PerformanceGroup = {
  group: string
  avg_ttft_ms: number
  avg_latency_ms: number
  success_rate: number
  avg_tps: number
  series: PerformanceSeriesPoint[]
}

export type PerformanceMetricsData = {
  success: boolean
  message?: string
  data: {
    model_name: string
    series_schema?: string
    groups: PerformanceGroup[]
  }
}

export type SuccessRatePoint = { ts: number; success_rate: number }

export type PerfModelSummary = {
  model_name: string
  avg_latency_ms: number
  success_rate: number
  avg_tps: number
  recent_success_series?: SuccessRatePoint[]
  request_count?: number
  // v2-only fields: present only when perf_metrics_setting.public_v2 is on.
  avg_ttft_ms?: number
  availability_24h?: number
  availability_7d?: number
  sample_count?: number
  probe_share?: number
  insufficient?: boolean
}

export type PerfSummaryAllData = {
  success: boolean
  message?: string
  data: {
    models: PerfModelSummary[]
  }
}

export type AdminCounterView = {
  requests: number
  success: number
  avg_latency_ms: number
  avg_ttft_ms: number
  avg_tps: number
}

export type AdminModelPerf = {
  model_name: string
  real: AdminCounterView
  probe: AdminCounterView
  internal: AdminCounterView
  client_error: AdminCounterView
  public_v1: PerfModelSummary
  public_v2: PerfModelSummary
  availability_24h: number
  availability_7d: number
  error_codes: Record<string, number>
  series: PerformanceSeriesPoint[]
  last_probe_at: number
  probe_tier: string
  probe_token_param: string
}

export type AdminPerfSummaryData = {
  success: boolean
  message?: string
  data: {
    models: AdminModelPerf[]
    probe_budget_used_quota: number
    probe_budget_quota: number
    generated_at: number
    error_codes_truncated?: boolean
  }
}
