package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbPath = "store/internet-monitor.db"

//go:embed schema.sql
var schema string

var db *sql.DB

// Init opens (creating if needed) the SQLite database at path and runs migrations.
func Init() error {
	if dir := filepath.Dir(dbPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	db = conn
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := addMissingColumns(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// addedColumns are speedtest_results columns introduced after the original
// schema. CREATE TABLE IF NOT EXISTS leaves existing databases untouched, so
// Init adds any of these that are missing.
var addedColumns = []struct{ name, def string }{
	{"download_bytes", "INTEGER"},
	{"upload_bytes", "INTEGER"},
}

func addMissingColumns() error {
	rows, err := db.Query(`SELECT name FROM pragma_table_info('speedtest_results')`)
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		existing[name] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, c := range addedColumns {
		if existing[c.name] {
			continue
		}
		if _, err := db.Exec(`ALTER TABLE speedtest_results ADD COLUMN ` + c.name + ` ` + c.def); err != nil {
			return fmt.Errorf("add column %s: %w", c.name, err)
		}
	}
	return nil
}

func Close() error {
	if db == nil {
		return nil
	}
	return db.Close()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(i int) any {
	if i == 0 {
		return nil
	}
	return i
}
