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
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

export interface PlaygroundHeaderProps {
  disabled?: boolean
  hasMessages: boolean
  onClearMessages: () => void
}

export function PlaygroundHeader(props: PlaygroundHeaderProps) {
  const { t } = useTranslation()
  const [clearConfirmOpen, setClearConfirmOpen] = useState(false)

  if (!props.hasMessages) {
    return null
  }

  const handleConfirmNewChat = () => {
    props.onClearMessages()
    setClearConfirmOpen(false)
    toast.success(t('Conversation cleared'))
  }

  return (
    <>
      <header className='flex h-12 shrink-0 items-center justify-end px-3 sm:px-5'>
        <Tooltip>
          <TooltipTrigger
            render={
              <Button
                aria-label={t('New chat')}
                className='text-muted-foreground hover:text-foreground hover:bg-muted/75 h-8 gap-1.5 rounded-xl px-2.5 font-medium'
                disabled={props.disabled}
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
