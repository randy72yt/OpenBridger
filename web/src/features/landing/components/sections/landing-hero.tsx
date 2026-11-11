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
import { HeroTerminalDemo } from '@/features/home/components/hero-terminal-demo'
import { useStatus } from '@/hooks/use-status'

import { HERO_HIGHLIGHT_PILLS, HERO_METRICS } from '../../constants'
import { HeroBackground } from '../hero-background'
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
    <section className='relative isolate overflow-hidden pt-28 pb-16 sm:pt-32 sm:pb-24 lg:pt-36 lg:pb-28'>
      <HeroBackground />

      <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
        <div className='grid grid-cols-1 items-center gap-12 lg:grid-cols-12 lg:gap-10'>
          {/* Left Column: Brand pitch, CTAs, and capability pills */}
          <div className='flex min-w-0 flex-col items-start text-left lg:col-span-6'>
            {/* Status Pill */}
            <div className='border-primary/30 bg-primary/10 text-primary mb-6 inline-flex items-center gap-2 rounded-full border px-3.5 py-1.5 text-xs font-semibold shadow-2xs backdrop-blur-sm'>
              <span className='relative flex size-2'>
                <span className='bg-primary absolute inline-flex h-full w-full animate-ping rounded-full opacity-75' />
                <span className='bg-primary relative inline-flex size-2 rounded-full' />
              </span>
              <span>{t('landing.hero.badge')}</span>
            </div>

            <h1 className='text-[clamp(2.25rem,4.5vw,3.75rem)] leading-[1.1] font-extrabold tracking-tight'>
              {t('landing.hero.titleLine1')}
              <br />
              <span className='from-primary via-primary bg-gradient-to-r to-emerald-400 bg-clip-text text-transparent'>
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
                  className='border-border/70 bg-card/75 text-foreground/85 inline-flex items-center gap-1.5 rounded-xl border px-3 py-1.5 text-xs font-medium backdrop-blur-sm'
                >
                  {idx === 0 && <Sparkles className='text-primary size-3.5' />}
                  {idx === 1 && (
                    <ShieldCheck className='text-primary size-3.5' />
                  )}
                  {idx === 2 && <Zap className='text-primary size-3.5' />}
                  {t(pillKey)}
                </span>
              ))}
            </div>

            {/* Primary & Secondary CTAs */}
            <div className='mt-8 flex flex-wrap items-center gap-3.5'>
              {props.isAuthenticated ? (
                <Button
                  size='lg'
                  className='group h-12 rounded-xl px-6 text-sm font-semibold shadow-md'
                  render={<Link to='/dashboard' />}
                >
                  {t('Go to Dashboard')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
              ) : (
                <Button
                  size='lg'
                  className='group h-12 rounded-xl px-6 text-sm font-semibold shadow-md'
                  render={<Link to='/sign-up' />}
                >
                  {t('landing.hero.ctaPrimary')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
              )}

              <Button
                variant='outline'
                size='lg'
                className='border-border/80 bg-card/60 hover:bg-muted/60 h-12 rounded-xl px-5 text-sm font-semibold backdrop-blur-sm'
                render={<Link to='/pricing' />}
              >
                {t('Model Square')}
              </Button>

              {isExternalDocs ? (
                <Button
                  variant='ghost'
                  size='lg'
                  className='h-12 rounded-xl px-4 text-sm font-medium'
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
                  variant='ghost'
                  size='lg'
                  className='h-12 rounded-xl px-4 text-sm font-medium'
                  render={<Link to={docsUrl} />}
                >
                  <BookOpen className='mr-1.5 size-4' />
                  {t('Practical tutorials')}
                </Button>
              )}
            </div>
          </div>

          {/* Right Column: Live API Request/Response Terminal Demo */}
          <div className='w-full min-w-0 lg:col-span-6'>
            <HeroTerminalDemo />
          </div>
        </div>

        {/* Bottom Metric Bar */}
        <ScrollReveal delay={0.2} className='mt-14 sm:mt-16'>
          <Card className='border-border/60 bg-card/65 rounded-2xl py-0 shadow-none backdrop-blur-md'>
            <CardContent className='grid grid-cols-2 gap-6 p-6 sm:grid-cols-4 sm:p-8'>
              {HERO_METRICS.map((item) => (
                <div
                  key={item.labelKey}
                  className='flex flex-col items-start sm:items-center sm:text-center'
                >
                  <span className='text-primary font-mono text-2xl font-extrabold tracking-tight tabular-nums sm:text-3xl'>
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
