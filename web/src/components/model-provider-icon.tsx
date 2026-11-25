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
import { SparklesIcon } from 'lucide-react'

import { getLobeIcon } from '@/lib/lobe-icon'
import { resolveModelProvider } from '@/lib/model-provider'
import { cn } from '@/lib/utils'

export interface ModelProviderIconProps {
  modelName?: string | null
  size?: number
  monochrome?: boolean
  fallbackToOpenAI?: boolean
  className?: string
}

export function ModelProviderIcon(props: ModelProviderIconProps) {
  const size = props.size ?? 18
  const provider = resolveModelProvider(props.modelName)

  if (provider) {
    const iconKey = props.monochrome ? provider.monoIcon : provider.icon
    return (
      <span
        className={cn(
          'inline-flex shrink-0 items-center justify-center',
          props.className
        )}
        data-provider={provider.label}
        aria-hidden='true'
      >
        {getLobeIcon(iconKey, size)}
      </span>
    )
  }

  if (props.fallbackToOpenAI) {
    return (
      <span
        className={cn(
          'inline-flex shrink-0 items-center justify-center',
          props.className
        )}
        data-provider='OpenAI'
        aria-hidden='true'
      >
        {getLobeIcon('OpenAI', size)}
      </span>
    )
  }

  return (
    <span
      className={cn(
        'text-muted-foreground inline-flex shrink-0 items-center justify-center',
        props.className
      )}
      data-provider='Generic'
      aria-hidden='true'
    >
      <SparklesIcon style={{ width: size * 0.85, height: size * 0.85 }} />
    </span>
  )
}
