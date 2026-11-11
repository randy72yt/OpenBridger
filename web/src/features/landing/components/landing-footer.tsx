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

import { HtmlContent } from '@/components/html-content'
import { HeaderLogo } from '@/components/layout/components/header-logo'
import { PlatformContact } from '@/components/platform-contact'
import { useSystemConfig } from '@/hooks/use-system-config'
import { DEFAULT_SYSTEM_NAME } from '@/lib/constants'
import { cn } from '@/lib/utils'

interface LandingFooterProps {
  className?: string
}

export function LandingFooter(props: LandingFooterProps) {
  const { t } = useTranslation()
  const {
    systemName,
    logo: systemLogo,
    loading,
    logoLoaded,
    footerHtml,
  } = useSystemConfig()
  const displayName = systemName || DEFAULT_SYSTEM_NAME
  const currentYear = new Date().getFullYear()

  return (
    <footer
      className={cn(
        'ob-footer-dark relative z-10 border-t px-4 pt-14 pb-10 sm:px-6 lg:px-8',
        props.className
      )}
    >
      <div className='mx-auto max-w-7xl'>
        <div className='grid grid-cols-1 gap-10 pb-10 md:grid-cols-12'>
          {/* Brand and Description */}
          <div className='space-y-4 md:col-span-5'>
            <div className='flex items-center gap-2.5'>
              <div className='flex size-8 items-center justify-center rounded-xl border border-white/15 bg-white/5'>
                <HeaderLogo
                  src={systemLogo}
                  loading={loading}
                  logoLoaded={logoLoaded}
                  className='size-5 rounded-md object-contain'
                />
              </div>
              <span className='ob-footer-title text-base font-bold tracking-tight'>
                {displayName}
              </span>
            </div>
            <p className='max-w-sm text-sm leading-relaxed'>
              {t('landing.footer.tagline')}
            </p>
            <div className='flex flex-wrap items-center gap-x-5 gap-y-2 text-xs'>
              <PlatformContact />
            </div>
          </div>

          {/* Product Navigation */}
          <div className='space-y-3 md:col-span-3'>
            <h3 className='ob-footer-title text-xs font-semibold tracking-wider uppercase'>
              {t('landing.footer.productHeading')}
            </h3>
            <ul className='space-y-2.5 text-sm'>
              <li>
                <Link to='/pricing' className='ob-footer-link'>
                  {t('Model Square')}
                </Link>
              </li>
              <li>
                <Link to='/dashboard' className='ob-footer-link'>
                  {t('Console')}
                </Link>
              </li>
              <li>
                <Link to='/docs' className='ob-footer-link'>
                  {t('Practical tutorials')}
                </Link>
              </li>
              <li>
                <Link to='/about' className='ob-footer-link'>
                  {t('About')}
                </Link>
              </li>
            </ul>
          </div>

          {/* Protocols & Compliance */}
          <div className='space-y-3 md:col-span-4'>
            <h3 className='ob-footer-title text-xs font-semibold tracking-wider uppercase'>
              {t('landing.footer.legalHeading')}
            </h3>
            <ul className='space-y-2.5 text-sm'>
              <li>
                <Link to='/user-agreement' className='ob-footer-link'>
                  {t('User Agreement')}
                </Link>
              </li>
              <li>
                <Link to='/privacy-policy' className='ob-footer-link'>
                  {t('Privacy Policy')}
                </Link>
              </li>
            </ul>
            {footerHtml ? (
              <HtmlContent content={footerHtml} className='pt-2 text-xs' />
            ) : null}
          </div>
        </div>

        {/* Bottom Attribution Bar — Mandatory NOTICE preservation */}
        <div
          className='flex flex-col items-start justify-between gap-4 border-t pt-6 text-xs sm:flex-row sm:items-center'
          style={{ borderColor: 'var(--ob-footer-line)' }}
        >
          <div>
            &copy; {currentYear} {displayName}. {t('footer.defaultCopyright')}
          </div>
          <div className='flex flex-wrap items-center gap-1.5'>
            <span>Powered by</span>
            <a
              href='https://github.com/QuantumNous/new-api'
              target='_blank'
              rel='noopener noreferrer'
              className='text-primary font-medium underline-offset-4 transition-colors hover:underline'
            >
              New API
            </a>
            <span>·</span>
            <span>
              Frontend design and development by New API contributors.
            </span>
          </div>
        </div>
      </div>
    </footer>
  )
}
