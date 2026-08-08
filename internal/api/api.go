package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"internet-monitor/internal/database"
	"internet-monitor/internal/scheduler"
	"internet-monitor/internal/speedtest"
)

// defaultRange is used by /api/results and /api/stats when no `from` is given.
const defaultRange = 7 * 24 * time.Hour

// maxLimit caps how many rows /api/results will return.
const maxLimit = 10000

// serversTimeout caps how long listing nearby servers may take.
const serversTimeout = 30 * time.Second

// ListResults handles GET /api/results?from=<iso8601>&to=<iso8601>&limit=<n>.
func ListResults(w http.ResponseWriter, r *http.Request) {
	from, to := parseRange(r)
	limit := parseLimit(r, maxLimit)

	results, err := database.ListBetween(from, to, limit)
	if err != nil {
		http.Error(w, "failed to load results", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// LatestResult handles GET /api/results/latest.
func LatestResult(w http.ResponseWriter, r *http.Request) {
	latest, err := database.Latest()
	if err != nil {
		http.Error(w, "failed to load latest result", http.StatusInternalServerError)
		return
	}
	if latest == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	writeJSON(w, http.StatusOK, latest)
}

// ListServers handles GET /api/servers — the nearest Ookla servers to pin.
func ListServers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), serversTimeout)
	defer cancel()

	servers, err := speedtest.ListServers(ctx)
	if err != nil {
		http.Error(w, "failed to list servers", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, servers)
}

// GetSettings handles GET /api/settings — the current server selection.
func GetSettings(w http.ResponseWriter, r *http.Request) {
	sel, err := database.GetServerSelection()
	if err != nil {
		http.Error(w, "failed to load settings", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, sel)
}

// UpdateSettings handles PUT /api/settings — set (or clear) the pinned server.
func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var sel database.ServerSelection
	if err := json.NewDecoder(r.Body).Decode(&sel); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	sel = sel.Normalized()

	if err := database.SetServerSelection(sel); err != nil {
		http.Error(w, "failed to save settings", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, sel)
}

// IPHistory handles GET /api/ip-history — the public IP over time, newest first.
func IPHistory(w http.ResponseWriter, r *http.Request) {
	segments, err := database.IPHistory()
	if err != nil {
		http.Error(w, "failed to load ip history", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, segments)
}

// RunNow returns a handler for POST /api/run that triggers an on-demand test
// via the scheduler and returns the result.
func RunNow(sched *scheduler.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := sched.RunOnce(r.Context())
		status := http.StatusCreated
		if !res.Success {
			status = http.StatusBadGateway
		}
		writeJSON(w, status, res)
	}
}

// parseRange reads `from`/`to` ISO-8601 query params, defaulting to the last
// defaultRange window ending now.
func parseRange(r *http.Request) (from, to time.Time) {
	now := time.Now().UTC()
	to = now
	from = now.Add(-defaultRange)

	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}
	return from, to
}

func parseLimit(r *http.Request, fallback int) int {
	v := r.URL.Query().Get("limit")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 || n > fallback {
		return fallback
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
