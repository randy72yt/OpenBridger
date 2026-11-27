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
import { useQuery } from '@tanstack/react-query'
import { VChart } from '@visactor/react-vchart'
import type { ISpec } from '@visactor/vchart'
import {
  Activity,
  AlertTriangle,
  Coins,
  CreditCard,
  DollarSign,
  Landmark,
  Loader2,
  PiggyBank,
  Receipt,
  Scale,
  TrendingUp,
  Users,
  Wallet,
  type LucideIcon,
} from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { ModelProviderIcon } from '@/components/model-provider-icon'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { IconBadge, type IconBadgeTone } from '@/components/ui/icon-badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { getAdminBusinessDashboard } from '@/features/dashboard/api'
import type {
  BusinessDailyPoint,
  BusinessDashboardSummary,
  ModelMarginStat,
  PaymentChannelStat,
} from '@/features/dashboard/types'
import { getRollingDateRange } from '@/lib/time'
import { useChartTheme } from '@/lib/use-chart-theme'
import { cn } from '@/lib/utils'
import { VCHART_OPTION } from '@/lib/vchart'

const BUSINESS_RANGE_OPTIONS = [
  { days: 1, labelKey: '1D' },
  { days: 7, labelKey: '7D' },
  { days: 14, labelKey: '14D' },
  { days: 30, labelKey: '30D' },
  { days: 90, labelKey: '90D' },
] as const

const KPI_SKELETON_KEYS = [
  'cash-rev',
  'consumption',
  'est-cost',
  'gross-profit',
  'paying-users',
  'user-growth',
  'wallet-pool',
  'channel-bal',
] as const

