/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the License,
or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Dialog } from '@/components/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { getChannelUpstreamCost, type UpstreamCostRow } from '../../api'

type UpstreamCostDialogProps = {
  open: boolean
  channelId: number
  channelName: string
  onOpenChange: (open: boolean) => void
}

function formatUSD(value: number | null | undefined) {
  if (value == null || Number.isNaN(value)) return '—'
  return `$${value.toLocaleString(undefined, { maximumFractionDigits: 6 })}`
}

function formatMargin(bps: number | null | undefined) {
  if (bps == null) return '—'
  return `${(bps / 100).toFixed(2)}%`
}

export function UpstreamCostDialog(props: UpstreamCostDialogProps) {
  const { t } = useTranslation()
  const [rows, setRows] = useState<UpstreamCostRow[]>([])
  const [keyGroup, setKeyGroup] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!props.open) return
    let cancelled = false
    setLoading(true)
    setRows([])
    setKeyGroup('')
    void getChannelUpstreamCost(props.channelId)
      .then((response) => {
        if (cancelled) return
        if (!response.success) {
          toast.error(
            response.message
              ? t(response.message)
              : t('Failed to load upstream cost')
          )
          return
        }
        setKeyGroup(response.data?.key_group || '')
        setRows(response.data?.models || [])
      })
      .catch(() => {
        if (!cancelled) {
          toast.error(t('Failed to load upstream cost'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [props.open, props.channelId, t])

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={t('Upstream cost vs sell price')}
      description={t(
        'Read-only. Upstream cost uses the effective group ratio and does not change sell prices. Channel: {{name}}',
        { name: props.channelName }
      )}
      contentClassName='sm:max-w-5xl'
      contentHeight='min(70vh, 36rem)'
    >
      <p className='text-muted-foreground mb-3 text-sm'>
        {t('Upstream key group: {{group}}', {
          group: keyGroup || t('auto'),
        })}
      </p>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Model')}</TableHead>
            <TableHead>{t('Effective group')}</TableHead>
            <TableHead>{t('Group ratio')}</TableHead>
            <TableHead>{t('Upstream input')}</TableHead>
            <TableHead>{t('Sell input')}</TableHead>
            <TableHead>{t('Input margin')}</TableHead>
            <TableHead>{t('Upstream output')}</TableHead>
            <TableHead>{t('Sell output')}</TableHead>
            <TableHead>{t('Output margin')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading && (
            <TableRow>
              <TableCell colSpan={9} className='text-muted-foreground py-8 text-center'>
                {t('Loading...')}
              </TableCell>
            </TableRow>
          )}
          {!loading && rows.length === 0 && (
            <TableRow>
              <TableCell colSpan={9} className='text-muted-foreground py-8 text-center'>
                {t('No models on this channel')}
              </TableCell>
            </TableRow>
          )}
          {rows.map((row) => (
            <TableRow key={row.model}>
              <TableCell className='font-medium'>
                {row.model}
                {row.note && (
                  <span className='text-muted-foreground mt-1 block text-xs'>
                    {t(row.note)}
                  </span>
                )}
              </TableCell>
              <TableCell>{row.effective_group || '—'}</TableCell>
              <TableCell className='tabular-nums'>
                {row.group_ratio ? `${row.group_ratio}×` : '—'}
              </TableCell>
              <TableCell className='tabular-nums'>
                {formatUSD(row.upstream_input_usd)}
              </TableCell>
              <TableCell className='tabular-nums'>
                {formatUSD(row.sell_input_usd)}
              </TableCell>
              <TableCell className='tabular-nums'>
                {formatMargin(row.input_margin_bps)}
              </TableCell>
              <TableCell className='tabular-nums'>
                {formatUSD(row.upstream_output_usd)}
              </TableCell>
              <TableCell className='tabular-nums'>
                {formatUSD(row.sell_output_usd)}
              </TableCell>
              <TableCell className='tabular-nums'>
                {formatMargin(row.output_margin_bps)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Dialog>
  )
}
