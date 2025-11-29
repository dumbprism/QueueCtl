package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all job states and active workers",
	Run: func(cmd *cobra.Command, args []string) {

		// Ensure database folder exists
		os.MkdirAll("data", 0755)

		// Connect to database
		db, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("Database error:", err)
			return
		}
		defer db.Close()

		// ----------------------------------------------------
		// 1. FETCH AND DISPLAY ALL JOB STATES
		// ----------------------------------------------------
		fmt.Println("\n===== JOB STATES =====")

		jobRows, err := db.Query(`SELECT Id, State FROM jobs ORDER BY Created_at DESC`)
		if err != nil {
			fmt.Println("Error fetching jobs:", err)
			return
		}
		defer jobRows.Close()

		hasJobs := false

		for jobRows.Next() {
			hasJobs = true
			var Id string
			var State string

			err = jobRows.Scan(&Id, &State)
			if err != nil {
				fmt.Println("Error reading job:", err)
				continue
			}

			fmt.Printf("ID: %s   State: %s\n", Id, State)
		}

		if !hasJobs {
			fmt.Println("No jobs found.")
		}

		// ----------------------------------------------------
		// 2. FETCH AND DISPLAY ACTIVE WORKERS
		// ----------------------------------------------------
		fmt.Println("\n===== ACTIVE WORKERS =====")

		workerRows, err := db.Query(`
			SELECT WorkerId, Last_heartbeat 
			FROM workers 
			ORDER BY Last_heartbeat DESC
		`)
		if err != nil {
			fmt.Println("Error fetching workers:", err)
			return
		}
		defer workerRows.Close()

		hasWorkers := false

		for workerRows.Next() {
			hasWorkers = true

			var WorkerId string
			var LastHeartbeat string

			err = workerRows.Scan(&WorkerId, &LastHeartbeat)
			if err != nil {
				fmt.Println("Error reading worker:", err)
				continue
			}

			fmt.Printf("Worker: %s   Heartbeat: %s\n", WorkerId, LastHeartbeat)
		}

		if !hasWorkers {
			fmt.Println("No active workers.")
		}

		fmt.Println("===========================")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
