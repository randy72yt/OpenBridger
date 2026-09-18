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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { ErrorState } from '@/components/error-state'
import { LoadingState } from '@/components/loading-state'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'

import { SettingsCard } from '../components/settings-card'
import { SettingsSection } from '../components/settings-section'
import {
  getPricingControlData,
  importPricingOffers,
  recalculatePricing,
  savePricingPolicy,
  syncPricingOffers,
  updatePricingProposal,
} from './pricing-control-api'
import { PricingControlInventory } from './pricing-control-inventory'
import { PricingControlPolicyForm } from './pricing-control-policy-form'
import type {
  ModelPricePolicy,
  ModelPriceProposal,
  UpstreamModelOffer,
} from './pricing-control-types'
import { PricingProposalReview } from './pricing-proposal-review'

const pricingControlKey = ['pricing-control'] as const

export function PricingControlSection() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [offerJson, setOfferJson] = useState('[]')
  const [publishTarget, setPublishTarget] = useState<ModelPriceProposal | null>(
    null
  )
  const query = useQuery({
    queryKey: pricingControlKey,
    queryFn: getPricingControlData,
    refetchInterval: 15000,
  })
  const refresh = async () => {
    await queryClient.invalidateQueries({ queryKey: pricingControlKey })
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
    onSuccess: () => toast.success(t('Pricing recalculation queued')),
    onError: (error: Error) => toast.error(error.message),
  })
  const syncMutation = useMutation({
    mutationFn: syncPricingOffers,
    onSuccess: async () => {
      toast.success(t('Upstream prices synchronized'))
      await refresh()
    },
    onError: (error: Error) => toast.error(error.message),
  })

  if (query.isLoading) return <LoadingState />
  if (query.isError || !query.data) {
    return (
      <ErrorState
        description={t('Failed to load pricing control data')}
        onRetry={() => query.refetch()}
      />
    )
  }

  const pending = query.data.proposals.filter(
    (proposal) => proposal.status === 'pending'
  ).length
  const summary = [
    [Database, t('Current upstream offers'), query.data.offers.length],
    [Calculator, t('Active pricing policies'), query.data.policies.length],
    [ShieldCheck, t('Pending proposals'), pending],
    [AlertTriangle, t('Pricing risks'), query.data.risks.length],
  ] as const

  return (
    <SettingsSection title={t('Pricing Control')}>
      <div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>
        {summary.map(([Icon, label, value]) => (
          <Card key={label}>
            <CardHeader className='flex flex-row items-center justify-between pb-2'>
              <CardTitle className='text-sm font-medium'>{label}</CardTitle>
              <Icon
                className='text-muted-foreground size-4'
                aria-hidden='true'
              />
            </CardHeader>
            <CardContent className='text-2xl font-semibold tabular-nums'>
              {value}
            </CardContent>
          </Card>
        ))}
      </div>

      <SettingsCard
        title={t('Pricing policy')}
        description={t(
          'Set the main and backup channels, target margin, operating overhead, and fallback probability for each public model.'
        )}
      >
        <PricingControlPolicyForm
          saving={policyMutation.isPending}
          onSave={(policy) => policyMutation.mutate(policy)}
        />
      </SettingsCard>

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

      <PricingControlInventory
        offers={query.data.offers}
        policies={query.data.policies}
        risks={query.data.risks}
      />

      <PricingProposalReview
        proposals={query.data.proposals}
        recalculating={recalculateMutation.isPending}
        updating={proposalMutation.isPending}
        onRecalculate={() => recalculateMutation.mutate()}
        onAction={(id, action) => proposalMutation.mutate({ id, action })}
        onPublish={setPublishTarget}
      />

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
