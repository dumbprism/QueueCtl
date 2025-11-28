package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Connect() error {
	var err error
	DB, err := sql.Open("sqlite", "./data/queue.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return err
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
		return err
	}

	log.Println("Connected to SQLite database successfully")
	return nil
}

func Close() error {
	return DB.Close()
}
