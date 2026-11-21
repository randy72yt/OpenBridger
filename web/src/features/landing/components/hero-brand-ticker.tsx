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
import {
  Claude,
  DeepSeek,
  Doubao,
  Gemini,
  Grok,
  Jimeng,
  Kling,
  Minimax,
  Moonshot,
  OpenAI,
  Qwen,
  Zhipu,
} from '@lobehub/icons'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

interface BrandItem {
  id: string
  name: string
  icon: ReactNode
}

export function HeroBrandTicker() {
  const { t } = useTranslation()
  const brands: BrandItem[] = [
    { id: 'openai', name: t('OpenAI'), icon: <OpenAI size={18} /> },
    { id: 'claude', name: t('Claude'), icon: <Claude.Color size={18} /> },
    { id: 'gemini', name: t('Gemini'), icon: <Gemini.Color size={18} /> },
    { id: 'deepseek', name: t('DeepSeek'), icon: <DeepSeek.Color size={18} /> },
    { id: 'qwen', name: t('Alibaba Qwen'), icon: <Qwen.Color size={18} /> },
    { id: 'grok', name: t('xAI Grok'), icon: <Grok size={18} /> },
    { id: 'zhipu', name: t('Zhipu GLM'), icon: <Zhipu.Color size={18} /> },
    {
      id: 'minimax',
      name: t('MiniMax Hailuo'),
      icon: <Minimax.Color size={18} />,
    },
    { id: 'moonshot', name: t('Kimi Moonshot'), icon: <Moonshot size={18} /> },
    {
      id: 'doubao',
      name: t('ByteDance Doubao'),
      icon: <Doubao.Color size={18} />,
    },
    { id: 'kling', name: t('Kuaishou Kling'), icon: <Kling.Color size={18} /> },
    { id: 'jimeng', name: t('Jimeng AI'), icon: <Jimeng.Color size={18} /> },
  ]

  return (
    <div
      aria-label={t('Supported AI providers')}
      className='mt-12 border-t border-black/[0.06] pt-6 sm:mt-16 dark:border-white/[0.08]'
    >
      <div className='text-muted-foreground/80 flex flex-wrap items-center justify-between gap-x-6 gap-y-4 sm:gap-x-8'>
        {brands.map((b) => (
          <div
            key={b.id}
            className='hover:text-foreground flex shrink-0 items-center gap-2 text-xs font-medium transition-colors'
          >
            <span className='flex size-5 shrink-0 items-center justify-center'>
              {b.icon}
            </span>
            <span className='tracking-tight'>{b.name}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
