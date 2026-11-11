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
import { Activity, Clock, RefreshCw, Shield } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { cn } from '@/lib/utils'

import { DEFAULT_MODEL_HEALTH_ITEMS } from '../../constants'
import type { ModelStatusLevel } from '../../types'
import { ScrollReveal } from '../scroll-reveal'

function getSlotColorClass(level: ModelStatusLevel): string {
  if (level === 'healthy') {
    return 'ob-status-bar-healthy'
  }
  if (level === 'degraded') {
    return 'ob-status-bar-degraded'
  }
  return 'ob-status-bar-down'
}

function getStatusTextColorClass(level: ModelStatusLevel): string {
  if (level === 'healthy') {
    return 'ob-status-text-healthy'
  }
  if (level === 'degraded') {
    return 'ob-status-text-degraded'
  }
  return 'ob-status-text-down'
}

export function ModelStatusSection() {
  const { t } = useTranslation()

  const architecturePoints = [
    {
      icon: RefreshCw,
      titleKey: 'landing.status.feature.failoverTitle',
      descKey: 'landing.status.feature.failoverDesc',
    },
    {
      icon: Activity,
      titleKey: 'landing.status.feature.probeTitle',
      descKey: 'landing.status.feature.probeDesc',
    },
    {
      icon: Shield,
      titleKey: 'landing.status.feature.isolationTitle',
      descKey: 'landing.status.feature.isolationDesc',
    },
  ]

  return (
    <section className='ob-section-warm relative z-10 py-20 sm:py-28'>
      <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
        <div className='grid grid-cols-1 items-start gap-10 lg:grid-cols-12'>
          {/* Left Column: Reliability & Architecture */}
          <ScrollReveal className='min-w-0 space-y-6 lg:col-span-5'>
            <div className='ob-chip-primary inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold'>
              <Activity className='size-3.5' />
              <span>{t('landing.status.eyebrow')}</span>
            </div>
            <h2 className='ob-serif-heading text-2xl font-bold tracking-tight sm:text-4xl'>
              {t('landing.status.title')}
            </h2>
            <p className='text-muted-foreground text-sm leading-relaxed sm:text-base'>
              {t('landing.status.description')}
            </p>

            <div className='space-y-4 pt-2'>
              {architecturePoints.map((point) => {
                const Icon = point.icon
                return (
                  <Card
                    key={point.titleKey}
                    className='ob-card ob-card-interactive rounded-2xl py-0 shadow-none'
                  >
                    <CardContent className='flex items-start gap-3.5 p-4'>
                      <div className='ob-chip-primary mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-xl'>
                        <Icon className='size-4' />
                      </div>
                      <div className='space-y-1'>
                        <h3 className='text-sm font-semibold'>
                          {t(point.titleKey)}
                        </h3>
                        <p className='text-muted-foreground text-xs leading-relaxed'>
                          {t(point.descKey)}
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                )
              })}
            </div>
          </ScrollReveal>

          {/* Right Column: Real-Time Model Health Telemetry Board */}
          <ScrollReveal delay={0.1} className='min-w-0 lg:col-span-7'>
            <Card className='ob-card overflow-hidden rounded-2xl py-0 shadow-none'>
              <CardHeader className='ob-section-warm border-border flex flex-row items-center justify-between border-b px-5 py-4'>
                <div className='flex items-center gap-2.5'>
                  <span className='relative flex size-2.5'>
                    <span className='ob-dot-success absolute inline-flex h-full w-full animate-ping rounded-full opacity-75' />
                    <span className='ob-dot-success relative inline-flex size-2.5 rounded-full' />
                  </span>
                  <span className='text-sm font-semibold'>
                    {t('landing.status.boardTitle')}
                  </span>
                </div>
                <div className='text-muted-foreground ob-mono flex items-center gap-1.5 text-xs'>
                  <Clock className='size-3.5' />
                  <span>{t('landing.status.window24h')}</span>
                </div>
              </CardHeader>

              <CardContent className='divide-border divide-y p-0'>
                {DEFAULT_MODEL_HEALTH_ITEMS.map((item) => (
                  <div
                    key={item.modelName}
                    className='flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-center sm:justify-between'
                  >
                    <div className='min-w-0 space-y-1'>
                      <div className='flex items-center gap-2'>
                        <span className='ob-mono truncate text-sm font-semibold'>
                          {item.modelName}
                        </span>
                        <span className='ob-chip-primary ob-mono rounded px-1.5 py-0.5 text-[10px] font-semibold'>
                          {item.groupName}
                        </span>
                      </div>
                      <div className='text-muted-foreground ob-mono flex items-center gap-3 text-xs tabular-nums'>
                        <span>
                          TTFT{' '}
                          <strong className='text-foreground'>
                            {item.ttftMs}ms
                          </strong>
                        </span>
                        <span>·</span>
                        <span>
                          TPS{' '}
                          <strong className='text-foreground'>
                            {item.tps}
                          </strong>
                        </span>
                      </div>
                    </div>

                    {/* 24-hour status bars + success rate */}
                    <div className='flex items-center justify-between gap-4 sm:justify-end'>
                      <div
                        className='flex items-center gap-0.5'
                        aria-label={`${item.modelName} 24h status`}
                      >
                        {item.hourlySlots.map((slot) => (
                          <span
                            key={slot.id}
                            className={cn(
                              'h-5 w-1.5 rounded-[2px] transition-opacity hover:opacity-80',
                              getSlotColorClass(slot.level)
                            )}
                          />
                        ))}
                      </div>
                      <span
                        className={cn(
                          'ob-mono w-16 text-right text-xs font-bold tabular-nums',
                          getStatusTextColorClass(item.status)
                        )}
                      >
                        {item.successRate.toFixed(2)}%
                      </span>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          </ScrollReveal>
        </div>
      </div>
    </section>
  )
}
