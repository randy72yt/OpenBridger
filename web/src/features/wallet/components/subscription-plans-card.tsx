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
import {
  Crown,
  RefreshCw,
  Sparkles,
  Check,
  Coins,
  BadgePercent,
  Wallet,
  ArrowRightLeft,
  CreditCard,
  ReceiptText,
} from "lucide-react";
import { useState, useEffect, useMemo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

import {
  StatusBadge,
  dotColorMap,
  textColorMap,
} from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { TitledCard } from "@/components/ui/titled-card";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  getPublicPlans,
  getSelfSubscriptionFull,
  updateBillingPreference,
} from "@/features/subscriptions/api";
import { SubscriptionPurchaseDialog } from "@/features/subscriptions/components/dialogs/subscription-purchase-dialog";
import {
  formatDuration,
  formatResetPeriod,
} from "@/features/subscriptions/lib";
import type {
  PlanRecord,
  UserSubscriptionRecord,
} from "@/features/subscriptions/types";
import { formatBillingCurrencyFromUSD } from "@/lib/currency";
import { formatQuota } from "@/lib/format";
import { cn } from "@/lib/utils";
import { useSystemConfig } from "@/hooks/use-system-config";

import {
  getPricing,
} from "@/features/pricing/api";
import {
  formatDynamicUnitPrice,
  getDynamicPricingSummary,
} from "@/features/pricing/lib/dynamic-price";
import type { PricingModel } from "@/features/pricing/types";
import { ChevronDown } from "lucide-react";

import type { PaymentMethod, TopupInfo } from "../types";

// Models whose upstream offers no channel discount: sold at cost + service
// fee (~official list price) instead of the 60%-of-list rule. Keep in sync
// with pricing plan §5 when the upstream offer set changes.
const CHANNEL_PRICE_MODELS = new Set([
  "glm-5.2",
  "glm-5.3",
  "deepseek-v4-pro",
  "deepseek-v4.1-flash",
  "kimi-k3",
  "gpt-6-luna",
]);

interface SubscriptionPlansCardProps {
  topupInfo: TopupInfo | null;
  onAvailabilityChange?: (available: boolean) => void;
  userQuota?: number;
  onPurchaseSuccess?: () => void | Promise<void>;
}

function getEpayMethods(payMethods: PaymentMethod[] = []): PaymentMethod[] {
  return payMethods.filter(
    (m) => m?.type && m.type !== "stripe" && m.type !== "creem",
  );
}

type RateCell = {
  label: string;
  price: number;
  currency: string | undefined;
  months: number;
};

type RateRow = {
  title: string;
  total: number;
  cells: RateCell[];
};

// Convert a plan price from its settlement currency to system-USD so it can
// be rendered in whatever display currency the site is configured for.
function planAmountToDisplayUSD(
  amount: number,
  planCurrency: string | undefined,
  usdExchangeRate: number,
): number {
  const rate = usdExchangeRate > 0 ? usdExchangeRate : 1;
  return (planCurrency || "USD").toUpperCase() === "CNY"
    ? amount / rate
    : amount;
}

function getBillingPreferenceLabel(
  preference: string,
  t: (key: string) => string,
): string {
  switch (preference) {
    case "subscription_first":
      return t("Subscription First");
    case "wallet_first":
      return t("Wallet First");
    case "subscription_only":
      return t("Subscription Only");
    case "wallet_only":
      return t("Wallet Only");
    default:
      return preference;
  }
}

