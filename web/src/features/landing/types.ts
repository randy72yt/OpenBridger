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
export type ModelStatusLevel = 'healthy' | 'degraded' | 'down'

export type LandingProviderItem = {
  id: string
  name: string
  protocol: string
  models: string[]
  tagKey: string
  iconKey: string
}

export type LandingPricingPreviewRow = {
  modelName: string
  vendorName: string
  groupName: string
  groupRatio: number
  inputPerMillionUsd: number
  outputPerMillionUsd: number
  cacheReadPerMillionUsd: number
  tags: string[]
}

export type LandingModelHealthSlot = {
  id: string
  level: ModelStatusLevel
}

export type LandingModelHealthItem = {
  modelName: string
  provider: string
  groupName: string
  status: ModelStatusLevel
  successRate: number
  ttftMs: number
  tps: number
  hourlySlots: LandingModelHealthSlot[]
}

export type CodeLanguage = 'python' | 'typescript' | 'curl'

export type CodeSnippetItem = {
  id: CodeLanguage
  label: string
  filename: string
  code: string
  highlightLines: number[]
}

export type LandingFaqItem = {
  id: string
  questionKey: string
  answerKey: string
}
