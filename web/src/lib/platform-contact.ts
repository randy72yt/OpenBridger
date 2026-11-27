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
import type { SystemStatus } from '@/features/auth/types'

export const DEFAULT_SUPPORT_EMAIL = 'support@openbridger.com'

export const PLATFORM_CONTACT = {
  email: DEFAULT_SUPPORT_EMAIL,
  wechatId: '',
  wechatQrCode: '',
}

export interface ResolvedPlatformContact {
  email: string
  wechatId: string
  wechatQrCode: string
}

function pickStatusString(
  status: SystemStatus | null | undefined,
  key: 'contact_email' | 'contact_wechat_id' | 'contact_wechat_qrcode'
): string | undefined {
  const directValue = status?.[key]
  if (typeof directValue === 'string') {
    return directValue.trim()
  }

  const nestedValue = status?.data?.[key]
  if (typeof nestedValue === 'string') {
    return nestedValue.trim()
  }

  return undefined
}

export function resolvePlatformContact(
  status?: SystemStatus | null
): ResolvedPlatformContact {
  const configuredEmail = pickStatusString(status, 'contact_email')
  const configuredWeChatId = pickStatusString(status, 'contact_wechat_id')
  const configuredWeChatQr = pickStatusString(status, 'contact_wechat_qrcode')

  return {
    email:
      configuredEmail !== undefined ? configuredEmail : PLATFORM_CONTACT.email,
    wechatId: configuredWeChatId ?? PLATFORM_CONTACT.wechatId,
    wechatQrCode: configuredWeChatQr ?? PLATFORM_CONTACT.wechatQrCode,
  }
}

export function isQrImageSource(raw: string): boolean {
  const trimmed = raw.trim().toLowerCase()
  if (!trimmed) {
    return false
  }

  if (trimmed.startsWith('data:image/') || trimmed.startsWith('/')) {
    return true
  }

  return /\.(png|jpe?g|webp|gif)(\?.*)?$/i.test(trimmed)
}
