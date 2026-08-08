<template>
  <div class="card bg-base-100 shadow-sm">
    <div class="card-body">
      <h2 class="card-title text-base">Results</h2>
      <div class="overflow-x-auto max-h-[28rem] overflow-y-auto">
        <table class="table table-sm table-pin-rows tabular">
          <thead>
            <tr>
              <th>Time</th>
              <th class="text-right">Download (Mbps)</th>
              <th class="text-right">Upload (Mbps)</th>
              <th class="text-right">Ping (ms)</th>
              <th class="text-right">Jitter (ms)</th>
              <th class="text-right">Packet loss (%)</th>
              <th>Server</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="results.length === 0">
              <td colspan="8" class="text-center text-base-content/50">No data for this range</td>
            </tr>
            <tr v-for="r in results" :key="r.id" :class="{ 'text-error': !r.success }">
              <td class="whitespace-nowrap">{{ formatTime(r.timestamp) }}</td>
              <template v-if="r.success">
                <td class="text-right">{{ fmt(r.download_mbps) }}</td>
                <td class="text-right">{{ fmt(r.upload_mbps) }}</td>
                <td class="text-right">{{ fmt(r.ping_ms) }}</td>
                <td class="text-right">{{ fmt(r.jitter_ms) }}</td>
                <td class="text-right">{{ fmt(r.packet_loss_pct) }}</td>
                <td class="whitespace-nowrap">{{ r.server_name || '—' }}</td>
                <td>
                  <a v-if="r.result_url" :href="r.result_url" target="_blank" rel="noopener" class="link link-primary">↗</a>
                </td>
              </template>
              <template v-else>
                <td colspan="7">Failed: {{ r.error }}</td>
              </template>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SpeedtestResult } from '@/types/result'

defineProps<{ results: SpeedtestResult[] }>()

function fmt(v: number | null): string {
  return v == null ? '—' : v.toFixed(1)
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleString()
}
</script>
