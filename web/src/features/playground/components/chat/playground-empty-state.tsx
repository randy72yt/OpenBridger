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
  BarChartIcon,
  CodeSquareIcon,
  GraduationCapIcon,
  NotepadTextIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { ModelProviderIcon } from '@/components/model-provider-icon'
import { Button } from '@/components/ui/button'
import { resolveModelProvider } from '@/lib/model-provider'

export interface PlaygroundHeroGreetingProps {
  modelName: string
}

export interface PlaygroundStarterPillsProps {
  disabled?: boolean
  onSelectPrompt: (prompt: string) => void
}

const STARTER_PROMPTS = [
  {
    icon: NotepadTextIcon,
    iconColorClass: 'text-orange-500',
    text: 'Summarize text',
  },
  {
    icon: CodeSquareIcon,
    iconColorClass: 'text-blue-500',
    text: 'Code',
  },
  {
    icon: BarChartIcon,
    iconColorClass: 'text-emerald-500',
    text: 'Analyze data',
  },
  {
    icon: GraduationCapIcon,
    iconColorClass: 'text-violet-500',
    text: 'Get advice',
  },
] as const

export function PlaygroundHeroGreeting(props: PlaygroundHeroGreetingProps) {
  const { t } = useTranslation()
  const provider = resolveModelProvider(props.modelName)

  return (
    <div className='flex flex-col items-center gap-4 text-center'>
      <div
        className='border-border/70 bg-card flex size-13 items-center justify-center rounded-full border shadow-xs transition-transform duration-200'
        data-testid='playground-model-emblem'
      >
        <ModelProviderIcon
          modelName={props.modelName}
          size={26}
          fallbackToOpenAI={!props.modelName}
        />
      </div>

      <div className='space-y-1.5'>
        <h1 className='text-foreground text-2xl font-semibold tracking-tight text-balance sm:text-3xl'>
          {t('What can I help with?')}
        </h1>
        {provider && (
          <p className='text-muted-foreground text-xs font-medium'>
            {provider.label} · {props.modelName}
          </p>
        )}
      </div>
    </div>
  )
}

export function PlaygroundStarterPills(props: PlaygroundStarterPillsProps) {
  const { t } = useTranslation()

  return (
    <div className='flex flex-wrap items-center justify-center gap-2'>
      {STARTER_PROMPTS.map((item) => {
        const Icon = item.icon
        const prompt = t(item.text)

        return (
          <Button
            className='border-border/75 bg-background/80 hover:bg-muted/70 text-muted-foreground hover:text-foreground h-9 gap-2 rounded-full px-4 text-[13px] font-normal shadow-2xs transition-colors'
            disabled={props.disabled}
            key={item.text}
            onClick={() => props.onSelectPrompt(prompt)}
            type='button'
            variant='outline'
          >
            <Icon
              className={`size-4 shrink-0 ${item.iconColorClass}`}
              aria-hidden='true'
            />
            <span>{prompt}</span>
          </Button>
        )
      })}
    </div>
  )
}

export function PlaygroundEmptyState(props: {
  modelName?: string
  onSelectPrompt: (prompt: string) => void
}) {
  return (
    <div className='flex min-h-[min(460px,calc(100svh-18rem))] flex-col items-center justify-center gap-6 px-2 py-8'>
      <PlaygroundHeroGreeting modelName={props.modelName ?? ''} />
      <PlaygroundStarterPills onSelectPrompt={props.onSelectPrompt} />
    </div>
  )
}
