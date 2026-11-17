const STORAGE_KEY = 'hammond-modern-theme'

export function currentTheme() {
  return document.documentElement.getAttribute('data-theme') || 'dark'
}

export function applySavedTheme() {
  let saved = 'dark'
  try { saved = localStorage.getItem(STORAGE_KEY) || 'dark' } catch (_) {}
  const theme = saved === 'light' ? 'light' : 'dark'
  document.documentElement.setAttribute('data-theme', theme)
  return theme
}

export function toggleTheme() {
  const theme = currentTheme() === 'dark' ? 'light' : 'dark'
  document.documentElement.setAttribute('data-theme', theme)
  try { localStorage.setItem(STORAGE_KEY, theme) } catch (_) {}
  return theme
}
