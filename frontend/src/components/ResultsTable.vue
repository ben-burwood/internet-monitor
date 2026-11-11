<template>
  <div class="rounded-xl border border-border bg-card">
    <div class="flex flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-3">
      <h2 class="text-sm font-semibold text-ink">Results</h2>
      <div class="flex items-center gap-3 text-xs">
        <span v-if="summary.bytes != null" class="text-muted" title="Data used by tests in this range">{{ formatBytes(summary.bytes) }} used</span>
        <span class="text-muted">{{ summary.total }} test{{ summary.total === 1 ? '' : 's' }}</span>
        <span class="inline-flex items-center gap-1.5 text-muted">
          <span class="h-2 w-2 rounded-full bg-accent"></span>
          {{ summary.success }} succeeded
        </span>
        <span
          class="inline-flex items-center gap-1.5"
          :class="summary.fail > 0 ? 'text-danger' : 'text-muted'"
        >
          <span class="h-2 w-2 rounded-full" :class="summary.fail > 0 ? 'bg-danger' : 'bg-muted'"></span>
          {{ summary.fail }} failed
        </span>
      </div>
    </div>
    <div class="max-h-[28rem] overflow-auto">
      <table class="w-full min-w-[720px] border-collapse text-sm tabular">
        <thead>
          <tr class="sticky top-0 bg-card text-xs uppercase tracking-wide text-muted">
            <th class="px-4 py-2 text-left font-medium">Time</th>
            <th class="px-4 py-2 text-right font-medium">Runtime (s)</th>
            <th class="px-4 py-2 text-right font-medium">Download (Mbps)</th>
            <th class="px-4 py-2 text-right font-medium">Upload (Mbps)</th>
            <th class="px-4 py-2 text-right font-medium">Ping (ms)</th>
            <th class="px-4 py-2 text-right font-medium">Jitter (ms)</th>
            <th class="px-4 py-2 text-right font-medium">Packet loss (%)</th>
            <th class="px-4 py-2 text-right font-medium">Data used</th>
            <th class="px-4 py-2 text-left font-medium">Server</th>
            <th class="px-4 py-2"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="results.length === 0">
            <td colspan="10" class="px-4 py-6 text-center text-muted">No data for this range</td>
          </tr>
          <tr
            v-for="r in results"
            :key="r.id"
            class="border-t border-border"
            :class="{ 'text-danger': !r.success }"
          >
            <td class="whitespace-nowrap px-4 py-2 text-left">{{ formatTime(r.timestamp) }}</td>
            <td class="px-4 py-2 text-right">{{ fmtDuration(r.duration_ms) }}</td>
            <template v-if="r.success">
              <td class="px-4 py-2 text-right">{{ fmt(r.download_mbps) }}</td>
              <td class="px-4 py-2 text-right">{{ fmt(r.upload_mbps) }}</td>
              <td class="px-4 py-2 text-right">{{ fmt(r.ping_ms) }}</td>
              <td class="px-4 py-2 text-right">{{ fmt(r.jitter_ms) }}</td>
              <td class="px-4 py-2 text-right">{{ fmt(r.packet_loss_pct) }}</td>
              <td class="whitespace-nowrap px-4 py-2 text-right">{{ formatBytes(totalBytes(r)) }}</td>
              <td class="whitespace-nowrap px-4 py-2 text-left">{{ r.server_name || '—' }}</td>
              <td class="px-4 py-2 text-right">
                <a v-if="r.result_url" :href="r.result_url" target="_blank" rel="noopener" class="text-accent hover:underline">↗</a>
              </td>
            </template>
            <template v-else>
              <td colspan="8" class="px-4 py-2 text-left">Failed: {{ r.error }}</td>
            </template>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { SpeedtestResult } from '@/types/result'
import { formatBytes, formatDateTime as formatTime, formatMetric as fmt } from '@/utils/format'

const props = defineProps<{ results: SpeedtestResult[] }>()

// bytes is null when no result in range has byte counts (e.g. all pre-date
// them), so the header hides the total rather than showing 0 B.
const summary = computed(() => {
  const success = props.results.filter((r) => r.success).length
  let bytes: number | null = null
  for (const r of props.results) {
    const b = totalBytes(r)
    if (b != null) bytes = (bytes ?? 0) + b
  }
  return { total: props.results.length, success, fail: props.results.length - success, bytes }
})

// Runtime is stored in milliseconds; show it in seconds.
function fmtDuration(ms: number | null): string {
  return fmt(ms == null ? null : ms / 1000)
}

// Total data used by a test; null for results recorded before bytes were stored.
function totalBytes(r: SpeedtestResult): number | null {
  if (r.download_bytes == null && r.upload_bytes == null) return null
  return (r.download_bytes ?? 0) + (r.upload_bytes ?? 0)
}
</script>
