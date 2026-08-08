CREATE TABLE IF NOT EXISTS speedtest_results (
    id              TEXT PRIMARY KEY,
    timestamp       DATETIME NOT NULL,          -- ISO-8601 UTC (e.g. 2026-08-08T12:00:00Z)
    success         INTEGER NOT NULL DEFAULT 1, -- 0 on a failed run
    error           TEXT,                       -- populated when success = 0
    download_mbps   REAL,
    upload_mbps     REAL,
    ping_ms         REAL,
    jitter_ms       REAL,
    packet_loss_pct REAL,
    duration_ms     REAL,                       -- how long the test took to run
    server_id       INTEGER,
    server_name     TEXT,
    server_location TEXT,
    isp             TEXT,
    external_ip     TEXT,
    result_url      TEXT
);

CREATE INDEX IF NOT EXISTS idx_speedtest_timestamp ON speedtest_results (timestamp);
