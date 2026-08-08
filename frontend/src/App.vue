<template>
  <div class="min-h-screen bg-base-200">
    <div class="navbar bg-base-100 shadow-sm">
      <div class="flex-1 px-2">
        <span class="text-xl font-bold">📡 Internet Monitor</span>
      </div>
      <div class="flex-none gap-2 px-2">
        <button class="btn btn-primary btn-sm" :disabled="running" @click="runTest">
          <span v-if="running" class="loading loading-spinner loading-xs"></span>
          {{ running ? 'Testing…' : 'Run test now' }}
        </button>
        <ThemeSwitcher />
      </div>
    </div>

    <div class="mx-auto max-w-6xl p-4 flex flex-col gap-4">
      <div class="card bg-base-100 shadow-sm">
        <div class="card-body flex-row flex-wrap gap-x-12 gap-y-4 py-4">
          <div>
            <div class="text-xs uppercase tracking-wide text-base-content/60">Public IP</div>
            <div class="text-2xl font-semibold tabular">{{ connection.ip || '—' }}</div>
          </div>
          <div>
            <div class="text-xs uppercase tracking-wide text-base-content/60">ISP</div>
            <div class="text-2xl font-semibold">{{ connection.isp || '—' }}</div>
          </div>
        </div>
      </div>

      <StatCards :latest="latest" />

      <div class="flex items-center justify-between">
        <TimeRange v-model="range" />
        <button class="btn btn-ghost btn-sm" :disabled="loading" @click="refresh">
          <span v-if="loading" class="loading loading-spinner loading-xs"></span>
          Refresh
        </button>
      </div>

      <div v-if="errorMessage" role="alert" class="alert alert-error">
        <span>{{ errorMessage }}</span>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <LineChart
          title="Throughput"
          unit="Mbps"
          :results="results"
          :metrics="[
            { key: 'download_mbps', label: 'Download', color: 'series1' },
            { key: 'upload_mbps', label: 'Upload', color: 'series2' },
          ]"
        />
        <LineChart
          title="Latency"
          unit="ms"
          :results="results"
          :metrics="[
            { key: 'ping_ms', label: 'Ping', color: 'series1' },
            { key: 'jitter_ms', label: 'Jitter', color: 'series2' },
          ]"
        />
        <LineChart
          title="Packet loss"
          unit="%"
          :results="results"
          :metrics="[{ key: 'packet_loss_pct', label: 'Packet loss', color: 'series1' }]"
        />
      </div>

      <ResultsTable :results="results" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import StatCards from '@/components/StatCards.vue'
import TimeRange, { type RangeKey } from '@/components/TimeRange.vue'
import LineChart from '@/components/LineChart.vue'
import ResultsTable from '@/components/ResultsTable.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'
import { getLatest, getResults, runNow } from '@/services/api'
import type { SpeedtestResult } from '@/types/result'

const latest = ref<SpeedtestResult | null>(null)
const results = ref<SpeedtestResult[]>([])
const range = ref<RangeKey>('7d')
const loading = ref(false)
const running = ref(false)
const errorMessage = ref('')

// Latest known public IP / ISP: prefer the most recent result, falling back to
// the most recent run that actually captured them (a failed run has neither).
const connection = computed(() => {
  if (latest.value?.external_ip) {
    return { ip: latest.value.external_ip, isp: latest.value.isp ?? '' }
  }
  const lastGood = results.value.find((r) => r.external_ip)
  return { ip: lastGood?.external_ip ?? '', isp: lastGood?.isp ?? '' }
})

function rangeToWindow(key: RangeKey): { from: Date; to: Date } {
  const to = new Date()
  const from = new Date(to)
  switch (key) {
    case '24h':
      from.setHours(from.getHours() - 24)
      break
    case '7d':
      from.setDate(from.getDate() - 7)
      break
    case '30d':
      from.setDate(from.getDate() - 30)
      break
    case 'all':
      return { from: new Date(0), to }
  }
  return { from, to }
}

async function refresh() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { from, to } = rangeToWindow(range.value)
    const [latestRes, resultsRes] = await Promise.all([getLatest(), getResults(from, to)])
    latest.value = latestRes
    results.value = resultsRes
  } catch (err) {
    errorMessage.value = `Failed to load data: ${(err as Error).message}`
  } finally {
    loading.value = false
  }
}

async function runTest() {
  running.value = true
  errorMessage.value = ''
  try {
    await runNow()
    await refresh()
  } catch (err) {
    errorMessage.value = `Test failed: ${(err as Error).message}`
  } finally {
    running.value = false
  }
}

watch(range, refresh)
onMounted(refresh)
</script>
