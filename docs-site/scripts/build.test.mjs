import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile, readdir } from 'node:fs/promises'
import path from 'node:path'

import { routeFor } from './build.mjs'
import { site, siteEn } from '../src/site.mjs'

const root = path.resolve(import.meta.dirname, '..')

function slugsOf(currentSite) {
  return currentSite.groups.flatMap((group) => group.pages.map(([slug]) => slug))
}

function hrefToSlug(href) {
  const trimmed = href.replace(/^\//, '').replace(/\/$/, '')
  if (!trimmed) return 'introduction'
  return trimmed.replaceAll('/', '--')
}

test('creates clean documentation routes', () => {
  assert.equal(routeFor('zh', 'introduction'), '/zh/')
  assert.equal(routeFor('en', 'guide--clients--codex-cli'), '/en/guide/clients/codex-cli/')
  assert.equal(routeFor('zh', 'legal--privacy'), '/zh/legal/privacy/')
  assert.equal(routeFor('zh', 'guide--errors'), '/zh/guide/errors/')
  assert.equal(routeFor('en', 'guide--clients--lobechat'), '/en/guide/clients/lobechat/')
})

test('Chinese and English sidebars list the same pages', async () => {
  const zh = slugsOf(site)
  const en = slugsOf(siteEn)
  assert.deepEqual(en, zh)
  for (const slug of zh) {
    await readFile(path.join(root, 'src', 'pages', `${slug}.md`))
    await readFile(path.join(root, 'src', 'pages-en', `${slug}.md`))
  }
  const pageFiles = await readdir(path.join(root, 'src', 'pages'))
  const englishFiles = await readdir(path.join(root, 'src', 'pages-en'))
  assert.deepEqual(englishFiles.filter((name) => name.endsWith('.md')).sort(), pageFiles.filter((name) => name.endsWith('.md')).sort())
})

test('internal links point at sidebar pages and banned phrasing stays out', async () => {
  const known = new Set(slugsOf(site))
  const banned = /翻墙|绕过限制|科学上网|魔法上网/
  for (const folder of ['pages', 'pages-en']) {
    for (const slug of known) {
      const markdown = await readFile(path.join(root, 'src', folder, `${slug}.md`), 'utf8')
      assert.doesNotMatch(markdown, banned, `${folder}/${slug}.md`)
      for (const match of markdown.matchAll(/(?:\]\(|href=")(\/[^")\s]*)/g)) {
        const href = match[1]
        if (href.startsWith('//')) continue
        assert.equal(known.has(hrefToSlug(href)), true, `${folder}/${slug}.md -> ${href}`)
      }
    }
  }
})

test('build output uses OpenBridger production URLs and contains no source placeholders', async () => {
  const home = await readFile(path.join(root, 'dist', 'zh', 'index.html'), 'utf8')
  const english = await readFile(path.join(root, 'dist', 'en', 'index.html'), 'utf8')
  const index = await readFile(path.join(root, 'dist', 'search-index-zh.json'), 'utf8')
  assert.match(home, /https:\/\/openbridger\.com/)
  assert.match(home, /href="\/en\/"/)
  assert.match(home, /href="\/zh\/guide\/quickstart\/"/)
  assert.match(english, /href="\/zh\/"/)
  const errors = await readFile(path.join(root, 'dist', 'zh', 'guide', 'errors', 'index.html'), 'utf8')
  const lobechat = await readFile(path.join(root, 'dist', 'en', 'guide', 'clients', 'lobechat', 'index.html'), 'utf8')
  assert.match(errors, /insufficient_user_quota/)
  assert.match(lobechat, /https:\/\/openbridger\.com\/v1/)
  assert.doesNotMatch(home + english + index + errors, /\/zh\/zh\/|\/en\/en\//)
  assert.doesNotMatch(home + english + index, /\{\{(?:CONSOLE|API_BASE)_URL\}\}/)
  assert.doesNotMatch(home + english + index, /cun\.ai|dddai|wintoken|Waffo Pancake/i)
})
