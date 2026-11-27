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
import { Link, useNavigate, useRouterState } from '@tanstack/react-router'
import { ArrowUpRight, Menu, X } from 'lucide-react'
import { useCallback, useEffect, useId, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { LanguageSwitcher } from '@/components/language-switcher'
import { NotificationPopover } from '@/components/notification-popover'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useNotifications } from '@/hooks/use-notifications'
import { useSystemConfig } from '@/hooks/use-system-config'
import { useTopNavLinks } from '@/hooks/use-top-nav-links'
import { DEFAULT_SYSTEM_NAME } from '@/lib/constants'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import { defaultTopNavLinks } from '../config/top-nav.config'
import type { TopNavLink } from '../types'
import { HeaderLogo } from './header-logo'
import { PublicUserMenu } from './public-user-menu'

const AUTH_PROMPT_SECONDS = 5

type AuthPromptTarget = {
  title: string
  href: string
}

export interface PublicHeaderProps {
  navLinks?: TopNavLink[]
  mobileLinks?: TopNavLink[]
  navContent?: React.ReactNode
  showThemeSwitch?: boolean
  showLanguageSwitcher?: boolean
  logo?: React.ReactNode
  siteName?: string
  homeUrl?: string
  leftContent?: React.ReactNode
  rightContent?: React.ReactNode
  showNavigation?: boolean
  showAuthButtons?: boolean
  showNotifications?: boolean
  className?: string
}

