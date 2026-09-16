const root = document.documentElement
const themeButton = document.querySelector('[data-theme-button]')
const savedTheme = localStorage.getItem('openbridger-docs-theme')
if (savedTheme) root.dataset.theme = savedTheme
else if (matchMedia('(prefers-color-scheme: dark)').matches) root.dataset.theme = 'dark'

themeButton?.addEventListener('click', () => {
  root.dataset.theme = root.dataset.theme === 'dark' ? 'light' : 'dark'
  localStorage.setItem('openbridger-docs-theme', root.dataset.theme)
})

const closeMenu = () => document.body.classList.remove('menu-open')
document.querySelector('[data-menu-button]')?.addEventListener('click', () => document.body.classList.toggle('menu-open'))
document.querySelector('[data-backdrop]')?.addEventListener('click', closeMenu)
document.querySelectorAll('.sidebar a').forEach((link) => link.addEventListener('click', closeMenu))

const dialog = document.querySelector('[data-search-dialog]')
const input = document.querySelector('[data-search-input]')
const results = document.querySelector('[data-search-results]')
let searchIndex
const language = root.dataset.language || 'zh'
const emptyLabel = language === 'zh' ? '输入关键词开始搜索' : 'Type to search'
const noResultsLabel = language === 'zh' ? '没有找到相关教程' : 'No matching guides found'

async function openSearch() {
  dialog.showModal()
  input.focus()
  searchIndex ||= await fetch(`/search-index-${language}.json`).then((response) => response.json())
}

document.querySelector('[data-search-button]')?.addEventListener('click', openSearch)
document.addEventListener('keydown', (event) => {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    openSearch()
  }
})

input?.addEventListener('input', () => {
  const query = input.value.trim().toLocaleLowerCase()
  if (!query) {
    results.innerHTML = `<p>${emptyLabel}</p>`
    return
  }
  const matches = searchIndex.filter((item) => `${item.title} ${item.text}`.toLocaleLowerCase().includes(query)).slice(0, 10)
  results.replaceChildren(...matches.map((item) => {
    const link = document.createElement('a')
    const title = document.createElement('strong')
    const group = document.createElement('span')
    link.href = item.href
    title.textContent = item.title
    group.textContent = item.group
    link.append(title, group)
    return link
  }))
  if (!matches.length) results.innerHTML = `<p>${noResultsLabel}</p>`
})
