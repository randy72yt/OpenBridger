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
import { Link, useNavigate } from '@tanstack/react-router'
import { ArrowRight, Coins, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { ErrorState } from '@/components/error-state'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { PricingTable } from '@/features/pricing/components/pricing-table'
import { usePricingData } from '@/features/pricing/hooks/use-pricing-data'

import { ScrollReveal } from '../scroll-reveal'

export function PricingPreviewSection() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { models, isLoading, error, refetch } = usePricingData()

  return (
    <section className='ob-section-bg relative z-10 py-20 sm:py-28'>
      <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
        <ScrollReveal className='flex flex-col items-start justify-between gap-6 md:flex-row md:items-end'>
          <div className='max-w-2xl'>
            <div className='ob-chip-primary mb-3 inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold'>
              <Coins className='size-3.5' />
              <span>{t('landing.pricing.eyebrow')}</span>
            </div>
            <h2 className='ob-serif-heading text-2xl font-bold tracking-tight sm:text-4xl'>
              {t('landing.pricing.title')}
            </h2>
            <p className='text-muted-foreground mt-3 text-sm leading-relaxed sm:text-base'>
              {t('landing.pricing.description')}
            </p>
          </div>
          <Button
            variant='outline'
            className='ob-btn-secondary group shrink-0 rounded-xl px-5'
            render={<Link to='/pricing' />}
          >
            <span>{t('landing.pricing.exploreAll')}</span>
            <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
          </Button>
        </ScrollReveal>

        {/* Group Strategy Highlights */}
        <ScrollReveal
          delay={0.08}
          className='mt-10 grid grid-cols-1 gap-4 md:grid-cols-3'
        >
          <Card className='ob-card ob-card-interactive rounded-2xl py-0 shadow-none'>
            <CardHeader className='p-5 pb-2'>
              <div className='flex items-center justify-between'>
                <span className='ob-mono text-xs font-semibold'>default</span>
                <Badge
                  variant='secondary'
                  className='ob-chip-primary text-[11px]'
                >
                  {t('landing.pricing.group.standardBadge')}
                </Badge>
              </div>
              <h3 className='mt-1 text-base font-semibold'>
                {t('landing.pricing.group.standardTitle')}
              </h3>
            </CardHeader>
            <CardContent className='text-muted-foreground p-5 pt-1 text-xs leading-relaxed sm:text-sm'>
              {t('landing.pricing.group.standardDesc')}
            </CardContent>
          </Card>

          <Card className='ob-card-featured ob-card-interactive rounded-2xl py-0 shadow-none'>
            <CardHeader className='p-5 pb-2'>
              <div className='flex items-center justify-between'>
                <span className='text-primary ob-mono flex items-center gap-1 text-xs font-semibold'>
                  <Sparkles className='size-3.5' />
                  smart / auto
                </span>
                <Badge className='ob-btn-primary text-[11px]'>
                  {t('landing.pricing.group.smartBadge')}
                </Badge>
              </div>
              <h3 className='mt-1 text-base font-semibold'>
                {t('landing.pricing.group.smartTitle')}
              </h3>
            </CardHeader>
            <CardContent className='text-muted-foreground p-5 pt-1 text-xs leading-relaxed sm:text-sm'>
              {t('landing.pricing.group.smartDesc')}
            </CardContent>
          </Card>

          <Card className='ob-card ob-card-interactive rounded-2xl py-0 shadow-none'>
            <CardHeader className='p-5 pb-2'>
              <div className='flex items-center justify-between'>
                <span className='ob-mono text-xs font-semibold'>
                  codex / dedicated
                </span>
                <Badge
                  variant='secondary'
                  className='ob-chip-primary text-[11px]'
                >
                  {t('landing.pricing.group.dedicatedBadge')}
                </Badge>
              </div>
              <h3 className='mt-1 text-base font-semibold'>
                {t('landing.pricing.group.dedicatedTitle')}
              </h3>
            </CardHeader>
            <CardContent className='text-muted-foreground p-5 pt-1 text-xs leading-relaxed sm:text-sm'>
              {t('landing.pricing.group.dedicatedDesc')}
            </CardContent>
          </Card>
        </ScrollReveal>

        {/* Live Model Square Preview Table */}
        <ScrollReveal delay={0.14} className='mt-6'>
          <Card className='ob-card overflow-hidden rounded-2xl py-0 shadow-none'>
            {error && !models.length ? (
              <ErrorState onRetry={() => void refetch()} />
            ) : (
              <PricingTable
                models={models.slice(0, 5)}
                isLoading={isLoading}
                showPagination={false}
                onModelClick={(name) =>
                  navigate({ to: '/pricing', search: { search: name } })
                }
              />
            )}
          </Card>
        </ScrollReveal>
      </div>
    </section>
  )
}
