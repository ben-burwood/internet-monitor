<template>
  <div>
    <div v-if="!latest" class="alert">No speedtests recorded yet — the first runs shortly after startup.</div>

    <div v-else-if="!latest.success" role="alert" class="alert alert-error">
      <span>Last test failed: {{ latest.error }}</span>
    </div>

    <div v-else class="stats stats-vertical sm:stats-horizontal w-full shadow-sm bg-base-100">
      <div class="stat">
        <div class="stat-title">Download</div>
        <div class="stat-value text-primary tabular">{{ fmt(latest.download_mbps) }}</div>
        <div class="stat-desc">Mbps</div>
      </div>
      <div class="stat">
        <div class="stat-title">Upload</div>
        <div class="stat-value text-secondary tabular">{{ fmt(latest.upload_mbps) }}</div>
        <div class="stat-desc">Mbps</div>
      </div>
      <div class="stat">
        <div class="stat-title">Ping</div>
        <div class="stat-value tabular">{{ fmt(latest.ping_ms) }}</div>
        <div class="stat-desc">ms</div>
      </div>
      <div class="stat">
        <div class="stat-title">Jitter</div>
        <div class="stat-value tabular">{{ fmt(latest.jitter_ms) }}</div>
        <div class="stat-desc">ms</div>
      </div>
      <div class="stat">
        <div class="stat-title">Packet loss</div>
        <div class="stat-value tabular">{{ fmt(latest.packet_loss_pct) }}</div>
        <div class="stat-desc">%</div>
      </div>
    </div>

    <div v-if="latest?.success" class="mt-2 text-sm text-base-content/70 flex flex-wrap gap-x-4 gap-y-1">
      <span>Last run: {{ formatTime(latest.timestamp) }}</span>
      <span v-if="latest.server_name">Server: {{ latest.server_name }}<template v-if="latest.server_location">, {{ latest.server_location }}</template></span>
      <a v-if="latest.result_url" :href="latest.result_url" target="_blank" rel="noopener" class="link link-primary">
        View on Speedtest.net ↗
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SpeedtestResult } from '@/types/result'

defineProps<{ latest: SpeedtestResult | null }>()

function fmt(v: number | null): string {
  return v == null ? '—' : v.toFixed(1)
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleString()
}
</script>
