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
  language: string | undefined
): string {
  const prefix = docsLanguagePrefix(language)
  const trimmed = (docsLink ?? '').trim()
  const defaultHref = `${DEFAULT_DOCS_ORIGIN}${prefix}`
  if (trimmed === '' || trimmed === '/docs' || trimmed === '/docs/') {
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
  return trimmed
}
