<template>
  <div class="min-h-screen">
    <div class="mx-auto flex max-w-[85%] flex-col gap-4 p-4 sm:p-6">
      <header class="flex flex-wrap items-center justify-between gap-3">
        <span class="text-xl font-bold text-ink">📡 Internet Monitor</span>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-lg bg-accent px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-accent-strong disabled:opacity-60"
            :disabled="running"
            @click="runTest"
          >
            <Spinner v-if="running" />
            {{ running ? 'Testing…' : 'Run test now' }}
          </button>
          <ThemeSwitcher />
        </div>
      </header>

      <div class="flex flex-wrap items-center gap-x-8 gap-y-3 rounded-xl bg-accent/10 px-5 py-3 ring-1 ring-inset ring-accent/25">
        <div class="flex items-center gap-2 text-accent">
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <path d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
          </svg>
          <span class="text-sm font-semibold">Connection</span>
        </div>
        <div class="flex items-baseline gap-2">
          <span class="text-xs uppercase tracking-wide text-muted">Public IP</span>
          <span class="text-base font-semibold tabular text-ink">{{ connection.ip || '—' }}</span>
        </div>
        <div class="hidden h-4 w-px bg-accent/25 sm:block"></div>
        <div class="flex items-baseline gap-2">
          <span class="text-xs uppercase tracking-wide text-muted">ISP</span>
          <span class="text-base font-semibold text-ink">{{ connection.isp || '—' }}</span>
        </div>
      </div>

      <StatCards :latest="latest" />

      <div class="flex items-center justify-between">
        <TimeRange v-model="range" />
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-lg border border-border px-3 py-2 text-sm font-medium text-muted transition-colors hover:bg-card hover:text-ink disabled:opacity-60"
          :disabled="loading"
          @click="refresh"
        >
          <Spinner v-if="loading" />
          Refresh
        </button>
      </div>

      <div
        v-if="errorMessage"
        role="alert"
        class="rounded-xl border border-danger/40 bg-danger/10 px-4 py-3 text-sm text-danger"
      >
        {{ errorMessage }}
      </div>

      <div class="grid gap-4 lg:grid-cols-3">
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

      <IpHistory :segments="ipHistory" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import StatCards from '@/components/StatCards.vue'
import TimeRange, { type RangeKey } from '@/components/TimeRange.vue'
import LineChart from '@/components/LineChart.vue'
import ResultsTable from '@/components/ResultsTable.vue'
import IpHistory from '@/components/IpHistory.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'
import Spinner from '@/components/Spinner.vue'
import { getIpHistory, getLatest, getResults, runNow } from '@/services/api'
import type { IPSegment, SpeedtestResult } from '@/types/result'

const latest = ref<SpeedtestResult | null>(null)
const results = ref<SpeedtestResult[]>([])
const ipHistory = ref<IPSegment[]>([])
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
    const [latestRes, resultsRes, ipHistoryRes] = await Promise.all([
      getLatest(),
      getResults(from, to),
      getIpHistory(),
    ])
    latest.value = latestRes
    results.value = resultsRes
    ipHistory.value = ipHistoryRes
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
