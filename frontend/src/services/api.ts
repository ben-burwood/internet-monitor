import type { IPSegment, Server, ServerSelection, SpeedtestResult, TestProgress } from '@/types/result'

// All requests are same-origin; in dev, Vite proxies /api to the Go backend.
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<T>
}

function get<T>(path: string): Promise<T> {
  return request<T>(path)
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

export function getIpHistory(): Promise<IPSegment[]> {
  return get<IPSegment[]>('/api/ip-history')
}

// getServers lists the nearest Ookla servers. This calls the CLI, so it can
// take a couple of seconds.
export function getServers(): Promise<Server[]> {
  return get<Server[]>('/api/servers')
}

export function getSettings(): Promise<ServerSelection> {
  return get<ServerSelection>('/api/settings')
}

export function updateSettings(sel: ServerSelection): Promise<ServerSelection> {
  return request<ServerSelection>('/api/settings', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(sel),
  })
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

type StreamEvent =
  | { type: 'progress'; progress: TestProgress }
  | { type: 'result'; result: SpeedtestResult }

// runStream triggers an on-demand speedtest and reports live progress as it
// runs, resolving with the stored result (which may be a failure record).
// The backend streams newline-delimited JSON over a POST, so this reads the
// body incrementally rather than using EventSource (GET-only, auto-reconnects).
export async function runStream(onProgress: (p: TestProgress) => void): Promise<SpeedtestResult> {
  const res = await fetch('/api/run/stream', { method: 'POST' })
  if (!res.ok || !res.body) throw new Error(`HTTP ${res.status}`)

  const reader = res.body.pipeThrough(new TextDecoderStream()).getReader()
  let buffer = ''
  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += value

    const lines = buffer.split('\n')
    buffer = lines.pop() ?? ''
    for (const line of lines) {
      if (!line.trim()) continue
      const event = JSON.parse(line) as StreamEvent
      if (event.type === 'progress') onProgress(event.progress)
      else return event.result
    }
  }
  throw new Error('Connection closed before the test finished')
}
