import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'

import { routeFor } from './build.mjs'

const root = path.resolve(import.meta.dirname, '..')

test('creates clean documentation routes', () => {
  assert.equal(routeFor('zh', 'introduction'), '/zh/')
  assert.equal(routeFor('en', 'guide--clients--codex-cli'), '/en/guide/clients/codex-cli/')
  assert.equal(routeFor('zh', 'legal--privacy'), '/zh/legal/privacy/')
})

test('build output uses OpenBridger production URLs and contains no source placeholders', async () => {
  const home = await readFile(path.join(root, 'dist', 'zh', 'index.html'), 'utf8')
  const english = await readFile(path.join(root, 'dist', 'en', 'index.html'), 'utf8')
  const index = await readFile(path.join(root, 'dist', 'search-index-zh.json'), 'utf8')
  assert.match(home, /https:\/\/openbridger\.com/)
  assert.match(home, /href="\/en\/"/)
  assert.match(english, /href="\/zh\/"/)
  assert.doesNotMatch(home + english + index, /\{\{(?:CONSOLE|API_BASE)_URL\}\}/)
  assert.doesNotMatch(home + english + index, /cun\.ai|dddai|wintoken|Waffo Pancake/i)
})
