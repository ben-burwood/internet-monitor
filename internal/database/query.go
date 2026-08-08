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
			download_mbps, upload_mbps, ping_ms, jitter_ms, packet_loss_pct,
			server_id, server_name, server_location, isp, external_ip, result_url
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Timestamp.UTC().Format(time.RFC3339), boolToInt(r.Success), nullString(r.Error),
		r.DownloadMbps, r.UploadMbps, r.PingMs, r.JitterMs, r.PacketLossPct,
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
			download_mbps, upload_mbps, ping_ms, jitter_ms, packet_loss_pct,
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
			download_mbps, upload_mbps, ping_ms, jitter_ms, packet_loss_pct,
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
		&r.DownloadMbps, &r.UploadMbps, &r.PingMs, &r.JitterMs, &r.PacketLossPct,
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
