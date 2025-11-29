package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var dlqCmd = &cobra.Command{
	Use:   "dlq",
	Short: "Dead Letter Queue operations",
}

// ---------------------------------------------------------
// LIST DEAD JOBS → queuectl dlq list
// ---------------------------------------------------------
var dlqListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all dead jobs",
	Run: func(cmd *cobra.Command, args []string) {

		os.MkdirAll("data", 0755)

		db, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("DB error:", err)
			return
		}
		defer db.Close()

		fmt.Println("\n===== DEAD LETTER QUEUE =====")

		rows, err := db.Query(`
			SELECT Id, Command, Attempts, Max_retries, Updated_at
			FROM jobs
			WHERE State = 'dead'
			ORDER BY Updated_at DESC
		`)
		if err != nil {
			fmt.Println("Query error:", err)
			return
		}
		defer rows.Close()

		found := false

		for rows.Next() {
			found = true

			var Id, Command, Updated string
			var Attempts, MaxRetries int

			err = rows.Scan(&Id, &Command, &Attempts, &MaxRetries, &Updated)
			if err != nil {
				fmt.Println("Error reading job:", err)
				continue
			}

			fmt.Printf(
				"\nID: %s\nCommand: %s\nAttempts: %d\nMax_retries: %d\nUpdated: %s\n",
				Id, Command, Attempts, MaxRetries, Updated,
			)
		}

		if !found {
			fmt.Println("No dead jobs.")
		}

		fmt.Println("\n==============================\n")
	},
}

// ---------------------------------------------------------
// RETRY DEAD JOB → queuectl dlq retry <jobId>
// ---------------------------------------------------------
var dlqRetryCmd = &cobra.Command{
	Use:   "retry <jobId>",
	Short: "Retry a dead job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		jobId := args[0]

		os.MkdirAll("data", 0755)

		db, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("DB error:", err)
			return
		}
		defer db.Close()

		// Check if job exists & is dead
		var currentState string
		err = db.QueryRow(`
			SELECT State FROM jobs WHERE Id = ?
		`, jobId).Scan(&currentState)

		if err == sql.ErrNoRows {
			fmt.Println("No such job:", jobId)
			return
		}
		if err != nil {
			fmt.Println("Error reading job:", err)
			return
		}

		if currentState != "dead" {
			fmt.Println("Job is not dead. Cannot retry via DLQ.")
			return
		}

		// Reset job and requeue it
		_, err = db.Exec(`
			UPDATE jobs
			SET 
				State = 'pending',
				Attempts = 0,
				Max_retries = 3,
				Updated_at = ?
			WHERE Id = ?
		`, time.Now().Format("2006-01-02 15:04:05"), jobId)

		if err != nil {
			fmt.Println("Error retrying job:", err)
			return
		}

		fmt.Println("Job", jobId, "has been requeued with max_retries = 3.")
	},
}

func init() {
	rootCmd.AddCommand(dlqCmd)
	dlqCmd.AddCommand(dlqListCmd)
	dlqCmd.AddCommand(dlqRetryCmd)
}
