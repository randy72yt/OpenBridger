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
import { useMemo } from 'react'

import { usePricingData } from '@/features/pricing/hooks/use-pricing-data'
import { useStatus } from '@/hooks/use-status'

import {
  DEFAULT_MODEL_HEALTH_ITEMS,
  LANDING_PROVIDERS,
  getCodeSnippets,
} from '../constants'
import type {
  CodeSnippetItem,
  LandingModelHealthItem,
  LandingProviderItem,
} from '../types'

type ProviderBucketId =
  | 'openai'
  | 'anthropic'
  | 'gemini'
  | 'deepseek'
  | 'qwen'
  | 'xai'

export interface LiveModelInput {
  model_name: string
  enable_groups?: string[]
}

export interface LandingShowcaseData {
  heroModelsTagline: string | null
  providers: LandingProviderItem[]
  healthItems: LandingModelHealthItem[]
  terminalModels: {
    chatLabel: string
    chatModel: string
    codexModel: string
    claudeLabel: string
    claudeModel: string
    geminiLabel: string
    geminiModel: string
  }
  getSnippets: (baseUrl: string) => CodeSnippetItem[]
}

const EXCLUDED_MODEL_PATTERNS = [
  'embedding',
  'moderation',
  'rerank',
  'bge-',
  'gte-',
  'whisper',
  'tts',
  'transcribe',
  'dall-e-2',
  'davinci',
  'babbage',
  'curie',
  'text-ada',
  'gizmo',
  'aqa',
]

function isShowcaseCandidate(rawName: string): boolean {
  const lower = rawName.trim().toLowerCase()
  if (!lower || lower.includes('*')) return false
  return !EXCLUDED_MODEL_PATTERNS.some((kw) => lower.includes(kw))
}

function classifyProvider(rawName: string): ProviderBucketId | null {
  const lower = rawName.trim().toLowerCase()
  const base = lower.includes('/') ? (lower.split('/').pop() ?? lower) : lower

  if (
    base.startsWith('gpt-') ||
    base.startsWith('chatgpt-') ||
    base.includes('codex') ||
    base.startsWith('sora') ||
    /^o[1-9](?:-|$)/.test(base)
  ) {
    return 'openai'
  }
  if (base.includes('claude')) return 'anthropic'
  if (base.startsWith('gemini') || base.startsWith('veo-')) return 'gemini'
  if (base.includes('deepseek')) return 'deepseek'
  if (base.startsWith('qwen') || base.startsWith('qwq')) return 'qwen'
  if (base.startsWith('grok')) return 'xai'
  return null
}

function canonicalModelStem(rawName: string): string {
  const lower = rawName.trim().toLowerCase()
  const withoutOrg = lower.includes('/')
    ? (lower.split('/').pop() ?? lower)
    : lower
  return withoutOrg
    .replace(/@.*$/, '')
    .replace(/-\d{4}-\d{2}-\d{2}$/, '')
    .replace(/-\d{8}$/, '')
    .replaceAll(
      /-(?:thinking|xhigh|high|medium|low|none|latest|preview|beta|search|fast-reasoning|fast-non-reasoning)(?=-|$)/g,
      ''
    )
}

function extractVersionScore(rawName: string): number {
  const lower = rawName.trim().toLowerCase()
  const withoutOrg = lower.includes('/')
    ? (lower.split('/').pop() ?? lower)
    : lower

  // Strip parameter counts (e.g. 235b, 32b, 70b), context sizes (128k), and date stamps
  const cleaned = withoutOrg
    .replace(/@.*$/, '')
    .replaceAll(/-\d{4}-\d{2}-\d{2}/g, '')
    .replaceAll(/-\d{8}/g, '')
    .replaceAll(/-(?:0[1-9]|1[0-2])\d{2}(?=-|$)/g, '')
    .replaceAll(/-\d+(?:\.\d+)?[bk]\b/g, '')

  // Anthropic style: claude-opus-5-5, claude-fable-5-2, claude-3-7-sonnet
  const claudeDashVer =
    /claude-(?:[a-z]+-)?(\d+)-(\d+)(?!\d)/.exec(cleaned) ??
    /claude-(\d+)(?!\d)/.exec(cleaned)
  if (claudeDashVer) {
    const major = Number(claudeDashVer[1])
    const minor = claudeDashVer[2] ? Number(claudeDashVer[2]) : 0
    return major + minor / 10
  }

  // Grok / general dash-minor style: grok-4-1-fast -> 4.1
  const grokDashVer = /grok-(\d+)-(\d+)(?!\d)/.exec(cleaned)
  if (grokDashVer) {
    return Number(grokDashVer[1]) + Number(grokDashVer[2]) / 10
  }

  // OpenAI o-series reasoning models (o3, o4-mini -> map into modern GPT-5+ era)
  const oSeries = /^o([1-9])(?:-|$)/.exec(cleaned)
  if (oSeries) {
    return 4 + Number(oSeries[1]) * 0.4
  }

  // Standard dotted or integer version: gpt-6.2, gemini-4, deepseek-v4.1, qwen4-max
  const verMatch = /v?(\d+)(?:\.(\d+))?/.exec(cleaned)
  if (verMatch) {
    const major = Number(verMatch[1])
    const minor = verMatch[2] ? Number(verMatch[2]) : 0
    return major + minor / 10
  }

  // Unversioned aliases like deepseek-reasoner / deepseek-chat
  if (cleaned.includes('reasoner') || cleaned.includes('r1')) return 3.8
  return 1
}

