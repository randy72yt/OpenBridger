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
import { Link } from '@tanstack/react-router'
import {
  ArrowUpRight,
  BookOpen,
  Code2,
  GraduationCap,
  KeyRound,
  Rocket,
  Terminal,
} from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { PublicLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { useStatus } from '@/hooks/use-status'

const DEFAULT_DOCS_ORIGIN = 'https://docs.openbridger.com'

export function Tutorials() {
  const { t, i18n } = useTranslation()
  const { status } = useStatus()

  const apiOrigin = useMemo(() => {
    const raw = status?.server_address as string | undefined
    if (raw && raw.trim()) {
      return raw.trim().replace(/\/+$/, '')
    }
    if (typeof window !== 'undefined' && window.location?.origin) {
      return window.location.origin.replace(/\/+$/, '')
    }
    return 'https://openbridger.com'
  }, [status?.server_address])

  const docsLangPrefix = useMemo(() => {
    const lang = (i18n.resolvedLanguage || i18n.language || 'zh').toLowerCase()
    return lang.startsWith('en') ? '/en' : '/zh'
  }, [i18n.language, i18n.resolvedLanguage])

  const docsOrigin = useMemo(() => {
    const raw = status?.docs_link as string | undefined
    if (raw && /^https?:\/\//i.test(raw.trim())) {
      return raw
        .trim()
        .replace(/\/+$/, '')
        .replace(/\/(?:zh|en)$/i, '')
    }
    return DEFAULT_DOCS_ORIGIN
  }, [status?.docs_link])

  const buildDocHref = (slugPath: string) =>
    `${docsOrigin}${docsLangPrefix}${slugPath}`

  const endpointFormats = [
    {
      id: 'openai',
      badge: 'OpenAI / Codex / SDK',
      label: t('docsHub.url.openaiLabel'),
      value: `${apiOrigin}/v1`,
      hint: t('docsHub.url.openaiHint'),
    },
    {
      id: 'native',
      badge: 'Claude Code / Gemini CLI',
      label: t('docsHub.url.nativeLabel'),
      value: apiOrigin,
      hint: t('docsHub.url.nativeHint'),
    },
    {
      id: 'full',
      badge: 'HTTP / Translation',
      label: t('docsHub.url.fullLabel'),
      value: `${apiOrigin}/v1/chat/completions`,
      hint: t('docsHub.url.fullHint'),
    },
  ]

  const docGroups = [
    {
      id: 'quickstart',
      icon: Rocket,
      title: t('docsHub.group.quickstart.title'),
      description: t('docsHub.group.quickstart.desc'),
      items: [
        { label: t('docsHub.link.quickstart'), path: '/guide/quickstart/' },
        { label: t('docsHub.link.baseUrls'), path: '/guide/base-urls/' },
        { label: t('docsHub.link.apiKey'), path: '/guide/api-key/' },
        { label: t('docsHub.link.errors'), path: '/guide/errors/' },
      ],
    },
    {
      id: 'coding',
      icon: Terminal,
      title: t('docsHub.group.coding.title'),
      description: t('docsHub.group.coding.desc'),
      items: [
        { label: 'OpenAI Codex', path: '/guide/clients/codex-cli/' },
        { label: 'Claude Code', path: '/guide/clients/claude-code/' },
        { label: 'CC Switch', path: '/guide/clients/cc-switch/' },
        { label: 'Cursor', path: '/guide/clients/cursor/' },
        { label: 'Cline / Roo Code', path: '/guide/clients/cline/' },
        { label: 'Gemini CLI', path: '/guide/clients/gemini-cli/' },
      ],
    },
    {
      id: 'research',
      icon: GraduationCap,
      title: t('docsHub.group.research.title'),
      description: t('docsHub.group.research.desc'),
      items: [
        {
          label: t('docsHub.link.immersiveTranslate'),
          path: '/guide/clients/immersive-translate/',
        },
        { label: t('docsHub.link.zotero'), path: '/guide/clients/zotero/' },
        { label: 'Cherry Studio', path: '/guide/clients/cherry-studio/' },
        { label: 'Chatbox / LobeChat', path: '/guide/clients/chatbox/' },
        {
          label: t('docsHub.link.workflows'),
          path: '/guide/clients/dify-workflows/',
        },
      ],
    },
    {
      id: 'api',
      icon: Code2,
      title: t('docsHub.group.api.title'),
      description: t('docsHub.group.api.desc'),
      items: [
        {
          label: 'POST /v1/chat/completions',
          path: '/api/openai-chat/',
        },
        {
          label: 'POST /v1/responses (Codex)',
          path: '/api/openai-responses/',
        },
        {
          label: 'POST /v1/messages (Claude)',
          path: '/api/anthropic-messages/',
        },
        {
          label: 'POST /v1beta/models/* (Gemini)',
          path: '/api/gemini-native/',
        },
        {
          label: t('docsHub.link.multimodalApi'),
          path: '/api/multimodal-media/',
        },
      ],
    },
  ]

  return (
    <PublicLayout showMainContainer={false}>
      <main className='ob-landing-theme ob-section-bg min-h-[75vh] pt-28 pb-20 sm:pt-32'>
        <div className='mx-auto max-w-[92rem] space-y-12 px-4 sm:px-6 lg:px-8'>
          {/* Hero Banner */}
          <div className='flex flex-col items-start justify-between gap-6 lg:flex-row lg:items-end'>
            <div className='max-w-3xl space-y-3'>
              <div className='ob-chip-primary inline-flex items-center gap-1.5 rounded-full px-3.5 py-1 text-xs font-semibold'>
                <BookOpen className='size-3.5' />
                <span>{t('Practical tutorials')}</span>
              </div>
              <h1 className='ob-serif-heading text-3xl font-bold tracking-tight sm:text-4xl'>
                {t('docsHub.hero.title')}
              </h1>
              <p className='text-muted-foreground text-sm leading-relaxed sm:text-base'>
                {t('docsHub.hero.subtitle')}
              </p>
            </div>

            <div className='flex flex-wrap items-center gap-3'>
              <Button
                size='lg'
                className='ob-btn-primary h-11 rounded-full px-6 text-sm font-semibold'
                render={
                  <a
                    href={buildDocHref('/')}
                    target='_blank'
                    rel='noopener noreferrer'
                  />
                }
              >
                {t('docsHub.hero.openFullDocs')}
                <ArrowUpRight className='ml-1.5 size-4' />
              </Button>
              <Button
                variant='outline'
                size='lg'
                className='ob-btn-secondary h-11 rounded-full px-5 text-sm font-semibold'
                render={<Link to='/keys' />}
              >
                <KeyRound className='mr-1.5 size-4' />
                {t('API Keys')}
              </Button>
            </div>
          </div>

          {/* 3 Base URL Quick-Copy Cards */}
          <div className='grid grid-cols-1 gap-4 md:grid-cols-3'>
            {endpointFormats.map((item) => (
              <Card
                key={item.id}
                className='ob-card rounded-2xl py-0 shadow-none'
              >
                <CardContent className='space-y-2.5 p-5'>
                  <div className='flex items-center justify-between gap-2'>
                    <span className='text-xs font-semibold'>{item.label}</span>
                    <Badge
                      variant='secondary'
                      className='ob-chip-primary ob-mono text-[10px]'
                    >
                      {item.badge}
                    </Badge>
                  </div>
                  <div className='ob-section-warm border-border flex items-center justify-between gap-2 rounded-xl border px-3 py-2'>
                    <code className='ob-mono truncate text-xs font-medium'>
                      {item.value}
                    </code>
                    <CopyButton value={item.value} />
                  </div>
                  <p className='text-muted-foreground text-xs leading-relaxed'>
                    {item.hint}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* 4 Categorized Guide & API Reference Cards */}
          <div className='grid grid-cols-1 gap-5 md:grid-cols-2'>
            {docGroups.map((group) => {
              const Icon = group.icon
              return (
                <Card
                  key={group.id}
                  className='ob-card ob-card-interactive rounded-2xl py-0 shadow-none'
                >
                  <CardHeader className='flex flex-row items-start gap-3.5 p-6 pb-3'>
                    <div className='ob-chip-primary flex size-10 shrink-0 items-center justify-center rounded-xl'>
                      <Icon className='size-5' />
                    </div>
                    <div className='space-y-1'>
                      <h2 className='text-lg font-semibold'>{group.title}</h2>
                      <p className='text-muted-foreground text-xs leading-relaxed sm:text-sm'>
                        {group.description}
                      </p>
                    </div>
                  </CardHeader>
                  <CardContent className='grid grid-cols-1 gap-2 p-6 pt-2 sm:grid-cols-2'>
                    {group.items.map((doc) => (
                      <a
                        key={doc.path}
                        href={buildDocHref(doc.path)}
                        target='_blank'
                        rel='noopener noreferrer'
                        className='ob-section-warm border-border hover:border-primary/40 hover:text-primary group flex items-center justify-between gap-2 rounded-xl border px-3.5 py-2.5 text-xs font-medium transition-colors sm:text-sm'
                      >
                        <span className='truncate'>{doc.label}</span>
                        <ArrowUpRight className='text-muted-foreground group-hover:text-primary size-3.5 shrink-0 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5' />
                      </a>
                    ))}
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </div>
      </main>
    </PublicLayout>
  )
}
