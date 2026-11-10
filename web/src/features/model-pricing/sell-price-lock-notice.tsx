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

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

export function SellPriceLockNotice() {
  const { t } = useTranslation()

  return (
    <Alert>
      <AlertTitle>
        {t('Sell prices are published by dynamic pricing')}
      </AlertTitle>
      <AlertDescription>
        {t(
          'Approve a pricing proposal, then publish it. This editor cannot save model sell prices while dynamic pricing control is on.'
        )}{' '}
        <Link
          to='/system-settings/models/$section'
          params={{ section: 'pricing-control' }}
          className='text-primary underline-offset-4 hover:underline'
        >
          {t('Open dynamic pricing')}
        </Link>
      </AlertDescription>
    </Alert>
  )
}