export function SubscriptionPlansCard({
  topupInfo,
  onAvailabilityChange,
  userQuota,
  onPurchaseSuccess,
}: SubscriptionPlansCardProps) {
  const { t } = useTranslation();
  const { currency } = useSystemConfig();
  const usdExchangeRate =
    currency?.usdExchangeRate && currency.usdExchangeRate > 0
      ? currency.usdExchangeRate
      : 1;

  const [plans, setPlans] = useState<PlanRecord[]>([]);
  const [activeSubscriptions, setActiveSubscriptions] = useState<
    UserSubscriptionRecord[]
  >([]);
  const [allSubscriptions, setAllSubscriptions] = useState<
    UserSubscriptionRecord[]
  >([]);
  const [billingPreference, setBillingPreference] =
    useState("subscription_first");
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const [purchaseOpen, setPurchaseOpen] = useState(false);
  const [ratesOpen, setRatesOpen] = useState(false);
  const [pricingModels, setPricingModels] = useState<PricingModel[] | null>(
    null,
  );

  useEffect(() => {
    if (!ratesOpen || pricingModels) return;
    getPricing()
      .then((d) => setPricingModels(d?.data || []))
      .catch(() => setPricingModels([]));
  }, [ratesOpen, pricingModels]);

  // Per-model rate rows: our input/output price in USD per 1M tokens, plus
  // the official list price implied by the 60%-of-list rule.
  const modelRateRows = useMemo(() => {
    if (!pricingModels) return [];
    const rows = pricingModels
      .map((m) => {
        const isDynamic =
          m.billing_mode === "tiered_expr" && Boolean(m.billing_expr);
        let inputUsd = Number.NaN;
        let outputUsd = Number.NaN;
        if (isDynamic) {
          const summary = getDynamicPricingSummary(m, {
            tokenUnit: "1M",
            usdExchangeRate,
          });
          const entries = summary?.entries || [];
          const input = entries.find((e) => e.key === "p");
          const output = entries.find((e) => e.key === "c");
          inputUsd = input?.value ?? Number.NaN;
          outputUsd = output?.value ?? Number.NaN;
        } else if (Number(m.model_ratio) > 0) {
          inputUsd = Number(m.model_ratio) * 2;
          outputUsd = inputUsd * Number(m.completion_ratio || 1);
        }
        if (!Number.isFinite(inputUsd) || inputUsd <= 0) return null;
        return {
          name: m.model_name,
          inputUsd,
          outputUsd,
          channel: CHANNEL_PRICE_MODELS.has(m.model_name),
        };
      })
      .filter((r): r is NonNullable<typeof r> => !!r)
      .sort((a, b) => a.inputUsd - b.inputUsd);
    return rows;
  }, [pricingModels, usdExchangeRate]);
  const [selectedPlan, setSelectedPlan] = useState<PlanRecord | null>(null);
  const [periodKey, setPeriodKey] = useState("");

  const enableStripe = !!topupInfo?.enable_stripe_topup;
  const enableCreem = !!topupInfo?.enable_creem_topup;
  const enableWaffoPancake = !!topupInfo?.enable_waffo_pancake_topup;
  const enableOnlineTopUp = !!topupInfo?.enable_online_topup;
  const epayMethods = useMemo(
    () => getEpayMethods(topupInfo?.pay_methods),
    [topupInfo?.pay_methods],
  );

  const fetchPlans = useCallback(async () => {
    try {
      const res = await getPublicPlans();
      if (res.success) {
        setPlans(res.data || []);
      }
    } catch {
      setPlans([]);
    }
  }, []);

  const fetchSelfSubscription = useCallback(async () => {
    try {
      const res = await getSelfSubscriptionFull();
      if (res.success && res.data) {
        setBillingPreference(
          res.data.billing_preference || "subscription_first",
        );
        setActiveSubscriptions(res.data.subscriptions || []);
        setAllSubscriptions(res.data.all_subscriptions || []);
      }
    } catch {
      // ignore
    }
  }, []);

  useEffect(() => {
    const init = async () => {
      setLoading(true);
      await Promise.all([fetchPlans(), fetchSelfSubscription()]);
      setLoading(false);
    };
    init();
  }, [fetchPlans, fetchSelfSubscription]);

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await fetchSelfSubscription();
    } finally {
      setRefreshing(false);
    }
  };

  const handlePreferenceChange = async (pref: string) => {
    const previous = billingPreference;
    setBillingPreference(pref);
    try {
      const res = await updateBillingPreference(pref);
      if (res.success) {
        toast.success(t("Updated successfully"));
        const normalized = res.data?.billing_preference || pref;
        setBillingPreference(normalized);
      } else {
        toast.error(res.message || t("Update failed"));
        setBillingPreference(previous);
      }
    } catch {
      toast.error(t("Request failed"));
      setBillingPreference(previous);
    }
  };

  const hasActive = activeSubscriptions.length > 0;
  const hasAny = allSubscriptions.length > 0;
  const isAvailable = loading || plans.length > 0 || hasAny;
  const disablePref = !hasActive;
  const isSubPref =
    billingPreference === "subscription_first" ||
    billingPreference === "subscription_only";

  const planPurchaseCountMap = useMemo(() => {
    const map = new Map<number, number>();
    for (const sub of allSubscriptions) {
      const planId = sub?.subscription?.plan_id;
      if (!planId) continue;
      map.set(planId, (map.get(planId) || 0) + 1);
    }
    return map;
  }, [allSubscriptions]);

  useEffect(() => {
    onAvailabilityChange?.(isAvailable);
  }, [isAvailable, onAvailabilityChange]);

  const planTitleMap = useMemo(() => {
    const map = new Map<number, string>();
    for (const p of plans) {
      if (p?.plan?.id) {
        map.set(p.plan.id, p.plan.title || "");
      }
    }
    return map;
  }, [plans]);

  // Match the same tier across billing periods (e.g. 体验月付 vs 体验半年付).
  // Titles may carry period suffixes and quota totals differ when reset
  // periods differ, so pair by tier rank: position in sort_order within each
  // period group.
  const matchTier = useCallback(
    (candidates: PlanRecord[], p: PlanRecord): PlanRecord | undefined => {
      // Rank within the plan's own billing-period group, then pick the same
      // rank from the candidates group (each sorted by sort_order).
      const groupKey = (item: PlanRecord) =>
        `${item?.plan?.duration_unit || "month"}:${item?.plan?.duration_value || 1}`;
      const ownGroup = plans.filter((m) => groupKey(m) === groupKey(p));
      const wanted = [...ownGroup]
        .sort(
          (a, b) =>
            Number(a?.plan?.sort_order || 0) - Number(b?.plan?.sort_order || 0),
        )
        .findIndex((m) => m?.plan?.id === p?.plan?.id);
      if (wanted < 0) return undefined;
      const sorted = [...candidates].sort(
        (a, b) =>
          Number(a?.plan?.sort_order || 0) - Number(b?.plan?.sort_order || 0),
      );
      return sorted[wanted];
    },
    [plans],
  );

  // Group plans by billing period (e.g. monthly vs semi-annual) so each
  // cycle is shown on its own tab, mirroring the plan list layout.
  const periodGroups = useMemo(() => {
    const map = new Map<string, PlanRecord[]>();
    for (const p of plans) {
      const plan = p?.plan;
      if (!plan) continue;
      const unit = plan.duration_unit || "month";
      const value = Number(plan.duration_value || 1);
      const key = `${unit}:${value}`;
      const list = map.get(key) || [];
      list.push(p);
      map.set(key, list);
    }
    const monthlyPlans = map.get("month:1") || [];
    const groups = Array.from(map.entries())
      .sort(([a], [b]) =>
        a === "month:1" ? -1 : b === "month:1" ? 1 : a.localeCompare(b),
      )
      .map(([key, groupPlans]) => {
        const [unit, valueStr] = key.split(":");
        const value = Number(valueStr);
        let label = `${value} ${unit}`;
        if (unit === "month") {
          if (value === 1) label = t("Monthly");
          else if (value === 3) label = t("Quarterly");
          else if (value === 6) label = t("Semi-Annually");
          else if (value === 12) label = t("Yearly");
        } else if (unit === "year" && value === 1) {
          label = t("Yearly");
        }
        // Savings vs paying monthly: match tiers by identical total quota.
        let discountPercent: number | null = null;
        if (unit === "month" && value > 1 && monthlyPlans.length > 0) {
          for (const p of groupPlans) {
            const match = matchTier(monthlyPlans, p);
            if (!match) continue;
            const full = Number(match.plan.price_amount || 0) * value;
            const price = Number(p.plan.price_amount || 0);
            if (full > 0 && price > 0 && price < full) {
              const pct = Math.round((1 - price / full) * 100);
              if (discountPercent === null || pct > discountPercent) {
                discountPercent = pct;
              }
            }
          }
        }
        return { key, label, discountPercent, groupPlans };
      });
    return groups;
  }, [plans, matchTier, t]);

  const activePeriodKey = periodGroups.some((g) => g.key === periodKey)
    ? periodKey
    : periodGroups[0]?.key || "";

  // Rows for the rate details table: one per tier (matched across billing
  // periods by identical total quota), with each period's price.
  const rateRows = useMemo(() => {
    if (periodGroups.length === 0) return [];
    const monthlyGroup = periodGroups.find((g) => g.key === "month:1");
    const monthly = monthlyGroup?.groupPlans || [];
    const others = periodGroups.filter((g) => g.key !== "month:1");
    const base = monthly.length > 0 ? monthly : periodGroups[0].groupPlans;
    const rows = base.map((m) => {
      const plan = m?.plan;
      if (!plan) return null;
      const total = Number(plan.total_amount || 0);
      const cells = [
        {
          label: monthlyGroup ? monthlyGroup.label : periodGroups[0].label,
          price: Number(plan.price_amount || 0),
          currency: plan.currency,
          months: 1,
        },
      ];
      for (const g of others) {
        const match = matchTier(g.groupPlans, m);
        if (!match) continue;
        const [unit, valueStr] = g.key.split(":");
        const months = Number(valueStr || 1) * (unit === "year" ? 12 : 1);
        cells.push({
          label: g.label,
          price: Number(match.plan.price_amount || 0),
          currency: match.plan.currency,
          months,
        });
      }
      return { title: plan.title || "", total, cells };
    });
    return rows.filter((r): r is RateRow => !!r && r.total > 0);
  }, [periodGroups]);

  const getRemainingDays = (sub: UserSubscriptionRecord) => {
    const endTime = sub?.subscription?.end_time || 0;
    if (!endTime) return 0;
    const now = Date.now() / 1000;
    return Math.max(0, Math.ceil((endTime - now) / 86400));
  };

  const getUsagePercent = (sub: UserSubscriptionRecord) => {
    const total = Number(sub?.subscription?.amount_total || 0);
    const used = Number(sub?.subscription?.amount_used || 0);
    if (total <= 0) return 0;
    return Math.round((used / total) * 100);
  };

  const renderPlansGrid = (inputPlans: PlanRecord[]) => {
    // Cheapest plan first within each billing period.
    const gridPlans = [...inputPlans].sort(
      (a, b) =>
        Number(a?.plan?.price_amount || 0) - Number(b?.plan?.price_amount || 0),
    );
    // One row of up to four equal-height plans on wide screens.
    return (
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {gridPlans.map((p, index) => {
          const plan = p?.plan;
          if (!plan) return null;
          const totalAmount = Number(plan.total_amount || 0);
          const displayPrice = formatBillingCurrencyFromUSD(
            planAmountToDisplayUSD(
              Number(plan.price_amount || 0),
              plan.currency,
              usdExchangeRate,
            ),
          );
          const isMonthly =
            (plan.duration_unit || "month") === "month" &&
            Number(plan.duration_value || 1) === 1;
          const isPopular = index === 1 && gridPlans.length > 1;
          const limit = Number(plan.max_purchase_per_user || 0);
          const count = planPurchaseCountMap.get(plan.id) || 0;
          const reached = limit > 0 && count >= limit;

          const isWeeklyReset =
            (plan.quota_reset_period || "never") === "weekly";
          const benefits = [
            totalAmount > 0
              ? `${t(isWeeklyReset ? "Weekly quota" : "Monthly quota")}: ${formatQuota(totalAmount)}`
              : `${t("Total Quota")}: ${t("Unlimited")}`,
            `${t("Validity Period")}: ${formatDuration(plan, t)}`,
            formatResetPeriod(plan, t) !== t("No Reset")
              ? `${t("Quota Reset")}: ${formatResetPeriod(plan, t)}`
              : null,
            t("Covers all models at wallet rates"),
            t("Falls back to wallet balance when quota runs out"),
            limit > 0 ? `${t("Purchase Limit")}: ${limit}` : null,
            plan.upgrade_group
              ? `${t("Upgrade Group")}: ${plan.upgrade_group}`
              : null,
          ].filter(Boolean) as string[];

          return (
            <Card
              key={plan.id}
              data-card-hover="false"
              className={cn(
                "relative",
                isPopular && "border-primary/70 shadow-sm",
              )}
            >
              {isPopular && (
                <div className="pointer-events-none absolute -top-3 left-4 z-10">
                  <StatusBadge variant="info" copyable={false}>
                    <Sparkles className="h-3 w-3" />
                    {t("Recommended")}
                  </StatusBadge>
                </div>
              )}
              <CardContent className="flex h-full flex-col p-4 sm:p-5">

                <div className="mb-2 min-w-0">
                  <h4 className="truncate font-semibold">
                    {plan.title || t("Subscription Plans")}
                  </h4>
                  {plan.subtitle && (
                    <p className="text-muted-foreground line-clamp-2 text-xs leading-relaxed">
                      {plan.subtitle}
                    </p>
                  )}
                </div>

                <div className="flex items-baseline gap-1 py-2">
                  <span className="text-primary text-2xl font-bold">
                    {displayPrice}
                  </span>
                  {isMonthly && (
                    <span className="text-muted-foreground text-xs">
                      {t("per month")}
                    </span>
                  )}
                </div>

                <div className="flex-1 space-y-1.5 pb-3">
                  {benefits.map((label) => (
                    <div
                      key={label}
                      className="text-muted-foreground flex items-center gap-2 text-xs leading-relaxed"
                    >
                      <Check className="text-primary h-3 w-3 shrink-0" />
                      <span>{label}</span>
                    </div>
                  ))}
                </div>

                <Separator className="mb-3" />

                {reached ? (
                  <Tooltip>
                    <TooltipTrigger render={<div />}>
                      <Button variant="outline" className="w-full" disabled>
                        {t("Limit Reached")}
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                      {t("Purchase limit reached")} ({count}/{limit})
                    </TooltipContent>
                  </Tooltip>
                ) : (
                  <Button
                    variant="outline"
                    className="w-full"
                    onClick={() => {
                      setSelectedPlan(p);
                      setPurchaseOpen(true);
                    }}
                  >
                    {t("Subscribe Now")}
                  </Button>
                )}
              </CardContent>
            </Card>
          );
        })}
      </div>
    );
  };

  if (loading) {
    return (
      <Card data-card-hover="false" className="gap-0 overflow-hidden py-0">
        <CardHeader className="border-b p-3 !pb-3 sm:p-5 sm:!pb-5">
          <Skeleton className="h-6 w-32" />
        </CardHeader>
        <CardContent className="space-y-4 p-3 sm:p-5">
          <Skeleton className="h-20 w-full" />
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {["first", "second", "third", "fourth"].map((key) => (
              <Skeleton key={key} className="h-48 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (plans.length === 0 && !hasAny) {
    return null;
  }

  return (
    <>
      <TitledCard
        title={t("Subscription Plans")}
        description={t("Subscribe to a plan for model access")}
        icon={<Crown className="h-4 w-4" />}
        iconTone="warning"
        disableHoverEffect
        contentClassName="space-y-4 sm:space-y-5"
      >
        {/* My subscriptions & billing preference */}
        <div className="rounded-xl border p-3 sm:p-4">
          <div className="flex flex-wrap items-center justify-between gap-2.5 sm:gap-3">
            <div className="flex min-w-0 flex-wrap items-center gap-2">
              <span className="text-sm font-medium">
                {t("My Subscriptions")}
              </span>
              <span className="flex items-center gap-1.5 text-xs font-medium">
                <span
                  className={cn(
                    "size-1.5 shrink-0 rounded-full",
                    hasActive ? dotColorMap.success : dotColorMap.neutral,
                  )}
                  aria-hidden="true"
                />
                {hasActive ? (
                  <span className={cn(textColorMap.success)}>
                    {activeSubscriptions.length} {t("active")}
                  </span>
                ) : (
                  <span className="text-muted-foreground">
                    {t("No Active")}
                  </span>
                )}
                {allSubscriptions.length > activeSubscriptions.length && (
                  <>
                    <span className="text-muted-foreground/30">·</span>
                    <span className="text-muted-foreground">
                      {allSubscriptions.length - activeSubscriptions.length}{" "}
                      {t("expired")}
                    </span>
                  </>
                )}
              </span>
            </div>
            <div className="flex w-full items-center gap-2 sm:w-auto">
              <Select
                items={[
                  {
                    value: "subscription_first",
                    label: (
                      <>
                        {getBillingPreferenceLabel("subscription_first", t)}
                        {disablePref ? ` (${t("No Active")})` : ""}
                      </>
                    ),
                  },
                  {
                    value: "wallet_first",
                    label: getBillingPreferenceLabel("wallet_first", t),
                  },
                  {
                    value: "subscription_only",
                    label: (
                      <>
                        {getBillingPreferenceLabel("subscription_only", t)}
                        {disablePref ? ` (${t("No Active")})` : ""}
                      </>
                    ),
                  },
                  {
                    value: "wallet_only",
                    label: getBillingPreferenceLabel("wallet_only", t),
                  },
                ]}
                value={billingPreference}
                onValueChange={(v) => v !== null && handlePreferenceChange(v)}
              >
                <SelectTrigger className="h-8 flex-1 text-xs sm:w-[140px] sm:flex-none">
                  <SelectValue>
                    {getBillingPreferenceLabel(billingPreference, t)}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent alignItemWithTrigger={false}>
                  <SelectGroup>
                    <SelectItem
                      value="subscription_first"
                      disabled={disablePref}
                    >
                      {getBillingPreferenceLabel("subscription_first", t)}
                      {disablePref ? ` (${t("No Active")})` : ""}
                    </SelectItem>
                    <SelectItem value="wallet_first">
                      {getBillingPreferenceLabel("wallet_first", t)}
                    </SelectItem>
                    <SelectItem
                      value="subscription_only"
                      disabled={disablePref}
                    >
                      {getBillingPreferenceLabel("subscription_only", t)}
                      {disablePref ? ` (${t("No Active")})` : ""}
                    </SelectItem>
                    <SelectItem value="wallet_only">
                      {getBillingPreferenceLabel("wallet_only", t)}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={handleRefresh}
                disabled={refreshing}
              >
                <RefreshCw
                  className={`h-3.5 w-3.5 ${refreshing ? "animate-spin" : ""}`}
                />
              </Button>
            </div>
          </div>

          {disablePref && isSubPref && (
            <p className="text-muted-foreground mt-2 text-xs">
              {billingPreference === "subscription_only"
                ? t(
                    "Preference saved as {{pref}}, but no active subscription. Requests will be rejected.",
                    { pref: t("Subscription Only") },
                  )
                : t(
                    "Preference saved as {{pref}}, but no active subscription. Wallet will be used automatically.",
                    { pref: t("Subscription First") },
                  )}
            </p>
          )}

          {hasAny && (
            <>
              <Separator className="my-3" />
              <div className="max-h-64 space-y-3 overflow-y-auto pr-1">
                {allSubscriptions.map((sub) => {
                  const subscription = sub.subscription;
                  const totalAmount = Number(subscription?.amount_total || 0);
                  const usedAmount = Number(subscription?.amount_used || 0);
                  const remainAmount =
                    totalAmount > 0 ? Math.max(0, totalAmount - usedAmount) : 0;
                  const planTitle =
                    planTitleMap.get(subscription?.plan_id) || "";
                  const remainDays = getRemainingDays(sub);
                  const usagePercent = getUsagePercent(sub);
                  const now = Date.now() / 1000;
                  const isExpired = (subscription?.end_time || 0) < now;
                  const isCancelled = subscription?.status === "cancelled";
                  const isActive =
                    subscription?.status === "active" && !isExpired;
                  const nextResetTime = subscription?.next_reset_time ?? 0;
                  let statusBadge = (
                    <StatusBadge
                      label={t("Expired")}
                      variant="neutral"
                      copyable={false}
                    />
                  );
                  if (isActive) {
                    statusBadge = (
                      <StatusBadge
                        label={t("Active")}
                        variant="success"
                        copyable={false}
                      />
                    );
                  } else if (isCancelled) {
                    statusBadge = (
                      <StatusBadge
                        label={t("Cancelled")}
                        variant="neutral"
                        copyable={false}
                      />
                    );
                  }

                  let endTimeLabel = t("Expired at");
                  if (isActive) {
                    endTimeLabel = t("Until");
                  } else if (isCancelled) {
                    endTimeLabel = t("Cancelled at");
                  }

                  return (
                    <div
                      key={subscription?.id}
                      className="bg-background rounded-md border p-3 text-xs"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">
                            {planTitle
                              ? `${planTitle} · ${t("Subscription")} #${subscription?.id}`
                              : `${t("Subscription")} #${subscription?.id}`}
                          </span>
                          {statusBadge}
                        </div>
                        {isActive && (
                          <span className="text-muted-foreground">
                            {t("{{count}} days remaining", {
                              count: remainDays,
                            })}
                          </span>
                        )}
                      </div>
                      <div className="text-muted-foreground mt-1.5">
                        {endTimeLabel}{" "}
                        {new Date(
                          (subscription?.end_time || 0) * 1000,
                        ).toLocaleString()}
                      </div>
                      {isActive && nextResetTime > 0 && (
                        <div className="text-muted-foreground mt-1">
                          {t("Next reset")}:{" "}
                          {new Date(nextResetTime * 1000).toLocaleString()}
                        </div>
                      )}
                      <div className="text-muted-foreground mt-1">
                        {t("Total Quota")}:{" "}
                        {totalAmount > 0 ? (
                          <Tooltip>
                            <TooltipTrigger
                              render={<span className="cursor-help" />}
                            >
                              {formatQuota(usedAmount)}/
                              {formatQuota(totalAmount)} · {t("Remaining")}{" "}
                              {formatQuota(remainAmount)}
                            </TooltipTrigger>
                            <TooltipContent>
                              {t("Raw Quota")}: {usedAmount}/{totalAmount} ·{" "}
                              {t("Remaining")} {remainAmount}
                            </TooltipContent>
                          </Tooltip>
                        ) : (
                          t("Unlimited")
                        )}
                        {totalAmount > 0 && (
                          <span className="ml-2">
                            {t("Used")} {usagePercent}%
                          </span>
                        )}
                      </div>
                      {totalAmount > 0 && isActive && (
                        <Progress value={usagePercent} className="mt-2 h-1.5" />
                      )}
                    </div>
                  );
                })}
              </div>
            </>
          )}

          {!hasAny && (
            <p className="text-muted-foreground mt-2 text-xs">
              {t("Subscribe to a plan for model access")}
            </p>
          )}
        </div>

        {/* Available plans grid, grouped by billing period */}
        {plans.length > 0 ? (
          <div className="space-y-4">
            {periodGroups.length > 1 && (
              <Tabs value={activePeriodKey} onValueChange={setPeriodKey}>
                <TabsList>
                  {periodGroups.map((g) => (
                    <TabsTrigger key={g.key} value={g.key}>
                      {g.label}
                      {g.discountPercent !== null && (
                        <span className="text-xs font-semibold text-red-500">
                          -{g.discountPercent}%
                        </span>
                      )}
                    </TabsTrigger>
                  ))}
                </TabsList>
                {periodGroups.map((g) => (
                  <TabsContent key={g.key} value={g.key} className="mt-3">
                    {renderPlansGrid(g.groupPlans)}
                  </TabsContent>
                ))}
              </Tabs>
            )}
            {periodGroups.length <= 1 && renderPlansGrid(plans)}

            {/* Rate details */}
            <div className="rounded-xl border bg-muted/40 p-3 sm:p-4">
              <p className="text-sm font-medium">{t("Rate Details")}</p>
              <p className="text-muted-foreground mt-1 text-xs leading-relaxed">
                {t(
                  "Subscription quota covers all models on this site at the same per-token rates as wallet balance. See the pricing page for model-specific rates.",
                )}
              </p>
              {rateRows.length > 0 && (
                <div className="mt-3 overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead className="text-muted-foreground">
                      <tr className="border-b">
                        <th className="py-1.5 pr-3 font-medium">
                          {t("Plan Name")}
                        </th>
                        <th className="py-1.5 pr-3 font-medium">
                          {t("Monthly quota")}
                        </th>
                        <th className="py-1.5 pr-3 font-medium">
                          {t("≈ Codex API value")}
                        </th>
                        {rateRows[0].cells.map((c) => (
                          <th key={c.label} className="py-1.5 pr-3 font-medium">
                            {c.label}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {rateRows.map((row) => (
                        <tr key={row.title} className="border-b last:border-0">
                          <td className="py-1.5 pr-3 font-medium">
                            {row.title}
                          </td>
                          <td className="py-1.5 pr-3">
                            {formatQuota(row.total)}
                          </td>
                          <td className="py-1.5 pr-3">
                            {t("≈ {{price}} / month", {
                              price: formatBillingCurrencyFromUSD(
                                planAmountToDisplayUSD(
                                  row.total / 500000 / 0.6,
                                  "USD",
                                  usdExchangeRate,
                                ),
                              ),
                            })}
                          </td>
                          {row.cells.map((c) => (
                            <td key={c.label} className="py-1.5 pr-3">
                              {formatBillingCurrencyFromUSD(
                                planAmountToDisplayUSD(
                                  c.price,
                                  c.currency,
                                  usdExchangeRate,
                                ),
                              )}
                              {c.months > 1 && (
                                <span className="text-muted-foreground ml-1">
                                  {t("≈ {{price}} / month", {
                                    price: formatBillingCurrencyFromUSD(
                                      planAmountToDisplayUSD(
                                        c.price / c.months,
                                        c.currency,
                                        usdExchangeRate,
                                      ),
                                    ),
                                  })}
                                </span>
                              )}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

            {/* How credits work */}
            <div className="rounded-xl border p-3 sm:p-4">
              <p className="text-sm font-medium">{t("How credits work")}</p>
              <div className="text-muted-foreground mt-1.5 space-y-1 text-xs leading-relaxed">
                <p>{t("1 credit = ¥0.01. Credits are billed per token and work on every model — no per-model multipliers.")}</p>
                <p>{t("Monthly plans come with credits for the whole month; semi-annual plans release a fresh batch of credits every week. Base-pool models are available on every plan; the full model pool unlocks from the All-You-Can-Eat tier up.")}</p>
                <p>{t("When plan credits run out, usage continues on your wallet balance automatically — no interruption, no forced renewal.")}</p>
              </div>
            </div>

            {/* Features & services */}
            <div className="rounded-xl border p-3 sm:p-4">
              <p className="text-sm font-medium">{t("Features & Services")}</p>
              <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {[
                  {
                    icon: Coins,
                    title: t("All models, one subscription"),
                    desc: t(
                      "One plan unlocks every model on the site through a single API key.",
                    ),
                  },
                  {
                    icon: BadgePercent,
                    title: t("60% of official list price"),
                    desc: t(
                      "Standard-pool models bill at 60% of official list price. No hidden group multipliers.",
                    ),
                  },
                  {
                    icon: Wallet,
                    title: t("Top up 1:1, never expires"),
                    desc: t(
                      "¥10 buys 1,000 credits. Wallet balance never expires and is never devalued.",
                    ),
                  },
                  {
                    icon: ArrowRightLeft,
                    title: t("Seamless wallet fallback"),
                    desc: t(
                      "Plan credits first, wallet balance second — switching is automatic and instant.",
                    ),
                  },
                  {
                    icon: CreditCard,
                    title: t("Alipay & WeChat Pay"),
                    desc: t(
                      "Pay in seconds, credits arrive instantly and stay in your account.",
                    ),
                  },
                  {
                    icon: ReceiptText,
                    title: t("Transparent per-call billing"),
                    desc: t(
                      "Every request shows its token usage and cost. No minimum spend.",
                    ),
                  },
                ].map(({ icon: Icon, title, desc }) => (
                  <div key={title} className="flex items-start gap-2.5">
                    <Icon className="text-primary mt-0.5 h-4 w-4 shrink-0" />
                    <div className="min-w-0">
                      <p className="text-xs font-medium">{title}</p>
                      <p className="text-muted-foreground mt-0.5 text-xs leading-relaxed">
                        {desc}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Full model rate list */}
            <div className="rounded-xl border p-3 sm:p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm font-medium">
                  {t("Full model rate list")}
                </p>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 text-xs"
                  onClick={() => setRatesOpen((v) => !v)}
                >
                  {ratesOpen
                    ? t("Collapse rate list")
                    : t("Expand rate list", {
                        count: modelRateRows.length || (pricingModels ? 0 : 49),
                      })}
                  <ChevronDown
                    className={`h-3.5 w-3.5 transition-transform ${ratesOpen ? "rotate-180" : ""}`}
                  />
                </Button>
              </div>
              <p className="text-muted-foreground mt-1 text-xs leading-relaxed">
                {t(
                  "Every standard-pool model bills at 60% of its official list price — the struck-through number is what you would pay the official API, the bold number is what you pay here. Models marked with ◆ have no upstream channel discount and are offered at cost + service fee.",
                )}
              </p>
              {ratesOpen && (
                <div className="mt-3 max-h-72 overflow-y-auto">
                  {pricingModels === null ? (
                    <Skeleton className="h-40 w-full" />
                  ) : (
                    <table className="w-full text-left text-xs">
                      <thead className="text-muted-foreground sticky top-0 bg-background">
                        <tr className="border-b">
                          <th className="py-1.5 pr-3 font-medium">
                            {t("Model")}
                          </th>
                          <th className="py-1.5 pr-3 font-medium">
                            {t("Input")} / {t("per 1M tokens")}
                          </th>
                          <th className="py-1.5 pr-3 font-medium">
                            {t("Output")} / {t("per 1M tokens")}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {modelRateRows.map((row) => {
                          const priceOpts = {
                            tokenUnit: "1M" as const,
                            usdExchangeRate,
                          };
                          return (
                            <tr
                              key={row.name}
                              className="border-b last:border-0"
                            >
                              <td className="py-1.5 pr-3 font-medium">
                                {row.channel && (
                                  <span className="text-warning mr-1">◆</span>
                                )}
                                {row.name}
                              </td>
                              <td className="py-1.5 pr-3">
                                {!row.channel && (
                                  <span className="text-muted-foreground mr-1.5 line-through">
                                    {formatDynamicUnitPrice(
                                      row.inputUsd / 0.6,
                                      priceOpts,
                                    )}
                                  </span>
                                )}
                                <span className="font-semibold">
                                  {formatDynamicUnitPrice(row.inputUsd, priceOpts)}
                                </span>
                              </td>
                              <td className="py-1.5 pr-3">
                                {Number.isFinite(row.outputUsd) &&
                                row.outputUsd > 0
                                  ? formatDynamicUnitPrice(row.outputUsd, priceOpts)
                                  : "-"}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  )}
                </div>
              )}
            </div>
          </div>
        ) : (
          <p className="text-muted-foreground py-4 text-center text-sm">
            {t("No plans available")}
          </p>
        )}
      </TitledCard>

      <SubscriptionPurchaseDialog
        open={purchaseOpen}
        onOpenChange={(open) => {
          setPurchaseOpen(open);
          if (!open) {
            fetchSelfSubscription();
          }
        }}
        plan={selectedPlan}
        enableStripe={enableStripe}
        enableCreem={enableCreem}
        enableWaffoPancake={enableWaffoPancake}
        enableOnlineTopUp={enableOnlineTopUp}
        epayMethods={epayMethods}
        userQuota={userQuota}
        onPurchaseSuccess={onPurchaseSuccess}
        purchaseLimit={
          selectedPlan?.plan?.max_purchase_per_user
            ? Number(selectedPlan.plan.max_purchase_per_user)
            : undefined
        }
        purchaseCount={
          selectedPlan?.plan?.id
            ? planPurchaseCountMap.get(selectedPlan.plan.id)
            : undefined
        }
      />
    </>
  );
}
