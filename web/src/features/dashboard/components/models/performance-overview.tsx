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
import { useQuery } from '@tanstack/react-query'
import { HeartPulse } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import {
  StaticDataTable,
  staticDataTableClassNames as tableStyles,
} from '@/components/data-table'
import { IconBadge } from '@/components/ui/icon-badge'
import { Skeleton } from '@/components/ui/skeleton'
import { getAdminPerfSummary } from '@/features/performance-metrics/api'
import {
  formatLatency,
  formatThroughput,
  formatUptimePct,
  getSuccessRateTextClass,
} from '@/features/performance-metrics/lib/format'
import type { AdminModelPerf } from '@/features/performance-metrics/types'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'

import { UptimeSparkline } from '@/features/pricing/components/model-details-uptime-sparkline'

const ADMIN_WINDOW_HOURS = 24

type PerfSeriesPoint = AdminModelPerf['series'][number]

function toSparklineSeries(series: PerfSeriesPoint[]) {
  return series.map((point) => ({
    date: new Date(point.ts * 1000).toISOString(),
    uptime_pct: point.success_rate,
    incidents: 0,
    outage_minutes: 0,
  }))
}

function topErrorCodes(errorCodes: Record<string, number>, limit = 2): string[] {
  return Object.entries(errorCodes)
    .sort(([, a], [, b]) => b - a)
    .slice(0, limit)
    .map(([code, count]) => `${code} ×${count}`)
}

function formatRelativeTime(ts: number): string {
  if (!ts) return '—'
  const seconds = Math.max(0, Math.floor(Date.now() / 1000 - ts))
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}

