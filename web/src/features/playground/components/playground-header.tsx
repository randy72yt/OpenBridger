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
import { SquarePenIcon } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { ModelGroupSelector } from '@/components/model-group-selector'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

import type { GroupOption, ModelOption } from '../types'

export interface PlaygroundHeaderProps {
  disabled?: boolean
  groups: GroupOption[]
  groupValue: string
  hasMessages: boolean
  isModelLoading?: boolean
  models: ModelOption[]
  modelValue: string
  onClearMessages: () => void
  onGroupChange: (value: string) => void
  onModelChange: (value: string) => void
}

export function PlaygroundHeader(props: PlaygroundHeaderProps) {
  const { t } = useTranslation()
  const [clearConfirmOpen, setClearConfirmOpen] = useState(false)
  const isSelectorDisabled =
    Boolean(props.disabled) ||
    Boolean(props.isModelLoading) ||
    props.groups.length === 0

  const handleConfirmNewChat = () => {
    props.onClearMessages()
    setClearConfirmOpen(false)
    toast.success(t('Conversation cleared'))
  }

  return (
    <>
      <header className='flex h-13 shrink-0 items-center justify-between gap-2 px-3 sm:px-5'>
        <div className='flex min-w-0 items-center gap-1.5'>
          <ModelGroupSelector
            selectedModel={props.modelValue}
            models={props.models}
            onModelChange={props.onModelChange}
            selectedGroup={props.groupValue}
            groups={props.groups}
            onGroupChange={props.onGroupChange}
            disabled={isSelectorDisabled}
            popoverSide='bottom'
            popoverAlign='start'
            triggerVariant='ghost'
          />
        </div>

        <div className='flex shrink-0 items-center gap-1'>
          <Tooltip>
            <TooltipTrigger
              render={
                <Button
                  aria-label={t('New chat')}
                  className='text-muted-foreground hover:text-foreground hover:bg-muted/75 h-9 gap-1.5 rounded-xl px-2.5 font-medium'
                  disabled={props.disabled || !props.hasMessages}
                  onClick={() => setClearConfirmOpen(true)}
                  size='sm'
                  variant='ghost'
                >
                  <SquarePenIcon className='size-4' aria-hidden='true' />
                  <span className='hidden text-xs sm:inline'>
                    {t('New chat')}
                  </span>
                </Button>
              }
            />
            <TooltipContent>
              <p>{t('New chat')}</p>
            </TooltipContent>
          </Tooltip>
        </div>
      </header>

      <ConfirmDialog
        destructive
        desc={t(
          'All playground messages saved in this browser will be removed. This cannot be undone.'
        )}
        confirmText={t('Clear')}
        handleConfirm={handleConfirmNewChat}
        open={clearConfirmOpen}
        onOpenChange={setClearConfirmOpen}
        title={t('Clear chat history?')}
      />
    </>
  )
}