function scoreModelFreshness(rawName: string): number {
  const lower = rawName.trim().toLowerCase()
  let score = extractVersionScore(lower) * 10

  // Prefer clean canonical model names over dated snapshots, org-prefixed, or effort variants
  const hasDateStamp = /-\d{4}-\d{2}-\d{2}|-\d{8}/.test(lower)
  const hasEffortSuffix =
    /@|-(?:thinking|xhigh|high|medium|low|none)(?=-|$)/.test(lower)
  if (!hasDateStamp && !hasEffortSuffix && !lower.includes('/')) {
    score += 0.35
  }

  // Prefer flagship/coding tiers over lightweight tiers within the same generation
  if (
    /(?:astra|opus|fable|carbon|max|pro|ultra|codex)/.test(lower) &&
    !lower.includes('image')
  ) {
    score += 0.12
  } else if (/(?:sol|sonnet|neon|argon|plus|reasoner)/.test(lower)) {
    score += 0.08
  } else if (/(?:flash|haiku|hydrogen|mini)/.test(lower)) {
    score += 0.03
  } else if (/(?:nano|lite)/.test(lower)) {
    score -= 0.05
  }

  // Prefer primary text/multimodal chat endpoints over video/image-only models for top slot
  if (/(?:sora|veo-|image|imagine)/.test(lower)) {
    score -= 1.5
  }

  return score
}

function selectTopProviderModels(
  candidates: LiveModelInput[],
  limit: number
): LiveModelInput[] {
  const sorted = [...candidates].sort((a, b) => {
    const diff =
      scoreModelFreshness(b.model_name) - scoreModelFreshness(a.model_name)
    if (Math.abs(diff) > 1e-6) return diff
    return a.model_name.length - b.model_name.length
  })

  const picked: LiveModelInput[] = []
  const seenStems = new Set<string>()

  for (const item of sorted) {
    const stem = canonicalModelStem(item.model_name)
    if (seenStems.has(stem)) continue
    seenStems.add(stem)
    picked.push(item)
    if (picked.length >= limit) break
  }

  return picked
}

function formatProviderVersionLabel(
  bucket: ProviderBucketId,
  modelName: string
): string {
  const ver = extractVersionScore(modelName)
  const formattedVer = Number.isInteger(ver) ? String(ver) : ver.toFixed(1)

  switch (bucket) {
    case 'openai':
      return ver >= 3 ? `GPT-${formattedVer}` : 'GPT'
    case 'anthropic':
      return ver >= 3 ? `Claude ${formattedVer}` : 'Claude'
    case 'gemini':
      return ver >= 1.5 ? `Gemini ${formattedVer}` : 'Gemini'
    case 'deepseek':
      return ver >= 2 ? `DeepSeek V${formattedVer}` : 'DeepSeek'
    case 'qwen':
      return ver >= 2 ? `Qwen ${formattedVer}` : 'Qwen'
    case 'xai':
      return ver >= 2 ? `Grok ${formattedVer}` : 'Grok'
  }
}

