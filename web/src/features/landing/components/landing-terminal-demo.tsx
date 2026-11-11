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
import { useEffect, useRef, useState, type ReactNode } from 'react'

import { cn } from '@/lib/utils'

interface ApiDemoConfig {
  id: string
  label: string
  method: 'POST' | 'GET'
  endpoint: string
  headers: string[]
  request: string[]
  response: string[]
  tokens: number
  latency: number
}

const API_DEMOS: ApiDemoConfig[] = [
  {
    id: 'gpt-chat',
    label: 'Chat',
    method: 'POST',
    endpoint: '/v1/chat/completions',
    headers: ['"Authorization: Bearer sk-••••"'],
    request: [
      '"model": "your-model",',
      '"messages": [',
      '  { "role": "user", "content": "..." }',
      ']',
    ],
    response: [
      '{',
      '  "choices": [{ "message": { "content": <text> } }],',
      '  "usage": { "total_tokens": <tokens> }',
      '}',
    ],
    tokens: 27,
    latency: 142,
  },
  {
    id: 'responses',
    label: 'Responses',
    method: 'POST',
    endpoint: '/v1/responses',
    headers: ['"Authorization: Bearer sk-••••"'],
    request: ['"model": "your-model",', '"input": "..."'],
    response: [
      '{',
      '  "output": [{ "type": "output_text", "text": <text> }],',
      '  "usage": { "total_tokens": <tokens> }',
      '}',
    ],
    tokens: 31,
    latency: 168,
  },
  {
    id: 'claude',
    label: 'Claude',
    method: 'POST',
    endpoint: '/v1/messages',
    headers: ['"x-api-key: sk-••••"', '"anthropic-version: 2023-06-01"'],
    request: [
      '"model": "your-model",',
      '"max_tokens": 1024,',
      '"messages": [',
      '  { "role": "user", "content": "..." }',
      ']',
    ],
    response: [
      '{',
      '  "content": [{ "type": "text", "text": <text> }],',
      '  "usage": { "input_tokens": <in>, "output_tokens": <out> }',
      '}',
    ],
    tokens: 29,
    latency: 156,
  },
  {
    id: 'gemini',
    label: 'Gemini',
    method: 'POST',
    endpoint: '/v1beta/models/{model}:generateContent',
    headers: ['"x-goog-api-key: sk-••••"'],
    request: [
      '"contents": [',
      '  { "role": "user",',
      '    "parts": [{ "text": "..." }] }',
      ']',
    ],
    response: [
      '{',
      '  "candidates": [{ "content": { "parts": [{ "text": <text> }] } }],',
      '  "usageMetadata": { "totalTokenCount": <tokens> }',
      '}',
    ],
    tokens: 25,
    latency: 93,
  },
]

const CYCLE_INTERVAL = 4500
const TRANSITION_MS = 220
const STRING_RE = /"[^"]*"/g
const PLACEHOLDER_RE = /<[a-z]+>/gi

interface LandingTerminalDemoProps {
  className?: string
}

