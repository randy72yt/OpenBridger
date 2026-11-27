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
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import {
  Conversation,
  ConversationContent,
  ConversationScrollButton,
} from '@/components/ai-elements/conversation'
import { Loader } from '@/components/ai-elements/loader'
import { Message } from '@/components/ai-elements/message'
import { ModelProviderIcon } from '@/components/model-provider-icon'

import { MESSAGE_ROLES } from '../../constants'
import {
  getChatMessageRenderState,
  getEditingMessageContent,
  getMessageAlignment,
  getPreviousUserMessage,
  isErrorMessage,
} from '../../lib'
import type {
  Message as MessageType,
  PlaygroundMessageLayoutMode,
} from '../../types'
import { MessageActions } from '../message/message-actions'
import { MessageErrorActions } from '../message/message-error-actions'
import { PlaygroundMessageContent } from '../message/playground-message-content'
import { PlaygroundMessageEditor } from '../message/playground-message-editor'
import { PlaygroundEmptyState } from './playground-empty-state'

const MAX_RENDERED_HISTORY_MESSAGES = 24

interface PlaygroundChatProps {
  currentModel?: string
  messages: MessageType[]
  onCopyMessage?: (message: MessageType) => void
  onRegenerateMessage?: (message: MessageType) => void
  onEditMessage?: (message: MessageType) => void
  onDeleteMessage?: (message: MessageType) => void
  onSelectPrompt?: (prompt: string) => void
  isGenerating?: boolean
  isLoadingMessages?: boolean
  editingKey?: string | null
  onSaveEdit?: (newContent: string) => void
  onCancelEdit?: (open: boolean) => void
  onSaveEditAndSubmit?: (newContent: string) => void
  messageLayoutMode?: PlaygroundMessageLayoutMode
}

export function PlaygroundChat(props: PlaygroundChatProps) {
  const { t } = useTranslation()
  const [editText, setEditText] = useState('')
  const [originalText, setOriginalText] = useState('')
  const [sourceMessageKeys, setSourceMessageKeys] = useState<
    ReadonlySet<string>
  >(() => new Set())
  const messageLayoutMode = props.messageLayoutMode ?? 'alternating'
  const isGenerating = props.isGenerating ?? false
  const isLoadingMessages = props.isLoadingMessages ?? false
  const visibleMessageOffset = Math.max(
    0,
    props.messages.length - MAX_RENDERED_HISTORY_MESSAGES
  )
  const visibleMessages = props.messages.slice(visibleMessageOffset)

  function handleToggleMessageSource(message: MessageType): void {
    setSourceMessageKeys((currentKeys) => {
      const nextKeys = new Set(currentKeys)

      if (nextKeys.has(message.key)) {
        nextKeys.delete(message.key)
      } else {
        nextKeys.add(message.key)
      }

      return nextKeys
    })
  }

  useEffect(() => {
    if (!props.editingKey) return
    const content = getEditingMessageContent(props.messages, props.editingKey)
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setEditText(content)

    setOriginalText(content)
  }, [props.editingKey, props.messages])

  let chatContent = visibleMessages.map((message, visibleMessageIndex) => {
    const messageIndex = visibleMessageOffset + visibleMessageIndex
    const { alwaysShowActions, content, isEditing } = getChatMessageRenderState(
      props.messages,
      message,
      messageIndex,
      props.editingKey
    )
    const isError = isErrorMessage(message)
    const previousUserMessage = isError
      ? getPreviousUserMessage(props.messages, messageIndex)
      : null
    const alignment = getMessageAlignment(message, messageLayoutMode)
    const isSourceVisible = sourceMessageKeys.has(message.key)
    const isAssistant = message.from === MESSAGE_ROLES.ASSISTANT
    const effectiveModelName = message.model || props.currentModel

    return (
      <Message
        className='group items-start gap-3 py-3.5'
        from={message.from}
        key={message.key}
      >
        <div className='min-w-0 flex-1'>
          {isEditing ? (
            <PlaygroundMessageEditor
              editText={editText}
              message={message}
              onCancelEdit={props.onCancelEdit}
              onEditTextChange={setEditText}
              onSaveEdit={props.onSaveEdit}
              onSaveEditAndSubmit={props.onSaveEditAndSubmit}
              originalText={originalText}
            />
          ) : (
            <PlaygroundMessageContent
              alignment={alignment}
              actions={
                <MessageActions
                  message={message}
                  onCopy={props.onCopyMessage}
                  onRegenerate={props.onRegenerateMessage}
                  onToggleSource={handleToggleMessageSource}
                  onEdit={props.onEditMessage}
                  onDelete={props.onDeleteMessage}
                  isSourceVisible={isSourceVisible}
                  isGenerating={isGenerating}
                  alwaysVisible={alwaysShowActions}
                  className='mt-1.5'
                />
              }
              isSourceVisible={isSourceVisible}
              message={message}
              errorActions={
                isError ? (
                  <MessageErrorActions
                    disabled={isGenerating}
                    onRetry={
                      props.onRegenerateMessage
                        ? () => props.onRegenerateMessage?.(message)
                        : undefined
                    }
                    onEditPrompt={
                      props.onEditMessage && previousUserMessage
                        ? () => props.onEditMessage?.(previousUserMessage)
                        : undefined
                    }
                    onDelete={
                      props.onDeleteMessage
                        ? () => props.onDeleteMessage?.(message)
                        : undefined
                    }
                  />
                ) : undefined
              }
              versionContent={content}
            />
          )}
        </div>

        {isAssistant && (
          <div
            className='border-border/70 bg-card mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full border shadow-2xs'
            data-testid='assistant-model-avatar'
          >
            <ModelProviderIcon
              modelName={effectiveModelName}
              size={15}
              fallbackToOpenAI={!effectiveModelName}
            />
          </div>
        )}
      </Message>
    )
  })

  if (visibleMessages.length === 0 && props.onSelectPrompt) {
    chatContent = [
      <PlaygroundEmptyState
        key='empty'
        modelName={props.currentModel}
        onSelectPrompt={props.onSelectPrompt}
      />,
    ]
  }

  if (isLoadingMessages) {
    chatContent = [
      <div
        className='text-muted-foreground flex min-h-[min(520px,calc(100svh-18rem))] items-center justify-center gap-2 text-sm'
        key='loading'
      >
        <Loader />
        <span>{t('Loading conversation...')}</span>
      </div>,
    ]
  }

  return (
    <Conversation>
      <ConversationContent className='p-0'>
        <div className='mx-auto w-full max-w-3xl px-4 py-4'>{chatContent}</div>
      </ConversationContent>
      <ConversationScrollButton />
    </Conversation>
  )
}
