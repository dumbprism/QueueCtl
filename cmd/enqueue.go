package cmd

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var userCommand string

var enqueueCmd = &cobra.Command{
	Use:   "enqueue",
	Short: "Add a job to the queue",
	Long:  "This command adds a new job to the SQLite database with all required fields.",
	Run: func(cmd *cobra.Command, args []string) {

		os.MkdirAll("data", 0755)

		database, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("Error connecting to SQLite database:", err)
			return
		}
		defer database.Close()

		// Create table with all columns including WorkerId and Next_run_at
		createTableQuery := `
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
		);`

		_, err = database.Exec(createTableQuery)
		if err != nil {
			fmt.Println("Error creating jobs table:", err)
			return
		}

		var charset = "abcdefghijklmnopqrstuvwxyz1234567890"
		var jobId = ""

		for i := 0; i < 8; i++ {
			jobId += string(charset[rand.Intn(len(charset))])
		}

		var jobAttempts = 0
		var job_maxRetries = 0

		var job jobSpec
		job.Id = jobId

		if userCommand == "" {
			job.Command = "command not found"
		} else {
			job.Command = userCommand
		}

		job.State = "pending"
		job.Attempts = jobAttempts
		job.Max_retries = job_maxRetries
		job.Created_at = time.Now().Format("2006-01-02 15:04:05")
		job.Updated_at = time.Now().Format("2006-01-02 15:04:05")

		insertQuery := `
			INSERT INTO jobs (
				Id, Command, State, Attempts, Max_retries, WorkerId, Next_run_at, Created_at, Updated_at
			) VALUES (?, ?, ?, ?, ?, NULL, datetime('now'), ?, ?)
		`

		_, err = database.Exec(
			insertQuery,
			job.Id,
			job.Command,
			job.State,
			job.Attempts,
			job.Max_retries,
			job.Created_at,
			job.Updated_at,
		)

		if err != nil {
			fmt.Println("Error inserting values:", err)
			return
		}

		fmt.Println("Job added successfully with ID:", job.Id)
	},
}

func init() {
	rootCmd.AddCommand(enqueueCmd)
	enqueueCmd.Flags().StringVarP(&userCommand, "command", "c", "", "Command for the job")
}
