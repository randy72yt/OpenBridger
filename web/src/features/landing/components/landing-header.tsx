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
import { Link, useRouterState } from '@tanstack/react-router'
import { ArrowRight, Menu, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '@/components/language-switcher'
import { HeaderLogo } from '@/components/layout/components/header-logo'
import { NotificationPopover } from '@/components/notification-popover'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useNotifications } from '@/hooks/use-notifications'
import { useSystemConfig } from '@/hooks/use-system-config'
import { useTopNavLinks } from '@/hooks/use-top-nav-links'
import { DEFAULT_SYSTEM_NAME } from '@/lib/constants'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

interface LandingHeaderProps {
  className?: string
}

export function LandingHeader(props: LandingHeaderProps) {
  const { t } = useTranslation()
  const [scrolled, setScrolled] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const user = useAuthStore((s) => s.auth.user)
  const isAuthenticated = Boolean(user)
  const {
    systemName,
    logo: systemLogo,
    loading,
    logoLoaded,
  } = useSystemConfig()
  const dynamicLinks = useTopNavLinks()
  const notifications = useNotifications()
  const routerState = useRouterState()
  const pathname = routerState.location.pathname

  const displaySiteName = systemName || DEFAULT_SYSTEM_NAME

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 16)
    handleScroll()
    window.addEventListener('scroll', handleScroll, { passive: true })
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  const fallbackLinks = [
    { title: t('Home'), href: '/' },
    { title: t('Model Square'), href: '/pricing' },
    { title: t('Practical tutorials'), href: '/docs' },
    { title: t('About'), href: '/about' },
  ]

  const navLinks = dynamicLinks.length > 0 ? dynamicLinks : fallbackLinks

  return (
    <header
      className={cn(
        'fixed inset-x-0 top-0 z-50 transition-all duration-300',
        scrolled ? 'py-2.5' : 'py-4',
        props.className
      )}
    >
      <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
        <div
          className={cn(
            'flex h-14 items-center justify-between rounded-2xl border px-4 transition-all duration-300 sm:px-5',
            scrolled
              ? 'border-border/70 bg-background/85 shadow-[0_8px_32px_-8px_rgba(0,0,0,0.12)] backdrop-blur-xl dark:shadow-[0_8px_32px_-8px_rgba(0,0,0,0.5)]'
              : 'border-border/40 bg-background/60 backdrop-blur-md'
          )}
        >
          {/* Brand Logo */}
          <Link
            to='/'
            className='group flex shrink-0 items-center gap-2.5 focus-visible:outline-none'
          >
            <div className='bg-primary/10 ring-primary/25 flex size-8 shrink-0 items-center justify-center rounded-xl ring-1 transition-transform duration-200 group-hover:scale-105'>
              {loading ? (
                <Skeleton className='size-5 rounded-md' />
              ) : (
                <HeaderLogo
                  src={systemLogo}
                  loading={loading}
                  logoLoaded={logoLoaded}
                  className='size-5 rounded-md object-contain'
                />
              )}
            </div>
            <div className='flex items-center gap-2'>
              <span className='text-base font-bold tracking-tight'>
                {loading ? <Skeleton className='h-4 w-24' /> : displaySiteName}
              </span>
              <span className='border-primary/30 bg-primary/10 text-primary hidden rounded-full border px-2 py-0.5 font-mono text-[10px] font-semibold tracking-wider uppercase sm:inline-block'>
                API Gateway
              </span>
            </div>
          </Link>

          {/* Desktop Navigation Links */}
          <nav
            aria-label={t('landing.nav.ariaLabel')}
            className='hidden items-center gap-1 md:flex'
          >
            {navLinks.map((link) => {
              const isActive = pathname === link.href
              if ('external' in link && link.external) {
                return (
                  <a
                    key={link.href}
                    href={link.href}
                    target='_blank'
                    rel='noopener noreferrer'
                    className='text-muted-foreground hover:text-foreground hover:bg-muted/50 rounded-lg px-3.5 py-1.5 text-sm font-medium transition-colors'
                  >
                    {link.title}
                  </a>
                )
              }
              return (
                <Link
                  key={link.href}
                  to={link.href}
                  className={cn(
                    'rounded-lg px-3.5 py-1.5 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-primary/10 text-primary font-semibold'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
                  )}
                >
                  {link.title}
                </Link>
              )
            })}
          </nav>

          {/* Right Actions */}
          <div className='hidden items-center gap-1.5 md:flex'>
            <LanguageSwitcher />
            <ThemeSwitch />
            <NotificationPopover
              open={notifications.popoverOpen}
              onOpenChange={notifications.setPopoverOpen}
              unreadCount={notifications.unreadCount}
              activeTab={notifications.activeTab}
              onTabChange={notifications.setActiveTab}
              notice={notifications.notice}
              announcements={notifications.announcements}
              loading={notifications.loading}
            />
            <div className='bg-border/60 mx-1.5 h-4 w-px' />
            {loading && <Skeleton className='h-9 w-24 rounded-xl' />}
            {!loading && isAuthenticated && (
              <div className='flex items-center gap-2'>
                <Button
                  size='sm'
                  className='h-9 rounded-xl px-4 text-xs font-semibold'
                  render={<Link to='/dashboard' />}
                >
                  {t('Go to Dashboard')}
                  <ArrowRight className='ml-1 size-3.5' />
                </Button>
                <ProfileDropdown />
              </div>
            )}
            {!loading && !isAuthenticated && (
              <div className='flex items-center gap-2'>
                <Button
                  variant='ghost'
                  size='sm'
                  className='h-9 rounded-xl px-3.5 text-xs font-medium'
                  render={<Link to='/sign-in' />}
                >
                  {t('Sign in')}
                </Button>
                <Button
                  size='sm'
                  className='h-9 rounded-xl px-4 text-xs font-semibold shadow-sm'
                  render={<Link to='/sign-up' />}
                >
                  {t('Get Started')}
                  <ArrowRight className='ml-1 size-3.5' />
                </Button>
              </div>
            )}
          </div>

          {/* Mobile Controls */}
          <div className='flex items-center gap-1.5 md:hidden'>
            <LanguageSwitcher />
            <ThemeSwitch />
            <Button
              type='button'
              variant='ghost'
              size='icon'
              className='size-9 rounded-xl'
              onClick={() => setMobileMenuOpen((prev) => !prev)}
              aria-expanded={mobileMenuOpen}
              aria-label={t('Toggle navigation menu')}
            >
              {mobileMenuOpen ? (
                <X className='size-4' />
              ) : (
                <Menu className='size-4' />
              )}
            </Button>
          </div>
        </div>

        {/* Mobile Dropdown Panel */}
        {mobileMenuOpen && (
          <div className='border-border/70 bg-background/95 mt-2 rounded-2xl border p-4 shadow-xl backdrop-blur-xl md:hidden'>
            <nav className='flex flex-col gap-1'>
              {navLinks.map((link) => {
                if ('external' in link && link.external) {
                  return (
                    <a
                      key={link.href}
                      href={link.href}
                      target='_blank'
                      rel='noopener noreferrer'
                      onClick={() => setMobileMenuOpen(false)}
                      className='text-muted-foreground hover:text-foreground hover:bg-muted/50 rounded-xl px-3 py-2.5 text-sm font-medium'
                    >
                      {link.title}
                    </a>
                  )
                }
                return (
                  <Link
                    key={link.href}
                    to={link.href}
                    onClick={() => setMobileMenuOpen(false)}
                    className='text-muted-foreground hover:text-foreground hover:bg-muted/50 rounded-xl px-3 py-2.5 text-sm font-medium'
                  >
                    {link.title}
                  </Link>
                )
              })}
            </nav>
            <div className='border-border/60 mt-3 flex flex-col gap-2 border-t pt-3'>
              {isAuthenticated ? (
                <Button
                  className='w-full rounded-xl'
                  render={<Link to='/dashboard' />}
                  onClick={() => setMobileMenuOpen(false)}
                >
                  {t('Go to Dashboard')}
                </Button>
              ) : (
                <>
                  <Button
                    variant='outline'
                    className='w-full rounded-xl'
                    render={<Link to='/sign-in' />}
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    {t('Sign in')}
                  </Button>
                  <Button
                    className='w-full rounded-xl'
                    render={<Link to='/sign-up' />}
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    {t('Get Started')}
                  </Button>
                </>
              )}
            </div>
          </div>
        )}
      </div>
    </header>
  )
}