export function PublicHeader(props: PublicHeaderProps) {
  const {
    navLinks = defaultTopNavLinks,
    showThemeSwitch = true,
    showLanguageSwitcher = true,
    logo: customLogo,
    siteName: customSiteName,
    homeUrl = '/',
    showAuthButtons = true,
    showNotifications = true,
  } = props

  const { t } = useTranslation()
  const navigate = useNavigate()
  const [scrolled, setScrolled] = useState(false)
  const mobileNavigationId = useId()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [authPromptTarget, setAuthPromptTarget] =
    useState<AuthPromptTarget | null>(null)
  const [authPromptSecondsLeft, setAuthPromptSecondsLeft] =
    useState(AUTH_PROMPT_SECONDS)

  const user = useAuthStore((s) => s.auth.user)
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

  const isAuthenticated = Boolean(user)
  const displaySiteName = customSiteName || systemName || DEFAULT_SYSTEM_NAME
  const links = dynamicLinks.length > 0 ? dynamicLinks : navLinks

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 16)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    if (!authPromptTarget) return

    const intervalId = window.setInterval(() => {
      setAuthPromptSecondsLeft((seconds) => Math.max(seconds - 1, 0))
    }, 1000)

    const timeoutId = window.setTimeout(() => {
      const redirect = authPromptTarget.href
      setAuthPromptTarget(null)
      navigate({ to: '/sign-in', search: { redirect } })
    }, AUTH_PROMPT_SECONDS * 1000)

    return () => {
      window.clearInterval(intervalId)
      window.clearTimeout(timeoutId)
    }
  }, [authPromptTarget, navigate])

  const closeAuthPrompt = useCallback(() => {
    setAuthPromptTarget(null)
    setAuthPromptSecondsLeft(AUTH_PROMPT_SECONDS)
  }, [])

  const navigateToSignIn = useCallback(() => {
    const redirect = authPromptTarget?.href || '/'
    setAuthPromptTarget(null)
    navigate({ to: '/sign-in', search: { redirect } })
  }, [authPromptTarget?.href, navigate])

  const handleNavLinkClick = useCallback(
    (
      event: React.MouseEvent<HTMLAnchorElement>,
      link: TopNavLink,
      closeMobile = false
    ) => {
      if (link.disabled) {
        event.preventDefault()
        return
      }

      if (link.requiresAuth) {
        event.preventDefault()
        if (closeMobile) {
          setMobileOpen(false)
        }
        setAuthPromptSecondsLeft(AUTH_PROMPT_SECONDS)
        setAuthPromptTarget({
          title: t(link.title),
          href: link.href,
        })
        return
      }

      if (closeMobile) {
        setMobileOpen(false)
      }
    },
    [t]
  )

  return (
    <>
      <header
        className={cn(
          'fixed inset-x-0 top-0 z-50 transition-all duration-300',
          scrolled ? 'py-2.5' : 'py-3.5',
          props.className
        )}
      >
        <div className='mx-auto max-w-[75rem] px-4 sm:px-6 lg:px-10'>
          <div
            className={cn(
              'ob-header-bar flex h-14 items-center justify-between rounded-full border border-black/[0.06] px-4 backdrop-blur-xl transition-all duration-300 sm:px-5 dark:border-white/[0.08]',
              scrolled
                ? 'shadow-[0_8px_32px_-12px_rgba(20,17,15,0.12)] dark:shadow-[0_8px_32px_-12px_rgba(0,0,0,0.45)]'
                : 'shadow-[0_4px_24px_rgba(20,17,15,0.04)]'
            )}
          >
            {/* Brand Logo */}
            <Link
              to={homeUrl}
              className='group flex shrink-0 items-center gap-2.5 focus-visible:outline-none'
            >
              <div className='ob-chip-primary flex size-8 shrink-0 items-center justify-center rounded-full transition-transform duration-200 group-hover:scale-105'>
                {loading && <Skeleton className='size-5 rounded-full' />}
                {!loading && customLogo && customLogo}
                {!loading && !customLogo && (
                  <HeaderLogo
                    src={systemLogo}
                    loading={loading}
                    logoLoaded={logoLoaded}
                    className='size-5 rounded-md object-contain'
                  />
                )}
              </div>
              <div className='flex items-center gap-2'>
                <span className='text-foreground text-base font-bold tracking-tight'>
                  {loading ? (
                    <Skeleton className='h-4 w-24' />
                  ) : (
                    displaySiteName
                  )}
                </span>
                <span className='ob-chip-primary ob-mono hidden rounded-full px-2 py-0.5 text-[10px] font-semibold tracking-wider uppercase 2xl:inline-block'>
                  API Gateway
                </span>
              </div>
            </Link>

            {/* Desktop Navigation Links */}
            <nav
              aria-label={t('landing.nav.ariaLabel')}
              className='hidden items-center gap-0.5 md:flex'
            >
              {links.map((link) => {
                const isActive =
                  link.href === '/'
                    ? pathname === '/'
                    : pathname === link.href ||
                      pathname.startsWith(`${link.href}/`)
                const badgeNode = link.badge ? (
                  <span
                    aria-hidden='true'
                    className='inline-flex items-center rounded-full bg-gradient-to-r from-[#FF6B00] to-[#FF2E00] px-1.5 py-0.5 text-[9px] leading-none font-extrabold tracking-wider text-white uppercase shadow-[0_2px_6px_rgba(255,77,0,0.35)]'
                  >
                    {link.badge}
                  </span>
                ) : null
                if (link.external) {
                  return (
                    <a
                      key={link.href}
                      href={link.href}
                      target='_blank'
                      rel='noopener noreferrer'
                      aria-disabled={link.disabled}
                      tabIndex={link.disabled ? -1 : undefined}
                      onClick={(event) => handleNavLinkClick(event, link)}
                      className={cn(
                        'text-muted-foreground hover:text-primary inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-sm font-medium transition-colors xl:px-3.5',
                        link.disabled && 'pointer-events-none opacity-50'
                      )}
                    >
                      <span>{t(link.title)}</span>
                      {badgeNode}
                    </a>
                  )
                }
                return (
                  <Link
                    key={link.href}
                    to={link.href}
                    disabled={link.disabled}
                    onClick={(event) => handleNavLinkClick(event, link)}
                    className={cn(
                      'inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-sm font-medium transition-all duration-200 xl:px-3.5',
                      isActive
                        ? 'text-primary bg-[var(--ob-primary-soft)] font-semibold shadow-xs'
                        : 'text-muted-foreground hover:text-primary',
                      link.disabled && 'pointer-events-none opacity-50'
                    )}
                  >
                    <span>{t(link.title)}</span>
                    {badgeNode}
                  </Link>
                )
              })}
            </nav>

            {/* Desktop Right Actions */}
            <div className='hidden items-center gap-1.5 md:flex'>
              {showLanguageSwitcher && <LanguageSwitcher />}
              {showThemeSwitch && <ThemeSwitch />}
              {showNotifications && (
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
              )}

              {showAuthButtons && (
                <>
                  <div className='bg-border mx-1 h-4 w-px' />
                  {loading && <Skeleton className='h-9 w-24 rounded-full' />}
                  {!loading && isAuthenticated && (
                    <div className='flex items-center gap-2'>
                      <Button
                        size='sm'
                        className='ob-btn-primary h-9 rounded-full px-4 text-xs font-semibold'
                        render={<Link to='/dashboard' />}
                      >
                        {t('Console')}
                        <ArrowUpRight className='ml-1 size-3.5' />
                      </Button>
                      <PublicUserMenu />
                    </div>
                  )}
                  {!loading && !isAuthenticated && (
                    <div className='flex items-center gap-2'>
                      <Button
                        variant='ghost'
                        size='sm'
                        className='text-foreground hover:text-primary h-9 rounded-full px-3.5 text-xs font-medium'
                        render={<Link to='/sign-in' />}
                      >
                        {t('Sign in')}
                      </Button>
                      <Button
                        size='sm'
                        className='ob-btn-primary h-9 rounded-full px-4 text-xs font-semibold'
                        render={<Link to='/sign-up' />}
                      >
                        {t('Console')}
                        <ArrowUpRight className='ml-1 size-3.5' />
                      </Button>
                    </div>
                  )}
                </>
              )}
            </div>

            {/* Mobile Controls */}
            <div className='flex items-center gap-1.5 md:hidden'>
              {showLanguageSwitcher && <LanguageSwitcher />}
              {showThemeSwitch && <ThemeSwitch />}
              {showAuthButtons && !loading && isAuthenticated && (
                <PublicUserMenu />
              )}
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='size-9 rounded-full'
                onClick={() => setMobileOpen((prev) => !prev)}
                aria-expanded={mobileOpen}
                aria-controls={mobileNavigationId}
                aria-label={t('Toggle navigation menu')}
              >
                {mobileOpen ? (
                  <X className='size-4' />
                ) : (
                  <Menu className='size-4' />
                )}
              </Button>
            </div>
          </div>

          {/* Mobile Dropdown Panel */}
          {mobileOpen && (
            <div
              id={mobileNavigationId}
              className='ob-card mt-2 rounded-2xl p-4 shadow-xl backdrop-blur-xl md:hidden'
            >
              <nav className='flex flex-col gap-1'>
                {links.map((link) => {
                  const isActive =
                    link.href === '/'
                      ? pathname === '/'
                      : pathname === link.href ||
                        pathname.startsWith(`${link.href}/`)
                  const mobileBadgeNode = link.badge ? (
                    <span
                      aria-hidden='true'
                      className='inline-flex items-center rounded-full bg-gradient-to-r from-[#FF6B00] to-[#FF2E00] px-1.5 py-0.5 text-[9px] leading-none font-extrabold tracking-wider text-white uppercase shadow-[0_2px_6px_rgba(255,77,0,0.35)]'
                    >
                      {link.badge}
                    </span>
                  ) : null
                  if (link.external) {
                    return (
                      <a
                        key={link.href}
                        href={link.href}
                        target='_blank'
                        rel='noopener noreferrer'
                        aria-disabled={link.disabled}
                        tabIndex={link.disabled ? -1 : undefined}
                        onClick={(event) =>
                          handleNavLinkClick(event, link, true)
                        }
                        className={cn(
                          'text-muted-foreground hover:text-primary flex items-center gap-2 rounded-xl px-3 py-2.5 text-sm font-medium',
                          link.disabled && 'pointer-events-none opacity-50'
                        )}
                      >
                        <span>{t(link.title)}</span>
                        {mobileBadgeNode}
                      </a>
                    )
                  }
                  return (
                    <Link
                      key={link.href}
                      to={link.href}
                      disabled={link.disabled}
                      onClick={(event) => handleNavLinkClick(event, link, true)}
                      className={cn(
                        'flex items-center gap-2 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors',
                        isActive
                          ? 'text-primary bg-[var(--ob-primary-soft)] font-semibold'
                          : 'text-muted-foreground hover:text-primary',
                        link.disabled && 'pointer-events-none opacity-50'
                      )}
                    >
                      <span>{t(link.title)}</span>
                      {mobileBadgeNode}
                    </Link>
                  )
                })}
              </nav>
              {showAuthButtons && (
                <div className='border-border mt-3 flex flex-col gap-2 border-t pt-3'>
                  {isAuthenticated ? (
                    <Button
                      className='ob-btn-primary w-full rounded-full'
                      render={<Link to='/dashboard' />}
                      onClick={() => setMobileOpen(false)}
                    >
                      {t('Console')}
                      <ArrowUpRight className='ml-1.5 size-4' />
                    </Button>
                  ) : (
                    <>
                      <Button
                        variant='outline'
                        className='ob-btn-secondary w-full rounded-full'
                        render={<Link to='/sign-in' />}
                        onClick={() => setMobileOpen(false)}
                      >
                        {t('Sign in')}
                      </Button>
                      <Button
                        className='ob-btn-primary w-full rounded-full'
                        render={<Link to='/sign-up' />}
                        onClick={() => setMobileOpen(false)}
                      >
                        {t('Console')}
                        <ArrowUpRight className='ml-1.5 size-4' />
                      </Button>
                    </>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </header>

      <Dialog
        open={Boolean(authPromptTarget)}
        onOpenChange={(open) => {
          if (!open) {
            closeAuthPrompt()
          }
        }}
        title={t('Sign in required')}
        description={t('Please sign in to view {{module}}.', {
          module: authPromptTarget?.title || '',
        })}
        contentClassName='sm:max-w-md'
        contentHeight='auto'
        footer={
          <>
            <Button variant='outline' onClick={closeAuthPrompt}>
              {t('Cancel')}
            </Button>
            <Button onClick={navigateToSignIn}>{t('Sign in now')}</Button>
          </>
        }
      >
        <div className='bg-muted/40 text-muted-foreground rounded-lg px-3 py-2 text-sm'>
          {t('Redirecting to sign in in {{seconds}} seconds.', {
            seconds: authPromptSecondsLeft,
          })}
        </div>
      </Dialog>
    </>
  )
}
