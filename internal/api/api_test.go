package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"internet-monitor/internal/database"
	"internet-monitor/internal/speedtest"
)

func f(v float64) *float64 { return &v }

func setup(t *testing.T) http.Handler {
	t.Helper()
	// database.Init uses a fixed relative path; isolate it in a temp working directory.
	t.Chdir(t.TempDir())
	if err := database.Init(); err != nil {
		t.Fatalf("db init: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Mirrors the API routes wired up in main.go (RunNow needs a scheduler and is not exercised here).
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/results", ListResults)
	mux.HandleFunc("GET /api/results/latest", LatestResult)
	mux.HandleFunc("GET /api/stats", Stats)
	return mux
}

func seed(t *testing.T) {
	t.Helper()
	base := time.Now().UTC().Add(-2 * time.Hour)
	rows := []speedtest.Result{
		{ID: "a", Timestamp: base, Success: true, DownloadMbps: f(90), UploadMbps: f(20), PingMs: f(10), JitterMs: f(1), PacketLossPct: f(0), ServerName: "S1"},
		{ID: "b", Timestamp: base.Add(time.Hour), Success: true, DownloadMbps: f(100), UploadMbps: f(25), PingMs: f(12), JitterMs: f(2), PacketLossPct: f(1)},
		{ID: "c", Timestamp: base.Add(90 * time.Minute), Success: false, Error: "no servers"},
	}
	for _, r := range rows {
		if err := database.Insert(r); err != nil {
			t.Fatalf("seed insert %s: %v", r.ID, err)
		}
	}
}

func TestListResults(t *testing.T) {
	h := setup(t)
	seed(t)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/results", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var results []speedtest.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	// Newest first: the failed run at +90m.
	if results[0].ID != "c" || results[0].Success {
		t.Errorf("first result = %+v, want failed run c", results[0])
	}
	if results[0].DownloadMbps != nil {
		t.Errorf("failed run download = %v, want null", *results[0].DownloadMbps)
	}
}

func TestLatestResult(t *testing.T) {
	h := setup(t)
	seed(t)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/results/latest", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var latest speedtest.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &latest); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if latest.ID != "c" {
		t.Errorf("latest = %s, want c", latest.ID)
	}
}

func TestStats(t *testing.T) {
	h := setup(t)
	seed(t)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/stats", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var stats statsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &stats); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stats.Count != 3 {
		t.Errorf("Count = %d, want 3", stats.Count)
	}
	if stats.Successful != 2 {
		t.Errorf("Successful = %d, want 2", stats.Successful)
	}
	// Only successful runs feed aggregates: download 90 and 100.
	if stats.Download.Min == nil || *stats.Download.Min != 90 {
		t.Errorf("download min = %v, want 90", stats.Download.Min)
	}
	if stats.Download.Max == nil || *stats.Download.Max != 100 {
		t.Errorf("download max = %v, want 100", stats.Download.Max)
	}
	if stats.Download.Avg == nil || *stats.Download.Avg != 95 {
		t.Errorf("download avg = %v, want 95", stats.Download.Avg)
	}
}

func TestLatestEmpty(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/results/latest", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); body != "null\n" {
		t.Errorf("body = %q, want null", body)
	}
}
