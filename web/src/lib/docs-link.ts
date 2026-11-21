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
import { convertDetectedLanguage } from '@/i18n/languages'

export const DEFAULT_DOCS_ORIGIN = 'https://docs.openbridger.com'

export function docsLanguagePrefix(language: string | undefined): '/zh/' | '/en/' {
  const detected = convertDetectedLanguage(language?.trim() || 'en')
  const normalized = detected.replaceAll('_', '-').toLowerCase()
  if (normalized.startsWith('zh')) return '/zh/'
  return '/en/'
}

function isHttpUrl(value: string): boolean {
  return /^https?:\/\//i.test(value)
}

export function isExternalHref(href: string): boolean {
  return isHttpUrl(href)
}

/**
 * Resolve the public docs href.
 *
 * The default docs site is language-prefixed (`/zh/` or `/en/`). zh-TW uses
 * the Chinese docs because this repo has no Traditional Chinese docs site.
 * A custom docs_link is returned unchanged, including its path, query, and hash.
 * The in-app placeholder `/docs` is not a destination.
 */
export function resolveDocsHref(
  docsLink: string | undefined | null,
  language: string | undefined,
  currentOrigin?: string
): string {
  const prefix = docsLanguagePrefix(language)
  const trimmed = (docsLink ?? '').trim()
  const defaultHref = `${DEFAULT_DOCS_ORIGIN}${prefix}`
  if (trimmed === '' || isAppDocsPath(trimmed)) {
    return defaultHref
  }
  if (trimmed.startsWith('/') && !trimmed.startsWith('//')) {
    return trimmed
  }
  if (!isHttpUrl(trimmed)) return defaultHref

  let url: URL
  try {
    url = new URL(trimmed)
  } catch {
    return defaultHref
  }
  const bareRoot =
    url.protocol === 'https:' &&
    url.hostname === 'docs.openbridger.com' &&
    (url.pathname === '/' || url.pathname === '') &&
    url.search === '' &&
    url.hash === ''
  if (bareRoot) return `${url.origin}${prefix}`
  if (isSameOriginAppDocs(url, currentOrigin)) return defaultHref
  return trimmed
}

function isAppDocsPath(value: string): boolean {
  if (!value.startsWith('/') || value.startsWith('//')) return false
  try {
    return isDocsPathname(new URL(value, 'https://local.invalid').pathname)
  } catch {
    return false
  }
}

function isDocsPathname(pathname: string): boolean {
  const normalized = pathname.length > 1 ? pathname.replace(/\/+$/, '') : pathname
  return normalized === '/docs'
}

function isSameOriginAppDocs(url: URL, currentOrigin: string | undefined): boolean {
  if (!currentOrigin || !isDocsPathname(url.pathname)) return false
  try {
    return url.origin === new URL(currentOrigin).origin
  } catch {
    return false
  }
}
