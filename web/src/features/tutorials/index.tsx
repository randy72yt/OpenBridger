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
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { PublicShell } from '@/features/landing/components/public-shell'
import { useStatus } from '@/hooks/use-status'
import { resolveDocsHref } from '@/lib/docs-link'

export function Tutorials() {
  const { t, i18n } = useTranslation()
  const { status, loading } = useStatus()
  const href = resolveDocsHref(
    status?.docs_link as string | undefined,
    i18n.language,
    typeof window === 'undefined' ? undefined : window.location.origin
  )

  useEffect(() => {
    if (loading || typeof window === 'undefined') return
    const next = new URL(href, window.location.origin)
    const here = new URL(window.location.href)
    if (next.origin === here.origin && next.pathname.replace(/\/+$/, '') === '/docs') {
      return
    }
    if (next.href === here.href) return
    window.location.replace(href)
  }, [href, loading])

  return (
    <PublicShell>
      <main className='mx-auto min-h-[70vh] max-w-4xl px-6 pt-32 pb-16'>
        <a href={href}>{t('Practical tutorials')}</a>
      </main>
    </PublicShell>
  )
}
