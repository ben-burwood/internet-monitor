package config

import (
	"log/slog"
	"os"
	"time"
)

const (
	// TestTimeout caps how long a single speedtest may run.
	TestTimeout = 5 * time.Minute
	// StartupDelay is how long after boot the first test runs.
	StartupDelay = 1 * time.Minute
	// defaultSpeedtestInterval is how often to run a Speedtest in the background.
	defaultSpeedtestInterval = 1 * time.Hour
	// defaultDataRetention is set to 0, meaning old results are never pruned.
	defaultDataRetention = 0
)

// Config holds all runtime configuration, sourced from environment variables.
type Config struct {
	// Interval is how long to wait between speedtests.
	Interval time.Duration
	// Retention, when > 0, prunes results older than this after each run.
	// Zero means keep results forever.
	Retention time.Duration
}

// Load reads configuration from the environment, applying defaults.
func Load() Config {
	return Config{
		Interval:  getDuration("SPEEDTEST_INTERVAL", defaultSpeedtestInterval),
		Retention: getDuration("DATA_RETENTION", defaultDataRetention),
	}
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		slog.Warn("invalid duration, using default", "key", key, "value", v, "default", fallback, "error", err)
		return fallback
	}
	return d
}
