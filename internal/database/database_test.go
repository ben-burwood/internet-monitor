package database

import (
	"testing"
	"time"

	"internet-monitor/internal/speedtest"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	// Init uses a fixed relative path; isolate it in a temp working directory.
	t.Chdir(t.TempDir())
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(func() { Close() })
}

func f(v float64) *float64 { return &v }

func successResult(id string, ts time.Time, dl float64) speedtest.Result {
	return speedtest.Result{
		ID:            id,
		Timestamp:     ts,
		Success:       true,
		DownloadMbps:  f(dl),
		UploadMbps:    f(20),
		PingMs:        f(10),
		JitterMs:      f(1),
		PacketLossPct: f(0),
		ServerID:      1234,
		ServerName:    "Example",
		ISP:           "Example ISP",
		ExternalIP:    "203.0.113.5",
		ResultURL:     "https://example/result/" + id,
	}
}

func TestInsertAndLatest(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	if err := Insert(successResult("a", base, 90)); err != nil {
		t.Fatalf("Insert a: %v", err)
	}
	if err := Insert(successResult("b", base.Add(time.Hour), 100)); err != nil {
		t.Fatalf("Insert b: %v", err)
	}

	latest, err := Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest == nil || latest.ID != "b" {
		t.Fatalf("Latest = %+v, want id b", latest)
	}
	if latest.DownloadMbps == nil || *latest.DownloadMbps != 100 {
		t.Errorf("latest download = %v, want 100", latest.DownloadMbps)
	}
	if latest.ServerName != "Example" || latest.ISP != "Example ISP" {
		t.Errorf("context fields not round-tripped: %+v", latest)
	}
}

func TestLatestEmpty(t *testing.T) {
	setupTestDB(t)
	latest, err := Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest != nil {
		t.Errorf("Latest on empty db = %+v, want nil", latest)
	}
}

func TestInsertFailureRoundTrip(t *testing.T) {
	setupTestDB(t)

	ts := time.Date(2026, 8, 8, 13, 0, 0, 0, time.UTC)
	fail := speedtest.Result{ID: "fail", Timestamp: ts, Success: false, Error: "no servers"}
	if err := Insert(fail); err != nil {
		t.Fatalf("Insert failure: %v", err)
	}

	latest, err := Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest.Success {
		t.Error("Success = true, want false")
	}
	if latest.Error != "no servers" {
		t.Errorf("Error = %q, want %q", latest.Error, "no servers")
	}
	if latest.DownloadMbps != nil {
		t.Errorf("DownloadMbps = %v, want nil for a failed run", *latest.DownloadMbps)
	}
}

func TestListBetween(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC)
	for i := range 5 {
		id := string(rune('a' + i))
		if err := Insert(successResult(id, base.Add(time.Duration(i)*time.Hour), 90)); err != nil {
			t.Fatalf("Insert %s: %v", id, err)
		}
	}

	// Window covering the middle three rows (hours 1..3).
	from := base.Add(time.Hour)
	to := base.Add(3 * time.Hour)
	got, err := ListBetween(from, to, 100)
	if err != nil {
		t.Fatalf("ListBetween: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("ListBetween returned %d rows, want 3", len(got))
	}
	// Newest first.
	if got[0].ID != "d" || got[2].ID != "b" {
		t.Errorf("order wrong: got %s..%s, want d..b", got[0].ID, got[2].ID)
	}

	// Limit is honoured.
	limited, err := ListBetween(base.Add(-time.Hour), base.Add(10*time.Hour), 2)
	if err != nil {
		t.Fatalf("ListBetween limited: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("limit not honoured: got %d rows, want 2", len(limited))
	}
}

func TestPruneOlderThan(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC)
	for i := range 5 {
		id := string(rune('a' + i))
		if err := Insert(successResult(id, base.Add(time.Duration(i)*time.Hour), 90)); err != nil {
			t.Fatalf("Insert %s: %v", id, err)
		}
	}

	n, err := PruneOlderThan(base.Add(2 * time.Hour))
	if err != nil {
		t.Fatalf("PruneOlderThan: %v", err)
	}
	if n != 2 { // hours 0 and 1 removed
		t.Errorf("pruned %d rows, want 2", n)
	}

	remaining, err := ListBetween(base.Add(-time.Hour), base.Add(10*time.Hour), 100)
	if err != nil {
		t.Fatalf("ListBetween: %v", err)
	}
	if len(remaining) != 3 {
		t.Errorf("remaining = %d rows, want 3", len(remaining))
	}
}
