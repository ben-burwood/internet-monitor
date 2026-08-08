package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// serverSettingKey is the settings row holding the pinned speedtest server.
const serverSettingKey = "speedtest_server"

// ServerSelection is the speedtest server the tests should use. A nil (or
// non-positive) ID means "Automatic" — let the Ookla CLI pick the closest server.
type ServerSelection struct {
	ID       *int   `json:"server_id"`
	Name     string `json:"server_name,omitempty"`
	Location string `json:"server_location,omitempty"`
}

// EffectiveServerID is the id to pass to the speedtest CLI: 0 means Automatic.
// This is the single source of truth for the nil/non-positive → auto rule.
func (s ServerSelection) EffectiveServerID() int {
	if s.ID == nil || *s.ID <= 0 {
		return 0
	}
	return *s.ID
}

// Normalized collapses any "Automatic" selection to the canonical zero value so
// a blank/invalid id never carries a stale name or location.
func (s ServerSelection) Normalized() ServerSelection {
	if s.EffectiveServerID() == 0 {
		return ServerSelection{}
	}
	return s
}

// getSetting returns the stored value for key and whether it was present.
func getSetting(key string) (string, bool, error) {
	var value string
	err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

// setSetting upserts a settings key/value pair.
func setSetting(key, value string) error {
	_, err := db.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

// GetServerSelection returns the pinned speedtest server, or a zero value
// (Automatic) when none has been configured.
func GetServerSelection() (ServerSelection, error) {
	raw, ok, err := getSetting(serverSettingKey)
	if err != nil || !ok {
		return ServerSelection{}, err
	}
	var sel ServerSelection
	if err := json.Unmarshal([]byte(raw), &sel); err != nil {
		return ServerSelection{}, fmt.Errorf("parse stored server selection: %w", err)
	}
	return sel, nil
}

// SetServerSelection stores the pinned speedtest server, enforcing the
// Automatic invariant regardless of caller.
func SetServerSelection(sel ServerSelection) error {
	raw, err := json.Marshal(sel.Normalized())
	if err != nil {
		return err
	}
	return setSetting(serverSettingKey, string(raw))
}
