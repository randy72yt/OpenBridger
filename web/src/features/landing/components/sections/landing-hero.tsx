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
import { ArrowRight, BookOpen, ShieldCheck, Sparkles, Zap } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { useStatus } from '@/hooks/use-status'

import { HERO_HIGHLIGHT_PILLS, HERO_METRICS } from '../../constants'
import { HeroBackground } from '../hero-background'
import { LandingTerminalDemo } from '../landing-terminal-demo'
import { ScrollReveal } from '../scroll-reveal'

interface LandingHeroProps {
  isAuthenticated: boolean
}

export function LandingHero(props: LandingHeroProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const docsUrl = (status?.docs_link as string | undefined) || '/docs'
  const isExternalDocs = docsUrl.startsWith('http')

  return (
    <section className='ob-section-bg relative isolate overflow-hidden pt-28 pb-16 sm:pt-32 sm:pb-24 lg:pt-36 lg:pb-28'>
      <HeroBackground />

      <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
        <div className='grid grid-cols-1 items-center gap-12 lg:grid-cols-12 lg:gap-10'>
          {/* Left Column: Brand pitch, CTAs, and capability pills */}
          <div className='flex min-w-0 flex-col items-start text-left lg:col-span-6'>
            {/* Status Pill */}
            <div className='ob-chip-primary mb-6 inline-flex items-center gap-2 rounded-full px-3.5 py-1.5 text-xs font-semibold'>
              <span className='relative flex size-2'>
                <span className='ob-dot-primary absolute inline-flex h-full w-full animate-ping rounded-full opacity-75' />
                <span className='ob-dot-primary relative inline-flex size-2 rounded-full' />
              </span>
              <span>{t('landing.hero.badge')}</span>
            </div>

            <h1 className='ob-serif-heading text-foreground text-[clamp(2.25rem,4.4vw,3.65rem)] leading-[1.14] font-bold tracking-tight'>
              {t('landing.hero.titleLine1')}
              <br />
              <span className='ob-hero-gradient-text'>
                {t('landing.hero.titleLine2')}
              </span>
            </h1>

            <p className='text-muted-foreground mt-5 max-w-xl text-base leading-relaxed sm:text-lg'>
              {t('landing.hero.subtitle')}
            </p>

            {/* Capability Badges */}
            <div className='mt-6 flex flex-wrap items-center gap-2'>
              {HERO_HIGHLIGHT_PILLS.map((pillKey, idx) => (
                <span
                  key={pillKey}
                  className='ob-chip-primary inline-flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs font-medium'
                >
                  <span className='ob-dot-primary size-1.5 rounded-full' />
                  {idx === 0 && <Sparkles className='size-3.5' />}
                  {idx === 1 && <ShieldCheck className='size-3.5' />}
                  {idx === 2 && <Zap className='size-3.5' />}
                  {t(pillKey)}
                </span>
              ))}
            </div>

            {/* Primary & Secondary CTAs */}
            <div className='mt-8 flex flex-wrap items-center gap-3.5'>
              {props.isAuthenticated ? (
                <Button
                  size='lg'
                  className='ob-btn-primary group h-12 rounded-xl px-6 text-sm font-semibold'
                  render={<Link to='/dashboard' />}
                >
                  {t('Go to Dashboard')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
              ) : (
                <Button
                  size='lg'
                  className='ob-btn-primary group h-12 rounded-xl px-6 text-sm font-semibold'
                  render={<Link to='/sign-up' />}
                >
                  {t('landing.hero.ctaPrimary')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
              )}

              <Button
                variant='outline'
                size='lg'
                className='ob-btn-secondary h-12 rounded-xl px-5 text-sm font-semibold'
                render={<Link to='/pricing' />}
              >
                {t('Model Square')}
              </Button>

              {isExternalDocs ? (
                <Button
                  variant='outline'
                  size='lg'
                  className='ob-btn-secondary h-12 rounded-xl px-4 text-sm font-medium'
                  render={
                    <a
                      href={docsUrl}
                      target='_blank'
                      rel='noopener noreferrer'
                    />
                  }
                >
                  <BookOpen className='mr-1.5 size-4' />
                  {t('Practical tutorials')}
                </Button>
              ) : (
                <Button
                  variant='outline'
                  size='lg'
                  className='ob-btn-secondary h-12 rounded-xl px-4 text-sm font-medium'
                  render={<Link to={docsUrl} />}
                >
                  <BookOpen className='mr-1.5 size-4' />
                  {t('Practical tutorials')}
                </Button>
              )}
            </div>
          </div>

          {/* Right Column: Live API Request/Response Terminal Demo (#14110F Dark Card) */}
          <div className='w-full min-w-0 lg:col-span-6'>
            <LandingTerminalDemo />
          </div>
        </div>

        {/* Bottom Metric Bar */}
        <ScrollReveal delay={0.2} className='mt-14 sm:mt-16'>
          <Card className='ob-card rounded-2xl py-0 shadow-none'>
            <CardContent className='grid grid-cols-2 gap-6 p-6 sm:grid-cols-4 sm:p-8'>
              {HERO_METRICS.map((item) => (
                <div
                  key={item.labelKey}
                  className='flex flex-col items-start sm:items-center sm:text-center'
                >
                  <span className='text-primary ob-mono text-2xl font-extrabold tracking-tight tabular-nums sm:text-3xl'>
                    {item.value}
                  </span>
                  <span className='text-muted-foreground mt-1 text-xs font-medium sm:text-sm'>
                    {t(item.labelKey)}
                  </span>
                </div>
              ))}
            </CardContent>
          </Card>
        </ScrollReveal>
      </div>
    </section>
  )
}
