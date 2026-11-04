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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  AlertTriangle,
  Calculator,
  Database,
  RefreshCw,
  ShieldCheck,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { useDebounce } from '@/hooks/use-debounce'

import { SettingsCard } from '../components/settings-card'
import { SettingsSection } from '../components/settings-section'
import {
  getPricingOffers,
  getPricingPolicies,
  getPricingProposals,
  getPricingRisks,
  getPricingSyncRuns,
  importPricingOffers,
  recalculatePricing,
  savePricingPolicy,
  syncPricingOffers,
  updatePricingProposal,
} from './pricing-control-api'
import { PricingControlInventory } from './pricing-control-inventory'
import { PricingControlPolicyPanel } from './pricing-control-policy-panel'
import type {
  ModelPricePolicy,
  ModelPriceProposal,
  PricingApiResponse,
  PricingOfferSyncResult,
  PricingSyncRun,
  UpstreamModelOffer,
} from './pricing-control-types'
import { PricingOfferSyncSummary } from './pricing-offer-sync-summary'
import { PricingProposalReview } from './pricing-proposal-review'

const pricingControlKey = ['pricing-control'] as const
const pageSize = 50

function sectionClass(active: boolean, extra = '') {
  const base = `scroll-mt-24 rounded-xl transition-shadow ${extra}`
  return active ? `${base} ring-primary/40 ring-2` : base
}

function PersistedSyncSummary(props: { run: PricingSyncRun }) {
  const { t } = useTranslation()
  const variant = props.run.status === 'succeeded' ? 'default' : 'destructive'
  return (
    <Alert variant={variant}>
      <AlertTitle>
        {t('Last sync')} ·{' '}
        {new Date(props.run.finished_at * 1000).toLocaleString()} ·{' '}
        {t(props.run.trigger === 'scheduled' ? 'Scheduled' : 'Manual')}
      </AlertTitle>
      <AlertDescription className='text-current'>
        {t(
          'Imported {{imported}} offers; kept {{kept}} higher-priority offers; skipped {{skipped}} models; created {{proposals}} proposals.',
          {
            imported: props.run.imported,
            kept: props.run.kept,
            skipped: props.run.skipped,
            proposals: props.run.created_proposals,
          }
        )}
        {props.run.error && <p className='mt-2'>{props.run.error}</p>}
      </AlertDescription>
    </Alert>
  )
}

