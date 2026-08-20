<template>
  <div>
    <div
      v-if="!latest"
      class="rounded-xl border border-border bg-card px-4 py-3 text-sm text-muted"
    >
      No speedtests recorded yet — the first runs shortly after startup.
    </div>

    <div
      v-else-if="!latest.success"
      role="alert"
      class="rounded-xl border border-danger/40 bg-danger/10 px-4 py-3 text-sm text-danger"
    >
      Last test failed: {{ latest.error }}
    </div>

    <div v-else class="grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-border bg-border sm:grid-cols-3 lg:grid-cols-5">
      <div
        v-for="s in tiles"
        :key="s.label"
        class="bg-card px-4 py-3 last:odd:col-span-2 sm:last:odd:col-span-1"
      >
        <div class="text-xs font-medium uppercase tracking-wide text-muted">{{ s.label }}</div>
        <div class="mt-1 flex items-baseline gap-1">
          <span class="text-2xl font-semibold tabular" :class="s.accent">{{ fmt(s.value) }}</span>
          <span class="text-xs text-muted">{{ s.unit }}</span>
        </div>
      </div>
    </div>

    <div
      v-if="latest?.success"
      class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted"
    >
      <span>Last run: {{ formatTime(latest.timestamp) }}</span>
      <span v-if="latest.server_name">
        Server: {{ latest.server_name }}<template v-if="latest.server_location">, {{ latest.server_location }}</template>
      </span>
      <a
        v-if="latest.result_url"
        :href="latest.result_url"
        target="_blank"
        rel="noopener"
        class="text-accent hover:underline"
      >
        View on Speedtest.net ↗
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { SpeedtestResult } from '@/types/result'
import { formatDateTime as formatTime, formatMetric as fmt } from '@/utils/format'

const props = defineProps<{ latest: SpeedtestResult | null }>()

const tiles = computed(() => {
  const l = props.latest
  return [
    { label: 'Download', value: l?.download_mbps ?? null, unit: 'Mbps', accent: 'text-accent' },
    { label: 'Upload', value: l?.upload_mbps ?? null, unit: 'Mbps', accent: 'text-accent-2' },
    { label: 'Ping', value: l?.ping_ms ?? null, unit: 'ms', accent: 'text-ink' },
    { label: 'Jitter', value: l?.jitter_ms ?? null, unit: 'ms', accent: 'text-ink' },
    { label: 'Packet loss', value: l?.packet_loss_pct ?? null, unit: '%', accent: 'text-ink' },
  ]
})
</script>
