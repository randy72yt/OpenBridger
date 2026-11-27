const root = document.documentElement
const themeButton = document.querySelector('[data-theme-button]')
const savedTheme = localStorage.getItem('openbridger-docs-theme')
if (savedTheme === 'dark' || savedTheme === 'light') {
  root.dataset.theme = savedTheme
} else {
  root.dataset.theme = 'light'
}

themeButton?.addEventListener('click', () => {
  root.dataset.theme = root.dataset.theme === 'dark' ? 'light' : 'dark'
  localStorage.setItem('openbridger-docs-theme', root.dataset.theme)
})

const closeMenu = () => document.body.classList.remove('menu-open')
document.querySelector('[data-menu-button]')?.addEventListener('click', () => document.body.classList.toggle('menu-open'))
document.querySelector('[data-backdrop]')?.addEventListener('click', closeMenu)
document.querySelectorAll('.sidebar a').forEach((link) => link.addEventListener('click', closeMenu))

const language = root.dataset.language || 'zh'
const copyLabel = language === 'zh' ? '复制' : 'Copy'
const copiedLabel = language === 'zh' ? '已复制 ✓' : 'Copied ✓'

document.querySelectorAll('article pre').forEach((pre) => {
  const wrapper = document.createElement('div')
  wrapper.className = 'code-block-wrap'
  pre.parentNode?.insertBefore(wrapper, pre)
  wrapper.append(pre)

  const button = document.createElement('button')
  button.type = 'button'
  button.className = 'copy-code-btn'
  button.textContent = copyLabel
  button.addEventListener('click', async () => {
    const codeText = pre.querySelector('code')?.innerText ?? pre.innerText
    try {
      await navigator.clipboard.writeText(codeText)
      button.dataset.copied = 'true'
      button.textContent = copiedLabel
      setTimeout(() => {
        button.dataset.copied = 'false'
        button.textContent = copyLabel
      }, 1800)
    } catch {
      /* ignore clipboard permission errors */
    }
  })
  wrapper.append(button)
})

const dialog = document.querySelector('[data-search-dialog]')
const input = document.querySelector('[data-search-input]')
const results = document.querySelector('[data-search-results]')
let searchIndex
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
  const matches = searchIndex.filter((item) => `${item.title} ${item.text}`.toLocaleLowerCase().includes(query)).slice(0, 12)
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
