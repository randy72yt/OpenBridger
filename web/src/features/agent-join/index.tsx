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
import {
  ArrowUpRight,
  BarChart3,
  Clock,
  Coins,
  Globe,
  Headset,
  Mail,
  QrCode,
  Sliders,
  Zap,
} from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import type { ComponentType } from 'react'
import { useTranslation } from 'react-i18next'

import { IconWeChat } from '@/assets/brand-icons/icon-wechat'

import { CopyButton } from '@/components/copy-button'
import { PublicLayout } from '@/components/layout'
import { PageTransition } from '@/components/page-transition'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ScrollReveal } from '@/features/landing/components/scroll-reveal'
import { useStatus } from '@/hooks/use-status'
import { useSystemConfig } from '@/hooks/use-system-config'
import { DEFAULT_SYSTEM_NAME } from '@/lib/constants'
import {
  isQrImageSource,
  resolvePlatformContact,
} from '@/lib/platform-contact'

interface AdvantageItem {
  id: string
  icon: ComponentType<{ className?: string }>
  titleKey: string
  descKey: string
}

const ADVANTAGE_ITEMS: readonly AdvantageItem[] = [
  {
    id: 'margin',
    icon: Coins,
    titleKey: 'agentJoin.advantages.margin.title',
    descKey: 'agentJoin.advantages.margin.desc',
  },
  {
    id: 'brand',
    icon: Globe,
    titleKey: 'agentJoin.advantages.brand.title',
    descKey: 'agentJoin.advantages.brand.desc',
  },
  {
    id: 'pricing',
    icon: Sliders,
    titleKey: 'agentJoin.advantages.pricing.title',
    descKey: 'agentJoin.advantages.pricing.desc',
  },
  {
    id: 'turnkey',
    icon: Zap,
    titleKey: 'agentJoin.advantages.turnkey.title',
    descKey: 'agentJoin.advantages.turnkey.desc',
  },
  {
    id: 'analytics',
    icon: BarChart3,
    titleKey: 'agentJoin.advantages.analytics.title',
    descKey: 'agentJoin.advantages.analytics.desc',
  },
  {
    id: 'support',
    icon: Headset,
    titleKey: 'agentJoin.advantages.support.title',
    descKey: 'agentJoin.advantages.support.desc',
  },
]

interface StepItem {
  number: string
  titleKey: string
  descKey: string
}

const ONBOARDING_STEPS: readonly StepItem[] = [
  {
    number: '01',
    titleKey: 'agentJoin.steps.step1.title',
    descKey: 'agentJoin.steps.step1.desc',
  },
  {
    number: '02',
    titleKey: 'agentJoin.steps.step2.title',
    descKey: 'agentJoin.steps.step2.desc',
  },
  {
    number: '03',
    titleKey: 'agentJoin.steps.step3.title',
    descKey: 'agentJoin.steps.step3.desc',
  },
  {
    number: '04',
    titleKey: 'agentJoin.steps.step4.title',
    descKey: 'agentJoin.steps.step4.desc',
  },
]

