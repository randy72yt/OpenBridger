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

const initialPolicy: ModelPricePolicy = {
  public_model: '',
  service_tier: 'default',
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

type PricingControlPolicyFormProps = {
  saving: boolean
  onSave: (policy: ModelPricePolicy) => void
}

export function PricingControlPolicyForm(props: PricingControlPolicyFormProps) {
  const { t } = useTranslation()
  const [policy, setPolicy] = useState<ModelPricePolicy>(initialPolicy)
  const setNumber = (key: keyof ModelPricePolicy, value: string) => {
    setPolicy((current) => ({ ...current, [key]: Number(value) }))
  }

  return (
    <form
      className='grid gap-4 md:grid-cols-2 xl:grid-cols-4'
      onSubmit={(event) => {
        event.preventDefault()
        props.onSave(policy)
      }}
    >
      <div className='space-y-2 md:col-span-2'>
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
        <Label htmlFor='pricing-primary-channel'>{t('Primary channel')}</Label>
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
        <Label htmlFor='pricing-backup-channel'>{t('Backup channel')}</Label>
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
      {[
        ['target_margin_bps', 'Target margin (BPS)'],
        ['minimum_margin_bps', 'Minimum margin (BPS)'],
        ['overhead_bps', 'Operating overhead (BPS)'],
        ['fallback_probability_bps', 'Fallback probability (BPS)'],
      ].map(([key, label]) => (
        <div className='space-y-2' key={key}>
          <Label htmlFor={`pricing-${key}`}>{t(label)}</Label>
          <Input
            id={`pricing-${key}`}
            type='number'
            min={0}
            max={10000}
            value={Number(policy[key as keyof ModelPricePolicy])}
            onChange={(event) =>
              setNumber(key as keyof ModelPricePolicy, event.target.value)
            }
          />
        </div>
      ))}
      <div className='flex items-center gap-3 md:col-span-2'>
        <Switch
          id='pricing-policy-enabled'
          checked={policy.enabled}
          onCheckedChange={(enabled) =>
            setPolicy((current) => ({ ...current, enabled }))
          }
        />
        <Label htmlFor='pricing-policy-enabled'>{t('Enable policy')}</Label>
      </div>
      <div className='flex justify-end md:col-span-2'>
        <Button type='submit' disabled={props.saving}>
          {t('Save pricing policy')}
        </Button>
      </div>
    </form>
  )
}
