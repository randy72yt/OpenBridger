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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'

import type { ModelPricePolicy } from './pricing-control-types'

const globalDefaults: ModelPricePolicy = {
  public_model: '*',
  service_tier: 'default',
  scope: 'global',
  primary_channel_id: 0,
  backup_channel_id: 0,
  target_margin_bps: 3500,
  minimum_margin_bps: 2000,
  overhead_bps: 500,
  fallback_probability_bps: 1000,
  auto_publish_change_bps: 0,
  price_locked: false,
  enabled: true,
}

const modelDefaults: ModelPricePolicy = {
  public_model: '',
  service_tier: 'default',
  scope: 'model',
  primary_channel_id: 1,
  backup_channel_id: 0,
  target_margin_bps: 0,
  minimum_margin_bps: 0,
  overhead_bps: 0,
  fallback_probability_bps: 0,
  auto_publish_change_bps: 0,
  price_locked: false,
  enabled: true,
}

type PricingControlPolicyFormProps = {
  scope: 'global' | 'model'
  initial?: ModelPricePolicy | null
  saving: boolean
  onSave: (policy: ModelPricePolicy) => void
  onCancel?: () => void
}

export function PricingControlPolicyForm(props: PricingControlPolicyFormProps) {
  const { t } = useTranslation()
  const defaults = props.scope === 'global' ? globalDefaults : modelDefaults
  const [policy, setPolicy] = useState<ModelPricePolicy>({
    ...defaults,
    ...props.initial,
    scope: props.scope,
  })
  const setNumber = (key: keyof ModelPricePolicy, value: string) => {
    setPolicy((current) => ({ ...current, [key]: Number(value) }))
  }

  const knobs: [keyof ModelPricePolicy, string][] = [
    ['target_margin_bps', 'Target margin (BPS)'],
    ['minimum_margin_bps', 'Minimum margin (BPS)'],
    ['overhead_bps', 'Operating overhead (BPS)'],
    ['fallback_probability_bps', 'Fallback probability (BPS)'],
  ]

  return (
    <form
      className='space-y-4'
      onSubmit={(event) => {
        event.preventDefault()
        props.onSave({ ...policy, scope: props.scope })
      }}
    >
      {props.scope === 'model' && (
        <div className='grid gap-4 md:grid-cols-3'>
          <div className='space-y-2'>
            <Label htmlFor='pricing-public-model'>{t('Public model')}</Label>
            <Input
              id='pricing-public-model'
              value={policy.public_model}
              onChange={(event) =>
                setPolicy((current) => ({
                  ...current,
                  public_model: event.target.value,
                }))
              }
              placeholder='gpt-5.4'
              required
            />
          </div>
          <div className='space-y-2'>
            <Label htmlFor='pricing-primary-channel'>
              {t('Primary channel')}
            </Label>
            <Input
              id='pricing-primary-channel'
              type='number'
              min={1}
              value={policy.primary_channel_id}
              onChange={(event) =>
                setNumber('primary_channel_id', event.target.value)
              }
              required
            />
          </div>
          <div className='space-y-2'>
            <Label htmlFor='pricing-backup-channel'>
              {t('Backup channel')}
            </Label>
            <Input
              id='pricing-backup-channel'
              type='number'
              min={0}
              value={policy.backup_channel_id}
              onChange={(event) =>
                setNumber('backup_channel_id', event.target.value)
              }
            />
          </div>
        </div>
      )}

      <div className='grid gap-4 md:grid-cols-2 xl:grid-cols-4'>
        {knobs.map(([key, label]) => (
          <div className='space-y-2' key={key}>
            <Label htmlFor={`pricing-${String(key)}`}>{t(label)}</Label>
            <Input
              id={`pricing-${String(key)}`}
              type='number'
              min={0}
              max={10000}
              value={Number(policy[key] ?? 0)}
              onChange={(event) => setNumber(key, event.target.value)}
            />
          </div>
        ))}
      </div>

      <div className='space-y-2 md:max-w-md'>
        <Label htmlFor='pricing-auto-publish-change'>
          {t('Auto-publish change (BPS)')}
        </Label>
        <Input
          id='pricing-auto-publish-change'
          type='number'
          min={0}
          max={5000}
          value={policy.auto_publish_change_bps}
          onChange={(event) =>
            setNumber('auto_publish_change_bps', event.target.value)
          }
        />
        <p className='text-muted-foreground text-xs'>
          {t(
            '0 keeps manual approval. A positive value publishes increases within that many basis points.'
          )}
        </p>
      </div>

      {props.scope === 'model' && (
        <p className='text-muted-foreground text-xs'>
          {t(
            'Leave a knob at 0 to inherit the global policy. Only this model is affected.'
          )}
        </p>
      )}

      <div className='flex flex-wrap items-center justify-between gap-3'>
        <div className='flex items-center gap-3'>
          <Switch
            id={`pricing-policy-enabled-${props.scope}`}
            checked={policy.enabled}
            onCheckedChange={(enabled) =>
              setPolicy((current) => ({ ...current, enabled }))
            }
          />
          <Label htmlFor={`pricing-policy-enabled-${props.scope}`}>
            {t('Enable policy')}
          </Label>
        </div>
        <div className='flex items-center gap-2'>
          {props.onCancel && (
            <Button type='button' variant='ghost' onClick={props.onCancel}>
              {t('Cancel')}
            </Button>
          )}
          <Button type='submit' disabled={props.saving}>
            {t('Save pricing policy')}
          </Button>
        </div>
      </div>
    </form>
  )
}
