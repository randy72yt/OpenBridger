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

import { Button } from '@/components/ui/button'

export function MaintenanceError() {
  const { t } = useTranslation()
  return (
    <div className='ob-landing-theme bg-background text-foreground h-svh'>
      <div className='m-auto flex h-full w-full flex-col items-center justify-center gap-2 px-4'>
        <h1 className='ob-serif-heading ob-hero-gradient-text text-[7rem] leading-tight font-bold'>
          503
        </h1>
        <span className='text-lg font-semibold'>
          {t('Website is under maintenance!')}
        </span>
        <p className='text-muted-foreground text-center text-sm sm:text-base'>
          {t('The site is not available at the moment.')} <br />
          {t("We'll be back online shortly.")}
        </p>
        <div className='mt-6 flex gap-3'>
          <Button variant='outline' className='ob-btn-secondary rounded-full px-5'>
            {t('Learn more')}
          </Button>
        </div>
      </div>
    </div>
  )
}
