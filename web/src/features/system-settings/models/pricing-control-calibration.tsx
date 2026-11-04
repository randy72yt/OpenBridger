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
import { ChevronLeft, ChevronRight, Gauge, Pin, Search } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
  PricingCalibration,
  PricingCalibrationSummary,
} from './pricing-control-types'

type PricingControlCalibrationProps = {
  items: PricingCalibration[]
  summary?: PricingCalibrationSummary
  total: number
  limit: number
  offset: number
  loading: boolean
  syncing: boolean
  keyword: string
  onKeywordChange: (value: string) => void
  onPageChange: (offset: number) => void
  onRecalibrate: () => void
  onPin: (row: PricingCalibration) => void
}

function formatRatio(value: number) {
  if (!value || !Number.isFinite(value)) return '—'
  if (value >= 1) return value.toFixed(4)
  return value.toFixed(6)
}

// deviation > 1 means the published group ratio prices the cost higher than the
// invoice, which directly inflates the sale price.
function formatDeviation(calibration: PricingCalibration) {
  if (!calibration.policy_ratio || !calibration.effective_ratio) return '—'
  return `${(calibration.effective_ratio / calibration.policy_ratio).toFixed(2)}×`
}

function isStale(calibration: PricingCalibration, staleHours: number) {
  if (!calibration.window_end || !staleHours) return false
  return Date.now() / 1000 - calibration.window_end > staleHours * 3600
}

