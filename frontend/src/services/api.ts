import type { SpeedtestResult, Stats } from '@/types/result'

// All requests are same-origin; in dev, Vite proxies /api to the Go backend.
async function get<T>(path: string): Promise<T> {
  const res = await fetch(path)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<T>
}

export function getResults(from: Date, to: Date): Promise<SpeedtestResult[]> {
  const params = new URLSearchParams({
    from: from.toISOString(),
    to: to.toISOString(),
  })
  return get<SpeedtestResult[]>(`/api/results?${params}`)
}

export function getLatest(): Promise<SpeedtestResult | null> {
  return get<SpeedtestResult | null>('/api/results/latest')
}

export function getStats(from: Date, to: Date): Promise<Stats> {
  const params = new URLSearchParams({
    from: from.toISOString(),
    to: to.toISOString(),
  })
  return get<Stats>(`/api/stats?${params}`)
}

// runNow triggers an on-demand speedtest. This blocks until the test finishes,
// so it can take ~30s.
export async function runNow(): Promise<SpeedtestResult> {
  const res = await fetch('/api/run', { method: 'POST' })
  // 502 is returned when the test itself failed but the body still holds the
  // failure record, so we parse regardless of a non-ok gateway status.
  if (!res.ok && res.status !== 502) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<SpeedtestResult>
}
