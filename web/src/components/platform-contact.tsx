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

import { CopyButton } from '@/components/copy-button'
import { useStatus } from '@/hooks/use-status'
import { resolvePlatformContact } from '@/lib/platform-contact'

export function PlatformContact() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const contact = resolvePlatformContact(status)

  return (
    <>
      <span className='flex flex-wrap items-center gap-1.5 break-words'>
        <span>{t('Contact email')}:</span>
        {contact.email ? (
          <>
            <a
              className='text-foreground hover:text-primary font-medium underline-offset-4 hover:underline'
              href={`mailto:${contact.email}`}
            >
              {contact.email}
            </a>
            <CopyButton value={contact.email} />
          </>
        ) : (
          <span>{t('To be announced')}</span>
        )}
      </span>
      <span className='flex flex-wrap items-center gap-1.5 break-words'>
        <span>{t('WeChat')}:</span>
        {contact.wechatId ? (
          <>
            <span className='text-foreground font-medium'>
              {contact.wechatId}
            </span>
            <CopyButton value={contact.wechatId} />
          </>
        ) : (
          <span>{t('To be announced')}</span>
        )}
      </span>
    </>
  )
}
