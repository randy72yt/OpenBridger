import type { Env } from './types'

const WECOM_TOKEN_CACHE_KEY = 'wecom:access_token'
const WECOM_TOKEN_TTL_SECONDS = 7000

export function isWeComConfigured(env: Env): boolean {
  const hasApp =
    Boolean((env.WECOM_CORP_ID || '').trim()) &&
    Boolean((env.WECOM_AGENT_ID || '').trim()) &&
    Boolean((env.WECOM_CORP_SECRET || '').trim())
  const hasWebhook = Boolean((env.WECOM_WEBHOOK_URL || '').trim())
  return hasApp || hasWebhook
}

async function fetchWeComAccessToken(env: Env, forceRefresh: boolean): Promise<string> {
  if (!forceRefresh) {
    const cached = await env.OPS_STATE.get(WECOM_TOKEN_CACHE_KEY)
    if (cached) return cached
  }

  const corpId = encodeURIComponent((env.WECOM_CORP_ID || '').trim())
  const secret = encodeURIComponent((env.WECOM_CORP_SECRET || '').trim())
  const resp = await fetch(
    `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=${corpId}&corpsecret=${secret}`,
    { signal: AbortSignal.timeout(8000) }
  )
  if (!resp.ok) {
    throw new Error(`WeCom gettoken HTTP ${resp.status}`)
  }
  const body = (await resp.json()) as { errcode?: number; errmsg?: string; access_token?: string }
  if (body.errcode !== 0 || !body.access_token) {
    throw new Error(`WeCom gettoken errcode=${body.errcode ?? -1}`)
  }

  await env.OPS_STATE.put(WECOM_TOKEN_CACHE_KEY, body.access_token, {
    expirationTtl: WECOM_TOKEN_TTL_SECONDS,
  })
  return body.access_token
}

export interface WeComCardMessage {
  title: string
  description: string
  url?: string
  btntxt?: string
}

export async function sendWeComCard(env: Env, card: WeComCardMessage): Promise<void> {
  const cardUrl = (card.url || env.DASHBOARD_PUBLIC_URL || env.OB_BASE_URL || 'https://openbridger.com').trim()

  if (
    (env.WECOM_CORP_ID || '').trim() &&
    (env.WECOM_AGENT_ID || '').trim() &&
    (env.WECOM_CORP_SECRET || '').trim()
  ) {
    const agentId = Number.parseInt((env.WECOM_AGENT_ID || '').trim(), 10)
    const toUser = (env.WECOM_TO_USER || '@all').trim() || '@all'

    for (const force of [false, true]) {
      const token = await fetchWeComAccessToken(env, force)
      const resp = await fetch(
        `https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=${encodeURIComponent(token)}`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            touser: toUser,
            msgtype: 'textcard',
            agentid: agentId,
            textcard: {
              title: card.title.slice(0, 128),
              description: card.description.slice(0, 512),
              url: cardUrl,
              btntxt: card.btntxt || '打开总控台',
            },
          }),
          signal: AbortSignal.timeout(8000),
        }
      )
      if (!resp.ok) {
        throw new Error(`WeCom send HTTP ${resp.status}`)
      }
      const data = (await resp.json()) as { errcode?: number; errmsg?: string }
      if (data.errcode === 0) {
        return
      }
      // Expired or invalid access_token: refresh once and retry
      if (!force && (data.errcode === 40014 || data.errcode === 42001)) {
        continue
      }
      throw new Error(`WeCom send errcode=${data.errcode ?? -1}`)
    }
    return
  }

  const webhook = (env.WECOM_WEBHOOK_URL || '').trim()
  if (webhook) {
    const resp = await fetch(webhook, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        msgtype: 'markdown',
        markdown: {
          content: `### ${card.title}\n${card.description}\n[打开掌上总控台](${cardUrl})`,
        },
      }),
      signal: AbortSignal.timeout(8000),
    })
    if (!resp.ok) {
      throw new Error(`WeCom webhook HTTP ${resp.status}`)
    }
    return
  }

  throw new Error('WeCom push credentials are not configured')
}
