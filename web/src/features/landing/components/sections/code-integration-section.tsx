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
import { Code2, CheckCircle2 } from 'lucide-react'
import { useMemo, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { CardContent, CardHeader } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useStatus } from '@/hooks/use-status'
import { cn } from '@/lib/utils'

import { getCodeSnippets } from '../../constants'
import type { CodeLanguage } from '../../types'
import { ScrollReveal } from '../scroll-reveal'

const CODE_TOKEN_RE =
  /(#.*$)|("[^"]*"|'[^']*')|\b(from|import|const|new|await|curl|POST)\b/g

function highlightCodeLine(line: string): ReactNode {
  if (!line.trim()) return ' '
  const trimmed = line.trimStart()
  if (trimmed.startsWith('#')) {
    return <span className='ob-code-iron'>{line}</span>
  }

  const segments: ReactNode[] = []
  let cursor = 0
  const matches = [...line.matchAll(CODE_TOKEN_RE)]

  if (matches.length === 0) {
    return <span>{line}</span>
  }

  matches.forEach((match) => {
    const start = match.index ?? 0
    if (start > cursor) {
      segments.push(
        <span key={`plain-${start}`}>{line.slice(cursor, start)}</span>
      )
    }
    const [full, comment, strLiteral, keyword] = match
    if (comment) {
      segments.push(
        <span key={`com-${start}`} className='ob-code-iron'>
          {full}
        </span>
      )
    } else if (strLiteral) {
      segments.push(
        <span key={`str-${start}`} className='ob-code-str'>
          {full}
        </span>
      )
    } else if (keyword) {
      segments.push(
        <span key={`kw-${start}`} className='ob-code-kw font-semibold'>
          {full}
        </span>
      )
    } else {
      segments.push(<span key={`tok-${start}`}>{full}</span>)
    }
    cursor = start + full.length
  })

  if (cursor < line.length) {
    segments.push(<span key='tail'>{line.slice(cursor)}</span>)
  }

  return segments
}

export function CodeIntegrationSection() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const [activeTab, setActiveTab] = useState<CodeLanguage>('python')

  const serverAddress = useMemo(() => {
    const fromStatus = status?.server_address as string | undefined
    if (fromStatus && fromStatus.trim()) {
      return fromStatus.trim()
    }
    if (typeof window !== 'undefined' && window.location?.origin) {
      return window.location.origin
    }
    return 'https://api.openbridger.com'
  }, [status?.server_address])

  const snippets = useMemo(
    () => getCodeSnippets(serverAddress),
    [serverAddress]
  )

  const steps = [
    {
      num: '01',
      titleKey: 'landing.code.step1Title',
      descKey: 'landing.code.step1Desc',
    },
    {
      num: '02',
      titleKey: 'landing.code.step2Title',
      descKey: 'landing.code.step2Desc',
    },
    {
      num: '03',
      titleKey: 'landing.code.step3Title',
      descKey: 'landing.code.step3Desc',
    },
  ]

  return (
    <section className='ob-section-bg relative z-10 py-20 sm:py-28'>
      <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
        <div className='grid grid-cols-1 items-center gap-10 lg:grid-cols-12'>
          {/* Left Column: 3-Step Drop-In Guide */}
          <ScrollReveal className='min-w-0 space-y-6 lg:col-span-5'>
            <div className='ob-chip-primary inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold'>
              <Code2 className='size-3.5' />
              <span>{t('landing.code.eyebrow')}</span>
            </div>
            <h2 className='ob-serif-heading text-2xl font-bold tracking-tight sm:text-4xl'>
              {t('landing.code.title')}
            </h2>
            <p className='text-muted-foreground text-sm leading-relaxed sm:text-base'>
              {t('landing.code.description')}
            </p>

            <div className='space-y-4 pt-2'>
              {steps.map((step) => (
                <div
                  key={step.num}
                  className='ob-card ob-card-interactive flex items-start gap-3.5 rounded-2xl p-4'
                >
                  <span className='ob-chip-primary ob-mono flex size-8 shrink-0 items-center justify-center rounded-lg text-xs font-bold'>
                    {step.num}
                  </span>
                  <div className='space-y-1'>
                    <h3 className='flex items-center gap-1.5 text-sm font-semibold'>
                      <span>{t(step.titleKey)}</span>
                      <CheckCircle2 className='text-primary size-3.5' />
                    </h3>
                    <p className='text-muted-foreground text-xs leading-relaxed'>
                      {t(step.descKey)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </ScrollReveal>

          {/* Right Column: Interactive Code Window (#14110F Dark Card) */}
          <ScrollReveal delay={0.12} className='min-w-0 lg:col-span-7'>
            <div className='ob-code-card overflow-hidden rounded-2xl'>
              <Tabs
                value={activeTab}
                onValueChange={(value) => setActiveTab(value as CodeLanguage)}
                className='gap-0'
              >
                <CardHeader className='ob-code-header flex flex-row flex-wrap items-center justify-between gap-3 border-b px-4 py-3'>
                  <TabsList className='h-9 bg-black/30 p-1'>
                    {snippets.map((snippet) => (
                      <TabsTrigger
                        key={snippet.id}
                        value={snippet.id}
                        className='px-3 text-xs text-[var(--ob-iron)] data-[active]:bg-[var(--ob-primary)] data-[active]:text-white'
                      >
                        {snippet.label}
                      </TabsTrigger>
                    ))}
                  </TabsList>

                  {snippets.map((snippet) =>
                    snippet.id === activeTab ? (
                      <div key={snippet.id} className='flex items-center gap-2'>
                        <span className='ob-mono ob-code-iron hidden text-xs sm:inline'>
                          {snippet.filename}
                        </span>
                        <CopyButton
                          value={snippet.code}
                          variant='outline'
                          size='sm'
                          className='h-8 gap-1.5 rounded-lg border-white/15 bg-white/5 px-2.5 text-xs text-[var(--ob-code-ink)] hover:border-[var(--ob-primary)] hover:bg-white/10 hover:text-white'
                          iconClassName='size-3.5'
                        >
                          <span>{t('Copy')}</span>
                        </CopyButton>
                      </div>
                    ) : null
                  )}
                </CardHeader>

                <CardContent className='p-0'>
                  {snippets.map((snippet) => (
                    <TabsContent
                      key={snippet.id}
                      value={snippet.id}
                      className='m-0'
                    >
                      <pre className='ob-mono overflow-x-auto p-5 text-xs leading-6 sm:text-[13px]'>
                        <code>
                          {snippet.code.split('\n').map((line, index) => {
                            const lineNum = index + 1
                            const isHighlighted =
                              snippet.highlightLines.includes(lineNum)
                            return (
                              <div
                                key={lineNum}
                                className={cn(
                                  '-mx-5 flex px-5',
                                  isHighlighted &&
                                    'border-l-2 border-[var(--ob-primary)] bg-[rgba(255,107,0,0.12)]'
                                )}
                              >
                                <span
                                  aria-hidden='true'
                                  className='ob-code-iron mr-4 w-5 shrink-0 text-right select-none'
                                >
                                  {lineNum}
                                </span>
                                <span className='whitespace-pre'>
                                  {highlightCodeLine(line)}
                                </span>
                              </div>
                            )
                          })}
                        </code>
                      </pre>
                    </TabsContent>
                  ))}
                </CardContent>
              </Tabs>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>
  )
}
