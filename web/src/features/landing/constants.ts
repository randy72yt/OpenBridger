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
import type {
  CodeSnippetItem,
  LandingFaqItem,
  LandingModelHealthItem,
  LandingModelHealthSlot,
  LandingPricingPreviewRow,
  LandingProviderItem,
} from './types'

export const HERO_HIGHLIGHT_PILLS = [
  'landing.hero.pill.multiProtocol',
  'landing.hero.pill.liveProbe',
  'landing.hero.pill.payAsYouGo',
] as const

export const HERO_METRICS = [
  {
    value: '40+',
    labelKey: 'landing.hero.metric.providers',
  },
  {
    value: '99.9%',
    labelKey: 'landing.hero.metric.availability',
  },
  {
    value: '< 120ms',
    labelKey: 'landing.hero.metric.gatewayOverhead',
  },
  {
    value: '100%',
    labelKey: 'landing.hero.metric.transparentPricing',
  },
] as const

export const LANDING_PROVIDERS: LandingProviderItem[] = [
  {
    id: 'openai',
    name: 'OpenAI',
    protocol: '/v1/chat/completions · /v1/responses',
    models: ['gpt-5', 'gpt-4.1', 'o3', 'o4-mini'],
    tagKey: 'landing.multiModel.tag.reasoning',
    iconKey: 'OpenAI',
  },
  {
    id: 'anthropic',
    name: 'Anthropic Claude',
    protocol: '/v1/messages · /v1/chat/completions',
    models: ['claude-opus-4-6', 'claude-sonnet-4-6', 'claude-haiku-4-5'],
    tagKey: 'landing.multiModel.tag.coding',
    iconKey: 'Claude.Color',
  },
  {
    id: 'gemini',
    name: 'Google Gemini',
    protocol: '/v1beta/models · /v1/chat/completions',
    models: ['gemini-2.5-pro', 'gemini-2.5-flash'],
    tagKey: 'landing.multiModel.tag.longContext',
    iconKey: 'Gemini.Color',
  },
  {
    id: 'deepseek',
    name: 'DeepSeek',
    protocol: '/v1/chat/completions',
    models: ['deepseek-r1', 'deepseek-v3'],
    tagKey: 'landing.multiModel.tag.costEffective',
    iconKey: 'DeepSeek.Color',
  },
  {
    id: 'qwen',
    name: 'Alibaba Qwen',
    protocol: '/v1/chat/completions',
    models: ['qwen3-235b-a22b', 'qwen-max', 'qwen-coder-plus'],
    tagKey: 'landing.multiModel.tag.multilingual',
    iconKey: 'Qwen.Color',
  },
  {
    id: 'xai',
    name: 'xAI Grok',
    protocol: '/v1/chat/completions',
    models: ['grok-4', 'grok-3-mini'],
    tagKey: 'landing.multiModel.tag.realtime',
    iconKey: 'Grok',
  },
]

export const SUPPORTED_CLI_TOOLS = [
  { name: 'Claude Code', protocol: 'Anthropic / OpenAI' },
  { name: 'Codex CLI', protocol: 'OpenAI Responses' },
  { name: 'Cursor / Windsurf', protocol: 'OpenAI Chat' },
  { name: 'Cline / Roo Code', protocol: 'OpenAI / Anthropic' },
  { name: 'LangChain / Vercel AI SDK', protocol: 'Multi-SDK' },
] as const

export const FALLBACK_PRICING_ROWS: LandingPricingPreviewRow[] = [
  {
    modelName: 'claude-sonnet-4-6',
    vendorName: 'Anthropic',
    groupName: 'default',
    groupRatio: 1.0,
    inputPerMillionUsd: 3.0,
    outputPerMillionUsd: 15.0,
    cacheReadPerMillionUsd: 0.3,
    tags: ['Coding', 'Agent', '200K'],
  },
  {
    modelName: 'gpt-4.1',
    vendorName: 'OpenAI',
    groupName: 'default',
    groupRatio: 1.0,
    inputPerMillionUsd: 2.0,
    outputPerMillionUsd: 8.0,
    cacheReadPerMillionUsd: 0.5,
    tags: ['General', '1M Context', 'Tools'],
  },
  {
    modelName: 'gemini-2.5-pro',
    vendorName: 'Google',
    groupName: 'default',
    groupRatio: 1.0,
    inputPerMillionUsd: 1.25,
    outputPerMillionUsd: 10.0,
    cacheReadPerMillionUsd: 0.31,
    tags: ['Multimodal', '1M Context', 'Reasoning'],
  },
  {
    modelName: 'deepseek-r1',
    vendorName: 'DeepSeek',
    groupName: 'default',
    groupRatio: 1.0,
    inputPerMillionUsd: 0.55,
    outputPerMillionUsd: 2.19,
    cacheReadPerMillionUsd: 0.14,
    tags: ['Reasoning', 'Math', 'Open Weights'],
  },
  {
    modelName: 'gpt-4.1-mini',
    vendorName: 'OpenAI',
    groupName: 'default',
    groupRatio: 1.0,
    inputPerMillionUsd: 0.4,
    outputPerMillionUsd: 1.6,
    cacheReadPerMillionUsd: 0.1,
    tags: ['Fast', 'High Throughput'],
  },
]

