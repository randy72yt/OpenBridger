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
import { lazy, Suspense } from 'react'

import { useTheme } from '@/context/theme-provider'
import { useMediaQuery } from '@/hooks/use-media-query'
import { cn } from '@/lib/utils'

const LazyHeroScene3D = lazy(() => import('./hero-scene-3d'))

interface HeroBackgroundProps {
  className?: string
}

function CssGradientFallback(props: {
  className?: string
  reducedMotion?: boolean
}) {
  return (
    <div
      data-testid='landing-hero-css-fallback'
      aria-hidden='true'
      className={cn(
        'pointer-events-none absolute inset-0 overflow-hidden',
        props.className
      )}
    >
      {/* Primary teal ambient glow */}
      <div
        className={cn(
          'absolute -top-28 left-1/2 h-[420px] w-[720px] -translate-x-1/2 rounded-full opacity-35 blur-3xl dark:opacity-25',
          !props.reducedMotion && 'animate-pulse'
        )}
        style={{
          background:
            'radial-gradient(circle, color-mix(in srgb, var(--primary) 65%, transparent) 0%, transparent 72%)',
        }}
      />
      {/* Secondary offset orb */}
      <div
        className='absolute top-1/3 right-[-10%] h-[320px] w-[320px] rounded-full opacity-25 blur-3xl dark:opacity-20'
        style={{
          background:
            'radial-gradient(circle, color-mix(in srgb, var(--primary) 45%, #38bdf8) 0%, transparent 70%)',
        }}
      />
      {/* Perspective grid lines */}
      <div className='absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_75%_55%_at_50%_35%,black_20%,transparent_100%)] bg-[size:3.5rem_3.5rem] opacity-[0.18] dark:opacity-[0.12]' />
    </div>
  )
}

export function HeroBackground(props: HeroBackgroundProps) {
  const isDesktop = useMediaQuery('(min-width: 768px)')
  const prefersReducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const { resolvedTheme } = useTheme()
  const isDark = resolvedTheme === 'dark'

  const shouldRender3D = isDesktop && !prefersReducedMotion

  return (
    <div
      aria-hidden='true'
      className={cn(
        'pointer-events-none absolute inset-0 -z-10 overflow-hidden',
        props.className
      )}
    >
      <CssGradientFallback reducedMotion={prefersReducedMotion} />
      {shouldRender3D && (
        <Suspense fallback={null}>
          <LazyHeroScene3D isDark={isDark} />
        </Suspense>
      )}
      {/* Bottom vignette fade into page background */}
      <div className='from-background via-background/40 pointer-events-none absolute inset-x-0 bottom-0 h-32 bg-gradient-to-t to-transparent' />
    </div>
  )
}
