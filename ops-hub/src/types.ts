export interface KVStoreLike {
  get(key: string): Promise<string | null>
  put(key: string, value: string, options?: { expirationTtl?: number }): Promise<void>
  delete(key: string): Promise<void>
}

export interface Env {
  OPS_STATE: KVStoreLike
  OB_BASE_URL: string
  NETGO_OPS_URL: string
  DASHBOARD_PUBLIC_URL?: string
  DASHBOARD_PASSPHRASE?: string
  SESSION_SIGNING_SECRET?: string
  OB_OPS_TOKEN?: string
  NETGO_OPS_TOKEN?: string
  WECOM_CORP_ID?: string
  WECOM_AGENT_ID?: string
  WECOM_CORP_SECRET?: string
  WECOM_TO_USER?: string
  WECOM_WEBHOOK_URL?: string
  MODEL_ALERT_THRESHOLD?: string
  MIN_MODEL_SAMPLES?: string
  ALERT_COOLDOWN_SECONDS?: string
}

export interface ObPeriodKPI {
  cash_revenue_cny: number
  topup_revenue_cny: number
  subscription_revenue_cny: number
  consumption_cny: number
  estimated_cost_cny: number
  estimated_gross_profit_cny: number
  wallet_liability_cny: number
  estimated_gross_margin_pct: number | null
  paid_order_count: number
  paying_users: number
  redemption_count: number
  total_requests: number
  total_tokens: number
  total_users: number
  new_users: number
  active_users: number
}

export interface ObModelPerf {
  model_name: string
  real: {
    requests: number
    success: number
    avg_latency_ms: number
    avg_ttft_ms: number
    avg_tps: number
  }
  probe: {
    requests: number
    success: number
    avg_latency_ms: number
    avg_ttft_ms: number
  }
  client_error: {
    requests: number
  }
  public_v2?: {
    recent_success_series?: Array<{ ts: number; success_rate: number }>
  }
  availability_24h: number
  availability_7d: number
  error_codes: Record<string, number>
  last_probe_at: number
  probe_tier: string
}

export interface ObSnapshot {
  generated_at: number
  version: string
  uptime_seconds: number
  db_healthy: boolean
  pricing_control_enabled: boolean
  pending_pricing_proposals: number
  currency: {
    display_currency: string
    usd_exchange_rate: number
    price_ratio: number
  }
  today: ObPeriodKPI
  last_7d: ObPeriodKPI
  daily_trend_7d: Array<{
    date: string
    cash_revenue_cny: number
    consumption_cny: number
    gross_profit_cny: number
    paid_orders: number
    new_users: number
  }>
  channels: {
    total: number
    enabled: number
    auto_disabled: number
    manually_disabled: number
    low_balance_count: number
    total_balance_usd: number
  }
  models: {
    models: ObModelPerf[]
    probe_budget_used_quota: number
    probe_budget_quota: number
    generated_at: number
  }
}

export interface NetGoNodeItem {
  id: number
  name: string
  type: string
  visible: boolean
  rate: string
  online: boolean
  last_check_at: number
  heartbeat_age_seconds: number | null
}

export interface NetGoSnapshot {
  generated_at: number
  db_ok: boolean
  host: {
    overall: string
    timestamp_utc: string
    disk_percent: number | null
    memory_available_percent: number | null
    backup_age_seconds: number | null
    checks: Array<{ item: string; ok: boolean }>
    failures: string[]
  }
  today: {
    cash_revenue_cny: number
    paid_order_count: number
    paying_users: number
    new_users: number
  }
  last_7d: {
    cash_revenue_cny: number
    paid_order_count: number
  }
  users: {
    total_users: number
    active_subscribers: number
  }
  tickets: {
    open_tickets: number
    pending_reply_tickets: number
  }
  nodes_summary: {
    total: number
    visible: number
    online_visible: number
    offline_visible: number
  }
  nodes: NetGoNodeItem[]
}

export interface AlertEventRecord {
  ts: number
  level: 'critical' | 'warning' | 'recovery' | 'info'
  title: string
  summary: string
}

export interface CombinedHubPayload {
  fetched_at: number
  wecom_configured: boolean
  openbridger: {
    reachable: boolean
    error?: string
    data: ObSnapshot | null
  }
  netgo: {
    reachable: boolean
    error?: string
    data: NetGoSnapshot | null
  }
  recent_events: AlertEventRecord[]
}
