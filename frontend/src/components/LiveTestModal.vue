<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4 sm:items-center"
      @click.self="$emit('close')"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="live-test-title"
        class="my-auto w-full max-w-lg rounded-xl border border-border bg-card shadow-xl"
      >
        <!-- header -->
        <div class="flex items-center justify-between border-b border-border px-4 py-3">
          <h2 id="live-test-title" class="flex items-center gap-2 text-sm font-semibold text-ink">
            <Spinner v-if="running" class="text-accent" />
            {{ running ? 'Running speedtest…' : 'Speedtest' }}
          </h2>
          <button
            type="button"
            class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface hover:text-ink"
            aria-label="Close"
            @click="$emit('close')"
          >
            <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <path d="M18 6 6 18M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- body -->
        <div class="flex flex-col gap-4 px-4 py-4">
          <!-- server -->
          <div class="text-sm text-muted">
            <template v-if="serverName">
              Server: <span class="text-ink">{{ serverName }}</span><template v-if="serverLocation">, {{ serverLocation }}</template>
            </template>
            <template v-else-if="running">Finding a server…</template>
          </div>

          <!-- phase steps -->
          <div class="grid grid-cols-3 gap-2">
            <div v-for="step in steps" :key="step.phase">
              <div class="mb-1 flex items-center justify-between text-xs font-medium uppercase tracking-wide" :class="step.state === 'pending' ? 'text-muted' : 'text-ink'">
                {{ step.label }}
                <svg v-if="step.state === 'done'" class="h-3.5 w-3.5 text-accent" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M20 6 9 17l-5-5" />
                </svg>
              </div>
              <div class="h-1.5 overflow-hidden rounded-full bg-surface">
                <div class="h-full rounded-full bg-accent transition-[width] duration-200" :style="{ width: `${step.fill * 100}%` }"></div>
              </div>
            </div>
          </div>

          <!-- live values -->
          <div class="grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-border bg-border">
            <div v-for="t in tiles" :key="t.label" class="bg-card px-4 py-3" :class="t.active ? 'bg-accent/5' : ''">
              <div class="text-xs font-medium uppercase tracking-wide text-muted">{{ t.label }}</div>
              <div class="mt-1 flex items-baseline gap-1">
                <span class="text-2xl font-semibold tabular" :class="t.accent">{{ fmt(t.value) }}</span>
                <span class="text-xs text-muted">{{ t.unit }}</span>
              </div>
            </div>
          </div>

          <!-- outcome -->
          <div
            v-if="error || (result && !result.success)"
            role="alert"
            class="rounded-xl border border-danger/40 bg-danger/10 px-4 py-3 text-sm text-danger"
          >
            Test failed: {{ error || result?.error }}
          </div>
          <div v-else-if="result" class="flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted">
            <span>Packet loss: <span class="tabular text-ink">{{ fmt(result.packet_loss_pct) }}%</span></span>
            <span v-if="result.duration_ms != null">Took {{ (result.duration_ms / 1000).toFixed(0) }}s</span>
            <a
              v-if="result.result_url"
              :href="result.result_url"
              target="_blank"
              rel="noopener"
              class="text-accent hover:underline"
            >
              View on Speedtest.net ↗
            </a>
          </div>
          <p v-else class="text-xs text-muted">You can hide this — the test keeps running as long as this page stays open.</p>
        </div>

        <!-- footer -->
        <div class="flex items-center justify-end border-t border-border px-4 py-3">
          <button
            type="button"
            class="rounded-lg border border-border px-3 py-2 text-sm font-medium text-muted transition-colors hover:bg-surface hover:text-ink"
            @click="$emit('close')"
          >
            {{ running ? 'Hide' : 'Close' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import Spinner from '@/components/Spinner.vue'
import type { SpeedtestResult, TestPhase, TestProgress } from '@/types/result'
import { formatMetric as fmt } from '@/utils/format'

const props = defineProps<{
  open: boolean
  running: boolean
  progress: TestProgress | null
  result: SpeedtestResult | null
  error: string
}>()
const emit = defineEmits<{ close: [] }>()

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) emit('close')
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))

// Once the final result arrives, show its values; until then, the live ones.
const values = computed(() => props.result ?? props.progress)
const serverName = computed(() => props.result?.server_name || props.progress?.server_name)
const serverLocation = computed(() => props.result?.server_location || props.progress?.server_location)

const phaseOrder: TestPhase[] = ['start', 'ping', 'download', 'upload']
const stepLabels = { ping: 'Ping', download: 'Download', upload: 'Upload' } as const

// Each measured phase is done once the test has moved past it (or finished
// successfully), active while running, and pending before then.
const steps = computed(() => {
  const current = props.progress ? phaseOrder.indexOf(props.progress.phase) : -1
  const finished = props.result?.success === true
  return (Object.keys(stepLabels) as (keyof typeof stepLabels)[]).map((phase) => {
    const idx = phaseOrder.indexOf(phase)
    let state: 'pending' | 'active' | 'done' = 'pending'
    let fill = 0
    if (finished || idx < current) {
      state = 'done'
      fill = 1
    } else if (idx === current) {
      state = 'active'
      fill = props.progress?.phase_progress ?? 0
    }
    return { phase, label: stepLabels[phase], state, fill }
  })
})

const tiles = computed(() => {
  const v = values.value
  const phase = props.running ? props.progress?.phase : undefined
  return [
    { label: 'Download', value: v?.download_mbps ?? null, unit: 'Mbps', accent: 'text-accent', active: phase === 'download' },
    { label: 'Upload', value: v?.upload_mbps ?? null, unit: 'Mbps', accent: 'text-accent-2', active: phase === 'upload' },
    { label: 'Ping', value: v?.ping_ms ?? null, unit: 'ms', accent: 'text-ink', active: phase === 'ping' },
    { label: 'Jitter', value: v?.jitter_ms ?? null, unit: 'ms', accent: 'text-ink', active: phase === 'ping' },
  ]
})
</script>
