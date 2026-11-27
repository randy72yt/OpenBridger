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
import { QueryClientContext } from '@tanstack/react-query'
import { useContext } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import type { SystemStatus } from '@/features/auth/types'
import { useStatus } from '@/hooks/use-status'
import { resolvePlatformContact } from '@/lib/platform-contact'

function PlatformContactContent(props: { status: SystemStatus | null }) {
  const { t } = useTranslation()
  const contact = resolvePlatformContact(props.status)

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

function PlatformContactConnected() {
  const { status } = useStatus()
  return <PlatformContactContent status={status} />
}

export function PlatformContact() {
  const queryClient = useContext(QueryClientContext)
  if (!queryClient) {
    return <PlatformContactContent status={null} />
  }
  return <PlatformContactConnected />
}
