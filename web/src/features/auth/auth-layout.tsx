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
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '@/components/language-switcher'
import { HeaderLogo } from '@/components/layout/components/header-logo'
import { ThemeSwitch } from '@/components/theme-switch'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'
import { DEFAULT_SYSTEM_NAME } from '@/lib/constants'
import { cn } from '@/lib/utils'

type AuthLayoutProps = {
  children: React.ReactNode
  showcase?: boolean
}

const SHOWCASE_PROVIDER_PILLS = [
  'OpenAI',
  'Claude',
  'Gemini',
  'DeepSeek',
  'Qwen',
  'Grok',
] as const

export function AuthLayout(props: AuthLayoutProps) {
  const showcase = props.showcase ?? false
  const { t } = useTranslation()
  const { systemName, logo, loading, logoLoaded } = useSystemConfig()
  const displayName = systemName || DEFAULT_SYSTEM_NAME

  return (
    <div className='ob-landing-theme bg-background text-foreground relative grid min-h-svh max-w-none overflow-hidden'>
      <div
        aria-hidden='true'
        className='pointer-events-none absolute inset-0'
        style={{
          background: [
            'radial-gradient(ellipse 55% 45% at 15% 20%, rgba(255, 107, 0, 0.10) 0%, transparent 70%)',
            'radial-gradient(ellipse 45% 40% at 85% 80%, rgba(245, 158, 11, 0.08) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      <Link
        to='/'
        className='absolute top-4 left-4 z-10 flex items-center gap-2.5 transition-opacity hover:opacity-85 sm:top-8 sm:left-8'
      >
        <div className='ob-chip-primary flex size-8 shrink-0 items-center justify-center rounded-full'>
          {loading ? (
            <Skeleton className='size-5 rounded-full' />
          ) : (
            <HeaderLogo
              src={logo}
              loading={loading}
              logoLoaded={logoLoaded}
              className='size-5 rounded-md object-contain'
            />
          )}
        </div>
        {loading ? (
          <Skeleton className='h-6 w-24' />
        ) : (
          <h1 className='text-lg font-bold tracking-tight'>{displayName}</h1>
        )}
      </Link>

      <div className='absolute top-4 right-4 z-10 flex items-center gap-1.5 sm:top-7 sm:right-8'>
        <LanguageSwitcher />
        <ThemeSwitch />
      </div>

      <div
        className={cn(
          'relative z-10 container mx-auto flex items-center px-5 pt-24 pb-10 sm:px-8',
          showcase && 'max-w-6xl gap-16 lg:grid lg:grid-cols-2'
        )}
      >
        {showcase && (
          <section className='hidden space-y-7 lg:block'>
            <p className='ob-chip-primary inline-flex items-center gap-2 rounded-full px-3.5 py-1 text-xs font-semibold tracking-wider'>
              <span className='ob-dot-primary size-1.5 rounded-full' />
              {t('AI application infrastructure')}
            </p>
            <h2 className='ob-serif-heading text-4xl leading-tight font-bold tracking-tight xl:text-5xl'>
              {t('Connect to world-class AI,')}
              <br />
              <span className='ob-hero-gradient-text'>
                {t('build limitless possibilities')}
              </span>
            </h2>
            <div className='max-w-md space-y-3'>
              <p className='text-foreground/90 text-lg leading-relaxed font-medium'>
                {t(
                  'From inspiration to creation, make AI your creative power.'
                )}
              </p>
              <p className='text-muted-foreground text-sm leading-7'>
                {t(
                  'Access multiple AI models with one account for coding assistance, learning and exploration, and application development.'
                )}
              </p>
            </div>
            <div className='flex flex-wrap items-center gap-2 pt-1'>
              {SHOWCASE_PROVIDER_PILLS.map((provider) => (
                <span
                  key={provider}
                  className='ob-chip-primary ob-mono rounded-full px-3 py-1 text-xs font-medium'
                >
                  {provider}
                </span>
              ))}
            </div>
          </section>
        )}
        <Card className='ob-card mx-auto w-full max-w-[480px] rounded-3xl py-0 shadow-[0_20px_50px_-24px_rgba(20,17,15,0.12)] dark:shadow-[0_20px_50px_-24px_rgba(0,0,0,0.5)]'>
          <CardContent className='px-6 py-8 sm:p-10'>
            {props.children}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
