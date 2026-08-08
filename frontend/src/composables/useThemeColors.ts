import { onMounted, onUnmounted, ref } from 'vue'

export interface ChartColors {
  series1: string
  series2: string
  grid: string
  axis: string
  text: string
}

function read(): ChartColors {
  const s = getComputedStyle(document.documentElement)
  const v = (name: string) => s.getPropertyValue(name).trim()
  return {
    series1: v('--series-1'),
    series2: v('--series-2'),
    grid: v('--chart-grid'),
    axis: v('--chart-axis'),
    text: v('--chart-text'),
  }
}

// useThemeColors exposes the current chart palette as a reactive value that
// updates when the DaisyUI theme (data-theme) or the OS colour scheme changes.
export function useThemeColors() {
  const colors = ref<ChartColors>({
    series1: '#2a78d6',
    series2: '#eb6834',
    grid: '#e1e0d9',
    axis: '#898781',
    text: '#52514e',
  })

  let observer: MutationObserver | null = null
  const media = window.matchMedia('(prefers-color-scheme: dark)')
  const refresh = () => {
    colors.value = read()
  }

  onMounted(() => {
    refresh()
    observer = new MutationObserver(refresh)
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
    media.addEventListener('change', refresh)
  })

  onUnmounted(() => {
    observer?.disconnect()
    media.removeEventListener('change', refresh)
  })

  return { colors }
}
