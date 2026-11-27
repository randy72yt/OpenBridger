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
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useMemo, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import * as z from 'zod'

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useResetForm } from '../hooks/use-reset-form'
import { useUpdateOption } from '../hooks/use-update-option'
import { safeNumberFieldProps } from '../utils/numeric-field'

const numericString = z.string().refine((value) => {
  const trimmed = value.trim()
  if (!trimmed) return true
  return !Number.isNaN(Number(trimmed)) && Number(trimmed) >= 0
}, 'Enter a non-negative number or leave empty')

const monitoringSchema = z.object({
  QuotaRemindThreshold: numericString,
  perf_metrics_setting: z.object({
    enabled: z.boolean(),
    flush_interval: z.coerce.number().min(1),
    bucket_time: z.enum(['minute', '5min', 'hour']),
    retention_days: z.coerce.number().min(0),
    public_v2: z.boolean(),
    min_public_samples: z.coerce.number().min(0),
    internal_token_names: z.string(),
    probe_enabled: z.boolean(),
    probe_scan_minutes: z.coerce.number().min(1),
    probe_max_per_run: z.coerce.number().min(0),
    probe_daily_budget_quota: z.coerce.number().min(0),
    probe_real_traffic_threshold: z.coerce.number().min(0),
    probe_models: z.string(),
  }),
})

type MonitoringFormInput = z.input<typeof monitoringSchema>
type MonitoringFormValues = z.output<typeof monitoringSchema>

type FlatMonitoringDefaults = {
  QuotaRemindThreshold: string
  'perf_metrics_setting.enabled': boolean
  'perf_metrics_setting.flush_interval': number
  'perf_metrics_setting.bucket_time': 'minute' | '5min' | 'hour'
  'perf_metrics_setting.retention_days': number
  'perf_metrics_setting.public_v2': boolean
  'perf_metrics_setting.min_public_samples': number
  'perf_metrics_setting.internal_token_names': string
  'perf_metrics_setting.probe_enabled': boolean
  'perf_metrics_setting.probe_scan_minutes': number
  'perf_metrics_setting.probe_max_per_run': number
  'perf_metrics_setting.probe_daily_budget_quota': number
  'perf_metrics_setting.probe_real_traffic_threshold': number
  'perf_metrics_setting.probe_models': string
}

type MonitoringSettingsSectionProps = {
  defaultValues: FlatMonitoringDefaults
}

const buildFormDefaults = (
  defaults: MonitoringSettingsSectionProps['defaultValues']
): MonitoringFormInput => ({
  QuotaRemindThreshold: defaults.QuotaRemindThreshold ?? '',
  perf_metrics_setting: {
    enabled: defaults['perf_metrics_setting.enabled'],
    flush_interval: defaults['perf_metrics_setting.flush_interval'],
    bucket_time: defaults['perf_metrics_setting.bucket_time'],
    retention_days: defaults['perf_metrics_setting.retention_days'],
    public_v2: defaults['perf_metrics_setting.public_v2'],
    min_public_samples: defaults['perf_metrics_setting.min_public_samples'],
    internal_token_names: defaults['perf_metrics_setting.internal_token_names'],
    probe_enabled: defaults['perf_metrics_setting.probe_enabled'],
    probe_scan_minutes: defaults['perf_metrics_setting.probe_scan_minutes'],
    probe_max_per_run: defaults['perf_metrics_setting.probe_max_per_run'],
    probe_daily_budget_quota:
      defaults['perf_metrics_setting.probe_daily_budget_quota'],
    probe_real_traffic_threshold:
      defaults['perf_metrics_setting.probe_real_traffic_threshold'],
    probe_models: defaults['perf_metrics_setting.probe_models'],
  },
})

const normalizeDefaults = (
  defaults: MonitoringSettingsSectionProps['defaultValues']
): FlatMonitoringDefaults => ({
  QuotaRemindThreshold: (defaults.QuotaRemindThreshold ?? '').trim(),
  'perf_metrics_setting.enabled': defaults['perf_metrics_setting.enabled'],
  'perf_metrics_setting.flush_interval':
    defaults['perf_metrics_setting.flush_interval'],
  'perf_metrics_setting.bucket_time':
    defaults['perf_metrics_setting.bucket_time'],
  'perf_metrics_setting.retention_days':
    defaults['perf_metrics_setting.retention_days'],
  'perf_metrics_setting.public_v2': defaults['perf_metrics_setting.public_v2'],
  'perf_metrics_setting.min_public_samples':
    defaults['perf_metrics_setting.min_public_samples'],
  'perf_metrics_setting.internal_token_names':
    defaults['perf_metrics_setting.internal_token_names'],
  'perf_metrics_setting.probe_enabled':
    defaults['perf_metrics_setting.probe_enabled'],
  'perf_metrics_setting.probe_scan_minutes':
    defaults['perf_metrics_setting.probe_scan_minutes'],
  'perf_metrics_setting.probe_max_per_run':
    defaults['perf_metrics_setting.probe_max_per_run'],
  'perf_metrics_setting.probe_daily_budget_quota':
    defaults['perf_metrics_setting.probe_daily_budget_quota'],
  'perf_metrics_setting.probe_real_traffic_threshold':
    defaults['perf_metrics_setting.probe_real_traffic_threshold'],
  'perf_metrics_setting.probe_models':
    defaults['perf_metrics_setting.probe_models'],
})

