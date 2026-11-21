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
import { Activity, RefreshCw, Shield } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Card, CardContent, CardHeader } from '@/components/ui/card'

import { ScrollReveal } from '../scroll-reveal'

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
              <CardHeader className='ob-section-warm border-border border-b px-5 py-4'>
                <span className='text-sm font-semibold'>
                  {t('landing.status.boardTitle')}
                </span>
              </CardHeader>
              <CardContent className='text-muted-foreground px-5 py-10 text-center text-sm'>
                {t('No sample data yet.')}
              </CardContent>
            </Card>
          </ScrollReveal>
        </div>
      </div>
    </section>
  )
}
