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
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'
import { Card, CardContent } from '@/components/ui/card'

export function Tutorials() {
  const { t } = useTranslation()
  return (
    <PublicLayout showMainContainer={false}>
      <main className='mx-auto min-h-[70vh] max-w-4xl px-6 pt-32 pb-16'>
        <Card className='rounded-3xl shadow-none'>
          <CardContent className='space-y-5 p-8 md:p-12'>
            <p className='text-primary text-xs font-semibold tracking-widest'>
              OpenBridger
            </p>
            <h1 className='text-3xl font-semibold'>
              {t('Practical tutorials')}
            </h1>
            <p className='text-muted-foreground leading-8'>
              {t(
                'Tool configuration guides and SDK examples are being prepared. This page will host our practical tutorials.'
              )}
            </p>
          </CardContent>
        </Card>
      </main>
      <Footer />
    </PublicLayout>
  )
}
