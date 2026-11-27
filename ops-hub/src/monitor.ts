import type {
  AlertEventRecord,
  CombinedHubPayload,
  Env,
  NetGoSnapshot,
  ObSnapshot,
} from './types'
import { isWeComConfigured, sendWeComCard } from './wecom'

const RECENT_EVENTS_KEY = 'alerts:recent_events'
const STATE_PREFIX = 'alerts:state:'
const MAX_RECENT_EVENTS = 30

interface StoredAlertState {
  active: boolean
  lastSentAt: number
  detail: string
}

function formatBeijingTime(epochSec: number): string {
  const d = new Date((epochSec + 8 * 3600) * 1000)
  const mm = String(d.getUTCMonth() + 1).padStart(2, '0')
  const dd = String(d.getUTCDate()).padStart(2, '0')
  const hh = String(d.getUTCHours()).padStart(2, '0')
  const mi = String(d.getUTCMinutes()).padStart(2, '0')
  return `${mm}-${dd} ${hh}:${mi}`
}

function fmtMoney(val: number | undefined | null): string {
  if (typeof val !== 'number' || !Number.isFinite(val)) return '0.00'
  return val.toFixed(2)
}

async function fetchOpenBridger(env: Env): Promise<{ reachable: boolean; error?: string; data: ObSnapshot | null }> {
  const baseUrl = (env.OB_BASE_URL || 'https://openbridger.com').replace(/\/+$/, '')
  const token = (env.OB_OPS_TOKEN || '').trim()
  if (!token) {
    return { reachable: false, error: '未配置 OB_OPS_TOKEN', data: null }
  }

  try {
    const resp = await fetch(`${baseUrl}/api/ops/snapshot`, {
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: 'application/json',
      },
      signal: AbortSignal.timeout(8500),
    })
    if (!resp.ok) {
      return { reachable: false, error: `HTTP ${resp.status}`, data: null }
    }
    const body = (await resp.json()) as { success?: boolean; data?: ObSnapshot }
    if (!body.success || !body.data) {
      return { reachable: false, error: '响应格式异常', data: null }
    }
    return { reachable: true, data: body.data }
  } catch (err) {
    return {
      reachable: false,
      error: err instanceof Error ? err.message : '连接超时或失败',
      data: null,
    }
  }
}

async function fetchNetGo(env: Env): Promise<{ reachable: boolean; error?: string; data: NetGoSnapshot | null }> {
  const url = (env.NETGO_OPS_URL || '').trim()
  const token = (env.NETGO_OPS_TOKEN || '').trim()
  if (!url || !token) {
    return { reachable: false, error: '未配置 NETGO_OPS_URL / NETGO_OPS_TOKEN', data: null }
  }

  try {
    const resp = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: 'application/json',
      },
      signal: AbortSignal.timeout(8500),
    })
    if (!resp.ok) {
      return { reachable: false, error: `HTTP ${resp.status}`, data: null }
    }
    const body = (await resp.json()) as { success?: boolean; data?: NetGoSnapshot }
    if (!body.success || !body.data) {
      return { reachable: false, error: '响应格式异常', data: null }
    }
    return { reachable: true, data: body.data }
  } catch (err) {
    return {
      reachable: false,
      error: err instanceof Error ? err.message : '连接超时或失败',
      data: null,
    }
  }
}

