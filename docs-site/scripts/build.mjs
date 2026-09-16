import { readFile, writeFile, mkdir, rm, cp } from 'node:fs/promises'
import { watch } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { marked } from 'marked'

import { site, siteEn } from '../src/site.mjs'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const publicDir = path.join(root, 'public')
const outputDir = path.join(root, 'dist')
const mainSiteUrl = (process.env.OPENBRIDGER_SITE_URL || 'https://openbridger.com').replace(/\/$/, '')

export function routeFor(language, slug) {
  if (slug === 'introduction') return `/${language}/`
  return `/${language}/${slug.replaceAll('--', '/')}/`
}

function escapeHtml(value) {
  return value.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('"', '&quot;')
}

function resolveSiteVariables(value) {
  return value
    .replaceAll('{{CONSOLE_URL}}', mainSiteUrl)
    .replaceAll('{{API_BASE_URL}}', `${mainSiteUrl}/v1`)
}

function renderNavigation(activeSlug, language, currentSite) {
  return currentSite.groups.map((group) => `
    <section class="nav-group">
      <h2>${escapeHtml(group.title)}</h2>
      ${group.pages.map(([slug, title]) => `<a href="${routeFor(language, slug)}"${slug === activeSlug ? ' aria-current="page"' : ''}>${escapeHtml(title)}</a>`).join('')}
    </section>`).join('')
}

function renderDocument({ slug, title, markdown, language, currentSite }) {
  const labels = language === 'zh'
    ? { guide: '实用教程', menu: '打开目录', search: '搜索教程', console: '进入控制台', theme: '切换主题', nav: '文档目录', help: '需要帮助？', contact: '联系我们', placeholder: '搜索教程、配置和错误码…', startSearch: '输入关键词开始搜索', close: '关闭搜索' }
    : { guide: 'Guides', menu: 'Open navigation', search: 'Search guides', console: 'Open console', theme: 'Toggle theme', nav: 'Documentation', help: 'Need help?', contact: 'Contact us', placeholder: 'Search guides, configuration, and errors…', startSearch: 'Type to search', close: 'Close search' }
  const alternateLanguage = language === 'zh' ? 'en' : 'zh'
  const source = resolveSiteVariables(markdown).replaceAll('href="/', `href="/${language}/`)
  const content = marked.parse(source, { gfm: true })
  const description = `${title}｜${currentSite.description}`
  return `<!doctype html>
<html lang="${language === 'zh' ? 'zh-CN' : 'en'}">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="description" content="${escapeHtml(description)}">
  <meta name="theme-color" content="#071218">
  <title>${escapeHtml(title)}｜${escapeHtml(currentSite.title)}</title>
  <link rel="icon" href="/openbridger-mark.svg" type="image/svg+xml">
  <link rel="stylesheet" href="/styles.css">
</head>
<body>
  <header class="topbar">
    <button class="icon-button menu-button" type="button" aria-label="${labels.menu}" data-menu-button><span></span><span></span><span></span></button>
    <a class="brand" href="/${language}/"><img src="/openbridger-mark.svg" alt=""><span>OpenBridger</span><small>${labels.guide}</small></a>
    <div class="top-actions">
      <button class="search-button" type="button" data-search-button><span>${labels.search}</span><kbd>⌘ K</kbd></button>
      <a class="language-link" href="${routeFor(alternateLanguage, slug)}" hreflang="${alternateLanguage}">${alternateLanguage === 'zh' ? '中文' : 'EN'}</a>
      <a class="console-link" href="${mainSiteUrl}">${labels.console}</a>
      <button class="icon-button" type="button" aria-label="${labels.theme}" data-theme-button>◐</button>
    </div>
  </header>
  <div class="shell">
    <aside class="sidebar" data-sidebar>
      <nav aria-label="${labels.nav}">${renderNavigation(slug, language, currentSite)}</nav>
      <div class="sidebar-foot"><span>${labels.help}</span><a href="${mainSiteUrl}/about">${labels.contact}</a></div>
    </aside>
    <button class="backdrop" type="button" aria-label="${labels.menu}" data-backdrop></button>
    <main class="content"><article>${content}</article>
      <footer><span>© 2026 OpenBridger</span><span>·</span><span>AI API Gateway Documentation</span></footer>
    </main>
  </div>
  <dialog class="search-dialog" data-search-dialog>
    <form method="dialog" class="search-panel"><div class="search-head"><input type="search" placeholder="${labels.placeholder}" aria-label="${labels.search}" autocomplete="off" data-search-input><button value="close" aria-label="${labels.close}">×</button></div><div class="search-results" data-search-results><p>${labels.startSearch}</p></div></form>
  </dialog>
  <script>document.documentElement.dataset.language=${JSON.stringify(language)}</script>
  <script src="/app.js" defer></script>
</body>
</html>`
}

export async function build() {
  await rm(outputDir, { recursive: true, force: true })
  await mkdir(outputDir, { recursive: true })
  await cp(publicDir, outputDir, { recursive: true })
  await cp(path.resolve(root, '..', 'web', 'public', 'openbridger-mark.svg'), path.join(outputDir, 'openbridger-mark.svg'))

  await writeFile(path.join(outputDir, 'index.html'), '<!doctype html><meta charset="utf-8"><meta http-equiv="refresh" content="0; url=/zh/"><link rel="canonical" href="/zh/"><title>OpenBridger Guides</title>')

  let pageCount = 0
  for (const [language, currentSite, pageFolder] of [['zh', site, 'pages'], ['en', siteEn, 'pages-en']]) {
    const searchIndex = []
    for (const group of currentSite.groups) {
      for (const [slug, title] of group.pages) {
        const markdown = await readFile(path.join(root, 'src', pageFolder, `${slug}.md`), 'utf8')
        const route = routeFor(language, slug)
        const destination = path.join(outputDir, route)
        await mkdir(destination, { recursive: true })
        await writeFile(path.join(destination, 'index.html'), renderDocument({ slug, title, markdown, language, currentSite }))
        searchIndex.push({ title, group: group.title, href: route, text: resolveSiteVariables(markdown).replace(/<[^>]+>|[#*`>|[\]()-]/g, ' ').replace(/\s+/g, ' ').trim() })
        pageCount += 1
      }
    }
    await writeFile(path.join(outputDir, `search-index-${language}.json`), JSON.stringify(searchIndex))
  }
  console.log(`Built ${pageCount} localized pages for ${mainSiteUrl}`)
}

await build()

if (process.argv.includes('--watch')) {
  let timer
  watch(path.join(root, 'src'), { recursive: true }, () => {
    clearTimeout(timer)
    timer = setTimeout(() => build().catch(console.error), 120)
  })
  console.log('Watching docs source files…')
}
