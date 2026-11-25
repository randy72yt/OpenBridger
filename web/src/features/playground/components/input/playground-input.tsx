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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import {
  PromptInput,
  PromptInputFooter,
  PromptInputTextarea,
  type PromptInputMessage,
} from '@/components/ai-elements/prompt-input'

import { getSubmittableInputText } from '../../lib'
import type {
  ModelOption,
  GroupOption,
  ParameterEnabled,
  PlaygroundConfig,
} from '../../types'
import { PlaygroundInputControls } from './playground-input-controls'
import { PlaygroundInputTools } from './playground-input-tools'

export interface PlaygroundInputProps {
  config: PlaygroundConfig
  onSubmit: (text: string) => void
  onStop?: () => void
  disabled?: boolean
  isGenerating?: boolean
  models: ModelOption[]
  isModelLoading?: boolean
  groups: GroupOption[]
  onConfigChange: <K extends keyof PlaygroundConfig>(
    key: K,
    value: PlaygroundConfig[K]
  ) => void
  onParameterEnabledChange: (
    key: keyof ParameterEnabled,
    value: boolean
  ) => void
  parameterEnabled: ParameterEnabled
  showDisclaimer?: boolean
}

export function PlaygroundInput(props: PlaygroundInputProps) {
  const { t } = useTranslation()
  const [text, setText] = useState('')

  const handleSubmit = (message: PromptInputMessage) => {
    const submittableText = getSubmittableInputText(message, props.disabled)

    if (!submittableText) return
    props.onSubmit(submittableText)
    setText('')
  }

  return (
    <div className='grid w-full shrink-0 gap-2'>
      <PromptInput
        className='relative'
        groupClassName='bg-muted/35 dark:bg-muted/20 border-border/80 hover:border-border focus-within:bg-background focus-within:border-border shadow-[0_6px_28px_-10px_rgba(0,0,0,0.08)] dark:shadow-[0_12px_34px_-12px_rgba(0,0,0,0.5)] rounded-[26px] overflow-hidden transition-all duration-200 focus-within:shadow-[0_10px_36px_-10px_rgba(0,0,0,0.12)]'
        onSubmit={handleSubmit}
      >
        <PromptInputTextarea
          autoComplete='off'
          autoCorrect='off'
          autoCapitalize='off'
          spellCheck={false}
          className='min-h-13 max-h-52 border-0 bg-transparent px-4 pt-3.5 pb-1.5 text-[15px] leading-6 shadow-none focus-visible:ring-0 md:text-base'
          disabled={props.disabled}
          onChange={(event) => setText(event.target.value)}
          placeholder={t('Ask anything')}
          value={text}
        />

        <PromptInputFooter className='border-0 bg-transparent px-3 pt-1 pb-2.5'>
          <PlaygroundInputControls
            disabled={props.disabled}
            groups={props.groups}
            isGenerating={props.isGenerating}
            isModelLoading={props.isModelLoading}
            models={props.models}
            onStop={props.onStop}
            text={text}
            tools={
              <PlaygroundInputTools
                config={props.config}
                disabled={props.disabled}
                onConfigChange={props.onConfigChange}
                onParameterEnabledChange={props.onParameterEnabledChange}
                parameterEnabled={props.parameterEnabled}
              />
            }
          />
        </PromptInputFooter>
      </PromptInput>

      {props.showDisclaimer && (
        <p className='text-muted-foreground/80 px-2 text-center text-[11px] leading-4'>
          {t('AI can make mistakes. Check important info.')}
        </p>
      )}
    </div>
  )
}
