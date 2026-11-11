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
import { ArrowRight, HelpCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

import { LANDING_FAQ_ITEMS } from '../../constants'
import { ScrollReveal } from '../scroll-reveal'

interface FaqSectionProps {
  isAuthenticated: boolean
}

export function FaqSection(props: FaqSectionProps) {
  const { t } = useTranslation()

  return (
    <section className='ob-section-warm relative z-10 py-20 sm:py-28'>
      <div className='mx-auto max-w-4xl px-4 sm:px-6 lg:px-8'>
        <ScrollReveal className='text-center'>
          <div className='ob-chip-primary mb-3 inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold'>
            <HelpCircle className='size-3.5' />
            <span>{t('landing.faq.eyebrow')}</span>
          </div>
          <h2 className='ob-serif-heading text-2xl font-bold tracking-tight sm:text-4xl'>
            {t('landing.faq.title')}
          </h2>
          <p className='text-muted-foreground mt-3 text-sm leading-relaxed sm:text-base'>
            {t('landing.faq.description')}
          </p>
        </ScrollReveal>

        <ScrollReveal delay={0.08} className='mt-10'>
          <Card className='ob-card rounded-2xl py-0 shadow-none'>
            <CardContent className='p-6 sm:p-8'>
              <Accordion defaultValue={['compatibility']}>
                {LANDING_FAQ_ITEMS.map((item) => (
                  <AccordionItem
                    key={item.id}
                    value={item.id}
                    className='border-border py-1'
                  >
                    <AccordionTrigger className='hover:text-primary py-4 text-base font-semibold hover:no-underline'>
                      {t(item.questionKey)}
                    </AccordionTrigger>
                    <AccordionContent className='text-muted-foreground pb-4 text-sm leading-relaxed'>
                      {t(item.answerKey)}
                    </AccordionContent>
                  </AccordionItem>
                ))}
              </Accordion>
            </CardContent>
          </Card>
        </ScrollReveal>

        {/* Bottom CTA Card */}
        <ScrollReveal delay={0.14} className='mt-14'>
          <Card className='ob-card relative overflow-hidden rounded-3xl py-0 shadow-none'>
            <CardContent className='flex flex-col items-center justify-between gap-6 p-8 text-center sm:flex-row sm:p-10 sm:text-left'>
              <div className='max-w-xl space-y-2'>
                <h3 className='ob-serif-heading text-xl font-bold tracking-tight sm:text-2xl'>
                  {t('landing.cta.title')}
                </h3>
                <p className='text-muted-foreground text-sm leading-relaxed'>
                  {t('landing.cta.description')}
                </p>
              </div>
              <div className='flex shrink-0 flex-wrap items-center gap-3'>
                <Button
                  size='lg'
                  className='ob-btn-primary h-11 rounded-xl px-6 text-sm font-semibold'
                  render={
                    <Link
                      to={props.isAuthenticated ? '/dashboard' : '/sign-up'}
                    />
                  }
                >
                  {props.isAuthenticated
                    ? t('Go to Dashboard')
                    : t('landing.hero.ctaPrimary')}
                  <ArrowRight className='ml-1.5 size-4' />
                </Button>
                <Button
                  variant='outline'
                  size='lg'
                  className='ob-btn-secondary h-11 rounded-xl px-5 text-sm font-medium'
                  render={<Link to='/pricing' />}
                >
                  {t('Model Square')}
                </Button>
              </div>
            </CardContent>
          </Card>
        </ScrollReveal>
      </div>
    </section>
  )
}
