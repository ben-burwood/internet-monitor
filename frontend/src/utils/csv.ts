// CSV export of speedtest results for the dashboard's download button.

import type { SpeedtestResult } from '@/types/result'

const COLUMNS: (keyof SpeedtestResult)[] = [
  'id',
  'timestamp',
  'success',
  'error',
  'download_mbps',
  'upload_mbps',
  'ping_ms',
  'jitter_ms',
  'packet_loss_pct',
  'duration_ms',
  'download_bytes',
  'upload_bytes',
  'server_id',
  'server_name',
  'server_location',
  'isp',
  'external_ip',
  'result_url',
]

// escapeCell quotes a value when it contains a delimiter, quote or newline,
// doubling any embedded quotes per RFC 4180. Missing values become empty.
function escapeCell(v: unknown): string {
  if (v == null) return ''
  const s = String(v)
  return /[",\r\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s
}

// resultsToCsv renders results as a CSV document with a header row.
export function resultsToCsv(results: SpeedtestResult[]): string {
  const rows = results.map((r) => COLUMNS.map((c) => escapeCell(r[c])).join(','))
  return [COLUMNS.join(','), ...rows].join('\r\n') + '\r\n'
}

// downloadCsv triggers a browser download of the given CSV text.
export function downloadCsv(csv: string, filename: string): void {
  const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