export async function getRecentEvents(env: Env): Promise<AlertEventRecord[]> {
  const raw = await env.OPS_STATE.get(RECENT_EVENTS_KEY)
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

async function appendRecentEvent(env: Env, ev: AlertEventRecord): Promise<void> {
  const list = await getRecentEvents(env)
  list.unshift(ev)
  if (list.length > MAX_RECENT_EVENTS) {
    list.length = MAX_RECENT_EVENTS
  }
  await env.OPS_STATE.put(RECENT_EVENTS_KEY, JSON.stringify(list))
}

export async function collectCombinedHubPayload(env: Env): Promise<CombinedHubPayload> {
  const [openbridger, netgo, recent_events] = await Promise.all([
    fetchOpenBridger(env),
    fetchNetGo(env),
    getRecentEvents(env),
  ])
  return {
    fetched_at: Math.floor(Date.now() / 1000),
    wecom_configured: isWeComConfigured(env),
    openbridger,
    netgo,
    recent_events,
  }
}

async function evaluateStatefulCheck(
  env: Env,
  checkId: string,
  isBad: boolean,
  alertTitle: string,
  alertDetail: string,
  recoveryTitle: string,
  nowSec: number,
  cooldownSec: number
): Promise<void> {
  const stateKey = `${STATE_PREFIX}${checkId}`
  const prevRaw = await env.OPS_STATE.get(stateKey)
  let prev: StoredAlertState = { active: false, lastSentAt: 0, detail: '' }
  if (prevRaw) {
    try {
      prev = JSON.parse(prevRaw) as StoredAlertState
    } catch {
      // ignore corrupted state
    }
  }

  if (isBad) {
    const shouldNotify = !prev.active || nowSec - prev.lastSentAt >= cooldownSec
    if (shouldNotify) {
      await appendRecentEvent(env, {
        ts: nowSec,
        level: 'critical',
        title: alertTitle,
        summary: alertDetail,
      })
      if (isWeComConfigured(env)) {
        await sendWeComCard(env, {
          title: alertTitle,
          description: `<div class="gray">${formatBeijingTime(nowSec)} 北京时间</div><div class="highlight">${alertDetail}</div>`,
        })
      }
      await env.OPS_STATE.put(
        stateKey,
        JSON.stringify({ active: true, lastSentAt: nowSec, detail: alertDetail })
      )
    }
    return
  }

  if (prev.active) {
    const summary = `故障已解除（原告警：${prev.detail}）`
    await appendRecentEvent(env, {
      ts: nowSec,
      level: 'recovery',
      title: recoveryTitle,
      summary,
    })
    if (isWeComConfigured(env)) {
      await sendWeComCard(env, {
        title: recoveryTitle,
        description: `<div class="gray">${formatBeijingTime(nowSec)} 北京时间</div><div class="normal">${summary}</div>`,
      })
    }
    await env.OPS_STATE.put(
      stateKey,
      JSON.stringify({ active: false, lastSentAt: nowSec, detail: '' })
    )
  }
}

export async function runHealthAndAlertPass(env: Env): Promise<CombinedHubPayload> {
  const payload = await collectCombinedHubPayload(env)
  const nowSec = payload.fetched_at
  const cooldownSec = Math.max(300, Number.parseInt(env.ALERT_COOLDOWN_SECONDS || '3600', 10) || 3600)
  const modelThreshold = Number.parseFloat(env.MODEL_ALERT_THRESHOLD || '85') || 85
  const minSamples = Number.parseInt(env.MIN_MODEL_SAMPLES || '5', 10) || 5

  // 1. OpenBridger service availability (only evaluate if OB_OPS_TOKEN is set)
  if ((env.OB_OPS_TOKEN || '').trim()) {
    const obBad = !payload.openbridger.reachable || payload.openbridger.data?.db_healthy === false
    const obReason = !payload.openbridger.reachable
      ? `OpenBridger 接口不可达 (${payload.openbridger.error || 'unknown'})`
      : 'OpenBridger 数据库健康检查失败'
    await evaluateStatefulCheck(
      env,
      'ob_service',
      obBad,
      '🚨 OpenBridger 服务异常',
      obReason,
      '✅ OpenBridger 服务已恢复',
      nowSec,
      cooldownSec
    )

    if (payload.openbridger.data) {
      const ob = payload.openbridger.data
      const degradedModels = (ob.models?.models || []).filter((m) => {
        const samples = (m.real?.requests || 0) + (m.probe?.requests || 0)
        return samples >= minSamples && m.availability_24h < modelThreshold
      })
      const modelAlertDetail = degradedModels
        .slice(0, 4)
        .map((m) => `${m.model_name} (${m.availability_24h.toFixed(1)}%)`)
        .join('、')
      await evaluateStatefulCheck(
        env,
        'ob_models',
        degradedModels.length > 0,
        `⚠️ OB 上游模型可用率告警 (${degradedModels.length} 个)`,
        `低于 ${modelThreshold}%：${modelAlertDetail}`,
        '✅ OB 上游模型可用率已全部恢复',
        nowSec,
        cooldownSec
      )

      const chBad = ob.channels.auto_disabled > 0 || ob.channels.low_balance_count > 0
      await evaluateStatefulCheck(
        env,
        'ob_channels',
        chBad,
        '⚠️ OB 上游渠道告警',
        `自动禁用渠道: ${ob.channels.auto_disabled} 个 ｜ 余额<$5渠道: ${ob.channels.low_balance_count} 个`,
        '✅ OB 上游渠道状态已恢复正常',
        nowSec,
        cooldownSec
      )
    }
  }

  // 2. NetGo host, nodes & tickets (only evaluate if NETGO_OPS_TOKEN is set)
  if ((env.NETGO_OPS_TOKEN || '').trim() && (env.NETGO_OPS_URL || '').trim()) {
    const ngReachBad = !payload.netgo.reachable || payload.netgo.data?.db_ok === false
    await evaluateStatefulCheck(
      env,
      'ng_service',
      ngReachBad,
      '🚨 NetGo 面板连接异常',
      !payload.netgo.reachable
        ? `NetGo 状态接口不可达 (${payload.netgo.error || 'unknown'})`
        : 'NetGo XBoard 数据库查询异常',
      '✅ NetGo 面板连接已恢复',
      nowSec,
      cooldownSec
    )

    if (payload.netgo.data) {
      const ng = payload.netgo.data
      const hostBad = ng.host.overall === 'unhealthy'
      await evaluateStatefulCheck(
        env,
        'ng_host',
        hostBad,
        '🚨 NetGo 面板主机体检告警',
        `异常项：${(ng.host.failures || []).join(', ') || 'healthcheck unhealthy'}`,
        '✅ NetGo 面板主机体检已恢复',
        nowSec,
        cooldownSec
      )

      const offlineNodes = (ng.nodes || []).filter((n) => n.visible && !n.online)
      await evaluateStatefulCheck(
        env,
        'ng_nodes',
        offlineNodes.length > 0,
        `🚨 NetGo 节点离线 (${offlineNodes.length}/${ng.nodes_summary.visible})`,
        `离线节点：${offlineNodes.map((n) => n.name).join('、')}`,
        '✅ NetGo 全部可见节点已恢复在线',
        nowSec,
        cooldownSec
      )

      // Notify when pending tickets count increases
      const prevTicketsRaw = await env.OPS_STATE.get('alerts:last_pending_tickets')
      const prevTickets = prevTicketsRaw ? Number.parseInt(prevTicketsRaw, 10) || 0 : 0
      const currTickets = ng.tickets?.pending_reply_tickets || 0
      if (currTickets > prevTickets && isWeComConfigured(env)) {
        await sendWeComCard(env, {
          title: `💬 NetGo 有待回复用户工单 (${currTickets})`,
          description: `<div class="gray">${formatBeijingTime(nowSec)} 北京时间</div><div class="normal">当前共有 ${currTickets} 张工单等待管理员回复</div>`,
        })
        await appendRecentEvent(env, {
          ts: nowSec,
          level: 'info',
          title: `💬 NetGo 新待回工单 (${currTickets})`,
          summary: `待回复工单由 ${prevTickets} 增加至 ${currTickets}`,
        })
      }
      await env.OPS_STATE.put('alerts:last_pending_tickets', String(currTickets))
    }
  }

  return payload
}

export function buildDailyBriefingCard(payload: CombinedHubPayload): { title: string; description: string } {
  const lines: string[] = [`<div class="gray">${formatBeijingTime(payload.fetched_at)} 实时汇总</div>`]

  const ob = payload.openbridger.data
  if (ob) {
    const models = ob.models?.models || []
    const healthyModels = models.filter((m) => m.availability_24h >= 95 || (m.real.requests + m.probe.requests) === 0).length
    lines.push(
      `<div class="normal">🟠 OpenBridger：今日营收 ¥${fmtMoney(ob.today.cash_revenue_cny)}（${ob.today.paid_order_count}单）｜消耗 ¥${fmtMoney(ob.today.consumption_cny)}｜毛利 ¥${fmtMoney(ob.today.estimated_gross_profit_cny)}</div>`,
      `<div class="normal">   7日营收 ¥${fmtMoney(ob.last_7d.cash_revenue_cny)}｜模型健康 ${healthyModels}/${models.length}｜渠道余额 $${fmtMoney(ob.channels.total_balance_usd)}</div>`
    )
  } else {
    lines.push(`<div class="highlight">🟠 OpenBridger：暂未连接 (${payload.openbridger.error || '未配置'})</div>`)
  }

  const ng = payload.netgo.data
  if (ng) {
    lines.push(
      `<div class="normal">🔵 NetGo：今日营收 ¥${fmtMoney(ng.today.cash_revenue_cny)}（${ng.today.paid_order_count}单）｜7日 ¥${fmtMoney(ng.last_7d.cash_revenue_cny)}</div>`,
      `<div class="normal">   有效订阅 ${ng.users.active_subscribers}人｜节点在线 ${ng.nodes_summary.online_visible}/${ng.nodes_summary.visible}｜待回工单 ${ng.tickets.pending_reply_tickets}</div>`
    )
  } else {
    lines.push(`<div class="highlight">🔵 NetGo：暂未连接 (${payload.netgo.error || '未配置'})</div>`)
  }

  return {
    title: '📊 OB + NetGo 掌上经营与状态简报',
    description: lines.join('\n'),
  }
}

export async function sendDailyBriefing(env: Env): Promise<void> {
  const payload = await collectCombinedHubPayload(env)
  if (!isWeComConfigured(env)) {
    throw new Error('企业微信尚未配置密钥')
  }
  const card = buildDailyBriefingCard(payload)
  await sendWeComCard(env, card)
  await appendRecentEvent(env, {
    ts: payload.fetched_at,
    level: 'info',
    title: '📊 已推送双平台经营简报',
    summary: '已向企业微信发送实时汇总卡片',
  })
}
