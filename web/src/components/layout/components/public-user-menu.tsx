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
import { useNavigate } from '@tanstack/react-router'
import { ChevronDown, CreditCard, Key, LogOut, User } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { SignOutDialog } from '@/components/sign-out-dialog'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import useDialogState from '@/hooks/use-dialog'
import { useIsSidebarModuleVisible } from '@/hooks/use-sidebar-config'
import { useUserDisplay } from '@/hooks/use-user-display'
import { getUserAvatarFallback, getUserAvatarStyle } from '@/lib/avatar'
import { useAuthStore } from '@/stores/auth-store'

export function PublicUserMenu() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [signOutOpen, setSignOutOpen] = useDialogState()
  const user = useAuthStore((s) => s.auth.user)
  const { displayName } = useUserDisplay(user)
  const isWalletVisible = useIsSidebarModuleVisible('/wallet')

  const avatarName = user?.username || displayName
  const avatarFallback = getUserAvatarFallback(avatarName)
  const avatarFallbackStyle = useMemo(
    () => getUserAvatarStyle(avatarName),
    [avatarName]
  )

  if (!user) return null

  return (
    <>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger
          render={
            <button
              type='button'
              className='text-foreground group flex h-9 items-center gap-1.5 rounded-full border border-black/[0.08] bg-black/[0.03] pr-2.5 pl-1 text-xs font-medium transition-all hover:bg-black/[0.06] focus-visible:outline-none dark:border-white/[0.1] dark:bg-white/5 dark:hover:bg-white/10'
            />
          }
        >
          <Avatar className='size-6 rounded-full'>
            <AvatarFallback
              className='text-[11px] font-semibold text-white'
              style={avatarFallbackStyle}
            >
              {avatarFallback}
            </AvatarFallback>
          </Avatar>
          <span className='max-w-[80px] truncate text-xs font-medium'>
            {avatarName}
          </span>
          <ChevronDown className='text-muted-foreground size-3 transition-transform duration-200 group-data-[state=open]:rotate-180' />
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align='end'
          sideOffset={8}
          className='w-44 rounded-xl border border-black/[0.08] p-1.5 shadow-xl backdrop-blur-xl dark:border-white/[0.1]'
        >
          <DropdownMenuItem
            onClick={() => navigate({ to: '/profile' })}
            className='flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-xs font-medium'
          >
            <User className='text-muted-foreground size-4' />
            <span>{t('Personal Settings')}</span>
          </DropdownMenuItem>

          <DropdownMenuItem
            onClick={() => navigate({ to: '/keys' })}
            className='flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-xs font-medium'
          >
            <Key className='text-muted-foreground size-4' />
            <span>{t('Token Management')}</span>
          </DropdownMenuItem>

          {isWalletVisible && (
            <DropdownMenuItem
              onClick={() => navigate({ to: '/wallet' })}
              className='flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-xs font-medium'
            >
              <CreditCard className='text-muted-foreground size-4' />
              <span>{t('Wallet Management')}</span>
            </DropdownMenuItem>
          )}

          <DropdownMenuSeparator className='my-1' />

          <DropdownMenuItem
            variant='destructive'
            onClick={() => setSignOutOpen(true)}
            className='flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-xs font-medium'
          >
            <LogOut className='size-4' />
            <span>{t('Sign out')}</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <SignOutDialog open={Boolean(signOutOpen)} onOpenChange={setSignOutOpen} />
    </>
  )
}
