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
import { ArrowUpIcon, SquareIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { PromptInputButton } from '@/components/ai-elements/prompt-input'
import { ModelGroupSelector } from '@/components/model-group-selector'

import { getInputControlState } from '../../lib'
import type { GroupOption, ModelOption } from '../../types'

type PlaygroundInputControlsProps = {
  disabled?: boolean
  groups: GroupOption[]
  groupValue: string
  isGenerating?: boolean
  isModelLoading?: boolean
  models: ModelOption[]
  modelValue: string
  onGroupChange: (value: string) => void
  onModelChange: (value: string) => void
  onStop?: () => void
  text: string
  tools: ReactNode
}

export function PlaygroundInputControls(props: PlaygroundInputControlsProps) {
  const { t } = useTranslation()
  const { canSubmit, isSelectorDisabled, shouldShowStop } =
    getInputControlState({
      disabled: props.disabled,
      groups: props.groups,
      hasStopHandler: Boolean(props.onStop),
      isGenerating: props.isGenerating,
      isModelLoading: props.isModelLoading,
      models: props.models,
      text: props.text,
    })

  return (
    <div className='flex w-full items-center justify-between gap-2'>
      <div className='flex min-w-0 items-center gap-1'>{props.tools}</div>

      <div className='flex min-w-0 shrink-0 items-center gap-1.5'>
        <ModelGroupSelector
          selectedModel={props.modelValue}
          models={props.models}
          onModelChange={props.onModelChange}
          selectedGroup={props.groupValue}
          groups={props.groups}
          onGroupChange={props.onGroupChange}
          disabled={isSelectorDisabled}
          popoverSide='top'
          popoverAlign='end'
          triggerVariant='ghost'
          className='hover:bg-muted/80 h-8 max-w-[10.5rem] rounded-full px-2.5 text-xs font-medium sm:max-w-[15rem]'
        />

        {shouldShowStop ? (
          <PromptInputButton
            aria-label={t('Stop')}
            className='bg-foreground text-background hover:bg-foreground/85 size-8 shrink-0 rounded-full p-0 shadow-xs transition-transform active:scale-95'
            onClick={props.onStop}
            variant='default'
          >
            <SquareIcon className='size-3.5 fill-current' aria-hidden='true' />
            <span className='sr-only'>{t('Stop')}</span>
          </PromptInputButton>
        ) : (
          <PromptInputButton
            aria-label={t('Send')}
            className='bg-foreground text-background hover:bg-foreground/85 disabled:bg-muted disabled:text-muted-foreground size-8 shrink-0 rounded-full p-0 shadow-xs transition-all active:scale-95 disabled:opacity-45'
            disabled={!canSubmit}
            type='submit'
            variant='default'
          >
            <ArrowUpIcon className='size-4 stroke-[2.5]' aria-hidden='true' />
            <span className='sr-only'>{t('Send')}</span>
          </PromptInputButton>
        )}
      </div>
    </div>
  )
}