export function deriveLandingShowcase(
  liveModels: LiveModelInput[],
  configuredHeroTagline?: string | null
): LandingShowcaseData {
  const buckets: Record<ProviderBucketId, LiveModelInput[]> = {
    openai: [],
    anthropic: [],
    gemini: [],
    deepseek: [],
    qwen: [],
    xai: [],
  }

  for (const item of liveModels) {
    if (!item?.model_name || !isShowcaseCandidate(item.model_name)) continue
    const bucket = classifyProvider(item.model_name)
    if (bucket) {
      buckets[bucket].push(item)
    }
  }

  const topByBucket: Record<ProviderBucketId, LiveModelInput[]> = {
    openai: selectTopProviderModels(buckets.openai, 4),
    anthropic: selectTopProviderModels(buckets.anthropic, 3),
    gemini: selectTopProviderModels(buckets.gemini, 4),
    deepseek: selectTopProviderModels(buckets.deepseek, 3),
    qwen: selectTopProviderModels(buckets.qwen, 3),
    xai: selectTopProviderModels(buckets.xai, 3),
  }

  const providers: LandingProviderItem[] = LANDING_PROVIDERS.map((preset) => {
    const bucket = preset.id as ProviderBucketId
    const liveTop = topByBucket[bucket]
    if (!liveTop || liveTop.length === 0) {
      return preset
    }
    return {
      ...preset,
      models: liveTop.map((m) => m.model_name),
    }
  })

  const topOpenAI = topByBucket.openai[0]
  const secondOpenAI = topByBucket.openai[1] ?? topOpenAI
  const topCodex =
    topByBucket.openai.find((m) =>
      m.model_name.toLowerCase().includes('codex')
    ) ?? topOpenAI
  const topClaude = topByBucket.anthropic[0]
  const topGemini = topByBucket.gemini[0]
  const topDeepSeek = topByBucket.deepseek[0]

  const healthBuckets: ProviderBucketId[] = [
    'openai',
    'anthropic',
    'gemini',
    'deepseek',
  ]
  const healthItems: LandingModelHealthItem[] = DEFAULT_MODEL_HEALTH_ITEMS.map(
    (preset, idx) => {
      const bucket = healthBuckets[idx]
      const liveTop = bucket ? topByBucket[bucket][0] : undefined
      if (!liveTop) return preset
      return {
        ...preset,
        modelName: liveTop.model_name,
        groupName: liveTop.enable_groups?.[0] || preset.groupName,
      }
    }
  )

  const chatModel = secondOpenAI?.model_name ?? 'gpt-6.2-sol'
  const codexModel = topCodex?.model_name ?? 'gpt-6.2-astra'
  const claudeModel = topClaude?.model_name ?? 'claude-opus-5-5'
  const geminiModel = topGemini?.model_name ?? 'gemini-4-neon'

  const terminalModels = {
    chatLabel: topOpenAI
      ? formatProviderVersionLabel('openai', topOpenAI.model_name)
      : 'GPT-6.2',
    chatModel,
    codexModel,
    claudeLabel: topClaude
      ? formatProviderVersionLabel('anthropic', topClaude.model_name)
      : 'Claude 5.5',
    claudeModel,
    geminiLabel: topGemini
      ? formatProviderVersionLabel('gemini', topGemini.model_name)
      : 'Gemini 4',
    geminiModel,
  }

  const trimmedOverride = configuredHeroTagline?.trim() ?? ''
  let heroModelsTagline: string | null = trimmedOverride || null

  if (!heroModelsTagline) {
    const activeLabels: string[] = []
    if (topOpenAI) {
      activeLabels.push(
        formatProviderVersionLabel('openai', topOpenAI.model_name)
      )
      activeLabels.push('Codex')
    }
    if (topClaude) {
      activeLabels.push(
        formatProviderVersionLabel('anthropic', topClaude.model_name)
      )
    }
    if (topGemini) {
      activeLabels.push(
        formatProviderVersionLabel('gemini', topGemini.model_name)
      )
    }
    if (topDeepSeek) {
      activeLabels.push(
        formatProviderVersionLabel('deepseek', topDeepSeek.model_name)
      )
    }
    // Use dynamic hero tagline when at least 2 upstream provider families are configured
    if (activeLabels.length >= 3) {
      heroModelsTagline = activeLabels.join(' · ')
    }
  }

  return {
    heroModelsTagline,
    providers,
    healthItems,
    terminalModels,
    getSnippets: (baseUrl: string) =>
      getCodeSnippets(baseUrl, {
        pythonModel: topOpenAI?.model_name,
        tsModel: topClaude?.model_name,
        curlModel: topGemini?.model_name,
      }),
  }
}

export function useLandingShowcase(): LandingShowcaseData {
  const { models } = usePricingData()
  const { status } = useStatus()
  const configuredHeroTagline =
    typeof status?.landing_featured_models === 'string'
      ? status.landing_featured_models
      : null

  return useMemo(
    () => deriveLandingShowcase(models, configuredHeroTagline),
    [models, configuredHeroTagline]
  )
}
