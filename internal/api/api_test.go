package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestSettingsRoundTrip(t *testing.T) {
	setup(t) // initialises a temp DB

	// Default is Automatic (null server_id).
	rec := httptest.NewRecorder()
	GetSettings(rec, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	var sel database.ServerSelection
	if err := json.Unmarshal(rec.Body.Bytes(), &sel); err != nil {
		t.Fatalf("decode default: %v", err)
	}
	if sel.ID != nil {
		t.Errorf("default server_id = %v, want null", *sel.ID)
	}

	// Pin a server.
	body := `{"server_id":28459,"server_name":"Oxford","server_location":"Oxford"}`
	rec = httptest.NewRecorder()
	UpdateSettings(rec, httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", rec.Code)
	}

	rec = httptest.NewRecorder()
	GetSettings(rec, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	if err := json.Unmarshal(rec.Body.Bytes(), &sel); err != nil {
		t.Fatalf("decode pinned: %v", err)
	}
	if sel.ID == nil || *sel.ID != 28459 || sel.Name != "Oxford" {
		t.Errorf("pinned selection = %+v", sel)
	}

	// A non-positive id resets to Automatic.
	rec = httptest.NewRecorder()
	UpdateSettings(rec, httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(`{"server_id":0}`)))
	rec = httptest.NewRecorder()
	GetSettings(rec, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	if err := json.Unmarshal(rec.Body.Bytes(), &sel); err != nil {
		t.Fatalf("decode reset: %v", err)
	}
	if sel.ID != nil {
		t.Errorf("reset server_id = %v, want null", *sel.ID)
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
