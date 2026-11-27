import {
  checkLoginRateLimit,
  clearLoginFailures,
  constantTimeSecretMatch,
  getClientKey,
  isAuthConfigured,
  issueSessionCookie,
  recordLoginFailure,
  revokeSession,
  verifySameOrigin,
  verifySession,
} from './auth'
import {
  collectCombinedHubPayload,
  runHealthAndAlertPass,
  sendDailyBriefing,
} from './monitor'
import { PWA_ICON_SVG, PWA_MANIFEST_JSON, renderAppHtml } from './pwa'
import type { Env } from './types'
import { sendWeComCard } from './wecom'

function jsonResponse(
  status: number,
  body: Record<string, unknown>,
  extraHeaders: Record<string, string> = {}
): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      'Content-Type': 'application/json; charset=utf-8',
      'Cache-Control': 'no-store',
      'X-Content-Type-Options': 'nosniff',
      'X-Frame-Options': 'DENY',
      ...extraHeaders,
    },
  })
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)

    if (request.method === 'GET' && url.pathname === '/manifest.webmanifest') {
      return new Response(PWA_MANIFEST_JSON, {
        status: 200,
        headers: { 'Content-Type': 'application/manifest+json; charset=utf-8' },
      })
    }

    if (request.method === 'GET' && url.pathname === '/icon.svg') {
      return new Response(PWA_ICON_SVG, {
        status: 200,
        headers: { 'Content-Type': 'image/svg+xml; charset=utf-8', 'Cache-Control': 'public, max-age=86400' },
      })
    }

    if (request.method === 'POST' && url.pathname === '/api/auth/login') {
      if (!isAuthConfigured(env)) {
        return jsonResponse(503, { success: false, message: '服务端尚未配置访问口令' })
      }
      if (!verifySameOrigin(request)) {
        return jsonResponse(403, { success: false, message: '非法跨站请求' })
      }
      const clientKey = getClientKey(request)
      if (!(await checkLoginRateLimit(env, clientKey))) {
        return jsonResponse(429, { success: false, message: '尝试次数过多，请 15 分钟后再试' })
      }

      let body: { passphrase?: string } = {}
      try {
        body = (await request.json()) as { passphrase?: string }
      } catch {
        return jsonResponse(400, { success: false, message: '请求格式错误' })
      }

      const candidate = typeof body.passphrase === 'string' ? body.passphrase.trim() : ''
      const expected = (env.DASHBOARD_PASSPHRASE || '').trim()
      const matched = candidate.length > 0 && (await constantTimeSecretMatch(expected, candidate))
      if (!matched) {
        await recordLoginFailure(env, clientKey)
        return jsonResponse(401, { success: false, message: '管理口令错误' })
      }

      await clearLoginFailures(env, clientKey)
      const cookie = await issueSessionCookie(request, env)
      return jsonResponse(200, { success: true }, { 'Set-Cookie': cookie })
    }

    const sid = await verifySession(request, env)

    if (request.method === 'POST' && url.pathname === '/api/auth/logout') {
      if (!verifySameOrigin(request)) {
        return jsonResponse(403, { success: false, message: '非法跨站请求' })
      }
      const clearCookie = sid ? await revokeSession(request, env, sid) : ''
      return jsonResponse(
        200,
        { success: true },
        clearCookie ? { 'Set-Cookie': clearCookie } : {}
      )
    }

    if (request.method === 'GET' && url.pathname === '/') {
      const html = renderAppHtml(Boolean(sid), isAuthConfigured(env))
      return new Response(html, {
        status: 200,
        headers: {
          'Content-Type': 'text/html; charset=utf-8',
          'Cache-Control': 'no-store',
          'X-Content-Type-Options': 'nosniff',
          'X-Frame-Options': 'DENY',
          'Content-Security-Policy':
            "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'",
        },
      })
    }

    if (!sid) {
      return jsonResponse(401, { success: false, message: '请先登录掌上总控' })
    }

    if (request.method === 'GET' && url.pathname === '/api/hub') {
      const data = await collectCombinedHubPayload(env)
      return jsonResponse(200, { success: true, data })
    }

    if (request.method === 'POST' && url.pathname === '/api/push/test') {
      if (!verifySameOrigin(request)) {
        return jsonResponse(403, { success: false, message: '非法跨站请求' })
      }
      try {
        await sendWeComCard(env, {
          title: '✅ OB · NetGo 微信推送通道已打通',
          description: '<div class="normal">测试成功！当节点掉线、上游模型异常或每日 09:00 简报生成时，你会在这里收到实时卡片。</div>',
          url: url.origin,
        })
        return jsonResponse(200, { success: true, message: '测试卡片已推送至微信！' })
      } catch (err) {
        return jsonResponse(500, {
          success: false,
          message: err instanceof Error ? err.message : '企业微信推送失败',
        })
      }
    }

    if (request.method === 'POST' && url.pathname === '/api/push/briefing') {
      if (!verifySameOrigin(request)) {
        return jsonResponse(403, { success: false, message: '非法跨站请求' })
      }
      try {
        await sendDailyBriefing(env)
        return jsonResponse(200, { success: true, message: '双平台经营简报已发往微信！' })
      } catch (err) {
        return jsonResponse(500, {
          success: false,
          message: err instanceof Error ? err.message : '推送简报失败',
        })
      }
    }

    return jsonResponse(404, { success: false, message: 'Not Found' })
  },

  async scheduled(controller: ScheduledController, env: Env): Promise<void> {
    if (controller.cron === '0 1 * * *') {
      await runHealthAndAlertPass(env)
      await sendDailyBriefing(env)
      return
    }
    await runHealthAndAlertPass(env)
  },
}
