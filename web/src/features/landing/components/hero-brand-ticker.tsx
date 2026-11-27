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

interface BrandItem {
  id: string
  name: string
  icon: ReactNode
}

const BRANDS: BrandItem[] = [
  { id: 'openai', name: 'OpenAI · Codex', icon: <OpenAI size={18} /> },
  { id: 'claude', name: 'Claude', icon: <Claude.Color size={18} /> },
  { id: 'gemini', name: 'Gemini', icon: <Gemini.Color size={18} /> },
  { id: 'deepseek', name: 'DeepSeek', icon: <DeepSeek.Color size={18} /> },
  { id: 'qwen', name: '阿里巴巴 Qwen', icon: <Qwen.Color size={18} /> },
  { id: 'grok', name: 'xAI Grok', icon: <Grok size={18} /> },
  { id: 'zhipu', name: '智谱 GLM', icon: <Zhipu.Color size={18} /> },
  { id: 'minimax', name: 'MiniMax 海螺', icon: <Minimax.Color size={18} /> },
  { id: 'moonshot', name: 'Kimi Moonshot', icon: <Moonshot size={18} /> },
  { id: 'doubao', name: '字节跳动 豆包', icon: <Doubao.Color size={18} /> },
  { id: 'kling', name: '快手可灵 Kling', icon: <Kling.Color size={18} /> },
  { id: 'jimeng', name: '即梦 AI', icon: <Jimeng.Color size={18} /> },
]

export function HeroBrandTicker() {
  return (
    <div
      aria-label='Supported AI Providers'
      className='mt-12 sm:mt-16 border-t border-black/[0.06] pt-6 dark:border-white/[0.08]'
    >
      <div className='flex flex-wrap items-center justify-between gap-y-4 gap-x-6 sm:gap-x-8 text-muted-foreground/80'>
        {BRANDS.map((b) => (
          <div
            key={b.id}
            className='flex items-center gap-2 text-xs font-medium transition-colors hover:text-foreground shrink-0'
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