export function LandingTerminalDemo(props: LandingTerminalDemoProps) {
  const [activeIndex, setActiveIndex] = useState(0)
  const [transitioning, setTransitioning] = useState(false)
  const intervalRef = useRef<ReturnType<typeof setInterval>>(undefined)
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
    if (mq.matches) return

    intervalRef.current = setInterval(() => {
      setTransitioning(true)
      timeoutRef.current = setTimeout(() => {
        setActiveIndex((prev) => (prev + 1) % API_DEMOS.length)
        setTransitioning(false)
      }, TRANSITION_MS)
    }, CYCLE_INTERVAL)

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
      if (timeoutRef.current) clearTimeout(timeoutRef.current)
    }
  }, [])

  const handleSelect = (index: number) => {
    if (index === activeIndex) return
    if (intervalRef.current) clearInterval(intervalRef.current)
    if (timeoutRef.current) clearTimeout(timeoutRef.current)
    setTransitioning(true)
    timeoutRef.current = setTimeout(() => {
      setActiveIndex(index)
      setTransitioning(false)
    }, TRANSITION_MS)
  }

  const demo = API_DEMOS[activeIndex]

  return (
    <div className={cn('mx-auto w-full max-w-2xl', props.className)}>
      <div className='ob-code-card overflow-hidden rounded-2xl'>
        {/* Tab strip */}
        <div
          className='flex items-center gap-1 border-b px-2 sm:gap-1.5 sm:px-3'
          style={{ borderColor: 'var(--ob-code-line)' }}
        >
          {API_DEMOS.map((item, index) => {
            const isActive = index === activeIndex
            return (
              <button
                key={item.id}
                type='button'
                onClick={() => handleSelect(index)}
                className={cn(
                  'relative -mb-px flex items-center gap-1.5 border-b-2 px-2.5 py-2.5 text-[11px] font-medium tracking-wide transition-colors sm:px-3 sm:text-xs',
                  isActive
                    ? 'border-[var(--ob-primary)] text-[var(--ob-primary)]'
                    : 'border-transparent text-[var(--ob-iron)] hover:text-[var(--ob-code-ink)]'
                )}
              >
                {item.label}
              </button>
            )
          })}
          <div className='ml-auto flex items-center gap-2 pr-2 sm:pr-3'>
            <span className='ob-dot-success inline-block size-1.5 rounded-full' />
            <span className='ob-mono ob-code-success text-[10px] font-semibold tracking-wider uppercase'>
              200 OK
            </span>
          </div>
        </div>

        {/* Endpoint row */}
        <div
          className='flex items-center gap-2.5 border-b px-5 py-3'
          style={{ borderColor: 'var(--ob-code-line)' }}
        >
          <span
            className='ob-mono rounded-md px-1.5 py-0.5 text-[10px] font-semibold tracking-wider'
            style={{
              backgroundColor: 'rgba(255, 107, 0, 0.16)',
              color: 'var(--ob-primary)',
            }}
          >
            {demo.method}
          </span>
          <code
            className={cn(
              'ob-mono truncate text-[12.5px] transition-opacity duration-200',
              transitioning ? 'opacity-0' : 'opacity-100'
            )}
            style={{ color: 'var(--ob-code-ink)' }}
          >
            {demo.endpoint}
          </code>
        </div>

        {/* Body */}
        <div className='ob-mono grid h-[400px] grid-rows-[235px_minmax(0,1fr)] text-[12.5px] leading-[1.55]'>
          {/* Request */}
          <div className='relative px-5 py-4'>
            <span className='ob-code-iron font-sans text-[10px] font-semibold tracking-[0.18em] uppercase'>
              Request
            </span>
            <div
              className={cn(
                'mt-2 transition-opacity duration-200',
                transitioning ? 'opacity-0' : 'opacity-100'
              )}
            >
              <CodeLine>
                <span className='ob-code-kw font-semibold'>curl</span>{' '}
                <span className='ob-code-kw'>-X</span>{' '}
                <span className='ob-code-kw'>POST</span>{' '}
                <span className='ob-code-str'>&quot;{demo.endpoint}&quot;</span>{' '}
                <span className='ob-code-iron'>{'\\'}</span>
              </CodeLine>
              {demo.headers.map((header) => (
                <CodeLine key={header} indent={2}>
                  <span className='ob-code-kw'>-H</span>{' '}
                  <span className='ob-code-str'>{header}</span>{' '}
                  <span className='ob-code-iron'>{'\\'}</span>
                </CodeLine>
              ))}
              <CodeLine indent={2}>
                <span className='ob-code-kw'>-d</span>{' '}
                <span className='ob-code-str'>&apos;{'{'}</span>
              </CodeLine>
              {demo.request.map((line) => (
                <CodeLine key={line} indent={4}>
                  {tokenize(line)}
                </CodeLine>
              ))}
              <CodeLine indent={2}>
                <span className='ob-code-str'>{'}'}&apos;</span>
              </CodeLine>
            </div>
          </div>

          {/* Response */}
          <div
            className='relative border-t px-5 py-4'
            style={{
              borderColor: 'var(--ob-code-line)',
              backgroundColor: 'rgba(255, 255, 255, 0.02)',
            }}
          >
            <span className='ob-code-iron font-sans text-[10px] font-semibold tracking-[0.18em] uppercase'>
              Response
            </span>
            <div
              className={cn(
                'mt-2 transition-opacity duration-200',
                transitioning ? 'opacity-0' : 'opacity-100'
              )}
            >
              {demo.response.map((line) => (
                <CodeLine key={line}>{renderResponseLine(line, demo)}</CodeLine>
              ))}
            </div>
          </div>
        </div>

        {/* Footer metrics */}
        <div
          className='ob-code-header flex items-center justify-between border-t px-5 py-2.5'
          style={{ borderColor: 'var(--ob-code-line)' }}
        >
          <div className='ob-code-iron flex items-center gap-3 text-[10px] tabular-nums'>
            <span className='flex items-center gap-1'>
              <span className='ob-mono'>{demo.latency}</span>
              <span className='tracking-wider uppercase'>ms</span>
            </span>
            <span className='size-1 rounded-full bg-[var(--ob-iron)] opacity-50' />
            <span className='flex items-center gap-1'>
              <span className='ob-mono'>{demo.tokens}</span>
              <span className='tracking-wider uppercase'>tokens</span>
            </span>
            <span className='size-1 rounded-full bg-[var(--ob-iron)] opacity-50' />
            <span className='flex items-center gap-1'>
              <span className='tracking-wider uppercase'>cost</span>
              <span className='ob-mono'>
                ${(demo.tokens * 0.00003).toFixed(5)}
              </span>
            </span>
          </div>
          <span className='ob-mono ob-code-iron text-[10px] tracking-wider uppercase'>
            stream · sse
          </span>
        </div>
      </div>
    </div>
  )
}

