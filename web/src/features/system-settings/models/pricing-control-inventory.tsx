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
import { ChevronLeft, ChevronRight, RefreshCw, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { SettingsCard } from '../components/settings-card'
import type { PricingRisk, UpstreamModelOffer } from './pricing-control-types'

type PricingControlInventoryProps = {
  offers: UpstreamModelOffer[]
  removing: boolean
  onRemoveModel: (model: string, tier: string) => void
  total: number
  limit: number
  offset: number
  loading: boolean
  keyword: string
  sourceType: string
  includeExpired: boolean
  risks: PricingRisk[]
  onKeywordChange: (value: string) => void
  onSourceTypeChange: (value: string) => void
  onIncludeExpiredChange: (value: boolean) => void
  onPageChange: (offset: number) => void
  onResync: () => void
  onFocusModel: (model: string) => void
}

function formatCost(value: number | null | undefined) {
  if (value === null || value === undefined) return '—'
  return `$${value.toLocaleString(undefined, { maximumFractionDigits: 6 })}`
}

// actualCost is the number that must be compared with the upstream bill:
// base unit price × the settlement multiplier of the group it was collected for.
function actualCost(cost: number, ratio: number | null | undefined) {
  if (ratio === null || ratio === undefined) return null
  return cost * ratio
}

export function PricingControlInventory(props: PricingControlInventoryProps) {
  const { t } = useTranslation()
  const from = props.total === 0 ? 0 : props.offset + 1
  const to = Math.min(props.offset + props.limit, props.total)
  const canPrev = props.offset > 0
  const canNext = props.offset + props.limit < props.total

  const riskLabel = (risk: PricingRisk) => {
    if (risk.code.endsWith('_cost_invalid')) {
      return t('Cost cannot be computed for this offer')
    }
    return t('No valid upstream offer')
  }

  return (
    <SettingsCard title={t('Pricing inventory')}>
      <div className='space-y-6'>
        {props.risks.length > 0 && (
          <div
            id='pricing-risks'
            className='border-warning/30 bg-warning/10 scroll-mt-24 rounded-lg border p-4'
          >
            <div className='flex flex-wrap items-center justify-between gap-2'>
              <p className='font-medium'>{t('Action required')}</p>
              <Button size='sm' variant='outline' onClick={props.onResync}>
                <RefreshCw />
                {t('Resync offers')}
              </Button>
            </div>
            <ul className='text-muted-foreground mt-2 space-y-2 text-sm'>
              {props.risks.map((risk) => (
                <li
                  key={`${risk.public_model}-${risk.service_tier}-${risk.code}-${risk.channel_id ?? 0}`}
                  className='flex flex-wrap items-center gap-2'
                >
                  <span className='text-foreground font-medium'>
                    {risk.public_model}
                  </span>
                  <Badge variant='outline'>{risk.service_tier}</Badge>
                  <Badge variant='destructive'>{riskLabel(risk)}</Badge>
                  {risk.other_channels && risk.other_channels.length > 0 && (
                    <Badge variant='secondary'>
                      {t('Still offered by channel {{channels}}', {
                        channels: risk.other_channels
                          .map((id) => `#${id}`)
                          .join(', '),
                      })}
                    </Badge>
                  )}
                  <span className='text-xs'>
                    {t('Channel')} #{risk.channel_id ?? '—'}
                    {risk.channel_role
                      ? ` · ${t(
                          risk.channel_role === 'backup' ? 'Backup' : 'Primary'
                        )}`
                      : ''}
                    {risk.expires_at
                      ? ` · ${t('Expired at')} ${new Date(
                          risk.expires_at * 1000
                        ).toLocaleString()}`
                      : ''}
                  </span>
                  {risk.code.endsWith('_offer_missing_or_expired') &&
                    typeof risk.missing_checks === 'number' &&
                    risk.missing_checks > 0 && (
                      <span className='text-warning text-xs font-medium'>
                        {t(
                          'Verified missing {{checks}}/{{threshold}} sweeps — auto-disables after the last one',
                          {
                            checks: risk.missing_checks,
                            threshold: risk.missing_disable_after,
                          }
                        )}
                      </span>
                    )}
                  <span className='ml-auto flex items-center gap-1'>
                    <Button
                      size='sm'
                      variant='ghost'
                      className='h-6 px-2 text-xs'
                      onClick={() => props.onFocusModel(risk.public_model)}
                    >
                      {t('Show offer')}
                    </Button>
                    <Button
                      size='sm'
                      variant='ghost'
                      className='text-destructive h-6 px-2 text-xs'
                      disabled={props.removing}
                      onClick={() =>
                        props.onRemoveModel(risk.public_model, risk.service_tier)
                      }
                    >
                      {t('Remove model')}
                    </Button>
                  </span>
                </li>
              ))}
            </ul>
            <p className='text-muted-foreground mt-3 text-xs'>
              {t(
                'Each successful sync verifies missing models across every channel. After {{threshold}} verified sweeps with no offer anywhere, the model is taken out of repricing automatically.',
                {
                  threshold:
                    props.risks.find((r) => r.missing_disable_after)
                      ?.missing_disable_after ?? 3,
                }
              )}
            </p>
          </div>
        )}

        <div className='flex flex-wrap items-center gap-3'>
          <div className='relative min-w-56 flex-1'>
            <Search className='text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2' />
            <Input
              value={props.keyword}
              onChange={(event) => props.onKeywordChange(event.target.value)}
              placeholder={t('Search model')}
              className='pl-9'
            />
          </div>
          <Select
            value={props.sourceType || 'all'}
            onValueChange={(value) =>
              props.onSourceTypeChange(!value || value === 'all' ? '' : value)
            }
          >
            <SelectTrigger className='w-44'>
              <SelectValue placeholder={t('All sources')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='all'>{t('All sources')}</SelectItem>
              <SelectItem value='new-api-pricing'>
                {t('Upstream sync')}
              </SelectItem>
              <SelectItem value='manual'>{t('Manual import')}</SelectItem>
            </SelectContent>
          </Select>
          <div className='flex items-center gap-2'>
            <Switch
              id='pricing-include-expired'
              checked={props.includeExpired}
              onCheckedChange={props.onIncludeExpiredChange}
            />
            <label
              htmlFor='pricing-include-expired'
              className='text-muted-foreground text-sm'
            >
              {t('Include expired')}
            </label>
          </div>
        </div>

        <div className='overflow-x-auto'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Model')}</TableHead>
                <TableHead>{t('Channel')}</TableHead>
                <TableHead>{t('Base input cost')}</TableHead>
                <TableHead>{t('Base output cost')}</TableHead>
                <TableHead>{t('Group ratio')}</TableHead>
                <TableHead>{t('Actual input cost')}</TableHead>
                <TableHead>{t('Actual output cost')}</TableHead>
                <TableHead>{t('Source')}</TableHead>
                <TableHead>{t('Expires')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {props.offers.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={9}
                    className='text-muted-foreground py-10 text-center'
                  >
                    {props.loading
                      ? t('Loading...')
                      : t(
                          'No valid upstream offers. Run a sync to collect current prices.'
                        )}
                  </TableCell>
                </TableRow>
              )}
              {props.offers.map((offer) => {
                const expired = offer.expires_at * 1000 <= Date.now()
                return (
                  <TableRow
                    key={offer.id}
                    className={
                      expired ? 'text-muted-foreground opacity-60' : ''
                    }
                  >
                    <TableCell className='font-medium'>
                      {offer.public_model}
                      {offer.upstream_model !== offer.public_model && (
                        <span className='text-muted-foreground mt-1 block text-xs'>
                          {offer.upstream_model}
                        </span>
                      )}
                    </TableCell>
                    <TableCell>{offer.channel_id}</TableCell>
                    <TableCell className='tabular-nums'>
                      {formatCost(offer.input_cost)}
                    </TableCell>
                    <TableCell className='tabular-nums'>
                      {formatCost(offer.output_cost)}
                    </TableCell>
                    <TableCell className='tabular-nums'>
                      {offer.upstream_group_ratio == null
                        ? '—'
                        : `${offer.upstream_group_ratio}×`}
                      {offer.upstream_group && (
                        <span className='text-muted-foreground mt-1 block text-xs'>
                          {offer.upstream_group}
                        </span>
                      )}
                    </TableCell>
                    <TableCell className='tabular-nums'>
                      {formatCost(
                        actualCost(offer.input_cost, offer.upstream_group_ratio)
                      )}
                    </TableCell>
                    <TableCell className='tabular-nums'>
                      {formatCost(
                        actualCost(
                          offer.output_cost,
                          offer.upstream_group_ratio
                        )
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant='outline' title={offer.source_url}>
                        {offer.source_type}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className='flex items-center gap-2'>
                        {expired && (
                          <Badge variant='destructive'>{t('Expired')}</Badge>
                        )}
                        {new Date(offer.expires_at * 1000).toLocaleString()}
                      </span>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>

        <div className='flex flex-wrap items-center justify-between gap-3'>
          <p className='text-muted-foreground text-sm'>
            {t('Showing {{from}}-{{to}} of {{total}}', {
              from,
              to,
              total: props.total,
            })}
          </p>
          <div className='flex items-center gap-2'>
            <Button
              size='sm'
              variant='outline'
              disabled={!canPrev}
              onClick={() =>
                props.onPageChange(Math.max(0, props.offset - props.limit))
              }
            >
              <ChevronLeft />
              {t('Previous')}
            </Button>
            <Button
              size='sm'
              variant='outline'
              disabled={!canNext}
              onClick={() => props.onPageChange(props.offset + props.limit)}
            >
              {t('Next')}
              <ChevronRight />
            </Button>
          </div>
        </div>
      </div>
    </SettingsCard>
  )
}
