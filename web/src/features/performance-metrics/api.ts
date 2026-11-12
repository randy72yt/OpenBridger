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
  AdminPerfSummaryData,
  PerformanceMetricsData,
  PerfSummaryAllData,
} from './types'

// bars is only honored by the backend when public_v2 is enabled; v1 responses
// ignore it, so callers can always send the wider status-series window.
export async function getPerfMetricsSummary(
  hours = 24,
  bars = 24
): Promise<PerfSummaryAllData> {
  const res = await api.get<PerfSummaryAllData>('/api/perf-metrics/summary', {
    params: { hours, bars },
  })
  return res.data
}

export async function getAdminPerfSummary(
  hours = 24
): Promise<AdminPerfSummaryData> {
  const res = await api.get<AdminPerfSummaryData>(
    '/api/perf-metrics/admin/summary',
    { params: { hours } }
  )
  return res.data
}

export async function getPerfMetrics(
  modelName: string,
  hours = 24
): Promise<PerformanceMetricsData> {
  const res = await api.get<PerformanceMetricsData>('/api/perf-metrics', {
    params: {
      model: modelName,
      hours,
    },
  })
  return res.data
}
