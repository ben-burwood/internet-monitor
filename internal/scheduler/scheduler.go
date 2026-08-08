package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"internet-monitor/internal/config"
	"internet-monitor/internal/database"
	"internet-monitor/internal/speedtest"
)

// Scheduler runs speedtests on an interval and persists the results.
type Scheduler struct {
	interval  time.Duration
	retention time.Duration // zero means keep results forever
}

// New creates a Scheduler that runs a test every interval and, when retention
// is > 0, prunes results older than retention after each run.
func New(interval, retention time.Duration) *Scheduler {
	return &Scheduler{interval: interval, retention: retention}
}

// Run blocks until ctx is cancelled. It performs the first test after a short
// startup delay, then repeats every configured interval.
func (s *Scheduler) Run(ctx context.Context) {
	slog.Info("scheduler started", "startup_delay", config.StartupDelay, "interval", s.interval)

	select {
	case <-time.After(config.StartupDelay):
	case <-ctx.Done():
		return
	}
	s.RunOnce(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.RunOnce(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// RunOnce executes a single speedtest, stores the outcome (success or failure),
// and prunes old rows if retention is configured. It returns the stored result.
func (s *Scheduler) RunOnce(ctx context.Context) speedtest.Result {
	runCtx, cancel := context.WithTimeout(ctx, config.TestTimeout)
	defer cancel()

	serverID := 0
	if sel, err := database.GetServerSelection(); err != nil {
		slog.Error("failed to load server selection, using automatic", "error", err)
	} else {
		serverID = sel.EffectiveServerID()
	}

	slog.Info("running speedtest", "server_id", serverID)
	start := time.Now()
	res, err := speedtest.Run(runCtx, serverID)
	durationMs := float64(time.Since(start).Milliseconds())

	if err != nil {
		slog.Error("speedtest failed", "error", err, "duration_ms", durationMs)
		res = speedtest.Result{
			ID:        uuid.NewString(),
			Timestamp: time.Now().UTC(),
			Success:   false,
			Error:     err.Error(),
		}
	} else {
		slog.Info("speedtest complete",
			"download_mbps", deref(res.DownloadMbps),
			"upload_mbps", deref(res.UploadMbps),
			"ping_ms", deref(res.PingMs),
			"jitter_ms", deref(res.JitterMs),
			"packet_loss_pct", deref(res.PacketLossPct),
			"duration_ms", durationMs,
			"server", res.ServerName,
			"isp", res.ISP,
		)
	}
	res.DurationMs = &durationMs

	if err := database.Insert(res); err != nil {
		slog.Error("failed to store result", "error", err)
	}

	s.prune()
	return res
}

func (s *Scheduler) prune() {
	if s.retention <= 0 {
		return
	}
	cutoff := time.Now().UTC().Add(-s.retention)
	n, err := database.PruneOlderThan(cutoff)
	if err != nil {
		slog.Error("prune failed", "error", err)
		return
	}
	if n > 0 {
		slog.Info("pruned old results", "count", n, "retention", s.retention)
	}
}

func deref(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
