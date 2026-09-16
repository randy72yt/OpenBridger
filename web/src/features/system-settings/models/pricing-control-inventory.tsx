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

import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { SettingsCard } from '../components/settings-card'
import type {
  ModelPricePolicy,
  PricingRisk,
  UpstreamModelOffer,
} from './pricing-control-types'

type PricingControlInventoryProps = {
  offers: UpstreamModelOffer[]
  policies: ModelPricePolicy[]
  risks: PricingRisk[]
}

export function PricingControlInventory(props: PricingControlInventoryProps) {
  const { t } = useTranslation()

  return (
    <SettingsCard title={t('Pricing inventory')}>
      <div className='space-y-6'>
        {props.risks.length > 0 && (
          <div className='border-warning/30 bg-warning/10 rounded-lg border p-4'>
            <p className='font-medium'>{t('Action required')}</p>
            <ul className='text-muted-foreground mt-2 space-y-1 text-sm'>
              {props.risks.map((risk) => (
                <li
                  key={`${risk.public_model}-${risk.service_tier}-${risk.code}`}
                >
                  {risk.public_model} · {risk.service_tier}: {risk.code}
                </li>
              ))}
            </ul>
          </div>
        )}
        <div className='overflow-x-auto'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Model')}</TableHead>
                <TableHead>{t('Channel')}</TableHead>
                <TableHead>{t('Input cost')}</TableHead>
                <TableHead>{t('Output cost')}</TableHead>
                <TableHead>{t('Source')}</TableHead>
                <TableHead>{t('Expires')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {props.offers.map((offer) => (
                <TableRow key={offer.id}>
                  <TableCell className='font-medium'>
                    {offer.public_model}
                  </TableCell>
                  <TableCell>{offer.channel_id}</TableCell>
                  <TableCell className='tabular-nums'>
                    ${offer.input_cost}
                  </TableCell>
                  <TableCell className='tabular-nums'>
                    ${offer.output_cost}
                  </TableCell>
                  <TableCell>
                    <Badge variant='outline'>{offer.source_type}</Badge>
                  </TableCell>
                  <TableCell>
                    {new Date(offer.expires_at * 1000).toLocaleString()}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        <p className='text-muted-foreground text-sm'>
          {t('{{count}} pricing policies configured.', {
            count: props.policies.length,
          })}
        </p>
      </div>
    </SettingsCard>
  )
}