const normalizeFormValues = (
  values: MonitoringFormValues
): FlatMonitoringDefaults => ({
  QuotaRemindThreshold: values.QuotaRemindThreshold.trim(),
  'perf_metrics_setting.enabled': values.perf_metrics_setting.enabled,
  'perf_metrics_setting.flush_interval':
    values.perf_metrics_setting.flush_interval,
  'perf_metrics_setting.bucket_time': values.perf_metrics_setting.bucket_time,
  'perf_metrics_setting.retention_days':
    values.perf_metrics_setting.retention_days,
  'perf_metrics_setting.public_v2': values.perf_metrics_setting.public_v2,
  'perf_metrics_setting.min_public_samples':
    values.perf_metrics_setting.min_public_samples,
  'perf_metrics_setting.internal_token_names':
    values.perf_metrics_setting.internal_token_names.trim(),
  'perf_metrics_setting.probe_enabled':
    values.perf_metrics_setting.probe_enabled,
  'perf_metrics_setting.probe_scan_minutes':
    values.perf_metrics_setting.probe_scan_minutes,
  'perf_metrics_setting.probe_max_per_run':
    values.perf_metrics_setting.probe_max_per_run,
  'perf_metrics_setting.probe_daily_budget_quota':
    values.perf_metrics_setting.probe_daily_budget_quota,
  'perf_metrics_setting.probe_real_traffic_threshold':
    values.perf_metrics_setting.probe_real_traffic_threshold,
  'perf_metrics_setting.probe_models': values.perf_metrics_setting.probe_models,
})

