package api

import (
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

// Stats handles GET /api/stats?from=&to= and returns min/max/avg per metric.
func Stats(w http.ResponseWriter, r *http.Request) {
	from, to := parseRange(r)
	results, err := database.ListBetween(from, to, maxLimit)
	if err != nil {
		http.Error(w, "failed to load results", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, computeStats(results))
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

// metricStat holds aggregate values for a single metric.
type metricStat struct {
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
	Avg *float64 `json:"avg"`
}

// statsResponse is returned by GET /api/stats.
type statsResponse struct {
	Count      int        `json:"count"`
	Successful int        `json:"successful"`
	Download   metricStat `json:"download_mbps"`
	Upload     metricStat `json:"upload_mbps"`
	Ping       metricStat `json:"ping_ms"`
	Jitter     metricStat `json:"jitter_ms"`
	PacketLoss metricStat `json:"packet_loss_pct"`
}

func computeStats(results []speedtest.Result) statsResponse {
	resp := statsResponse{Count: len(results)}
	dl := newAgg()
	ul := newAgg()
	ping := newAgg()
	jitter := newAgg()
	loss := newAgg()

	for _, r := range results {
		if !r.Success {
			continue
		}
		resp.Successful++
		dl.add(r.DownloadMbps)
		ul.add(r.UploadMbps)
		ping.add(r.PingMs)
		jitter.add(r.JitterMs)
		loss.add(r.PacketLossPct)
	}

	resp.Download = dl.stat()
	resp.Upload = ul.stat()
	resp.Ping = ping.stat()
	resp.Jitter = jitter.stat()
	resp.PacketLoss = loss.stat()
	return resp
}

type agg struct {
	min, max, sum float64
	n             int
}

func newAgg() *agg { return &agg{} }

func (a *agg) add(v *float64) {
	if v == nil {
		return
	}
	if a.n == 0 || *v < a.min {
		a.min = *v
	}
	if a.n == 0 || *v > a.max {
		a.max = *v
	}
	a.sum += *v
	a.n++
}

func (a *agg) stat() metricStat {
	if a.n == 0 {
		return metricStat{}
	}
	avg := a.sum / float64(a.n)
	min, max := a.min, a.max
	return metricStat{Min: &min, Max: &max, Avg: &avg}
}
