package db

import (
	"log"
)

func Migrate() error {
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
	`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
		return err
	}

	log.Println("Migration completed successfully")
	return nil
}
