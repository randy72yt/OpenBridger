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
import { memo, useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import {
  formatLatency,
  formatThroughput,
  getSuccessRateDotClass,
} from '@/features/performance-metrics/lib/format'
import type { SuccessRatePoint } from '@/features/performance-metrics/types'
import { cn } from '@/lib/utils'

export type ModelPerfBadgeData = {
  avg_latency_ms: number
  success_rate: number
  avg_tps: number
  recent_success_series?: SuccessRatePoint[]
  // v2-only fields; their presence switches the badge to the v2 caliber.
  avg_ttft_ms?: number
  availability_24h?: number
  sample_count?: number
  probe_share?: number
  insufficient?: boolean
}

export interface ModelPerfBadgeProps
  extends React.HTMLAttributes<HTMLDivElement> {
  perf: ModelPerfBadgeData | undefined
  statusSlots?: number
}

// Denser strips need thinner bars to fit the fixed 96px status block.
function statusBarWidthClass(slotCount: number): string {
  if (slotCount >= 72) return 'w-[1px]'
  if (slotCount >= 48) return 'w-[2px]'
  return 'w-[3px]'
}

export const ModelPerfBadge = memo(function ModelPerfBadge(
  props: ModelPerfBadgeProps
) {
  const { t } = useTranslation()
  const slotCount = props.statusSlots ?? 24
  const statusSlots = useMemo(
    () => Array.from({ length: slotCount }, (_, slot) => slot),
    [slotCount]
  )
  const barWidthClass = statusBarWidthClass(slotCount)

  const v2 = props.perf?.availability_24h != null
  const successRate = v2 ? props.perf?.availability_24h : props.perf?.success_rate
  const hasSuccessRate =
    successRate != null &&
    Number.isFinite(successRate) &&
    successRate >= 0 &&
    successRate <= 100

  const totalLatencyText = formatLatency(props.perf?.avg_latency_ms ?? 0)
  const ttftLatencyText = formatLatency(props.perf?.avg_ttft_ms ?? 0)
  let latencyText = totalLatencyText
  if (v2 && ttftLatencyText !== '—') {
    latencyText = `${ttftLatencyText} / ${totalLatencyText}`
  }
  const throughputText = formatThroughput(props.perf?.avg_tps ?? 0).replace(
    ' t/s',
    't/s'
  )

  const insufficient = v2 && Boolean(props.perf?.insufficient)
  const probeDriven = v2 && (props.perf?.probe_share ?? 0) > 0.5
  const hints: string[] = []
  if (insufficient) hints.push(t('Not enough samples'))
  if (probeDriven) hints.push(t('Mostly probe traffic'))

  // Hourly points with timestamps, anchored to the client's current hour.
  // Hours without traffic stay gray. The last slot is the current, partial hour.
  const statusRates = useMemo(() => {
    const currentHourStart = Math.floor(Date.now() / 1000 / 3600) * 3600
    const ratesByHour = new Map<number, number>()
    for (const point of props.perf?.recent_success_series ?? []) {
      ratesByHour.set(point.ts, point.success_rate)
    }
    return statusSlots.map((slot) => {
      const hourStart = currentHourStart - (slotCount - 1 - slot) * 3600
      return ratesByHour.get(hourStart)
    })
  }, [props.perf?.recent_success_series, slotCount, statusSlots])

  return (
    <div
      aria-label={t('Performance metrics for the last 24 hours')}
      className={cn(
        'flex w-full min-w-0 items-center justify-between gap-3',
        insufficient && 'opacity-55',
        props.className
      )}
    >
      <dl
        className={cn(
          'flex min-w-0 items-start gap-5 text-xs tabular-nums',
          insufficient && 'grayscale-[30%]'
        )}
      >
        <div className='w-24 shrink-0'>
          <dt
            title={t('Request success rate sampled over the last 24 hours')}
            className='text-muted-foreground flex items-center justify-between gap-1 text-[11px] leading-4'
          >
            <span>{t('Status')}</span>
            <span className='font-mono'>
              {hasSuccessRate
                ? `${successRate.toFixed(v2 ? 2 : 1)}%`
                : '—%'}
            </span>
          </dt>
          <dd
            role='img'
            aria-label={t(
              'Recent success-rate samples; gray bars indicate missing data.'
            )}
            title={t(
              'Recent success-rate samples; gray bars indicate missing data.'
            )}
            className='mt-1 flex h-3 w-24 items-center justify-between'
          >
            {statusSlots.map((slot) => {
              const rate = statusRates[slot]
              return (
                <span
                  key={slot}
                  aria-hidden
                  className={cn(
                    'h-full shrink-0 rounded-xs',
                    barWidthClass,
                    rate != null &&
                      Number.isFinite(rate) &&
                      rate >= 0 &&
                      rate <= 100
                      ? getSuccessRateDotClass(rate)
                      : 'bg-muted-foreground/15'
                  )}
                />
              )
            })}
          </dd>
          {hints.length > 0 && (
            <p className='text-muted-foreground/80 mt-1 text-[10px] leading-4'>
              {hints.join(' · ')}
            </p>
          )}
        </div>
        <div title={t('Average latency')} className='shrink-0'>
          <dt className='text-muted-foreground text-[11px] leading-4'>
            {v2 ? t('First token / total') : t('Latency short')}
          </dt>
          <dd className='mt-1 font-mono whitespace-nowrap'>
            {latencyText === '—' ? '—s' : latencyText}
          </dd>
        </div>
        <div title={t('Throughput')} className='shrink-0'>
          <dt className='text-muted-foreground text-[11px] leading-4'>
            {t('Throughput short')}
          </dt>
          <dd className='mt-1 font-mono whitespace-nowrap'>
            {throughputText === '—' ? '—t/s' : throughputText}
          </dd>
        </div>
      </dl>
      {props.children}
    </div>
  )
})
