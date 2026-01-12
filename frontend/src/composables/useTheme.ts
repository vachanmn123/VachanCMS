import { ref, watch } from 'vue'
import { useLocalStorage } from '@vueuse/core'

export type Theme = 'light' | 'dark' | 'system'

const theme = useLocalStorage<Theme>('vachancms-theme', 'dark')
const isDark = ref(true)

function getSystemTheme(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyTheme(dark: boolean) {
  isDark.value = dark
  if (dark) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

function updateTheme() {
  if (theme.value === 'system') {
    applyTheme(getSystemTheme())
  } else {
    applyTheme(theme.value === 'dark')
  }
}

// Watch for theme changes
watch(theme, updateTheme, { immediate: true })

// Listen for system theme changes
if (typeof window !== 'undefined') {
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    if (theme.value === 'system') {
      applyTheme(e.matches)
    }
  })
}

export function useTheme() {
  function setTheme(newTheme: Theme) {
    theme.value = newTheme
  }

  function toggleTheme() {
    if (theme.value === 'dark') {
      setTheme('light')
    } else {
      setTheme('dark')
    }
  }

  return {
    theme,
    isDark,
    setTheme,
    toggleTheme,
  }
}