const HEALTHY_24H: LandingModelHealthSlot[] = Array.from(
  { length: 24 },
  (_, index) => ({
    id: `h-${index}`,
    level: 'healthy',
  })
)
const DEGRADED_SLOT_24H: LandingModelHealthSlot[] = Array.from(
  { length: 24 },
  (_, index) => ({
    id: `d-${index}`,
    level: index === 17 ? 'degraded' : 'healthy',
  })
)

export const DEFAULT_MODEL_HEALTH_ITEMS: LandingModelHealthItem[] = [
  {
    modelName: 'claude-sonnet-4-6',
    provider: 'Anthropic',
    groupName: 'default',
    status: 'healthy',
    successRate: 99.94,
    ttftMs: 420,
    tps: 78,
    hourlySlots: HEALTHY_24H,
  },
  {
    modelName: 'gpt-4.1',
    provider: 'OpenAI',
    groupName: 'default',
    status: 'healthy',
    successRate: 99.96,
    ttftMs: 310,
    tps: 94,
    hourlySlots: HEALTHY_24H,
  },
  {
    modelName: 'gemini-2.5-pro',
    provider: 'Google',
    groupName: 'default',
    status: 'healthy',
    successRate: 99.89,
    ttftMs: 480,
    tps: 86,
    hourlySlots: HEALTHY_24H,
  },
  {
    modelName: 'deepseek-r1',
    provider: 'DeepSeek',
    groupName: 'default',
    status: 'degraded',
    successRate: 99.42,
    ttftMs: 690,
    tps: 52,
    hourlySlots: DEGRADED_SLOT_24H,
  },
]

export function getCodeSnippets(baseUrl: string): CodeSnippetItem[] {
  const normalizedBase = baseUrl.replace(/\/+$/, '')
  return [
    {
      id: 'python',
      label: 'Python (OpenAI SDK)',
      filename: 'quickstart.py',
      highlightLines: [4, 5],
      code: [
        'from openai import OpenAI',
        '',
        'client = OpenAI(',
        `    base_url="${normalizedBase}/v1",  # 1. Point to OpenBridger`,
        '    api_key="sk-ob-your-api-key",      # 2. Use your unified key',
        ')',
        '',
        'resp = client.chat.completions.create(',
        '    model="claude-sonnet-4-6",         # 3. Call any supported model',
        '    messages=[{"role": "user", "content": "Hello from OpenBridger!"}],',
        ')',
        'print(resp.choices[0].message.content)',
      ].join('\n'),
    },
    {
      id: 'typescript',
      label: 'TypeScript / Node.js',
      filename: 'quickstart.ts',
      highlightLines: [4, 5],
      code: [
        "import OpenAI from 'openai'",
        '',
        'const client = new OpenAI({',
        `  baseURL: '${normalizedBase}/v1', // 1. Point to OpenBridger`,
        '  apiKey: process.env.OPENBRIDGER_API_KEY, // 2. Unified API key',
        '})',
        '',
        'const response = await client.chat.completions.create({',
        "  model: 'gpt-4.1',                        // 3. Switch models freely",
        "  messages: [{ role: 'user', content: 'Explain streaming SSE in 3 lines.' }],",
        '})',
        'console.log(response.choices[0]?.message.content)',
      ].join('\n'),
    },
    {
      id: 'curl',
      label: 'cURL (HTTP)',
      filename: 'request.sh',
      highlightLines: [1, 3, 5],
      code: [
        `curl ${normalizedBase}/v1/chat/completions \\`,
        '  -H "Content-Type: application/json" \\',
        '  -H "Authorization: Bearer sk-ob-your-api-key" \\',
        "  -d '{",
        '    "model": "gemini-2.5-pro",',
        '    "messages": [{"role": "user", "content": "Ping!"}],',
        '    "stream": true',
        "  }'",
      ].join('\n'),
    },
  ]
}

export const LANDING_FAQ_ITEMS: LandingFaqItem[] = [
  {
    id: 'compatibility',
    questionKey: 'landing.faq.q1.question',
    answerKey: 'landing.faq.q1.answer',
  },
  {
    id: 'billing',
    questionKey: 'landing.faq.q2.question',
    answerKey: 'landing.faq.q2.answer',
  },
  {
    id: 'groups',
    questionKey: 'landing.faq.q3.question',
    answerKey: 'landing.faq.q3.answer',
  },
  {
    id: 'reliability',
    questionKey: 'landing.faq.q4.question',
    answerKey: 'landing.faq.q4.answer',
  },
  {
    id: 'privacy',
    questionKey: 'landing.faq.q5.question',
    answerKey: 'landing.faq.q5.answer',
  },
]