function truncateResponse(demo: ApiDemoConfig): string {
  const map: Record<string, string> = {
    'gpt-chat': 'Chat request routed.',
    responses: 'Response workflow ready.',
    claude: 'Claude message routed.',
    gemini: 'Gemini request served.',
  }
  return map[demo.id] ?? '...'
}

function tokenize(input: string): ReactNode {
  const segments: ReactNode[] = []
  let cursor = 0
  const matches = [...input.matchAll(STRING_RE)]

  matches.forEach((match) => {
    const start = match.index ?? 0
    if (start > cursor) {
      segments.push(
        <span key={`m-${start}`} className='ob-code-iron'>
          {input.slice(cursor, start)}
        </span>
      )
    }
    const text = match[0]
    const after = input.slice(start + text.length).trimStart()
    const isKey = after.startsWith(':')
    if (isKey) {
      segments.push(
        <span key={`k-${start}`} className='ob-code-kw font-medium'>
          {text}
        </span>
      )
    } else {
      segments.push(
        <span key={`s-${start}`} className='ob-code-str'>
          {text}
        </span>
      )
    }
    cursor = start + text.length
  })

  if (cursor < input.length) {
    segments.push(
      <span key={`tail-${cursor}`} className='ob-code-iron'>
        {input.slice(cursor)}
      </span>
    )
  }

  return segments
}

function renderResponseLine(line: string, demo: ApiDemoConfig): ReactNode {
  if (!line.trim()) return <span className='ob-code-iron'> </span>

  const segments: ReactNode[] = []
  let cursor = 0
  const matches = [...line.matchAll(PLACEHOLDER_RE)]

  if (matches.length === 0) return tokenize(line)

  matches.forEach((match) => {
    const start = match.index ?? 0
    if (start > cursor) {
      segments.push(
        <span key={`pre-${start}`}>{tokenize(line.slice(cursor, start))}</span>
      )
    }
    const placeholder = match[0]
    if (placeholder === '<text>') {
      segments.push(
        <span key={`ph-text-${start}`} className='ob-code-str font-medium'>
          {`"${truncateResponse(demo)}"`}
        </span>
      )
    } else if (placeholder === '<tokens>') {
      segments.push(
        <span key={`ph-tokens-${start}`} className='ob-code-num font-semibold'>
          {demo.tokens}
        </span>
      )
    } else if (placeholder === '<in>') {
      segments.push(
        <span key={`ph-in-${start}`} className='ob-code-num font-semibold'>
          {Math.floor(demo.tokens * 0.4)}
        </span>
      )
    } else if (placeholder === '<out>') {
      segments.push(
        <span key={`ph-out-${start}`} className='ob-code-num font-semibold'>
          {Math.ceil(demo.tokens * 0.6)}
        </span>
      )
    } else {
      segments.push(
        <span key={`ph-other-${start}`} className='ob-code-iron'>
          {placeholder}
        </span>
      )
    }
    cursor = start + placeholder.length
  })

  if (cursor < line.length) {
    segments.push(
      <span key={`tail-${cursor}`}>{tokenize(line.slice(cursor))}</span>
    )
  }

  return segments
}

function CodeLine(props: { children: ReactNode; indent?: number }) {
  return (
    <div className='break-words whitespace-pre-wrap'>
      {props.indent ? (
        <span
          aria-hidden='true'
          className='inline-block'
          style={{ width: `${props.indent}ch` }}
        />
      ) : null}
      {props.children}
    </div>
  )
}
