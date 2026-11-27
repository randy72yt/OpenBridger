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
export interface ModelProvider {
  icon: string
  monoIcon: string
  label: string
}

const MODEL_PROVIDER_RULES: ReadonlyArray<{
  matches: (model: string) => boolean
  provider: ModelProvider
}> = [
  {
    matches: (model) =>
      [
        'gpt-',
        'chatgpt',
        'text-embedding-',
        'omni-moderation',
        'dall-e',
        'whisper',
        'tts-',
        'sora',
        'codex',
      ].some((kw) => model.includes(kw)) || /\bo[134](?:-|$)/.test(model),
    provider: { icon: 'OpenAI', monoIcon: 'OpenAI', label: 'OpenAI' },
  },
  {
    matches: (model) =>
      model.includes('claude') || model.includes('anthropic'),
    provider: { icon: 'Claude.Color', monoIcon: 'Claude', label: 'Claude' },
  },
  {
    matches: (model) =>
      model.includes('gemini') ||
      model.includes('learnlm') ||
      model.includes('imagen') ||
      model.includes('veo-'),
    provider: { icon: 'Gemini.Color', monoIcon: 'Gemini', label: 'Gemini' },
  },
  {
    matches: (model) => model.includes('grok') || model.includes('xai-'),
    provider: { icon: 'Grok', monoIcon: 'Grok', label: 'Grok' },
  },
  {
    matches: (model) => model.includes('deepseek'),
    provider: {
      icon: 'DeepSeek.Color',
      monoIcon: 'DeepSeek',
      label: 'DeepSeek',
    },
  },
  {
    matches: (model) => model.includes('qwen') || model.includes('qwq-'),
    provider: { icon: 'Qwen.Color', monoIcon: 'Qwen', label: 'Qwen' },
  },
  {
    matches: (model) => model.includes('doubao') || model.includes('volcengine'),
    provider: { icon: 'Doubao.Color', monoIcon: 'Doubao', label: 'Doubao' },
  },
  {
    matches: (model) => model.includes('moonshot') || model.includes('kimi'),
    provider: {
      icon: 'Moonshot',
      monoIcon: 'Moonshot',
      label: 'Moonshot',
    },
  },
  {
    matches: (model) => model.includes('minimax') || model.includes('abab'),
    provider: {
      icon: 'Minimax.Color',
      monoIcon: 'Minimax',
      label: 'MiniMax',
    },
  },
  {
    matches: (model) =>
      ['glm-', 'chatglm', 'cogview', 'cogvideo'].some((kw) =>
        model.includes(kw)
      ),
    provider: { icon: 'Zhipu.Color', monoIcon: 'Zhipu', label: 'Zhipu' },
  },
  {
    matches: (model) => model.includes('mimo-'),
    provider: {
      icon: 'XiaomiMiMo',
      monoIcon: 'XiaomiMiMo',
      label: 'MiMo',
    },
  },
  {
    matches: (model) => model.includes('ernie'),
    provider: { icon: 'Wenxin.Color', monoIcon: 'Wenxin', label: 'Baidu' },
  },
  {
    matches: (model) => model.includes('spark'),
    provider: { icon: 'Spark.Color', monoIcon: 'Spark', label: 'iFlyTek' },
  },
  {
    matches: (model) => model.includes('hunyuan'),
    provider: {
      icon: 'Hunyuan.Color',
      monoIcon: 'Hunyuan',
      label: 'Tencent',
    },
  },
  {
    matches: (model) => model.includes('baichuan'),
    provider: {
      icon: 'Baichuan.Color',
      monoIcon: 'Baichuan',
      label: 'Baichuan',
    },
  },
  {
    matches: (model) => model.includes('internlm'),
    provider: {
      icon: 'InternLM.Color',
      monoIcon: 'InternLM',
      label: 'InternLM',
    },
  },
  {
    matches: (model) => model.includes('step-'),
    provider: {
      icon: 'Stepfun.Color',
      monoIcon: 'Stepfun',
      label: 'StepFun',
    },
  },
  {
    matches: (model) => model.includes('yi-'),
    provider: { icon: 'Yi.Color', monoIcon: 'Yi', label: 'Yi' },
  },
  {
    matches: (model) =>
      model.includes('mistral') ||
      model.includes('mixtral') ||
      model.includes('codestral'),
    provider: {
      icon: 'Mistral.Color',
      monoIcon: 'Mistral',
      label: 'Mistral',
    },
  },
  {
    matches: (model) => model.includes('llama') || model.includes('meta-'),
    provider: { icon: 'Meta.Color', monoIcon: 'Meta', label: 'Meta' },
  },
  {
    matches: (model) => model.includes('command-') || model.includes('cohere'),
    provider: { icon: 'Cohere.Color', monoIcon: 'Cohere', label: 'Cohere' },
  },
  {
    matches: (model) => model.includes('sonar') || model.includes('perplexity'),
    provider: {
      icon: 'Perplexity.Color',
      monoIcon: 'Perplexity',
      label: 'Perplexity',
    },
  },
]

export function resolveModelProvider(
  modelName: string | undefined | null
): ModelProvider | null {
  if (!modelName) {
    return null
  }

  const normalized = modelName.trim().toLowerCase()
  if (!normalized) {
    return null
  }

  for (const rule of MODEL_PROVIDER_RULES) {
    if (rule.matches(normalized)) {
      return rule.provider
    }
  }

  return null
}