export function PricingControlCalibration(
  props: PricingControlCalibrationProps
) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(true)
  const summary = props.summary
  const from = props.total === 0 ? 0 : props.offset + 1
  const to = Math.min(props.offset + props.limit, props.total)
  const canPrev = props.offset > 0
  const canNext = props.offset + props.limit < props.total
  const coverage =
    summary && summary.models > 0
      ? Math.round((summary.usable / summary.models) * 100)
      : 0

  const statusBadge = (row: PricingCalibration) => {
    if (isStale(row, summary?.stale_hours ?? 0)) {
      return <Badge variant='secondary'>{t('Calibration stale')}</Badge>
    }
    if (row.applied) {
      return <Badge>{t('Applied')}</Badge>
    }
    return <Badge variant='destructive'>{t('Not applied')}</Badge>
  }

  return (
    <SettingsCard
      title={t('Cost calibration')}
      description={t(
        'Measured against the upstream bill: our base unit price × this factor reproduces what the provider actually charged for the same tokens. It replaces the group-ratio guess wherever it is available.'
      )}
    >
      <div className='space-y-6'>
        {summary && (
          <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
            <div className='bg-muted/40 rounded-lg border p-3'>
              <p className='text-muted-foreground text-xs'>
                {t('Calibrated models')}
              </p>
              <p className='mt-1 text-xl font-semibold tabular-nums'>
                {summary.usable}
                <span className='text-muted-foreground text-sm font-normal'>
                  {' '}
                  / {summary.models}
                </span>
              </p>
              <p className='text-muted-foreground mt-1 text-xs'>
                {t('Coverage {{percent}}%', { percent: coverage })}
              </p>
            </div>
            <div className='bg-muted/40 rounded-lg border p-3'>
              <p className='text-muted-foreground text-xs'>
                {t('Policy overstatement')}
              </p>
              <p className='mt-1 text-xl font-semibold tabular-nums'>
                {summary.policy_overstatement
                  ? `${summary.policy_overstatement}×`
                  : '—'}
              </p>
              <p className='text-muted-foreground mt-1 text-xs'>
                {t('Published group ratio vs measured cost')}
              </p>
            </div>
            <div className='bg-muted/40 rounded-lg border p-3'>
              <p className='text-muted-foreground text-xs'>
                {t('Last calibrated')}
              </p>
              <p className='mt-1 text-sm font-medium'>
                {summary.last_calibrated_at
                  ? new Date(
                      summary.last_calibrated_at * 1000
                    ).toLocaleString()
                  : t('Never')}
              </p>
              <p className='text-muted-foreground mt-1 text-xs'>
                {t('Stale after {{hours}}h', { hours: summary.stale_hours })}
              </p>
            </div>
            <div className='bg-muted/40 rounded-lg border p-3'>
              <p className='text-muted-foreground text-xs'>
                {t('Observed settlement groups')}
              </p>
              <p className='mt-1 text-sm font-medium'>
                {summary.observed_groups.length > 0
                  ? summary.observed_groups
                      .slice(0, 3)
                      .map((group) => group.group)
                      .join(', ')
                  : '—'}
              </p>
              <p className='text-muted-foreground mt-1 text-xs'>
                {t('From real traffic, not the price list')}
              </p>
            </div>
          </div>
        )}

        <div className='flex flex-wrap items-center gap-3'>
          <div className='relative min-w-56 flex-1'>
            <Search className='text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2' />
            <Input
              value={props.keyword}
              onChange={(event) => props.onKeywordChange(event.target.value)}
              placeholder={t('Search model or group')}
              className='pl-9'
            />
          </div>
          <Button size='sm' variant='outline' onClick={props.onRecalibrate}>
            <Gauge />
            {props.syncing ? t('Calibrating...') : t('Recalibrate now')}
          </Button>
          <Button
            size='sm'
            variant='ghost'
            onClick={() => setExpanded((value) => !value)}
            aria-expanded={expanded}
          >
            {expanded ? t('Hide details') : t('Show details')}
          </Button>
        </div>

        {expanded && (
          <div className='overflow-x-auto'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Model')}</TableHead>
                  <TableHead>{t('Channel')}</TableHead>
                  <TableHead>{t('Measured ratio')}</TableHead>
                  <TableHead>{t('Policy ratio')}</TableHead>
                  <TableHead>{t('Ratio gap')}</TableHead>
                  <TableHead>{t('Settlement group')}</TableHead>
                  <TableHead>{t('Samples')}</TableHead>
                  <TableHead>{t('Dispersion')}</TableHead>
                  <TableHead>{t('Status')}</TableHead>
                  <TableHead>{t('Actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {props.items.length === 0 && (
                  <TableRow>
                    <TableCell
                      colSpan={10}
                      className='text-muted-foreground py-10 text-center'
                    >
                      {props.loading
                        ? t('Loading...')
                        : t(
                            'No calibration yet. Run a calibration to read the upstream consumption log.'
                          )}
                    </TableCell>
                  </TableRow>
                )}
                {props.items.map((row) => (
                  <TableRow key={`${row.channel_id}-${row.public_model}`}>
                    <TableCell>
                      <div className='font-medium'>{row.public_model}</div>
                      {row.upstream_model &&
                        row.upstream_model !== row.public_model && (
                          <div className='text-muted-foreground text-xs'>
                            {row.upstream_model}
                          </div>
                        )}
                    </TableCell>
                    <TableCell className='tabular-nums'>
                      #{row.channel_id}
                    </TableCell>
                    <TableCell className='font-medium tabular-nums'>
                      {formatRatio(row.effective_ratio)}
                    </TableCell>
                    <TableCell className='text-muted-foreground tabular-nums'>
                      {formatRatio(row.policy_ratio)}
                    </TableCell>
                    <TableCell className='tabular-nums'>
                      {formatDeviation(row)}
                    </TableCell>
                    <TableCell>
                      {row.observed_group ? (
                        <Badge variant='outline'>{row.observed_group}</Badge>
                      ) : (
                        '—'
                      )}
                    </TableCell>
                    <TableCell className='tabular-nums'>
                      {row.samples > 0 ? row.samples : '—'}
                    </TableCell>
                    <TableCell className='tabular-nums'>
                      {row.samples > 0
                        ? `${row.dispersion_pct.toFixed(1)}%`
                        : '—'}
                    </TableCell>
                    <TableCell>
                      <div className='flex flex-col items-start gap-1'>
                        {statusBadge(row)}
                        {row.reason && (
                          <span className='text-muted-foreground text-xs'>
                            {row.reason}
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Button
                        size='sm'
                        variant='ghost'
                        className='h-7 px-2 text-xs'
                        onClick={() => props.onPin(row)}
                        title={t('Pin a measured ratio')}
                      >
                        <Pin />
                        {row.source === 'manual' ? t('Change') : t('Pin')}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}

        {expanded && (
          <div className='flex items-center justify-between gap-2'>
            <p className='text-muted-foreground text-xs'>
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
                onClick={() => props.onPageChange(Math.max(0, props.offset - props.limit))}
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
        )}
      </div>
    </SettingsCard>
  )
}
