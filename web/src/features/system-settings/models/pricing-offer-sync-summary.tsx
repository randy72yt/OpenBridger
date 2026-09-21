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

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

import type {
  PricingApiResponse,
  PricingOfferSyncResult,
} from './pricing-control-types'

type PricingOfferSyncSummaryProps = {
  response: PricingApiResponse<PricingOfferSyncResult>
}

export function PricingOfferSyncSummary(props: PricingOfferSyncSummaryProps) {
  const { t } = useTranslation()
  const result = props.response.data
  const hasIssues =
    !props.response.success ||
    result.channels.some((channel) => channel.error || channel.warnings?.length)

  return (
    <Alert variant={hasIssues ? 'destructive' : 'default'}>
      <AlertTitle>
        {hasIssues
          ? t('Upstream price sync needs review')
          : t('Upstream price sync completed')}
      </AlertTitle>
      <AlertDescription className='space-y-2 text-current'>
        <p>
          {t(
            'Imported {{imported}} offers; skipped {{skipped}} models; created {{proposals}} proposals.',
            {
              imported: result.imported,
              skipped: result.skipped,
              proposals: result.created_proposals,
            }
          )}
        </p>
        {!props.response.success && <p>{props.response.message}</p>}
        <ul className='max-h-48 space-y-2 overflow-y-auto'>
          {result.channels.map((channel) => (
            <li key={channel.channel_id}>
              <strong>{t('Channel {{id}}', { id: channel.channel_id })}</strong>
              {' · '}
              {t('{{count}} imported', { count: channel.imported })}
              {channel.error && <p>{channel.error}</p>}
              {channel.warnings?.map((warning) => (
                <p key={warning}>{warning}</p>
              ))}
            </li>
          ))}
        </ul>
      </AlertDescription>
    </Alert>
  )
}
