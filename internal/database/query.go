package database

import (
	"database/sql"
	"fmt"
	"internet-monitor/internal/speedtest"
	"time"
)

// Insert stores a single speedtest result.
func Insert(r speedtest.Result) error {
	_, err := db.Exec(
		`INSERT INTO speedtest_results (
			id, timestamp, success, error,
			download_mbps, upload_mbps, ping_ms, jitter_ms, packet_loss_pct, duration_ms,
			server_id, server_name, server_location, isp, external_ip, result_url
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Timestamp.UTC().Format(time.RFC3339), boolToInt(r.Success), nullString(r.Error),
		r.DownloadMbps, r.UploadMbps, r.PingMs, r.JitterMs, r.PacketLossPct, r.DurationMs,
		nullInt(r.ServerID), nullString(r.ServerName), nullString(r.ServerLocation),
		nullString(r.ISP), nullString(r.ExternalIP), nullString(r.ResultURL),
	)
	return err
}

// ListBetween returns results with timestamp in [from, to], newest first,
// capped at limit rows.
func ListBetween(from, to time.Time, limit int) ([]speedtest.Result, error) {
	rows, err := db.Query(
		`SELECT id, timestamp, success, error,
			download_mbps, upload_mbps, ping_ms, jitter_ms, packet_loss_pct, duration_ms,
			server_id, server_name, server_location, isp, external_ip, result_url
		 FROM speedtest_results
		 WHERE timestamp BETWEEN ? AND ?
		 ORDER BY timestamp DESC
		 LIMIT ?`,
		from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []speedtest.Result{}
	for rows.Next() {
		r, err := scanResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// Latest returns the most recent result, or (nil, nil) if there are none.
func Latest() (*speedtest.Result, error) {
	row := db.QueryRow(
		`SELECT id, timestamp, success, error,
			download_mbps, upload_mbps, ping_ms, jitter_ms, packet_loss_pct, duration_ms,
			server_id, server_name, server_location, isp, external_ip, result_url
		 FROM speedtest_results
		 ORDER BY timestamp DESC
		 LIMIT 1`,
	)
	r, err := scanResult(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// PruneOlderThan deletes results with a timestamp before cutoff and returns
// the number of rows removed.
func PruneOlderThan(cutoff time.Time) (int64, error) {
	res, err := db.Exec(
		`DELETE FROM speedtest_results WHERE timestamp < ?`,
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// IPSegment is a contiguous period during which the public IP stayed constant.
type IPSegment struct {
	IP        string    `json:"ip"`
	ISP       string    `json:"isp,omitempty"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Count     int       `json:"count"` // number of successful tests on this IP
}

// IPHistory returns the public IP over time, collapsing consecutive tests on the
// same IP into a single segment. Segments are newest-first, so the current IP is
// first. Failed runs and rows without an external IP are ignored.
func IPHistory() ([]IPSegment, error) {
	rows, err := db.Query(
		`SELECT external_ip, COALESCE(isp, ''), timestamp
		 FROM speedtest_results
		 WHERE success = 1 AND external_ip IS NOT NULL AND external_ip <> ''
		 ORDER BY timestamp ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	segments := []IPSegment{}
	for rows.Next() {
		var ip, isp, ts string
		if err := rows.Scan(&ip, &isp, &ts); err != nil {
			return nil, err
		}
		parsed, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil, fmt.Errorf("parse stored timestamp %q: %w", ts, err)
		}

		if n := len(segments); n > 0 && segments[n-1].IP == ip {
			segments[n-1].LastSeen = parsed
			segments[n-1].Count++
			continue
		}
		segments = append(segments, IPSegment{IP: ip, ISP: isp, FirstSeen: parsed, LastSeen: parsed, Count: 1})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Reverse to newest-first.
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}
	return segments, nil
}

// scanRow is satisfied by both *sql.Row and *sql.Rows.
type scanRow interface {
	Scan(dest ...any) error
}

func scanResult(s scanRow) (speedtest.Result, error) {
	var (
		r         speedtest.Result
		ts        string
		success   int
		errText   sql.NullString
		serverID  sql.NullInt64
		serverN   sql.NullString
		serverLoc sql.NullString
		isp       sql.NullString
		extIP     sql.NullString
		resultURL sql.NullString
	)
	if err := s.Scan(
		&r.ID, &ts, &success, &errText,
		&r.DownloadMbps, &r.UploadMbps, &r.PingMs, &r.JitterMs, &r.PacketLossPct, &r.DurationMs,
		&serverID, &serverN, &serverLoc, &isp, &extIP, &resultURL,
	); err != nil {
		return speedtest.Result{}, err
	}

	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return speedtest.Result{}, fmt.Errorf("parse stored timestamp %q: %w", ts, err)
	}
	r.Timestamp = parsed
	r.Success = success != 0
	r.Error = errText.String
	r.ServerID = int(serverID.Int64)
	r.ServerName = serverN.String
	r.ServerLocation = serverLoc.String
	r.ISP = isp.String
	r.ExternalIP = extIP.String
	r.ResultURL = resultURL.String
	return r, nil
}
