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
import { useNavigate, useRouter } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

export function NotFoundError() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { history } = useRouter()
  return (
    <div className='ob-landing-theme bg-background text-foreground h-svh'>
      <div className='m-auto flex h-full w-full flex-col items-center justify-center gap-2 px-4'>
        <h1 className='ob-serif-heading ob-hero-gradient-text text-[7rem] leading-tight font-bold'>
          404
        </h1>
        <span className='text-lg font-semibold'>
          {t('Oops! Page Not Found!')}
        </span>
        <p className='text-muted-foreground text-center text-sm sm:text-base'>
          {t("It seems like the page you're looking for")} <br />
          {t('does not exist or might have been removed.')}
        </p>
        <div className='mt-6 flex gap-3'>
          <Button
            variant='outline'
            className='ob-btn-secondary rounded-full px-5'
            onClick={() => history.go(-1)}
          >
            {t('Go Back')}
          </Button>
          <Button
            className='ob-btn-primary rounded-full px-6'
            onClick={() => navigate({ to: '/' })}
          >
            {t('Back to Home')}
          </Button>
        </div>
      </div>
    </div>
  )
}
