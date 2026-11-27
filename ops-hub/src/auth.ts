import type { Env } from './types'

const SESSION_TTL_SECONDS = 12 * 3600
const LOGIN_WINDOW_SECONDS = 15 * 60
const MAX_LOGIN_FAILURES = 5
const MIN_PASSPHRASE_LEN = 16
const MIN_SIGNING_SECRET_LEN = 32

const encoder = new TextEncoder()

function toHex(buf: ArrayBuffer): string {
  const bytes = new Uint8Array(buf)
  let out = ''
  for (let i = 0; i < bytes.length; i++) {
    out += bytes[i].toString(16).padStart(2, '0')
  }
  return out
}

function fromHex(hex: string): Uint8Array | null {
  if (hex.length % 2 !== 0 || !/^[0-9a-f]+$/i.test(hex)) {
    return null
  }
  const out = new Uint8Array(hex.length / 2)
  for (let i = 0; i < out.length; i++) {
    out[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16)
  }
  return out
}

async function importHmacKey(secret: string): Promise<CryptoKey> {
  return crypto.subtle.importKey(
    'raw',
    encoder.encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign', 'verify']
  )
}

export async function constantTimeSecretMatch(expected: string, candidate: string): Promise<boolean> {
  const fixedKey = await importHmacKey('ops-hub-constant-time-domain-separator-v1')
  const expectedSig = await crypto.subtle.sign('HMAC', fixedKey, encoder.encode(expected))
  return crypto.subtle.verify('HMAC', fixedKey, expectedSig, encoder.encode(candidate))
}

export function isAuthConfigured(env: Env): boolean {
  const pass = (env.DASHBOARD_PASSPHRASE || '').trim()
  const sign = (env.SESSION_SIGNING_SECRET || '').trim()
  return pass.length >= MIN_PASSPHRASE_LEN && sign.length >= MIN_SIGNING_SECRET_LEN
}

export function verifySameOrigin(request: Request): boolean {
  const url = new URL(request.url)
  const origin = request.headers.get('Origin')
  if (!origin) {
    return false
  }
  return origin === url.origin
}

function cookieNameForRequest(request: Request): string {
  const isHttps = new URL(request.url).protocol === 'https:'
  return isHttps ? '__Host-ops_sid' : 'ops_sid'
}

function parseCookies(cookieHeader: string | null): Record<string, string> {
  if (!cookieHeader) return {}
  const out: Record<string, string> = {}
  for (const part of cookieHeader.split(';')) {
    const idx = part.indexOf('=')
    if (idx <= 0) continue
    const k = part.slice(0, idx).trim()
    const v = part.slice(idx + 1).trim()
    out[k] = v
  }
  return out
}

export function getClientKey(request: Request): string {
  return (request.headers.get('CF-Connecting-IP') || 'local').trim()
}

export async function checkLoginRateLimit(env: Env, ip: string): Promise<boolean> {
  const raw = await env.OPS_STATE.get(`ratelimit:login:${ip}`)
  if (!raw) return true
  const count = Number.parseInt(raw, 10)
  return Number.isNaN(count) || count < MAX_LOGIN_FAILURES
}

export async function recordLoginFailure(env: Env, ip: string): Promise<void> {
  const key = `ratelimit:login:${ip}`
  const raw = await env.OPS_STATE.get(key)
  const prev = raw ? Number.parseInt(raw, 10) || 0 : 0
  await env.OPS_STATE.put(key, String(prev + 1), { expirationTtl: LOGIN_WINDOW_SECONDS })
}

export async function clearLoginFailures(env: Env, ip: string): Promise<void> {
  await env.OPS_STATE.delete(`ratelimit:login:${ip}`)
}

export async function issueSessionCookie(request: Request, env: Env): Promise<string> {
  const signingSecret = (env.SESSION_SIGNING_SECRET || '').trim()
  const sidBytes = new Uint8Array(16)
  crypto.getRandomValues(sidBytes)
  const sid = toHex(sidBytes.buffer)
  const exp = Math.floor(Date.now() / 1000) + SESSION_TTL_SECONDS
  const payload = `${sid}.${exp}`

  const key = await importHmacKey(signingSecret)
  const sig = toHex(await crypto.subtle.sign('HMAC', key, encoder.encode(payload)))
  const token = `${payload}.${sig}`

  await env.OPS_STATE.put(`session:${sid}`, String(exp), { expirationTtl: SESSION_TTL_SECONDS })

  const name = cookieNameForRequest(request)
  const secureAttr = new URL(request.url).protocol === 'https:' ? '; Secure' : ''
  return `${name}=${token}; Max-Age=${SESSION_TTL_SECONDS}; Path=/; HttpOnly; SameSite=Strict${secureAttr}`
}

export async function verifySession(request: Request, env: Env): Promise<string | null> {
  if (!isAuthConfigured(env)) return null
  const cookies = parseCookies(request.headers.get('Cookie'))
  const token = cookies[cookieNameForRequest(request)]
  if (!token) return null

  const parts = token.split('.')
  if (parts.length !== 3) return null
  const [sid, expStr, sigHex] = parts
  if (!/^[0-9a-f]{32}$/.test(sid)) return null

  const exp = Number.parseInt(expStr, 10)
  const now = Math.floor(Date.now() / 1000)
  if (!Number.isFinite(exp) || exp <= now) return null

  const sigBytes = fromHex(sigHex)
  if (!sigBytes) return null

  const key = await importHmacKey((env.SESSION_SIGNING_SECRET || '').trim())
  const validSig = await crypto.subtle.verify(
    'HMAC',
    key,
    sigBytes,
    encoder.encode(`${sid}.${expStr}`)
  )
  if (!validSig) return null

  const active = await env.OPS_STATE.get(`session:${sid}`)
  if (!active) return null
  return sid
}

export async function revokeSession(request: Request, env: Env, sid: string): Promise<string> {
  await env.OPS_STATE.delete(`session:${sid}`)
  const name = cookieNameForRequest(request)
  const secureAttr = new URL(request.url).protocol === 'https:' ? '; Secure' : ''
  return `${name}=; Max-Age=0; Path=/; HttpOnly; SameSite=Strict${secureAttr}`
}
