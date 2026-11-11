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
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useStatus } from '@/hooks/use-status'
import { cn } from '@/lib/utils'

import { getCodeSnippets } from '../../constants'
import type { CodeLanguage } from '../../types'
import { ScrollReveal } from '../scroll-reveal'

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
    <section className='relative z-10 py-20 sm:py-28'>
      <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
        <div className='grid grid-cols-1 items-center gap-10 lg:grid-cols-12'>
          {/* Left Column: 3-Step Drop-In Guide */}
          <ScrollReveal className='min-w-0 space-y-6 lg:col-span-5'>
            <div className='border-primary/25 bg-primary/10 text-primary inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-semibold'>
              <Code2 className='size-3.5' />
              <span>{t('landing.code.eyebrow')}</span>
            </div>
            <h2 className='text-2xl font-bold tracking-tight sm:text-4xl'>
              {t('landing.code.title')}
            </h2>
            <p className='text-muted-foreground text-sm leading-relaxed sm:text-base'>
              {t('landing.code.description')}
            </p>

            <div className='space-y-4 pt-2'>
              {steps.map((step) => (
                <div
                  key={step.num}
                  className='border-border/60 bg-card/60 flex items-start gap-3.5 rounded-2xl border p-4'
                >
                  <span className='bg-primary/15 text-primary flex size-8 shrink-0 items-center justify-center rounded-lg font-mono text-xs font-bold'>
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

          {/* Right Column: Interactive Code Window */}
          <ScrollReveal delay={0.12} className='min-w-0 lg:col-span-7'>
            <Card className='border-border/70 bg-card/90 overflow-hidden rounded-2xl py-0 shadow-none backdrop-blur-md'>
              <Tabs
                value={activeTab}
                onValueChange={(value) => setActiveTab(value as CodeLanguage)}
                className='gap-0'
              >
                <CardHeader className='border-border/60 bg-muted/30 flex flex-row flex-wrap items-center justify-between gap-3 border-b px-4 py-3'>
                  <TabsList className='bg-background/80 h-9 p-1'>
                    {snippets.map((snippet) => (
                      <TabsTrigger
                        key={snippet.id}
                        value={snippet.id}
                        className='px-3 text-xs'
                      >
                        {snippet.label}
                      </TabsTrigger>
                    ))}
                  </TabsList>

                  {snippets.map((snippet) =>
                    snippet.id === activeTab ? (
                      <div key={snippet.id} className='flex items-center gap-2'>
                        <span className='text-muted-foreground hidden font-mono text-xs sm:inline'>
                          {snippet.filename}
                        </span>
                        <CopyButton
                          value={snippet.code}
                          variant='outline'
                          size='sm'
                          className='h-8 gap-1.5 rounded-lg px-2.5 text-xs'
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
                      <pre className='overflow-x-auto p-5 font-mono text-xs leading-6 sm:text-[13px]'>
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
                                    'bg-primary/10 border-primary border-l-2'
                                )}
                              >
                                <span
                                  aria-hidden='true'
                                  className='text-muted-foreground/50 mr-4 w-5 shrink-0 text-right select-none'
                                >
                                  {lineNum}
                                </span>
                                <span className='text-foreground/90'>
                                  {line || ' '}
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
            </Card>
          </ScrollReveal>
        </div>
      </div>
    </section>
  )
}