export function PerformanceOverview() {
  const { t } = useTranslation()
  const adminQuery = useQuery({
    queryKey: ['perf-metrics-admin', ADMIN_WINDOW_HOURS],
    queryFn: () => getAdminPerfSummary(ADMIN_WINDOW_HOURS),
    staleTime: 60 * 1000,
    retry: false,
  })

  const models = useMemo(() => adminQuery.data?.data.models ?? [], [
    adminQuery.data,
  ])
  const loading = adminQuery.isLoading

  if (loading) {
    return (
      <div className='flex items-center gap-3 rounded-lg border px-4 py-3'>
        <Skeleton className='h-4 w-40' />
        <Skeleton className='h-4 w-24' />
        <Skeleton className='h-4 w-24' />
      </div>
    )
  }

  if (models.length === 0) {
    return (
      <div className='text-muted-foreground overflow-hidden rounded-lg border px-4 py-3 text-center text-xs'>
        {t('No performance data available')}
      </div>
    )
  }

  return (
    <div className='overflow-hidden rounded-lg border'>
      <div className='flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-2.5 sm:px-5'>
        <div className='flex items-center gap-1.5'>
          <IconBadge tone='success' size='xs'>
            <HeartPulse />
          </IconBadge>
          <span className='text-xs font-semibold whitespace-nowrap'>
            {t('Model performance (admin)')}
          </span>
        </div>
        <span className='text-muted-foreground text-xs whitespace-nowrap'>
          {t('Sources: real / probe / internal / client errors')}
        </span>
        <span className='text-muted-foreground ml-auto font-mono text-xs tabular-nums'>
          {t('Probe budget')}: {formatQuota(adminQuery.data?.data.probe_budget_used_quota ?? 0)} /{' '}
          {formatQuota(adminQuery.data?.data.probe_budget_quota ?? 0)}
          {adminQuery.data?.data.error_codes_truncated && (
            <span
              className='ml-2'
              title={t(
                'Error-log scan hit its row cap; error codes cover only part of the window.'
              )}
            >
              {t('Error codes truncated')}
            </span>
          )}
        </span>
      </div>
      <StaticDataTable
        tableClassName='text-xs'
        headerRowClassName={tableStyles.compactHeaderRow}
        data={models}
        getRowKey={(model) => model.model_name}
        columns={[
          {
            id: 'model',
            header: t('Model'),
            className: tableStyles.compactHeaderCell,
            cellClassName: tableStyles.compactCell,
            cell: (model) => (
              <span className='max-w-[12rem] truncate font-mono'>
                {model.model_name}
              </span>
            ),
          },
          {
            id: 'requests',
            header: t('Requests (r/p/i/c)'),
            className: tableStyles.compactHeaderCell,
            cellClassName: tableStyles.compactCell,
            cell: (model) => (
              <span className='font-mono tabular-nums'>
                {model.real.requests}/{model.probe.requests}/
                {model.internal.requests}/{model.client_error.requests}
              </span>
            ),
          },
          {
            id: 'availability',
            header: t('Availability 24h / 7d'),
            className: tableStyles.compactHeaderCellRight,
            cellClassName: tableStyles.compactNumericCell,
            cell: (model) => (
              <span
                className={cn(
                  'font-mono tabular-nums',
                  getSuccessRateTextClass(model.availability_24h)
                )}
              >
                {formatUptimePct(model.availability_24h)}
                <span className='text-muted-foreground'>
                  {' / '}
                  {formatUptimePct(model.availability_7d)}
                </span>
              </span>
            ),
          },
          {
            id: 'ttft',
            header: t('First token / total'),
            className: tableStyles.compactHeaderCellRight,
            cellClassName: tableStyles.compactNumericCell,
            cell: (model) => (
              <span className='font-mono tabular-nums'>
                {formatLatency(model.real.avg_ttft_ms || model.probe.avg_ttft_ms)}
                <span className='text-muted-foreground'>
                  {' / '}
                  {formatLatency(
                    model.real.avg_latency_ms || model.probe.avg_latency_ms
                  )}
                </span>
              </span>
            ),
          },
          {
            id: 'tps',
            header: 'TPS',
            className: tableStyles.compactHeaderCellRight,
            cellClassName: tableStyles.compactNumericCell,
            cell: (model) => (
              <span className='font-mono tabular-nums'>
                {formatThroughput(
                  model.real.avg_tps || model.probe.avg_tps
                ).replace(' t/s', 't/s')}
              </span>
            ),
          },
          {
            id: 'series',
            header: t('24h status'),
            className: tableStyles.compactHeaderCell,
            cellClassName: tableStyles.compactCell,
            cell: (model) => (
              <UptimeSparkline size='sm' series={toSparklineSeries(model.series)} />
            ),
          },
          {
            id: 'public-preview',
            header: t('If public v2 is on'),
            className: tableStyles.compactHeaderCell,
            cellClassName: tableStyles.compactCell,
            cell: (model) => (
              <span
                className={cn(
                  'inline-flex items-center gap-1 font-mono text-[11px] tabular-nums',
                  getSuccessRateTextClass(model.public_v2.success_rate)
                )}
                title={t(
                  'What the public model square would show once public v2 is enabled'
                )}
              >
                {formatUptimePct(model.public_v2.success_rate)}
                {model.public_v2.insufficient && (
                  <span className='text-muted-foreground'>
                    · {t('Not enough samples')}
                  </span>
                )}
              </span>
            ),
          },
          {
            id: 'errors',
            header: t('Top error codes'),
            className: tableStyles.compactHeaderCell,
            cellClassName: tableStyles.compactCell,
            cell: (model) => {
              const codes = topErrorCodes(model.error_codes)
              if (codes.length === 0) return <span className='text-muted-foreground'>—</span>
              return (
                <span className='font-mono text-[11px]'>
                  {codes.join(', ')}
                </span>
              )
            },
          },
          {
            id: 'probe',
            header: t('Probe'),
            className: tableStyles.compactHeaderCell,
            cellClassName: tableStyles.compactCell,
            cell: (model) => (
              <span className='text-muted-foreground font-mono text-[11px]'>
                {model.probe_tier || '—'} · {formatRelativeTime(model.last_probe_at)}
              </span>
            ),
          },
        ]}
      />
    </div>
  )
}
