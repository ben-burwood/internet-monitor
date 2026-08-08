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

export interface IPSegment {
  ip: string
  isp?: string
  first_seen: string // ISO-8601
  last_seen: string // ISO-8601
  count: number
}

// A nearby Ookla server the user can pin.
export interface Server {
  id: number
  name: string
  location: string
  country?: string
  host?: string
  port?: number
}

// The configured speedtest server. server_id null means Automatic (closest).
export interface ServerSelection {
  server_id: number | null
  server_name?: string
  server_location?: string
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
