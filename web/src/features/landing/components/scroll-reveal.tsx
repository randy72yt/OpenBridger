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
import { motion } from 'motion/react'
import { useEffect, useRef, useState, type ReactNode } from 'react'

import { useMediaQuery } from '@/hooks/use-media-query'
import { cn } from '@/lib/utils'

interface ScrollRevealProps {
  children: ReactNode
  className?: string
  delay?: number
  yOffset?: number
  as?: 'div' | 'section'
}

export function ScrollReveal(props: ScrollRevealProps) {
  const delay = props.delay ?? 0
  const yOffset = props.yOffset ?? 28
  const Tag = props.as ?? 'div'
  const ref = useRef<HTMLDivElement>(null)
  const prefersReducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const [inView, setInView] = useState(false)

  useEffect(() => {
    if (prefersReducedMotion) {
      setInView(true)
      return
    }

    const element = ref.current
    if (!element) return

    const rect = element.getBoundingClientRect()
    if (rect.top <= window.innerHeight && rect.bottom >= 0) {
      setInView(true)
      return
    }

    if (typeof IntersectionObserver === 'undefined') {
      setInView(true)
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) {
          setInView(true)
          observer.unobserve(element)
        }
      },
      { threshold: 0.08, rootMargin: '0px 0px -24px 0px' }
    )

    observer.observe(element)
    return () => observer.disconnect()
  }, [prefersReducedMotion])

  const MotionTag = Tag === 'section' ? motion.section : motion.div

  if (prefersReducedMotion) {
    return (
      <Tag ref={ref as never} className={props.className}>
        {props.children}
      </Tag>
    )
  }

  return (
    <MotionTag
      ref={ref as never}
      initial={{ opacity: 0, y: yOffset }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: yOffset }}
      transition={{
        duration: 0.55,
        delay,
        ease: [0.16, 1, 0.3, 1],
      }}
      className={cn('will-change-[transform,opacity]', props.className)}
    >
      {props.children}
    </MotionTag>
  )
}
