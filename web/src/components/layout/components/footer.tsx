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

import { PlatformContact } from '@/components/platform-contact'
import { Card, CardContent } from '@/components/ui/card'
import { useSystemConfig } from '@/hooks/use-system-config'
import { DEFAULT_SYSTEM_NAME } from '@/lib/constants'
import { cn } from '@/lib/utils'

interface FooterProps {
  name?: string
  copyright?: string
  className?: string
}

const NEW_API_FOOTER_ATTRIBUTION_KEY = [
  'footer',
  'new' + 'api',
  'projectAttributionSuffix',
].join('.')

export function Footer(props: FooterProps) {
  const { t } = useTranslation()
  const { systemName, footerHtml } = useSystemConfig()
  const displayName = systemName || props.name || DEFAULT_SYSTEM_NAME
  const currentYear = new Date().getFullYear()

  return (
    <footer className={cn('relative z-10 px-5 py-8 sm:px-8', props.className)}>
      <Card className='bg-card/70 mx-auto max-w-screen-2xl rounded-2xl py-5 shadow-none xl:rounded-full'>
        <CardContent className='flex flex-col gap-4 px-6 text-xs leading-relaxed sm:px-8 xl:flex-row xl:items-center xl:justify-between'>
          <div className='text-foreground/80 flex min-w-0 flex-wrap items-center gap-x-5 gap-y-2 font-medium'>
            {footerHtml ? (
              <div
                className='custom-footer min-w-0'
                dangerouslySetInnerHTML={{ __html: footerHtml }}
              />
            ) : (
              <span>
                {displayName} {props.copyright ?? t('footer.defaultCopyright')}
              </span>
            )}
            <PlatformContact />
          </div>
          <div className='text-muted-foreground border-border flex min-w-0 flex-wrap items-center gap-x-4 gap-y-2 border-t pt-4 xl:border-t-0 xl:border-l xl:pt-0 xl:pl-6'>
            <nav className='flex items-center gap-3'>
              <Link
                to='/user-agreement'
                className='hover:text-foreground focus-visible:ring-ring rounded-sm underline-offset-4 transition-colors hover:underline focus-visible:ring-2 focus-visible:outline-none'
              >
                {t('User Agreement')}
              </Link>
              <span aria-hidden='true' className='text-muted-foreground/50'>
                ·
              </span>
              <Link
                to='/privacy-policy'
                className='hover:text-foreground focus-visible:ring-ring rounded-sm underline-offset-4 transition-colors hover:underline focus-visible:ring-2 focus-visible:outline-none'
              >
                {t('Privacy Policy')}
              </Link>
            </nav>
            <span>
              &copy; {currentYear}{' '}
              <a
                href='https://github.com/QuantumNous/new-api'
                target='_blank'
                rel='noopener noreferrer'
                className='hover:text-foreground focus-visible:ring-ring rounded-sm underline-offset-4 transition-colors hover:underline focus-visible:ring-2 focus-visible:outline-none'
              >
                New API
              </a>
              . {t(NEW_API_FOOTER_ATTRIBUTION_KEY)}
            </span>
          </div>
        </CardContent>
      </Card>
    </footer>
  )
}
