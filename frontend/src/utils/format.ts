// Shared display formatters used across the dashboard components.

// formatMetric renders a numeric metric to a fixed precision, or an em dash
// when the value is missing (failed runs serialise metrics as null).
export function formatMetric(v: number | null, digits = 1): string {
  return v == null ? '—' : v.toFixed(digits)
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
