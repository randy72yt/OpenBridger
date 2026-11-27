import { describe, expect, it, vi } from 'vitest'
import worker from '../src/index'
import { runHealthAndAlertPass } from '../src/monitor'
import type { Env, KVStoreLike } from '../src/types'

function createMemoryKV(): KVStoreLike {
  const map = new Map<string, string>()
  return {
    async get(key: string) {
      return map.get(key) ?? null
    },
    async put(key: string, value: string) {
      map.set(key, value)
    },
    async delete(key: string) {
      map.delete(key)
    },
  }
}

function createTestEnv(): Env {
  return {
    OPS_STATE: createMemoryKV(),
    OB_BASE_URL: 'https://openbridger.example',
    NETGO_OPS_URL: 'https://ops.netgo.example/api/netgo-ops/snapshot',
    DASHBOARD_PASSPHRASE: 'correct-horse-battery-staple-99',
    SESSION_SIGNING_SECRET: '0123456789abcdef0123456789abcdef0123456789abcdef',
    OB_OPS_TOKEN: 'ob-secret-token-0123456789abcdef0123456789',
    NETGO_OPS_TOKEN: 'ng-secret-token-0123456789abcdef0123456789',
    WECOM_WEBHOOK_URL: 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test-hook',
  }
}

describe('ops-hub authentication, CSRF, lockout, and alert engine', () => {
  it('enforces same-origin CSRF, brute-force lockout, HttpOnly cookie attributes, and server-side logout revocation', async () => {
    const env = createTestEnv()

    // Unauthenticated /api/hub returns 401
    const unauth = await worker.fetch(new Request('https://hub.example/api/hub'), env)
    expect(unauth.status).toBe(401)

    // Cross-origin login attempt is rejected with 403
    const crossOrigin = await worker.fetch(
      new Request('https://hub.example/api/auth/login', {
        method: 'POST',
        headers: { Origin: 'https://evil.example', 'Content-Type': 'application/json' },
        body: JSON.stringify({ passphrase: env.DASHBOARD_PASSPHRASE }),
      }),
      env
    )
    expect(crossOrigin.status).toBe(403)

    // 5 wrong login attempts trigger 429 lockout on the 6th attempt even with valid passphrase
    for (let i = 0; i < 5; i++) {
      const bad = await worker.fetch(
        new Request('https://hub.example/api/auth/login', {
          method: 'POST',
          headers: {
            Origin: 'https://hub.example',
            'CF-Connecting-IP': '203.0.113.9',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ passphrase: 'wrong-passphrase-attempt' }),
        }),
        env
      )
      expect(bad.status).toBe(401)
    }

    const lockedOut = await worker.fetch(
      new Request('https://hub.example/api/auth/login', {
        method: 'POST',
        headers: {
          Origin: 'https://hub.example',
          'CF-Connecting-IP': '203.0.113.9',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ passphrase: env.DASHBOARD_PASSPHRASE }),
      }),
      env
    )
    expect(lockedOut.status).toBe(429)

    // Valid login from clean IP issues __Host-ops_sid with HttpOnly; SameSite=Strict; Secure
    const loginRes = await worker.fetch(
      new Request('https://hub.example/api/auth/login', {
        method: 'POST',
        headers: {
          Origin: 'https://hub.example',
          'CF-Connecting-IP': '203.0.113.10',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ passphrase: env.DASHBOARD_PASSPHRASE }),
      }),
      env
    )
    expect(loginRes.status).toBe(200)
    const setCookie = loginRes.headers.get('Set-Cookie') || ''
    expect(setCookie).toContain('__Host-ops_sid=')
    expect(setCookie).toContain('HttpOnly')
    expect(setCookie).toContain('SameSite=Strict')
    expect(setCookie).toContain('Secure')

    const cookiePair = setCookie.split(';')[0]

    // Logout immediately revokes server-side KV session so replaying old cookie fails with 401
    const logoutRes = await worker.fetch(
      new Request('https://hub.example/api/auth/logout', {
        method: 'POST',
        headers: { Origin: 'https://hub.example', Cookie: cookiePair },
      }),
      env
    )
    expect(logoutRes.status).toBe(200)

    const replayAfterLogout = await worker.fetch(
      new Request('https://hub.example/api/hub', {
        headers: { Cookie: cookiePair },
      }),
      env
    )
    expect(replayAfterLogout.status).toBe(401)
  })

  it('pushes WeCom alert on node/model degradation, deduplicates during cooldown, and sends recovery card', async () => {
    const env = createTestEnv()
    const pushedCards: string[] = []

    let nodeOnline = false
    let modelAvail = 60.0

    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
      const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url
      if (url.includes('openbridger.example/api/ops/snapshot')) {
        return new Response(
          JSON.stringify({
            success: true,
            data: {
              generated_at: 1700000000,
              version: '0.1.6',
              uptime_seconds: 3600,
              db_healthy: true,
              pricing_control_enabled: true,
              pending_pricing_proposals: 0,
              currency: { display_currency: 'CNY', usd_exchange_rate: 7.2, price_ratio: 7.2 },
              today: { cash_revenue_cny: 88, consumption_cny: 30, estimated_gross_profit_cny: 18, paid_order_count: 3, new_users: 2 },
              last_7d: { cash_revenue_cny: 560, estimated_gross_profit_cny: 120 },
              daily_trend_7d: [],
              channels: { total: 3, enabled: 3, auto_disabled: 0, manually_disabled: 0, low_balance_count: 0, total_balance_usd: 50 },
              models: {
                models: [
                  {
                    model_name: 'gpt-6.2-astra',
                    real: { requests: 10, success: 6, avg_latency_ms: 900, avg_ttft_ms: 300, avg_tps: 40 },
                    probe: { requests: 0, success: 0, avg_latency_ms: 0, avg_ttft_ms: 0 },
                    client_error: { requests: 0 },
                    availability_24h: modelAvail,
                    availability_7d: 99,
                    error_codes: {},
                    last_probe_at: 0,
                    probe_tier: 'cheap',
                  },
                ],
                probe_budget_used_quota: 0,
                probe_budget_quota: 500000,
                generated_at: 1700000000,
              },
            },
          }),
          { status: 200 }
        )
      }
      if (url.includes('ops.netgo.example/api/netgo-ops/snapshot')) {
        return new Response(
          JSON.stringify({
            success: true,
            data: {
              generated_at: 1700000000,
              db_ok: true,
              host: { overall: 'healthy', timestamp_utc: '', disk_percent: 32, memory_available_percent: 55, backup_age_seconds: 900, checks: [], failures: [] },
              today: { cash_revenue_cny: 68, paid_order_count: 2, paying_users: 2, new_users: 5 },
              last_7d: { cash_revenue_cny: 420, paid_order_count: 14 },
              users: { total_users: 200, active_subscribers: 95 },
              tickets: { open_tickets: 0, pending_reply_tickets: 0 },
              nodes_summary: { total: 1, visible: 1, online_visible: nodeOnline ? 1 : 0, offline_visible: nodeOnline ? 0 : 1 },
              nodes: [{ id: 1, name: '🇯🇵 日本 01', type: 'vless', visible: true, rate: '1', online: nodeOnline, last_check_at: 1700000000, heartbeat_age_seconds: 10 }],
            },
          }),
          { status: 200 }
        )
      }
      if (url.includes('qyapi.weixin.qq.com/cgi-bin/webhook/send')) {
        const reqObj = input as Request
        pushedCards.push(url)
        void reqObj
        return new Response(JSON.stringify({ errcode: 0, errmsg: 'ok' }), { status: 200 })
      }
      throw new Error(`Unexpected fetch: ${url}`)
    })

    try {
      // Pass 1: fires 2 alerts (degraded OB model + offline NetGo node)
      await runHealthAndAlertPass(env)
      expect(pushedCards.length).toBe(2)

      // Pass 2 within cooldown: deduplicated (0 new pushes!)
      await runHealthAndAlertPass(env)
      expect(pushedCards.length).toBe(2)

      // Pass 3 after recovery: fires 2 recovery notifications!
      nodeOnline = true
      modelAvail = 99.5
      await runHealthAndAlertPass(env)
      expect(pushedCards.length).toBe(4)
    } finally {
      fetchSpy.mockRestore()
    }
  })
})
