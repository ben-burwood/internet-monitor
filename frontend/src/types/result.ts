// SpeedtestResult mirrors the JSON returned by the Go backend.
export interface SpeedtestResult {
  id: string
  timestamp: string // ISO-8601
  success: boolean
  error?: string

  download_mbps: number | null
  upload_mbps: number | null
  ping_ms: number | null
  jitter_ms: number | null
  packet_loss_pct: number | null
  duration_ms: number | null

  server_id?: number
  server_name?: string
  server_location?: string
  isp?: string
  external_ip?: string
  result_url?: string
}

export interface MetricStat {
  min: number | null
  max: number | null
  avg: number | null
}

export interface Stats {
  count: number
  successful: number
  download_mbps: MetricStat
  upload_mbps: MetricStat
  ping_ms: MetricStat
  jitter_ms: MetricStat
  packet_loss_pct: MetricStat
}