export function PricingControlSection() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [offerJson, setOfferJson] = useState('[]')
  const [lastSync, setLastSync] =
    useState<PricingApiResponse<PricingOfferSyncResult> | null>(null)
  const [publishTarget, setPublishTarget] = useState<ModelPriceProposal | null>(
    null
  )
  const [offerKeywordInput, setOfferKeywordInput] = useState('')
  const offerKeyword = useDebounce(offerKeywordInput, 300)
  const [offerSource, setOfferSource] = useState('')
  const [includeExpired, setIncludeExpired] = useState(false)
  const [offerOffset, setOfferOffset] = useState(0)
  const [proposalOffset, setProposalOffset] = useState(0)
  const [highlight, setHighlight] = useState<string | null>(null)

  useEffect(() => {
    setOfferOffset(0)
  }, [offerKeyword, offerSource, includeExpired])

  const overviewQuery = useQuery({
    queryKey: [...pricingControlKey, 'overview'],
    queryFn: async () => {
      const [policies, risks, syncRuns, pending] = await Promise.all([
        getPricingPolicies(),
        getPricingRisks(),
        getPricingSyncRuns(10),
        getPricingProposals({ status: 'pending', limit: 1, offset: 0 }),
      ])
      return { policies, risks, syncRuns, pendingTotal: pending.total }
    },
    refetchInterval: 60000,
  })

  const offersQuery = useQuery({
    queryKey: [
      ...pricingControlKey,
      'offers',
      { offerKeyword, offerSource, includeExpired, offerOffset },
    ],
    queryFn: () =>
      getPricingOffers({
        keyword: offerKeyword,
        sourceType: offerSource,
        includeExpired,
        limit: pageSize,
        offset: offerOffset,
      }),
    refetchInterval: 60000,
  })

  const proposalsQuery = useQuery({
    queryKey: [...pricingControlKey, 'proposals', proposalOffset],
    queryFn: () =>
      getPricingProposals({ limit: pageSize, offset: proposalOffset }),
    refetchInterval: 60000,
  })

  const refresh = async () => {
    await queryClient.invalidateQueries({ queryKey: pricingControlKey })
  }

  const jumpTo = (id: string) => {
    // "Pricing risks" only exists while there are risks; fall back to the
    // inventory card so the click never becomes a no-op.
    const element =
      document.querySelector(`#${id}`) ??
      document.querySelector('#pricing-offers')
    if (!element) return
    element.scrollIntoView({ behavior: 'smooth', block: 'start' })
    setHighlight(id)
    window.setTimeout(() => setHighlight(null), 1600)
  }

  const policyMutation = useMutation({
    mutationFn: (policy: ModelPricePolicy) => savePricingPolicy(policy),
    onSuccess: async () => {
      toast.success(t('Pricing policy saved'))
      await refresh()
    },
    onError: (error: Error) => toast.error(error.message),
  })
  const proposalMutation = useMutation({
    mutationFn: (request: {
      id: number
      action: 'approve' | 'reject' | 'publish'
    }) => updatePricingProposal(request.id, request.action),
    onSuccess: async (_, request) => {
      toast.success(
        t('Pricing proposal {{action}}', { action: request.action })
      )
      setPublishTarget(null)
      await refresh()
    },
    onError: (error: Error) => toast.error(error.message),
  })
  const importMutation = useMutation({
    mutationFn: async () => {
      const parsed: unknown = JSON.parse(offerJson)
      if (!Array.isArray(parsed)) throw new Error(t('Enter a JSON array'))
      return importPricingOffers(parsed as UpstreamModelOffer[])
    },
    onSuccess: async () => {
      toast.success(t('Upstream costs imported'))
      await refresh()
    },
    onError: (error: Error) => toast.error(error.message),
  })
  const recalculateMutation = useMutation({
    mutationFn: recalculatePricing,
    onSuccess: async () => {
      toast.success(t('Pricing recalculation queued'))
      await refresh()
    },
    onError: (error: Error) => toast.error(error.message),
  })
  const syncMutation = useMutation({
    mutationFn: syncPricingOffers,
    onSuccess: async (response) => {
      setLastSync(response)
      if (!response.success) {
        toast.error(response.message)
      } else if (
        response.data.channels.some(
          (channel) => channel.error || channel.warnings?.length
        )
      ) {
        toast.warning(t('Upstream price sync needs review'))
      } else {
        toast.success(t('Upstream prices synchronized'))
      }
      setOfferOffset(0)
      await refresh()
    },
    onError: (error: Error) => {
      setLastSync(null)
      toast.error(error.message)
    },
  })

  if (overviewQuery.isLoading) {
    return (
      <SettingsSection title={t('Pricing Control')}>
        <p className='text-muted-foreground text-sm'>{t('Loading...')}</p>
      </SettingsSection>
    )
  }
  if (overviewQuery.isError || !overviewQuery.data) {
    return (
      <SettingsSection title={t('Pricing Control')}>
        <Alert variant='destructive'>
          <AlertTitle>{t('Failed to load pricing control data')}</AlertTitle>
          <AlertDescription>
            <Button
              size='sm'
              variant='outline'
              onClick={() => overviewQuery.refetch()}
            >
              {t('Retry')}
            </Button>
          </AlertDescription>
        </Alert>
      </SettingsSection>
    )
  }

  const { policies, risks, syncRuns, pendingTotal } = overviewQuery.data
  const modelPolicies = policies.filter(
    (policy) => policy.scope !== 'global' && policy.public_model !== '*'
  )
  const activePolicies = modelPolicies.filter((policy) => policy.enabled).length
  const offersTotal = offersQuery.data?.total ?? 0

  const summary = [
    {
      id: 'pricing-offers',
      icon: Database,
      label: t('Current upstream offers'),
      value: offersTotal,
      hint: t('View inventory'),
    },
    {
      id: 'pricing-policy',
      icon: Calculator,
      label: t('Active pricing policies'),
      value: activePolicies,
      hint: t('View policies'),
    },
    {
      id: 'pricing-proposals',
      icon: ShieldCheck,
      label: t('Pending proposals'),
      value: pendingTotal,
      hint: t('View proposals'),
    },
    {
      id: 'pricing-risks',
      icon: AlertTriangle,
      label: t('Pricing risks'),
      value: risks.length,
      hint: t('View risks'),
    },
  ] as const

  return (
    <SettingsSection title={t('Pricing Control')}>
      <div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>
        {summary.map(({ id, icon: Icon, label, value, hint }) => (
          <Card
            key={id}
            role='button'
            tabIndex={0}
            aria-label={`${label}: ${value}. ${hint}`}
            onClick={() => jumpTo(id)}
            onKeyDown={(event) => {
              if (event.key === 'Enter' || event.key === ' ') {
                event.preventDefault()
                jumpTo(id)
              }
            }}
            className='hover:border-primary/40 cursor-pointer transition-colors'
          >
            <CardHeader className='flex flex-row items-center justify-between pb-2'>
              <CardTitle className='text-sm font-medium'>{label}</CardTitle>
              <Icon
                className='text-muted-foreground size-4'
                aria-hidden='true'
              />
            </CardHeader>
            <CardContent>
              <p className='text-2xl font-semibold tabular-nums'>{value}</p>
              <p className='text-muted-foreground mt-1 text-xs'>{hint}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      <div
        id='pricing-policy'
        className={sectionClass(highlight === 'pricing-policy')}
      >
        <SettingsCard
          title={t('Pricing policy')}
          description={t(
            'The global policy is always visible. Per-model overrides are collapsed and only affect the model you edit.'
          )}
        >
          <PricingControlPolicyPanel
            policies={policies}
            saving={policyMutation.isPending}
            onSave={(policy) => policyMutation.mutate(policy)}
          />
        </SettingsCard>
      </div>

      <div
        id='pricing-sync'
        className={sectionClass(highlight === 'pricing-sync')}
      >
        <SettingsCard
          title={t('Upstream cost import')}
          description={t(
            'Import normalized USD prices per million tokens. Every import is saved to the price history.'
          )}
        >
          <div className='space-y-3'>
            <div className='flex flex-wrap items-center justify-between gap-3 rounded-lg border p-3'>
              <p className='text-muted-foreground text-sm'>
                {t(
                  'Fetch current model prices and effective group ratios from configured upstream channels.'
                )}
              </p>
              <Button
                onClick={() => syncMutation.mutate()}
                disabled={syncMutation.isPending}
              >
                <RefreshCw
                  className={syncMutation.isPending ? 'animate-spin' : ''}
                />
                {t('Sync upstream prices')}
              </Button>
            </div>
            <Alert>
              <AlertTitle>{t('How cost is computed')}</AlertTitle>
              <AlertDescription>
                {t(
                  'Collected unit prices are the upstream base price. Actual procurement cost = base price × the settlement group ratio recorded on the offer. When an upstream exposes several groups per model, PRICING_GROUP_RATIO_POLICY decides which ratio is used (key_group / max / min).'
                )}
              </AlertDescription>
            </Alert>
            {lastSync && <PricingOfferSyncSummary response={lastSync} />}
            {!lastSync && syncRuns.latest && (
              <PersistedSyncSummary run={syncRuns.latest} />
            )}
            <Textarea
              value={offerJson}
              onChange={(event) => setOfferJson(event.target.value)}
              className='min-h-32 font-mono text-xs'
              aria-label={t('Upstream cost JSON')}
            />
            <div className='flex flex-wrap justify-between gap-2'>
              <p className='text-muted-foreground text-sm'>
                {t(
                  'Fields: channel_id, upstream_model, public_model, input_cost, output_cost, upstream_group_ratio, currency, source_type, enabled.'
                )}
              </p>
              <Button
                variant='outline'
                onClick={() => importMutation.mutate()}
                disabled={importMutation.isPending}
              >
                {t('Import costs')}
              </Button>
            </div>
          </div>
        </SettingsCard>
      </div>

      <div
        id='pricing-offers'
        className={sectionClass(highlight === 'pricing-offers')}
      >
        <PricingControlInventory
          offers={offersQuery.data?.items ?? []}
          total={offersTotal}
          limit={offersQuery.data?.limit ?? pageSize}
          offset={offersQuery.data?.offset ?? offerOffset}
          loading={offersQuery.isFetching}
          keyword={offerKeywordInput}
          sourceType={offerSource}
          includeExpired={includeExpired}
          risks={risks}
          onKeywordChange={setOfferKeywordInput}
          onSourceTypeChange={setOfferSource}
          onIncludeExpiredChange={setIncludeExpired}
          onPageChange={setOfferOffset}
          onResync={() => syncMutation.mutate()}
          onFocusModel={(model) => {
            setOfferKeywordInput(model)
            setIncludeExpired(true)
            setOfferOffset(0)
          }}
        />
      </div>

      <div
        id='pricing-proposals'
        className={sectionClass(highlight === 'pricing-proposals')}
      >
        <PricingProposalReview
          proposals={proposalsQuery.data?.items ?? []}
          total={proposalsQuery.data?.total ?? 0}
          limit={proposalsQuery.data?.limit ?? pageSize}
          offset={proposalsQuery.data?.offset ?? proposalOffset}
          loading={proposalsQuery.isFetching}
          recalculating={recalculateMutation.isPending}
          updating={proposalMutation.isPending}
          onRecalculate={() => recalculateMutation.mutate()}
          onAction={(id, action) => proposalMutation.mutate({ id, action })}
          onPublish={setPublishTarget}
          onPageChange={setProposalOffset}
        />
      </div>

      <ConfirmDialog
        open={publishTarget !== null}
        onOpenChange={(open) => !open && setPublishTarget(null)}
        title={t('Publish model price')}
        desc={t(
          'Review the final pricing change for {{model}} before publishing.',
          {
            model: publishTarget?.public_model ?? '',
          }
        )}
        confirmText={t('Publish')}
        isLoading={proposalMutation.isPending}
        handleConfirm={() =>
          publishTarget &&
          proposalMutation.mutate({ id: publishTarget.id, action: 'publish' })
        }
      >
        {publishTarget && (
          <div className='bg-muted/50 grid gap-3 rounded-lg border p-4 text-sm sm:grid-cols-2'>
            <p>
              {t('Current input price')}:{' '}
              <strong>${publishTarget.current_input_price ?? '—'}</strong>
            </p>
            <p>
              {t('Proposed input price')}:{' '}
              <strong>${publishTarget.proposed_input_price}</strong>
            </p>
            <p>
              {t('Current output price')}:{' '}
              <strong>${publishTarget.current_output_price ?? '—'}</strong>
            </p>
            <p>
              {t('Proposed output price')}:{' '}
              <strong>${publishTarget.proposed_output_price}</strong>
            </p>
            <p>
              {t('Expected margin')}:{' '}
              <strong>
                {(publishTarget.expected_margin_bps / 100).toFixed(2)}%
              </strong>
            </p>
            <p>
              {t('Worst margin')}:{' '}
              <strong>
                {(publishTarget.worst_margin_bps / 100).toFixed(2)}%
              </strong>
            </p>
          </div>
        )}
      </ConfirmDialog>
    </SettingsSection>
  )
}
