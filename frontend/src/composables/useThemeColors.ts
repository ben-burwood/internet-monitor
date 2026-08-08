import { ref } from 'vue'

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

// Shared across every chart: one reactive palette, refreshed when the theme
// (data-theme on <html>) changes. The theme only switches via that attribute
// (see style.css), so a single MutationObserver covers it.
const colors = ref<ChartColors>({ series1: '', series2: '', grid: '', axis: '', text: '' })
let started = false

function start() {
  if (started || typeof document === 'undefined') return
  started = true
  colors.value = read()
  new MutationObserver(() => {
    colors.value = read()
  }).observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
}

export function useThemeColors() {
  start()
  return { colors }
}
