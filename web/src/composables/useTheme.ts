import { ref, watch } from 'vue'

const STORAGE_KEY = 'multikube-theme'

type ThemePreference = 'light' | 'dark' | null

function getStoredPreference(): ThemePreference {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'light' || stored === 'dark') return stored
  return null
}

function getSystemPreference(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyTheme(dark: boolean) {
  document.documentElement.classList.toggle('dark', dark)
}

const isDark = ref(false)

let initialized = false
let stopSync: (() => void) | null = null

export function initTheme() {
  if (initialized) {
    return
  }

  initialized = true

  const stored = getStoredPreference()
  isDark.value = stored ? stored === 'dark' : getSystemPreference()
  applyTheme(isDark.value)

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    if (!getStoredPreference()) {
      isDark.value = e.matches
    }
  })

  stopSync = watch(isDark, (val) => {
    applyTheme(val)
    localStorage.setItem(STORAGE_KEY, val ? 'dark' : 'light')
  })
}

export function useTheme() {
  initTheme()

  function toggleTheme() {
    isDark.value = !isDark.value
  }

  return { isDark, toggleTheme }
}
