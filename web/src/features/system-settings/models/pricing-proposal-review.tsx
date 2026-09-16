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
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { SettingsCard } from '../components/settings-card'
import type { ModelPriceProposal } from './pricing-control-types'

type PricingProposalReviewProps = {
  proposals: ModelPriceProposal[]
  recalculating: boolean
  updating: boolean
  onRecalculate: () => void
  onAction: (id: number, action: 'approve' | 'reject') => void
  onPublish: (proposal: ModelPriceProposal) => void
}

function formatPrice(value: number | null) {
  if (value === null) return '—'
  return `$${value.toLocaleString(undefined, { maximumFractionDigits: 8 })}`
}

function formatBPS(value: number | null) {
  if (value === null) return '—'
  const sign = value > 0 ? '+' : ''
  return `${sign}${(value / 100).toFixed(2)}%`
}

function changeVariant(value: number | null) {
  if (value === null || value === 0) return 'outline' as const
  return value > 0 ? ('warning' as const) : ('secondary' as const)
}

export function PricingProposalReview(props: PricingProposalReviewProps) {
  const { t } = useTranslation()

  return (
    <SettingsCard
      title={t('Price proposals')}
      description={t(
        'Compare live prices, expected costs, and fallback risk before approval.'
      )}
    >
      <div className='mb-4 flex flex-wrap items-center justify-between gap-3'>
        <p className='text-muted-foreground text-sm'>
          {t('Prices and costs are USD per million tokens.')}
        </p>
        <Button onClick={props.onRecalculate} disabled={props.recalculating}>
          {t('Recalculate now')}
        </Button>
      </div>
      <div className='overflow-x-auto'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Model and channels')}</TableHead>
              <TableHead>{t('Quote freshness')}</TableHead>
              <TableHead>{t('Current price')}</TableHead>
              <TableHead>{t('Proposed price')}</TableHead>
              <TableHead>{t('Expected cost')}</TableHead>
              <TableHead>{t('Expected / worst margin')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
              <TableHead className='text-right'>{t('Actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {props.proposals.length === 0 && (
              <TableRow>
                <TableCell
                  colSpan={8}
                  className='text-muted-foreground py-10 text-center'
                >
                  {t(
                    'No pricing proposals yet. Recalculate after importing upstream costs.'
                  )}
                </TableCell>
              </TableRow>
            )}
            {props.proposals.map((proposal) => {
              const expired = proposal.primary_expires_at * 1000 <= Date.now()
              return (
                <TableRow key={proposal.id}>
                  <TableCell className='min-w-48'>
                    <p className='font-medium'>{proposal.public_model}</p>
                    <p className='text-muted-foreground mt-1 text-xs'>
                      {t('Primary')}: {proposal.primary_channel_name || '—'}
                      {proposal.backup_channel_name
                        ? ` · ${t('Backup')}: ${proposal.backup_channel_name}`
                        : ''}
                    </p>
                  </TableCell>
                  <TableCell className='min-w-44 text-xs'>
                    <Badge variant={expired ? 'destructive' : 'outline'}>
                      {expired ? t('Expired') : t('Valid')}
                    </Badge>
                    <p className='mt-2'>
                      {proposal.primary_collected_at
                        ? new Date(
                            proposal.primary_collected_at * 1000
                          ).toLocaleString()
                        : '—'}
                    </p>
                    <p className='text-muted-foreground mt-1'>
                      {[proposal.source_type, proposal.source_version]
                        .filter(Boolean)
                        .join(' · ') || '—'}
                    </p>
                  </TableCell>
                  <TableCell className='min-w-36 tabular-nums'>
                    <p>
                      {t('Input')}: {formatPrice(proposal.current_input_price)}
                    </p>
                    <p className='mt-1'>
                      {t('Output')}:{' '}
                      {formatPrice(proposal.current_output_price)}
                    </p>
                  </TableCell>
                  <TableCell className='min-w-40 tabular-nums'>
                    <p>
                      {t('Input')}: {formatPrice(proposal.proposed_input_price)}{' '}
                      <Badge
                        variant={changeVariant(proposal.input_price_change_bps)}
                      >
                        {formatBPS(proposal.input_price_change_bps)}
                      </Badge>
                    </p>
                    <p className='mt-1'>
                      {t('Output')}:{' '}
                      {formatPrice(proposal.proposed_output_price)}{' '}
                      <Badge
                        variant={changeVariant(
                          proposal.output_price_change_bps
                        )}
                      >
                        {formatBPS(proposal.output_price_change_bps)}
                      </Badge>
                    </p>
                  </TableCell>
                  <TableCell className='min-w-36 tabular-nums'>
                    <p>
                      {t('Input')}: {formatPrice(proposal.expected_input_cost)}
                    </p>
                    <p className='mt-1'>
                      {t('Output')}:{' '}
                      {formatPrice(proposal.expected_output_cost)}
                    </p>
                  </TableCell>
                  <TableCell className='min-w-36 tabular-nums'>
                    <p>{formatBPS(proposal.expected_margin_bps)}</p>
                    <p
                      className={
                        proposal.worst_margin_bps < 0
                          ? 'text-destructive mt-1'
                          : 'text-muted-foreground mt-1'
                      }
                    >
                      {t('Worst')}: {formatBPS(proposal.worst_margin_bps)}
                    </p>
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={proposal.pricing_changed ? 'warning' : 'outline'}
                    >
                      {proposal.pricing_changed
                        ? t('Stale')
                        : t(proposal.status)}
                    </Badge>
                  </TableCell>
                  <TableCell className='min-w-44 space-x-2 text-right'>
                    {proposal.status === 'pending' && (
                      <Button
                        size='sm'
                        variant='outline'
                        disabled={
                          props.updating || proposal.pricing_changed || expired
                        }
                        onClick={() => props.onAction(proposal.id, 'approve')}
                      >
                        {t('Approve')}
                      </Button>
                    )}
                    {proposal.status === 'approved' &&
                      proposal.service_tier === 'default' && (
                        <Button
                          size='sm'
                          disabled={
                            props.updating ||
                            proposal.pricing_changed ||
                            expired
                          }
                          onClick={() => props.onPublish(proposal)}
                        >
                          {t('Publish')}
                        </Button>
                      )}
                    {(proposal.status === 'pending' ||
                      proposal.status === 'approved') && (
                      <Button
                        size='sm'
                        variant='ghost'
                        disabled={props.updating}
                        onClick={() => props.onAction(proposal.id, 'reject')}
                      >
                        {t('Reject')}
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </div>
    </SettingsCard>
  )
}