function formatUSD(value: number | null | undefined, digits = 2): string {
  if (value == null || !Number.isFinite(value)) {
    return '-'
  }
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function formatOriginalMoney(amount: number, currency: string): string {
  const code = currency.toUpperCase() === 'CNY' ? 'CNY' : 'USD'
  return new Intl.NumberFormat(code === 'CNY' ? 'zh-CN' : 'en-US', {
    style: 'currency',
    currency: code,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount)
}

function formatCompactInt(value: number): string {
  return new Intl.NumberFormat(undefined, {
    notation: value >= 100_000 ? 'compact' : 'standard',
    maximumFractionDigits: 1,
  }).format(value)
}

interface KpiCardItem {
  id: string
  title: string
  value: string
  subtitle: string
  icon: LucideIcon
  tone: IconBadgeTone
  badgeText?: string
  badgeVariant?: 'default' | 'secondary' | 'destructive' | 'outline'
}

function buildFinancialTrendSpec(
  daily: BusinessDailyPoint[],
  labels: { revenue: string; consumption: string; cost: string }
): ISpec {
  const values = daily.flatMap((pt) => [
    { date: pt.date.slice(5), metric: labels.revenue, value: pt.cash_revenue_usd },
    {
      date: pt.date.slice(5),
      metric: labels.consumption,
      value: pt.consumption_usd,
    },
    {
      date: pt.date.slice(5),
      metric: labels.cost,
      value: pt.estimated_cost_usd,
    },
  ])

  return {
    type: 'area',
    data: [{ id: 'financial_trend', values }],
    xField: 'date',
    yField: 'value',
    seriesField: 'metric',
    color: ['#FF6B00', '#3B82F6', '#64748B'],
    line: {
      style: {
        curveType: 'monotone',
        lineWidth: 2,
      },
    },
    area: {
      style: {
        fillOpacity: 0.14,
        curveType: 'monotone',
      },
    },
    point: {
      visible: daily.length <= 14,
      style: { size: 5 },
    },
    legends: {
      visible: true,
      orient: 'top',
      position: 'end',
    },
    axes: [
      { orient: 'bottom', type: 'band' },
      {
        orient: 'left',
        type: 'linear',
        label: {
          formatMethod: (val: unknown) => `$${Number(val).toFixed(0)}`,
        },
      },
    ],
  } as unknown as ISpec
}

function buildGrowthBarSpec(
  daily: BusinessDailyPoint[],
  labels: { newUsers: string; paidOrders: string }
): ISpec {
  const values = daily.flatMap((pt) => [
    { date: pt.date.slice(5), metric: labels.newUsers, count: pt.new_users },
    {
      date: pt.date.slice(5),
      metric: labels.paidOrders,
      count: pt.paid_orders,
    },
  ])

  return {
    type: 'bar',
    data: [{ id: 'growth_trend', values }],
    xField: ['date', 'metric'],
    yField: 'count',
    seriesField: 'metric',
    color: ['#F59E0B', '#10B981'],
    bar: {
      style: {
        cornerRadius: [4, 4, 0, 0],
      },
    },
    legends: {
      visible: true,
      orient: 'top',
      position: 'end',
    },
    axes: [
      { orient: 'bottom', type: 'band' },
      { orient: 'left', type: 'linear' },
    ],
  } as unknown as ISpec
}

function getMarginBadgeClass(marginPct: number | null): string {
  if (marginPct == null) {
    return 'text-muted-foreground border-border bg-muted/40'
  }
  if (marginPct < 0) {
    return 'border-destructive/30 bg-destructive/10 text-destructive'
  }
  if (marginPct < 15) {
    return 'border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400'
  }
  return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
}

export function BusinessDashboard() {
  const { t } = useTranslation()
  const { resolvedTheme, themeReady } = useChartTheme()
  const [selectedDays, setSelectedDays] = useState<number>(30)

  const timeRange = useMemo(() => {
    const { start, end } = getRollingDateRange(selectedDays)
    return {
      start_timestamp: Math.floor(start.getTime() / 1000),
      end_timestamp: Math.floor(end.getTime() / 1000),
    }
  }, [selectedDays])

  const { data, isLoading, isError, error, isFetching } = useQuery({
    queryKey: ['dashboard', 'admin-business', timeRange],
    queryFn: async () => {
      const res = await getAdminBusinessDashboard(timeRange)
      if (!res.success || !res.data) {
        throw new Error(res.message || t('Failed to load'))
      }
      return res.data
    },
    staleTime: 60_000,
  })

  const summary: BusinessDashboardSummary | undefined = data?.summary
  const dailyTrend: BusinessDailyPoint[] = useMemo(
    () => data?.daily_trend ?? [],
    [data?.daily_trend]
  )
  const paymentChannels: PaymentChannelStat[] = useMemo(
    () => data?.payment_channels ?? [],
    [data?.payment_channels]
  )
  const modelMargins: ModelMarginStat[] = useMemo(
    () => data?.model_margins ?? [],
    [data?.model_margins]
  )

  const kpiCards = useMemo<KpiCardItem[]>(() => {
    if (!summary) return []
    const marginText =
      summary.estimated_gross_margin_pct != null
        ? `${summary.estimated_gross_margin_pct.toFixed(1)}%`
        : undefined
    const isNegativeMargin =
      summary.estimated_gross_margin_pct != null &&
      summary.estimated_gross_margin_pct < 0

    return [
      {
        id: 'cash-revenue',
        title: t('business.kpi.cashRevenue'),
        value: formatUSD(summary.cash_revenue_usd),
        subtitle: t('business.kpi.cashRevenueSub', {
          topup: formatUSD(summary.topup_revenue_usd),
          sub: formatUSD(summary.subscription_revenue_usd),
          orders: summary.paid_order_count,
        }),
        icon: DollarSign,
        tone: 'chart-1',
      },
      {
        id: 'consumption',
        title: t('business.kpi.consumption'),
        value: formatUSD(summary.consumption_usd),
        subtitle: t('business.kpi.consumptionSub', {
          requests: formatCompactInt(summary.total_requests),
          tokens: formatCompactInt(summary.total_tokens),
        }),
        icon: Activity,
        tone: 'info',
      },
      {
        id: 'estimated-cost',
        title: t('business.kpi.estimatedCost'),
        value: formatUSD(summary.estimated_cost_usd),
        subtitle: t('business.kpi.estimatedCostSub', {
          coverage: `${summary.cost_coverage_pct.toFixed(0)}%`,
          uncosted: formatUSD(summary.uncosted_consumption_usd),
        }),
        icon: Scale,
        tone: 'warning',
      },
      {
        id: 'gross-profit',
        title: t('business.kpi.grossProfit'),
        value: formatUSD(summary.estimated_gross_profit_usd),
        subtitle: t('business.kpi.grossProfitSub'),
        icon: TrendingUp,
        tone: isNegativeMargin ? 'destructive' : 'success',
        badgeText: marginText,
        badgeVariant: isNegativeMargin ? 'destructive' : 'secondary',
      },
      {
        id: 'paying-users',
        title: t('business.kpi.payingUsers'),
        value: formatCompactInt(summary.paying_users),
        subtitle: t('business.kpi.payingUsersSub', {
          arppu: formatUSD(summary.arppu_usd),
        }),
        icon: CreditCard,
        tone: 'chart-2',
      },
      {
        id: 'user-growth',
        title: t('business.kpi.userGrowth'),
        value: `+${formatCompactInt(summary.new_users)}`,
        subtitle: t('business.kpi.userGrowthSub', {
          total: formatCompactInt(summary.total_users),
          active: formatCompactInt(summary.active_users),
        }),
        icon: Users,
        tone: 'chart-3',
      },
      {
        id: 'wallet-liability',
        title: t('business.kpi.walletLiability'),
        value: formatUSD(summary.wallet_liability_usd),
        subtitle: t('business.kpi.walletLiabilitySub'),
        icon: Wallet,
        tone: 'chart-4',
      },
      {
        id: 'channel-promo',
        title: t('business.kpi.channelBalance'),
        value: formatUSD(summary.channel_balance_usd),
        subtitle: t('business.kpi.channelBalanceSub', {
          promo: formatUSD(summary.redemption_usd),
          count: summary.redemption_count,
        }),
        icon: PiggyBank,
        tone: 'chart-5',
      },
    ]
  }, [summary, t])

  const financialSpec = useMemo(
    () =>
      buildFinancialTrendSpec(dailyTrend, {
        revenue: t('business.chart.cashRevenue'),
        consumption: t('business.chart.consumption'),
        cost: t('business.chart.estimatedCost'),
      }),
    [dailyTrend, t]
  )

  const growthSpec = useMemo(
    () =>
      buildGrowthBarSpec(dailyTrend, {
        newUsers: t('business.chart.newUsers'),
        paidOrders: t('business.chart.paidOrders'),
      }),
    [dailyTrend, t]
  )

  const chartTheme = resolvedTheme === 'dark' ? 'dark' : 'light'

  if (isError) {
    return (
      <Alert variant='destructive'>
        <AlertTriangle className='size-4' />
        <AlertTitle>{t('Failed to load')}</AlertTitle>
        <AlertDescription>
          {error instanceof Error ? error.message : t('Please try again later.')}
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <div className='space-y-4'>
      {/* Range Toolbar & Accounting Basis Notice */}
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <div className='flex items-center gap-2'>
          <Tabs
            value={String(selectedDays)}
            onValueChange={(val) => setSelectedDays(Number(val))}
          >
            <TabsList>
              {BUSINESS_RANGE_OPTIONS.map((opt) => (
                <TabsTrigger
                  key={opt.days}
                  value={String(opt.days)}
                  className='px-2.5 text-xs'
                >
                  {t(opt.labelKey)}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
          {isFetching && (
            <Loader2 className='text-muted-foreground size-4 animate-spin' />
          )}
        </div>

        <Badge variant='outline' className='text-muted-foreground text-xs font-normal'>
          {t('business.accountingNotice')}
        </Badge>
      </div>

      {/* 8 Executive KPI Cards */}
      {isLoading || !summary ? (
        <div className='grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4'>
          {KPI_SKELETON_KEYS.map((key) => (
            <div key={key} className='bg-card rounded-xl border p-4'>
              <Skeleton className='h-4 w-28' />
              <Skeleton className='mt-3 h-7 w-24' />
              <Skeleton className='mt-2 h-3.5 w-40' />
            </div>
          ))}
        </div>
      ) : (
        <div
          className='grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4'
          data-testid='business-kpi-grid'
        >
          {kpiCards.map((card) => {
            const Icon = card.icon
            return (
              <div
                key={card.id}
                className='bg-card flex flex-col justify-between gap-2.5 rounded-xl border p-4 shadow-2xs'
              >
                <div className='flex items-center justify-between gap-2'>
                  <div className='flex min-w-0 items-center gap-2'>
                    <IconBadge tone={card.tone} size='sm'>
                      <Icon />
                    </IconBadge>
                    <span className='text-muted-foreground truncate text-xs font-medium'>
                      {card.title}
                    </span>
                  </div>
                  {card.badgeText && (
                    <Badge
                      variant={card.badgeVariant ?? 'secondary'}
                      className='font-mono text-[11px]'
                    >
                      {card.badgeText}
                    </Badge>
                  )}
                </div>

                <div>
                  <div className='text-foreground font-mono text-2xl font-semibold tracking-tight tabular-nums'>
                    {card.value}
                  </div>
                  <p className='text-muted-foreground mt-1 line-clamp-1 text-xs'>
                    {card.subtitle}
                  </p>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Dual Trend Charts */}
      <div className='grid grid-cols-1 gap-4 xl:grid-cols-2'>
        <div className='bg-card overflow-hidden rounded-xl border shadow-2xs'>
          <div className='flex items-center gap-2 border-b px-4 py-3'>
            <IconBadge tone='chart-1' size='sm'>
              <Coins />
            </IconBadge>
            <h3 className='text-sm font-semibold'>
              {t('business.section.financialTrend')}
            </h3>
          </div>
          <div className='h-72 p-2 sm:h-80'>
            {isLoading || !themeReady ? (
              <Skeleton className='h-full w-full' />
            ) : (
              <VChart
                key={`fin-${selectedDays}-${chartTheme}`}
                spec={{
                  ...financialSpec,
                  theme: chartTheme,
                  background: 'transparent',
                }}
                option={VCHART_OPTION}
              />
            )}
          </div>
        </div>

        <div className='bg-card overflow-hidden rounded-xl border shadow-2xs'>
          <div className='flex items-center gap-2 border-b px-4 py-3'>
            <IconBadge tone='success' size='sm'>
              <Users />
            </IconBadge>
            <h3 className='text-sm font-semibold'>
              {t('business.section.growthTrend')}
            </h3>
          </div>
          <div className='h-72 p-2 sm:h-80'>
            {isLoading || !themeReady ? (
              <Skeleton className='h-full w-full' />
            ) : (
              <VChart
                key={`growth-${selectedDays}-${chartTheme}`}
                spec={{
                  ...growthSpec,
                  theme: chartTheme,
                  background: 'transparent',
                }}
                option={VCHART_OPTION}
              />
            )}
          </div>
        </div>
      </div>

      {/* Dual Breakdown Tables: Payment Channels & Model Gross Margins */}
      <div className='grid grid-cols-1 gap-4 xl:grid-cols-2'>
        {/* Payment Channel Revenue Breakdown */}
        <div className='bg-card overflow-hidden rounded-xl border shadow-2xs'>
          <div className='flex items-center justify-between gap-2 border-b px-4 py-3'>
            <div className='flex items-center gap-2'>
              <IconBadge tone='info' size='sm'>
                <Landmark />
              </IconBadge>
              <h3 className='text-sm font-semibold'>
                {t('business.section.paymentChannels')}
              </h3>
            </div>
            {summary && summary.redemption_count > 0 && (
              <Badge variant='outline' className='text-xs font-normal'>
                {t('business.promoFooter', {
                  amount: formatUSD(summary.redemption_usd),
                  count: summary.redemption_count,
                })}
              </Badge>
            )}
          </div>

          <div className='max-h-96 overflow-auto'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('business.table.channel')}</TableHead>
                  <TableHead>{t('business.table.orderType')}</TableHead>
                  <TableHead className='text-right'>
                    {t('business.table.orders')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('business.table.originalAmount')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('business.table.usdRevenue')}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {paymentChannels.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={5}
                      className='text-muted-foreground h-28 text-center text-xs'
                    >
                      {t('No data available')}
                    </TableCell>
                  </TableRow>
                ) : (
                  paymentChannels.map((item) => (
                    <TableRow
                      key={`${item.order_type}-${item.channel}-${item.original_currency}`}
                    >
                      <TableCell className='font-medium'>
                        {item.channel}
                      </TableCell>
                      <TableCell>
                        <Badge variant='secondary' className='text-[11px]'>
                          {item.order_type === 'subscription'
                            ? t('Subscriptions')
                            : t('Top-up')}
                        </Badge>
                      </TableCell>
                      <TableCell className='text-right font-mono tabular-nums'>
                        {item.order_count}
                      </TableCell>
                      <TableCell className='text-muted-foreground text-right font-mono text-xs tabular-nums'>
                        {formatOriginalMoney(
                          item.original_amount,
                          item.original_currency
                        )}
                      </TableCell>
                      <TableCell className='text-right font-mono font-medium tabular-nums'>
                        {formatUSD(item.revenue_usd)}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>

        {/* Model & Channel Margin Analysis */}
        <div className='bg-card overflow-hidden rounded-xl border shadow-2xs'>
          <div className='flex items-center justify-between gap-2 border-b px-4 py-3'>
            <div className='flex items-center gap-2'>
              <IconBadge tone='warning' size='sm'>
                <Receipt />
              </IconBadge>
              <h3 className='text-sm font-semibold'>
                {t('business.section.modelMargins')}
              </h3>
            </div>
            <span className='text-muted-foreground text-xs'>
              {t('business.section.modelMarginsHint')}
            </span>
          </div>

          <div className='max-h-96 overflow-auto'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Model')}</TableHead>
                  <TableHead>{t('Channel')}</TableHead>
                  <TableHead className='text-right'>
                    {t('business.table.billedUSD')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('business.table.costUSD')}
                  </TableHead>
                  <TableHead className='text-right'>
                    {t('business.table.margin')}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {modelMargins.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={5}
                      className='text-muted-foreground h-28 text-center text-xs'
                    >
                      {t('No data available')}
                    </TableCell>
                  </TableRow>
                ) : (
                  modelMargins.slice(0, 30).map((row) => (
                    <TableRow key={`${row.channel_id}-${row.model_name}`}>
                      <TableCell>
                        <div className='flex max-w-44 items-center gap-1.5'>
                          <ModelProviderIcon
                            modelName={row.model_name}
                            size={15}
                          />
                          <span
                            className='truncate font-mono text-xs font-medium'
                            title={row.model_name}
                          >
                            {row.model_name}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className='text-muted-foreground max-w-32 truncate text-xs'>
                        {row.channel_name}
                      </TableCell>
                      <TableCell className='text-right font-mono text-xs tabular-nums'>
                        {formatUSD(row.consumption_usd)}
                      </TableCell>
                      <TableCell className='text-right font-mono text-xs tabular-nums'>
                        {row.cost_covered
                          ? formatUSD(row.estimated_cost_usd)
                          : t('business.uncosted')}
                      </TableCell>
                      <TableCell className='text-right'>
                        {row.cost_covered && row.gross_margin_pct != null ? (
                          <span
                            className={cn(
                              'inline-flex items-center rounded-md border px-1.5 py-0.5 font-mono text-[11px] font-medium tabular-nums',
                              getMarginBadgeClass(row.gross_margin_pct)
                            )}
                          >
                            {row.gross_margin_pct.toFixed(1)}%
                          </span>
                        ) : (
                          <span className='text-muted-foreground text-xs'>
                            -
                          </span>
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>
      </div>
    </div>
  )
}
