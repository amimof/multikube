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

export function useTheme() {
  if (!initialized) {
    initialized = true

    const stored = getStoredPreference()
    isDark.value = stored ? stored === 'dark' : getSystemPreference()
    applyTheme(isDark.value)

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      if (!getStoredPreference()) {
        isDark.value = e.matches
      }
    })

    watch(isDark, (val) => {
      applyTheme(val)
      localStorage.setItem(STORAGE_KEY, val ? 'dark' : 'light')
    })
  }

  function toggleTheme() {
    isDark.value = !isDark.value
  }

  return { isDark, toggleTheme }
}
