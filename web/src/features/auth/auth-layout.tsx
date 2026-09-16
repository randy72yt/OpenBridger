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

import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'
import { cn } from '@/lib/utils'

type AuthLayoutProps = {
  children: React.ReactNode
  showcase?: boolean
}

export function AuthLayout({ children, showcase = false }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <div className='bg-background relative grid min-h-svh max-w-none'>
      <Link
        to='/'
        className='absolute top-4 left-4 z-10 flex items-center gap-2 transition-opacity hover:opacity-80 sm:top-8 sm:left-8'
      >
        <div className='relative h-8 w-8'>
          {loading ? (
            <Skeleton className='absolute inset-0 rounded-full' />
          ) : (
            <img
              src={logo}
              alt={t('Logo')}
              className='h-8 w-8 rounded-full object-cover'
            />
          )}
        </div>
        {loading ? (
          <Skeleton className='h-6 w-24' />
        ) : (
          <h1 className='text-xl font-medium'>{systemName}</h1>
        )}
      </Link>
      <div
        className={cn(
          'container mx-auto flex items-center px-5 pt-24 pb-10 sm:px-8',
          showcase && 'max-w-6xl gap-16 lg:grid lg:grid-cols-2'
        )}
      >
        {showcase && (
          <section className='hidden space-y-8 lg:block'>
            <p className='text-primary text-sm font-semibold tracking-wider'>
              {t('AI application infrastructure')}
            </p>
            <h2 className='text-4xl leading-tight font-semibold tracking-tight xl:text-5xl'>
              {t('Connect to world-class AI,')}
              <br />
              {t('build limitless possibilities')}
            </h2>
            <div className='bg-primary h-1 w-12 rounded-full' />
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
          </section>
        )}
        <Card
          className={cn(
            'mx-auto w-full max-w-[480px] rounded-3xl py-0 shadow-none',
            !showcase && 'bg-transparent ring-0'
          )}
        >
          <CardContent className='px-6 py-8 sm:p-10'>{children}</CardContent>
        </Card>
      </div>
    </div>
  )
}
