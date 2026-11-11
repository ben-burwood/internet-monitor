package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbPath = "store/internet-monitor.db"

//go:embed migrations/*.sql
var migrations embed.FS

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
	if err := migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// migrate applies each migration newer than the database's user_version, in
// filename order. Migration N (1-based) sets user_version to N in the same
// transaction, so a failed migration rolls back and is retried next start.
//
// To change the schema, add the next numbered file to migrations/ — never edit
// one that has shipped. Databases created before versioning report 0; 001 uses
// IF NOT EXISTS so it is a no-op for them.
func migrate() error {
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}

	files, err := fs.Glob(migrations, "migrations/*.sql") // sorted by name
	if err != nil {
		return err
	}
	for i, name := range files {
		target := i + 1
		if target <= version {
			continue
		}
		stmts, err := migrations.ReadFile(name)
		if err != nil {
			return err
		}
		if err := applyMigration(string(stmts), target); err != nil {
			return fmt.Errorf("%s: %w", path.Base(name), err)
		}
	}
	return nil
}

func applyMigration(stmts string, version int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(stmts); err != nil {
		return err
	}
	// PRAGMA can't take a bound parameter; version is a trusted int.
	if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, version)); err != nil {
		return err
	}
	return tx.Commit()
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