export function AgentJoinPage() {
  const { t } = useTranslation()
  const { systemName } = useSystemConfig()
  const { status } = useStatus()
  const displaySiteName = systemName || DEFAULT_SYSTEM_NAME
  const contact = resolvePlatformContact(status)
  const hasQrCode = contact.wechatQrCode.length > 0
  const isQrImage = isQrImageSource(contact.wechatQrCode)

  const scrollToSection = (sectionId: string) => {
    const el = document.querySelector(`#${sectionId}`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  }

  return (
    <PublicLayout showMainContainer={false}>
      <PageTransition className='relative isolate overflow-hidden'>
        {/* Top Warm Radial Glow & Subtle Grid */}
        <div
          aria-hidden='true'
          className='ob-subtle-grid pointer-events-none absolute inset-x-0 top-0 h-[560px]'
        />
        <div
          aria-hidden='true'
          className='pointer-events-none absolute inset-x-0 top-0 h-[580px]'
          style={{
            background: [
              'radial-gradient(ellipse 65% 55% at 50% 12%, rgba(255, 107, 0, 0.14) 0%, transparent 72%)',
              'radial-gradient(ellipse 45% 35% at 80% 25%, rgba(245, 158, 11, 0.10) 0%, transparent 70%)',
            ].join(', '),
          }}
        />

        {/* 1. Centered Hero Section */}
        <section className='relative mx-auto max-w-4xl px-4 pt-32 pb-16 text-center sm:px-6 sm:pt-40 sm:pb-24 lg:px-8'>
          <div className='ob-chip-primary mb-6 inline-flex items-center gap-2 rounded-full px-4 py-1.5 text-xs font-semibold shadow-xs'>
            <span className='relative flex size-2'>
              <span className='ob-dot-primary absolute inline-flex h-full w-full animate-ping rounded-full opacity-75' />
              <span className='ob-dot-primary relative inline-flex size-2 rounded-full' />
            </span>
            <span>{t('agentJoin.hero.badge')}</span>
          </div>

          <h1 className='ob-serif-heading text-foreground text-[clamp(2.25rem,4.6vw,3.75rem)] leading-[1.15] font-bold tracking-tight'>
            {t('agentJoin.hero.titleLine1')}
            <br />
            <span className='ob-hero-gradient-text'>
              {t('agentJoin.hero.titleLine2')}
            </span>
          </h1>

          <p className='text-muted-foreground/90 mt-3 text-base font-medium tracking-tight italic sm:text-lg'>
            {t('agentJoin.hero.italicSubtitle')}
          </p>

          <p className='text-muted-foreground mx-auto mt-5 max-w-2xl text-sm leading-relaxed sm:text-base'>
            {t('agentJoin.hero.description')}
          </p>

          <div className='mt-9 flex flex-wrap items-center justify-center gap-3.5'>
            <Button
              size='lg'
              className='ob-btn-primary group h-11 rounded-full px-7 text-sm font-semibold'
              onClick={() => scrollToSection('partner-contact')}
            >
              {t('agentJoin.hero.ctaPrimary')}
              <ArrowUpRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5 group-hover:-translate-y-0.5' />
            </Button>

            <Button
              variant='outline'
              size='lg'
              className='ob-btn-secondary h-11 rounded-full px-6 text-sm font-semibold'
              onClick={() => scrollToSection('partner-advantages')}
            >
              {t('agentJoin.hero.ctaSecondary')}
            </Button>
          </div>
        </section>

        {/* 2. Six Core Advantages Section */}
        <section
          id='partner-advantages'
          className='relative mx-auto max-w-7xl scroll-mt-24 px-4 py-14 sm:px-6 sm:py-20 lg:px-8'
        >
          <ScrollReveal className='mx-auto max-w-2xl text-center'>
            <span className='ob-chip-primary inline-flex items-center gap-1.5 rounded-full px-3.5 py-1 text-xs font-semibold'>
              {t('agentJoin.advantages.eyebrow')}
            </span>
            <h2 className='ob-serif-heading text-foreground mt-4 text-3xl font-bold tracking-tight sm:text-4xl'>
              {t('agentJoin.advantages.title')}
            </h2>
            <p className='text-muted-foreground mt-3 text-sm leading-relaxed sm:text-base'>
              {t('agentJoin.advantages.subtitle')}
            </p>
          </ScrollReveal>

          <div className='mt-12 grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3'>
            {ADVANTAGE_ITEMS.map((item, idx) => {
              const Icon = item.icon
              return (
                <ScrollReveal key={item.id} delay={0.04 * idx}>
                  <Card className='ob-card ob-card-interactive h-full rounded-3xl py-0 shadow-none'>
                    <CardContent className='flex h-full flex-col gap-4 p-7 sm:p-8'>
                      <div className='ob-chip-primary flex size-12 items-center justify-center rounded-2xl'>
                        <Icon className='size-5' />
                      </div>
                      <h3 className='text-foreground text-lg font-bold tracking-tight'>
                        {t(item.titleKey)}
                      </h3>
                      <p className='text-muted-foreground text-sm leading-7'>
                        {t(item.descKey)}
                      </p>
                    </CardContent>
                  </Card>
                </ScrollReveal>
              )
            })}
          </div>
        </section>

        {/* 3. Four-Step Onboarding Workflow Section */}
        <section className='relative mx-auto max-w-7xl px-4 py-10 sm:px-6 sm:py-16 lg:px-8'>
          <ScrollReveal>
            <Card className='ob-card rounded-[2rem] py-0 shadow-[0_20px_50px_-28px_rgba(20,17,15,0.08)]'>
              <CardContent className='p-8 sm:p-12 lg:p-16'>
                <div className='mx-auto max-w-2xl text-center'>
                  <span className='ob-chip-primary inline-flex items-center gap-1.5 rounded-full px-3.5 py-1 text-xs font-semibold'>
                    {t('agentJoin.steps.eyebrow')}
                  </span>
                  <h2 className='ob-serif-heading text-foreground mt-4 text-3xl font-bold tracking-tight sm:text-4xl'>
                    {t('agentJoin.steps.title')}
                  </h2>
                  <p className='text-muted-foreground mt-3 text-sm leading-relaxed sm:text-base'>
                    {t('agentJoin.steps.subtitle')}
                  </p>
                </div>

                <ol className='mt-12 grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-4 lg:gap-6'>
                  {ONBOARDING_STEPS.map((step, index) => {
                    const isLast = index === ONBOARDING_STEPS.length - 1
                    return (
                      <li
                        key={step.number}
                        className='relative flex flex-col items-center text-center lg:items-start lg:text-left'
                      >
                        <div className='flex w-full items-center justify-center lg:justify-start'>
                          <span className='ob-btn-primary ob-mono flex size-12 shrink-0 items-center justify-center rounded-full text-sm font-bold'>
                            {step.number}
                          </span>
                          {!isLast && (
                            <span
                              aria-hidden='true'
                              className='border-primary/30 ml-4 hidden flex-1 border-t-2 border-dashed lg:block'
                            />
                          )}
                        </div>
                        <h3 className='text-foreground mt-5 text-base font-bold tracking-tight sm:text-lg'>
                          {t(step.titleKey)}
                        </h3>
                        <p className='text-muted-foreground mt-2 text-sm leading-relaxed'>
                          {t(step.descKey)}
                        </p>
                      </li>
                    )
                  })}
                </ol>
              </CardContent>
            </Card>
          </ScrollReveal>
        </section>

        {/* 4. Business Contact / WeChat QR Section */}
        <section
          id='partner-contact'
          className='relative mx-auto max-w-5xl scroll-mt-24 px-4 pt-12 pb-20 sm:px-6 sm:pt-16 sm:pb-28 lg:px-8'
        >
          <ScrollReveal className='mx-auto max-w-2xl text-center'>
            <span className='ob-chip-primary inline-flex items-center gap-1.5 rounded-full px-3.5 py-1 text-xs font-semibold'>
              {t('agentJoin.contact.eyebrow')}
            </span>
            <h2 className='ob-serif-heading text-foreground mt-4 text-3xl font-bold tracking-tight sm:text-4xl'>
              {t('agentJoin.contact.title')}
            </h2>
            <p className='text-muted-foreground mt-3 text-sm sm:text-base'>
              {t('agentJoin.contact.subtitle')}
            </p>
          </ScrollReveal>

          <ScrollReveal delay={0.1} className='mt-10'>
            <Card className='ob-card rounded-3xl py-0 shadow-[0_20px_50px_-24px_rgba(20,17,15,0.08)]'>
              <CardContent className='grid grid-cols-1 items-center gap-8 p-8 sm:p-12 md:grid-cols-[1fr_auto]'>
                {/* Left Column: Support Email + WeChat ID + Hours */}
                <div className='space-y-5'>
                  <div className='space-y-1.5'>
                    <p className='text-muted-foreground text-xs font-semibold tracking-wider uppercase'>
                      {t('agentJoin.contact.cardLabel')}
                    </p>
                    <h3 className='text-foreground text-2xl font-bold tracking-tight sm:text-3xl'>
                      {displaySiteName}
                    </h3>
                  </div>

                  <div className='space-y-3.5 pt-1 text-sm'>
                    <div className='flex items-center gap-3'>
                      <span className='ob-chip-primary flex size-9 shrink-0 items-center justify-center rounded-xl'>
                        <Mail className='size-4' />
                      </span>
                      <div className='flex min-w-0 flex-wrap items-center gap-2'>
                        <span className='text-muted-foreground'>
                          {t('Contact email')}:
                        </span>
                        {contact.email ? (
                          <>
                            <a
                              href={`mailto:${contact.email}`}
                              className='text-foreground hover:text-primary ob-mono truncate font-medium underline-offset-4 hover:underline'
                            >
                              {contact.email}
                            </a>
                            <CopyButton value={contact.email} />
                          </>
                        ) : (
                          <span className='text-foreground font-medium'>
                            {t('To be announced')}
                          </span>
                        )}
                      </div>
                    </div>

                    <div className='flex items-center gap-3'>
                      <span className='flex size-9 shrink-0 items-center justify-center rounded-xl bg-[#07C160]/12 text-[#07C160]'>
                        <IconWeChat className='size-4.5' aria-hidden='true' />
                      </span>
                      <div className='flex min-w-0 flex-wrap items-center gap-2'>
                        <span className='text-muted-foreground'>
                          {t('agentJoin.contact.wechatLabel')}:
                        </span>
                        {contact.wechatId ? (
                          <>
                            <span className='text-foreground ob-mono font-medium'>
                              {contact.wechatId}
                            </span>
                            <CopyButton value={contact.wechatId} />
                          </>
                        ) : (
                          <span className='text-muted-foreground font-medium'>
                            {t('agentJoin.contact.wechatScanHint')}
                          </span>
                        )}
                      </div>
                    </div>

                    <div className='flex items-center gap-3'>
                      <span className='flex size-9 shrink-0 items-center justify-center rounded-xl bg-amber-500/12 text-amber-600 dark:text-amber-400'>
                        <Clock className='size-4' />
                      </span>
                      <span className='text-foreground font-medium'>
                        {t('agentJoin.contact.hours')}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Right Column: Dedicated WeChat QR Code Card */}
                <div className='flex flex-col items-center gap-2.5 justify-self-center md:justify-self-end'>
                  {hasQrCode ? (
                    <div
                      className='rounded-2xl border border-black/[0.08] bg-white p-3.5 shadow-xs'
                      data-testid='partner-wechat-qr'
                    >
                      {isQrImage ? (
                        <img
                          src={contact.wechatQrCode}
                          alt={t('agentJoin.contact.scanWechatQr')}
                          className='size-38 rounded-lg object-contain'
                        />
                      ) : (
                        <QRCodeSVG
                          value={contact.wechatQrCode}
                          size={152}
                          level='M'
                        />
                      )}
                    </div>
                  ) : (
                    <div
                      className='ob-section-warm flex size-44 flex-col items-center justify-center gap-2.5 rounded-2xl border border-dashed border-[var(--ob-line)] p-4 text-center'
                      data-testid='partner-wechat-qr-placeholder'
                    >
                      <div className='ob-chip-primary flex size-10 items-center justify-center rounded-xl'>
                        <QrCode className='size-5' />
                      </div>
                      <span className='text-muted-foreground text-xs leading-relaxed'>
                        {t('agentJoin.contact.qrPending')}
                      </span>
                    </div>
                  )}
                  <span className='text-muted-foreground text-xs font-medium'>
                    {t('agentJoin.contact.scanWechatQr')}
                  </span>
                </div>
              </CardContent>
            </Card>
          </ScrollReveal>
        </section>
      </PageTransition>
    </PublicLayout>
  )
}
