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

	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migration completed successfully")
	return nil
}
