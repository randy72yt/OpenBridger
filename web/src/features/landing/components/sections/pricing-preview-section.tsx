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
import { Link } from '@tanstack/react-router'
import { ArrowRight, Coins, Sparkles } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { usePricingData } from '@/features/pricing/hooks/use-pricing-data'

import { FALLBACK_PRICING_ROWS } from '../../constants'
import type { LandingPricingPreviewRow } from '../../types'
import { ScrollReveal } from '../scroll-reveal'

export function PricingPreviewSection() {
  const { t } = useTranslation()
  const { models, groupRatio } = usePricingData()

  const previewRows = useMemo<LandingPricingPreviewRow[]>(() => {
    if (!models || models.length === 0) {
      return FALLBACK_PRICING_ROWS
    }

    const tokenModels = models.filter(
      (item) => item.quota_type === 0 && item.model_ratio > 0
    )
    if (tokenModels.length === 0) {
      return FALLBACK_PRICING_ROWS
    }

    return tokenModels.slice(0, 5).map((item) => {
      const firstGroup = item.enable_groups?.[0] || 'default'
      const ratio = groupRatio[firstGroup] ?? 1.0
      const inputUsd = Number((item.model_ratio * 2 * ratio).toFixed(3))
      const completionFactor =
        item.completion_ratio && item.completion_ratio > 0
          ? item.completion_ratio
          : 1
      const outputUsd = Number((inputUsd * completionFactor).toFixed(3))
      const cacheFactor =
        item.cache_ratio && item.cache_ratio > 0 ? item.cache_ratio : 0.1
      const cacheUsd = Number((inputUsd * cacheFactor).toFixed(3))
      const parsedTags = item.tags
        ? item.tags
            .split(',')
            .map((tag) => tag.trim())
            .filter(Boolean)
            .slice(0, 3)
        : ['API', 'Streaming']

      return {
        modelName: item.model_name,
        vendorName: item.vendor_name || 'AI Provider',
        groupName: firstGroup,
        groupRatio: ratio,
        inputPerMillionUsd: inputUsd,
        outputPerMillionUsd: outputUsd,
        cacheReadPerMillionUsd: cacheUsd,
        tags: parsedTags,
      }
    })
  }, [models, groupRatio])

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
            <div className='overflow-x-auto'>
              <table className='w-full border-collapse text-left text-sm'>
                <thead>
                  <tr className='ob-section-warm border-border text-muted-foreground border-b text-xs font-semibold tracking-wider uppercase'>
                    <th className='px-5 py-3.5'>
                      {t('landing.pricing.col.model')}
                    </th>
                    <th className='px-5 py-3.5'>
                      {t('landing.pricing.col.group')}
                    </th>
                    <th className='px-5 py-3.5 text-right'>
                      {t('landing.pricing.col.input')}
                    </th>
                    <th className='px-5 py-3.5 text-right'>
                      {t('landing.pricing.col.output')}
                    </th>
                    <th className='px-5 py-3.5 text-right'>
                      {t('landing.pricing.col.cacheRead')}
                    </th>
                  </tr>
                </thead>
                <tbody className='divide-border divide-y'>
                  {previewRows.map((row) => (
                    <tr
                      key={row.modelName}
                      className='transition-colors hover:bg-[var(--ob-surf-warm)]'
                    >
                      <td className='px-5 py-4'>
                        <div className='flex flex-col gap-1'>
                          <div className='flex items-center gap-2'>
                            <span className='text-foreground ob-mono text-sm font-semibold'>
                              {row.modelName}
                            </span>
                            <span className='text-muted-foreground text-xs'>
                              {row.vendorName}
                            </span>
                          </div>
                          <div className='flex flex-wrap gap-1'>
                            {row.tags.map((tag) => (
                              <span
                                key={tag}
                                className='ob-section-warm text-muted-foreground rounded px-1.5 py-0.5 text-[10px]'
                              >
                                {tag}
                              </span>
                            ))}
                          </div>
                        </div>
                      </td>
                      <td className='px-5 py-4'>
                        <span className='ob-chip-primary ob-mono inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs font-semibold'>
                          {row.groupName} ×{row.groupRatio}
                        </span>
                      </td>
                      <td className='text-foreground ob-mono px-5 py-4 text-right text-sm font-bold tabular-nums'>
                        ${row.inputPerMillionUsd.toFixed(2)}{' '}
                        <span className='text-muted-foreground text-[11px] font-normal'>
                          / 1M
                        </span>
                      </td>
                      <td className='text-foreground ob-mono px-5 py-4 text-right text-sm font-bold tabular-nums'>
                        ${row.outputPerMillionUsd.toFixed(2)}{' '}
                        <span className='text-muted-foreground text-[11px] font-normal'>
                          / 1M
                        </span>
                      </td>
                      <td className='px-5 py-4 text-right'>
                        <span className='ob-chip-primary ob-mono inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs font-bold tabular-nums'>
                          ${row.cacheReadPerMillionUsd.toFixed(3)} / 1M
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        </ScrollReveal>
      </div>
    </section>
  )
}
