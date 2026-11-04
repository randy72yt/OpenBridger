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
import { ChevronDown, Pencil, Plus, Search } from 'lucide-react'
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

import { PricingControlPolicyForm } from './pricing-control-policy-form'
import type { ModelPricePolicy } from './pricing-control-types'

type PricingControlPolicyPanelProps = {
  policies: ModelPricePolicy[]
  saving: boolean
  onSave: (policy: ModelPricePolicy) => void
}

const GLOBAL_MODEL = '*'

function isGlobalPolicy(policy: ModelPricePolicy) {
  return policy.scope === 'global' || policy.public_model === GLOBAL_MODEL
}

export function PricingControlPolicyPanel(
  props: PricingControlPolicyPanelProps
) {
  const { t } = useTranslation()
  const globalPolicy = props.policies.find(isGlobalPolicy) ?? null
  const modelPolicies = props.policies.filter(
    (policy) => !isGlobalPolicy(policy)
  )
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [editing, setEditing] = useState<ModelPricePolicy | null>(null)
  const [creating, setCreating] = useState(false)

  const keyword = search.trim().toLowerCase()
  const visible = keyword
    ? modelPolicies.filter((policy) =>
        policy.public_model.toLowerCase().includes(keyword)
      )
    : modelPolicies

  return (
    <div className='space-y-6'>
      <div className='space-y-3'>
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <div>
            <p className='font-medium'>{t('Global policy')}</p>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Applies to every model without its own override. Model overrides only affect that model.'
              )}
            </p>
          </div>
          <Badge variant={globalPolicy?.enabled ? 'default' : 'outline'}>
            {globalPolicy?.enabled ? t('Active') : t('Not configured')}
          </Badge>
        </div>
        <PricingControlPolicyForm
          key={globalPolicy?.id ?? 'global-new'}
          scope='global'
          initial={globalPolicy}
          saving={props.saving}
          onSave={props.onSave}
        />
      </div>

      <div className='space-y-4'>
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <Button
            variant='ghost'
            className='px-2'
            onClick={() => setOpen((value) => !value)}
            aria-expanded={open}
          >
            <ChevronDown
              className={`transition-transform ${open ? 'rotate-180' : ''}`}
            />
            {t('Per-model overrides')}
            <Badge variant='outline'>{modelPolicies.length}</Badge>
          </Button>
          <Button
            size='sm'
            variant='outline'
            onClick={() => {
              setCreating(true)
              setEditing(null)
              setOpen(true)
            }}
          >
            <Plus />
            {t('Add model override')}
          </Button>
        </div>
        {open && (
          <div className='space-y-4'>
            {(creating || editing) && (
              <div className='bg-muted/40 rounded-lg border p-4'>
                <PricingControlPolicyForm
                  key={editing?.id ?? 'model-new'}
                  scope='model'
                  initial={editing}
                  saving={props.saving}
                  onSave={(policy) => {
                    props.onSave(policy)
                    setEditing(null)
                    setCreating(false)
                  }}
                  onCancel={() => {
                    setEditing(null)
                    setCreating(false)
                  }}
                />
              </div>
            )}
            <div className='relative max-w-sm'>
              <Search className='text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2' />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder={t('Search model')}
                className='pl-9'
              />
            </div>
            <div className='overflow-x-auto'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('Model')}</TableHead>
                    <TableHead>{t('Primary channel')}</TableHead>
                    <TableHead>{t('Backup channel')}</TableHead>
                    <TableHead>{t('Target margin (BPS)')}</TableHead>
                    <TableHead>{t('Status')}</TableHead>
                    <TableHead className='text-right'>{t('Actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {visible.length === 0 && (
                    <TableRow>
                      <TableCell
                        colSpan={6}
                        className='text-muted-foreground py-8 text-center'
                      >
                        {t('No model overrides yet.')}
                      </TableCell>
                    </TableRow>
                  )}
                  {visible.map((policy) => (
                    <TableRow
                      key={`${policy.public_model}-${policy.service_tier}`}
                    >
                      <TableCell className='font-medium'>
                        {policy.public_model}
                        <span className='text-muted-foreground ml-2 text-xs'>
                          {policy.service_tier}
                        </span>
                      </TableCell>
                      <TableCell>{policy.primary_channel_id}</TableCell>
                      <TableCell>
                        {policy.backup_channel_id > 0
                          ? policy.backup_channel_id
                          : '—'}
                      </TableCell>
                      <TableCell className='tabular-nums'>
                        {policy.target_margin_bps > 0
                          ? `${(policy.target_margin_bps / 100).toFixed(2)}%`
                          : t('Inherited')}
                      </TableCell>
                      <TableCell>
                        <Badge variant={policy.enabled ? 'default' : 'outline'}>
                          {policy.enabled ? t('Active') : t('Disabled')}
                        </Badge>
                      </TableCell>
                      <TableCell className='text-right'>
                        <Button
                          size='sm'
                          variant='ghost'
                          onClick={() => {
                            setCreating(false)
                            setEditing(policy)
                            setOpen(true)
                          }}
                        >
                          <Pencil />
                          {t('Edit')}
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
