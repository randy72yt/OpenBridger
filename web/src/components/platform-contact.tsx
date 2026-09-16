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

import { PLATFORM_CONTACT } from '@/lib/platform-contact'

export function PlatformContact() {
  const { t } = useTranslation()
  return (
    <>
      <span className='break-words'>
        {t('Contact email')}:{' '}
        {PLATFORM_CONTACT.email ? (
          <a
            className='hover:text-foreground underline-offset-4 hover:underline'
            href={`mailto:${PLATFORM_CONTACT.email}`}
          >
            {PLATFORM_CONTACT.email}
          </a>
        ) : (
          t('To be announced')
        )}
      </span>
      <span className='break-words'>
        {t('Community')}:{' '}
        {PLATFORM_CONTACT.communityUrl ? (
          <a
            className='hover:text-foreground underline-offset-4 hover:underline'
            href={PLATFORM_CONTACT.communityUrl}
            target='_blank'
            rel='noopener noreferrer'
          >
            Telegram
          </a>
        ) : (
          t('To be announced')
        )}
      </span>
    </>
  )
}
