// Shared display formatters used across the dashboard components.

// formatMetric renders a numeric metric to a fixed precision, or an em dash
// when the value is missing (failed runs serialise metrics as null).
export function formatMetric(v: number | null, digits = 1): string {
  return v == null ? '—' : v.toFixed(digits)
}

// formatBytes renders a byte count with decimal units (1 MB = 1,000,000 bytes, as ISPs and data caps count).
export function formatBytes(v: number | null): string {
  if (v == null) return '—'
  const units = ['B', 'kB', 'MB', 'GB', 'TB']
  let i = 0
  while (v >= 1000 && i < units.length - 1) {
    v /= 1000
    i++
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

// formatDateTime renders an ISO-8601 timestamp as a locale date+time.
export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString()
}

// formatDate renders an ISO-8601 timestamp as a locale date (no time).
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}