export function MonitoringSettingsSection({
  defaultValues,
}: MonitoringSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const baselineRef = useRef<FlatMonitoringDefaults>(
    normalizeDefaults(defaultValues)
  )
  const baselineSerializedRef = useRef<string>(
    JSON.stringify(normalizeDefaults(defaultValues))
  )

  const formDefaults = useMemo(
    () => buildFormDefaults(defaultValues),
    [defaultValues]
  )

  const form = useForm<MonitoringFormInput, unknown, MonitoringFormValues>({
    resolver: zodResolver(monitoringSchema),
    defaultValues: formDefaults,
  })

  useResetForm(form, formDefaults)

  useEffect(() => {
    const normalized = normalizeDefaults(defaultValues)
    const serialized = JSON.stringify(normalized)
    if (serialized === baselineSerializedRef.current) return
    baselineRef.current = normalized
    baselineSerializedRef.current = serialized
  }, [defaultValues])

  const perfMetricsEnabled = form.watch('perf_metrics_setting.enabled')
  const probeEnabled = form.watch('perf_metrics_setting.probe_enabled')

  const onSubmit = async (values: MonitoringFormValues) => {
    const normalized = normalizeFormValues(values)
    const updates = (
      Object.keys(normalized) as Array<keyof FlatMonitoringDefaults>
    ).filter((key) => normalized[key] !== baselineRef.current[key])

    if (updates.length === 0) {
      toast.info(t('No changes to save'))
      return
    }

    for (const key of updates) {
      await updateOption.mutateAsync({
        key,
        value: normalized[key],
      })
    }

    baselineRef.current = normalized
    baselineSerializedRef.current = JSON.stringify(normalized)
  }

  return (
    <SettingsSection title={t('Monitoring & Alerts')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)}>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
          />
          <FormField
            control={form.control}
            name='QuotaRemindThreshold'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Quota reminder (tokens)')}</FormLabel>
                <FormControl>
                  <Input
                    type='number'
                    min={0}
                    step={1}
                    value={field.value}
                    onChange={(event) => field.onChange(event.target.value)}
                  />
                </FormControl>
                <FormDescription>
                  {t('Send email alerts when a user falls below this quota')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <div>
            <h4 className='font-medium'>{t('Model performance metrics')}</h4>
            <p className='text-muted-foreground mt-1 text-xs'>
              {t(
                'Collect relay latency and success-rate metrics for the model square.'
              )}
            </p>
          </div>

          <div className='grid grid-cols-1 gap-4 md:grid-cols-4'>
            <FormField
              control={form.control}
              name='perf_metrics_setting.enabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>
                      {t('Enable model performance metrics')}
                    </FormLabel>
                  </SettingsSwitchContent>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </SettingsSwitchItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.flush_interval'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Flush interval (minutes)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      step={1}
                      {...safeNumberFieldProps(field)}
                      disabled={!perfMetricsEnabled}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.bucket_time'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Aggregation bucket')}</FormLabel>
                  <Select
                    items={[
                      { value: 'minute', label: t('1 minute') },
                      { value: '5min', label: t('5 minutes') },
                      { value: 'hour', label: t('1 hour') },
                    ]}
                    value={field.value}
                    onValueChange={field.onChange}
                    disabled={!perfMetricsEnabled}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        <SelectItem value='minute'>{t('1 minute')}</SelectItem>
                        <SelectItem value='5min'>{t('5 minutes')}</SelectItem>
                        <SelectItem value='hour'>{t('1 hour')}</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.retention_days'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Retention days')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={0}
                      step={1}
                      {...safeNumberFieldProps(field)}
                      disabled={!perfMetricsEnabled}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('0 means data is kept permanently')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className='grid grid-cols-1 gap-4 md:grid-cols-4'>
            <FormField
              control={form.control}
              name='perf_metrics_setting.public_v2'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Public v2 metrics')}</FormLabel>
                    <FormDescription>
                      {t(
                        'Show availability, TTFT and sample counts on the public model square. Recommended after one week of ops review.'
                      )}
                    </FormDescription>
                  </SettingsSwitchContent>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                      disabled={!perfMetricsEnabled}
                    />
                  </FormControl>
                </SettingsSwitchItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.min_public_samples'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Minimum public samples')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={0}
                      step={1}
                      {...safeNumberFieldProps(field)}
                      disabled={!perfMetricsEnabled}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Models with fewer samples are grayed out as insufficient on the public square.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.internal_token_names'
              render={({ field }) => (
                <FormItem className='md:col-span-2'>
                  <FormLabel>{t('Internal token names')}</FormLabel>
                  <FormControl>
                    <Input
                      value={field.value}
                      onChange={(event) => field.onChange(event.target.value)}
                      disabled={!perfMetricsEnabled}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Comma-separated token names excluded from public metrics (plus the playground).'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div>
            <h4 className='font-medium'>{t('Model probing')}</h4>
            <p className='text-muted-foreground mt-1 text-xs'>
              {t(
                'Probe on-sale models without recent real traffic through the normal relay chain, budget-capped per day.'
              )}
            </p>
          </div>

          <div className='grid grid-cols-1 gap-4 md:grid-cols-4'>
            <FormField
              control={form.control}
              name='perf_metrics_setting.probe_enabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Enable model probing')}</FormLabel>
                  </SettingsSwitchContent>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                      disabled={!perfMetricsEnabled}
                    />
                  </FormControl>
                </SettingsSwitchItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.probe_scan_minutes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Probe scan interval (minutes)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      step={1}
                      {...safeNumberFieldProps(field)}
                      disabled={!probeEnabled}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.probe_max_per_run'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Max probes per run')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={0}
                      step={1}
                      {...safeNumberFieldProps(field)}
                      disabled={!probeEnabled}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.probe_daily_budget_quota'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Daily probe budget (quota)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={0}
                      step={1000}
                      {...safeNumberFieldProps(field)}
                      disabled={!probeEnabled}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('500000 quota equals $1 by default')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.probe_real_traffic_threshold'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Real traffic threshold (per hour)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={0}
                      step={1}
                      {...safeNumberFieldProps(field)}
                      disabled={!probeEnabled}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Models with at least this many real requests in the last hour are not probed.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='perf_metrics_setting.probe_models'
              render={({ field }) => (
                <FormItem className='md:col-span-3'>
                  <FormLabel>{t('Probe model overrides')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={6}
                      className='font-mono text-xs'
                      value={field.value}
                      onChange={(event) => field.onChange(event.target.value)}
                      disabled={!probeEnabled}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'One model per line: <model> <tier:cheap|medium|expensive|off> [token_param] [limit] [endpoint:chat|responses]. Lines starting with # are comments.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
