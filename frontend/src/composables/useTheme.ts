import { ref } from 'vue'

export type Theme = 'light' | 'dark'

// The initial theme is resolved before paint by the inline script in index.html,
// which stamps data-theme on <html>. Mirror that here so the toggle stays in sync.
function current(): Theme {
  return document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'
}

const theme = ref<Theme>(current())

function apply(next: Theme) {
  theme.value = next
  document.documentElement.dataset.theme = next
  localStorage.setItem('theme', next)
}

// Shared singleton — every caller reads and drives the same theme state.
export function useTheme() {
  function toggle() {
    apply(theme.value === 'dark' ? 'light' : 'dark')
  }
  return { theme, toggle }
}
