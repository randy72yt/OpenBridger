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
import { Claude, DeepSeek, Gemini, Grok, OpenAI, Qwen } from '@lobehub/icons'
import { KeyRound, Layers, Terminal } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader } from '@/components/ui/card'

import { LANDING_PROVIDERS, SUPPORTED_CLI_TOOLS } from '../../constants'
import { ScrollReveal } from '../scroll-reveal'

function renderProviderIcon(iconKey: string) {
  switch (iconKey) {
    case 'OpenAI':
      return <OpenAI size={22} />
    case 'Claude.Color':
      return <Claude.Color size={22} />
    case 'Gemini.Color':
      return <Gemini.Color size={22} />
    case 'DeepSeek.Color':
      return <DeepSeek.Color size={22} />
    case 'Qwen.Color':
      return <Qwen.Color size={22} />
    case 'Grok':
      return <Grok size={22} />
    default:
      return <OpenAI size={22} />
  }
}

export function MultiModelSection() {
  const { t } = useTranslation()

  return (
    <section className='ob-section-warm relative z-10 py-20 sm:py-28'>
      <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
        <ScrollReveal className='mx-auto max-w-3xl text-center'>
          <div className='ob-chip-primary mb-3 inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold'>
            <KeyRound className='size-3.5' />
            <span>{t('landing.multiModel.eyebrow')}</span>
          </div>
          <h2 className='ob-serif-heading text-2xl font-bold tracking-tight sm:text-4xl'>
            {t('landing.multiModel.title')}
          </h2>
          <p className='text-muted-foreground mt-3.5 text-sm leading-relaxed sm:text-base'>
            {t('landing.multiModel.description')}
          </p>
        </ScrollReveal>

        {/* Provider Cards Grid */}
        <div className='mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3'>
          {LANDING_PROVIDERS.map((provider, idx) => (
            <ScrollReveal key={provider.id} delay={idx * 0.06}>
              <Card className='ob-card ob-card-interactive h-full rounded-2xl py-0 shadow-none'>
                <CardHeader className='flex flex-row items-center justify-between gap-3 p-5 pb-3'>
                  <div className='flex items-center gap-3'>
                    <div className='ob-section-warm border-border flex size-10 shrink-0 items-center justify-center rounded-xl border'>
                      {renderProviderIcon(provider.iconKey)}
                    </div>
                    <div>
                      <h3 className='text-base font-semibold'>
                        {provider.name}
                      </h3>
                      <p className='text-muted-foreground ob-mono text-[11px]'>
                        {provider.protocol}
                      </p>
                    </div>
                  </div>
                  <Badge
                    variant='secondary'
                    className='ob-chip-primary shrink-0 text-[11px]'
                  >
                    {t(provider.tagKey)}
                  </Badge>
                </CardHeader>
                <CardContent className='p-5 pt-2'>
                  <div className='flex flex-wrap gap-1.5'>
                    {provider.models.map((modelName) => (
                      <span
                        key={modelName}
                        className='ob-section-bg border-border text-foreground/90 ob-mono rounded-lg border px-2.5 py-1 text-xs'
                      >
                        {modelName}
                      </span>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </ScrollReveal>
          ))}
        </div>

        {/* Bottom Compatibility Banner */}
        <ScrollReveal delay={0.18} className='mt-8'>
          <Card className='ob-card rounded-2xl py-0 shadow-none'>
            <CardContent className='flex flex-col justify-between gap-6 p-6 lg:flex-row lg:items-center'>
              <div className='space-y-1.5'>
                <div className='text-primary flex items-center gap-2 text-xs font-semibold tracking-wider uppercase'>
                  <Layers className='size-4' />
                  <span>{t('landing.multiModel.ecosystemBadge')}</span>
                </div>
                <h3 className='text-lg font-semibold'>
                  {t('landing.multiModel.ecosystemTitle')}
                </h3>
                <p className='text-muted-foreground max-w-xl text-xs leading-relaxed sm:text-sm'>
                  {t('landing.multiModel.ecosystemDesc')}
                </p>
              </div>
              <div className='flex flex-wrap gap-2'>
                {SUPPORTED_CLI_TOOLS.map((tool) => (
                  <div
                    key={tool.name}
                    className='ob-section-warm border-border flex items-center gap-2 rounded-xl border px-3.5 py-2 text-xs'
                  >
                    <Terminal className='text-primary size-3.5 shrink-0' />
                    <span className='font-semibold'>{tool.name}</span>
                    <span className='text-muted-foreground ob-mono text-[10px]'>
                      {tool.protocol}
                    </span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </ScrollReveal>
      </div>
    </section>
  )
}
