package db

import (
	"database/sql"
	"fmt"
)

func Migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS jobs (
		Id TEXT PRIMARY KEY,
		Command TEXT,
		State TEXT,
		Attempts INTEGER,
		Max_retries INTEGER,
		WorkerId TEXT,
		Next_run_at TEXT DEFAULT CURRENT_TIMESTAMP,
		Created_at TEXT,
		Updated_at TEXT
	);

	CREATE TABLE IF NOT EXISTS workers (
		WorkerId TEXT PRIMARY KEY,
		Started_at TEXT,
		Last_heartbeat TEXT
	);

	CREATE TABLE IF NOT EXISTS control (
		Key TEXT PRIMARY KEY,
		Value TEXT
	);

	CREATE TABLE IF NOT EXISTS config (
		Key TEXT PRIMARY KEY,
		Value TEXT
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Add missing columns to existing tables if needed
	db.Exec(`ALTER TABLE jobs ADD COLUMN WorkerId TEXT`)
	db.Exec(`ALTER TABLE jobs ADD COLUMN Next_run_at TEXT DEFAULT CURRENT_TIMESTAMP`)

	fmt.Println("Migration completed successfully")
	return nil
}
