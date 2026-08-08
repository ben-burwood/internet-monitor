<template>
  <div class="card bg-base-100 shadow-sm">
    <div class="card-body gap-2">
      <h2 class="card-title text-base">{{ title }}</h2>
      <div class="h-64">
        <Line v-if="hasData" :data="chartData" :options="chartOptions" />
        <div v-else class="flex h-full items-center justify-center text-sm text-base-content/50">
          No data for this range
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import type { ChartData, ChartOptions } from 'chart.js'
import type { SpeedtestResult } from '@/types/result'
import { useThemeColors } from '@/composables/useThemeColors'

interface Metric {
  key: keyof SpeedtestResult
  label: string
  color: 'series1' | 'series2'
}

const props = defineProps<{
  title: string
  unit: string
  results: SpeedtestResult[]
  metrics: Metric[]
}>()

const { colors } = useThemeColors()

// Oldest-first for a natural left-to-right time axis.
const sorted = computed(() =>
  [...props.results].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
  ),
)

const hasData = computed(() => sorted.value.length > 0)

const chartData = computed<ChartData<'line'>>(() => ({
  datasets: props.metrics.map((m) => {
    const color = m.color === 'series1' ? colors.value.series1 : colors.value.series2
    return {
      label: m.label,
      data: sorted.value.map((r) => ({
        x: new Date(r.timestamp).getTime(),
        y: r[m.key] as number | null,
      })),
      borderColor: color,
      backgroundColor: color,
      borderWidth: 2,
      pointRadius: 2,
      pointHoverRadius: 5,
      tension: 0.25,
      spanGaps: false, // failed runs (null) render as gaps
    }
  }),
}))

const chartOptions = computed<ChartOptions<'line'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index', intersect: false },
  plugins: {
    legend: {
      display: props.metrics.length > 1,
      labels: { color: colors.value.text, usePointStyle: true, boxWidth: 8 },
    },
    tooltip: {
      callbacks: {
        label: (ctx) => `${ctx.dataset.label}: ${formatValue(ctx.parsed.y)} ${props.unit}`,
      },
    },
  },
  scales: {
    x: {
      type: 'time',
      time: { tooltipFormat: 'PPpp' },
      grid: { color: colors.value.grid },
      ticks: { color: colors.value.axis, maxRotation: 0, autoSkip: true },
    },
    y: {
      beginAtZero: true,
      title: { display: true, text: props.unit, color: colors.value.axis },
      grid: { color: colors.value.grid },
      ticks: { color: colors.value.axis },
    },
  },
}))

function formatValue(v: number | null): string {
  return v == null ? '—' : v.toFixed(2)
}
</script>
