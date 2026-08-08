package speedtest

import (
	"encoding/json"
	"testing"
	"time"
)

// sampleJSON is a representative `speedtest -f json` result payload.
const sampleJSON = `{
  "type": "result",
  "timestamp": "2026-08-08T12:00:00Z",
  "ping": { "jitter": 1.25, "latency": 10.5, "low": 9.8, "high": 12.1 },
  "download": { "bandwidth": 11875000, "bytes": 100000000, "elapsed": 5000 },
  "upload": { "bandwidth": 2375000, "bytes": 20000000, "elapsed": 5000 },
  "packetLoss": 0.5,
  "isp": "Example ISP",
  "interface": { "externalIp": "203.0.113.5" },
  "server": { "id": 1234, "name": "Example Server", "location": "London", "country": "United Kingdom" },
  "result": { "id": "abc-123", "url": "https://www.speedtest.net/result/c/abc-123", "persisted": true }
}`

func TestBandwidthToMbps(t *testing.T) {
	tests := []struct {
		name  string
		bytes float64
		want  float64
	}{
		{"zero", 0, 0},
		{"1MBps is 8Mbps", 1_000_000, 8},
		{"12.5MBps is 100Mbps", 12_500_000, 100},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := bandwidthToMbps(tc.bytes); got != tc.want {
				t.Errorf("bandwidthToMbps(%v) = %v, want %v", tc.bytes, got, tc.want)
			}
		})
	}
}

func TestOoklaResultToResult(t *testing.T) {
	var o ooklaResult
	if err := json.Unmarshal([]byte(sampleJSON), &o); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	r := o.toResult("test-id")

	if r.ID != "test-id" {
		t.Errorf("ID = %q, want test-id", r.ID)
	}
	if !r.Success {
		t.Error("Success = false, want true")
	}
	wantTS := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	if !r.Timestamp.Equal(wantTS) {
		t.Errorf("Timestamp = %v, want %v", r.Timestamp, wantTS)
	}

	assertMetric(t, "DownloadMbps", r.DownloadMbps, 95.0) // 11_875_000 * 8 / 1e6
	assertMetric(t, "UploadMbps", r.UploadMbps, 19.0)     // 2_375_000 * 8 / 1e6
	assertMetric(t, "PingMs", r.PingMs, 10.5)
	assertMetric(t, "JitterMs", r.JitterMs, 1.25)
	assertMetric(t, "PacketLossPct", r.PacketLossPct, 0.5)

	if r.ServerID != 1234 {
		t.Errorf("ServerID = %d, want 1234", r.ServerID)
	}
	if r.ServerName != "Example Server" {
		t.Errorf("ServerName = %q", r.ServerName)
	}
	if r.ServerLocation != "London, United Kingdom" {
		t.Errorf("ServerLocation = %q, want %q", r.ServerLocation, "London, United Kingdom")
	}
	if r.ISP != "Example ISP" {
		t.Errorf("ISP = %q", r.ISP)
	}
	if r.ExternalIP != "203.0.113.5" {
		t.Errorf("ExternalIP = %q", r.ExternalIP)
	}
	if r.ResultURL != "https://www.speedtest.net/result/c/abc-123" {
		t.Errorf("ResultURL = %q", r.ResultURL)
	}
}

func assertMetric(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Errorf("%s = nil, want %v", name, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %v, want %v", name, *got, want)
	}
}
