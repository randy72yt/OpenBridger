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
import { BookOpen, Code2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Card, CardContent, CardHeader } from '@/components/ui/card'

export function Features() {
  const { t } = useTranslation()
  return (
    <section className='relative z-10 px-6 py-24 md:py-32'>
      <div className='mx-auto max-w-6xl'>
        <h2 className='mb-4 text-2xl font-semibold tracking-tight md:text-3xl'>
          {t('From your first experiment to your next application')}
        </h2>
        <p className='text-muted-foreground mb-10 max-w-2xl leading-relaxed'>
          {t(
            'Learn by building. Keep the same workflow as your projects grow.'
          )}
        </p>
        <div className='grid gap-6 md:grid-cols-2'>
          <Card className='rounded-2xl shadow-none'>
            <CardHeader className='gap-4 p-6 pb-0'>
              <BookOpen className='text-primary size-7' aria-hidden />
              <h3 className='text-xl font-semibold'>
                {t('Learning and course projects')}
              </h3>
            </CardHeader>
            <CardContent className='space-y-5 p-6 pt-2'>
              <p className='text-muted-foreground leading-7'>
                {t(
                  'Practice programming, explore model capabilities and turn course ideas into working prototypes.'
                )}
              </p>
              <ul className='text-muted-foreground list-inside list-disc space-y-3 text-sm'>
                <li>{t('Code explanations and debugging practice')}</li>
                <li>{t('Course prototypes and personal projects')}</li>
                <li>{t('Compare models and track your usage')}</li>
              </ul>
            </CardContent>
          </Card>
          <Card className='bg-primary/5 ring-primary/15 rounded-2xl shadow-none'>
            <CardHeader className='gap-4 p-6 pb-0'>
              <Code2 className='text-primary size-7' aria-hidden />
              <h3 className='text-xl font-semibold'>
                {t('Development and application integration')}
              </h3>
            </CardHeader>
            <CardContent className='space-y-5 p-6 pt-2'>
              <p className='text-muted-foreground leading-7'>
                {t(
                  'Connect coding agents, SDKs and applications through a familiar API workflow.'
                )}
              </p>
              <ul className='text-muted-foreground list-inside list-disc space-y-3 text-sm'>
                <li>{t('Agent tools and SDK integration')}</li>
                <li>{t('API key management and request logs')}</li>
                <li>{t('Iterate from prototype to application')}</li>
              </ul>
            </CardContent>
          </Card>
        </div>
      </div>
    </section>
  )
}
