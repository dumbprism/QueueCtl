package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var checkStateCmd string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs from the database",
	Run: func(cmd *cobra.Command, args []string) {

		os.MkdirAll("data", 0755)

		db, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("DB error:", err)
			return
		}
		defer db.Close()

		var query string
		if checkStateCmd == "" {
			query = `
				SELECT Id, Command, State, Attempts, Max_retries,
				       Created_at, Updated_at, Next_run_at, WorkerId
				FROM jobs
			`
		} else {
			query = `
				SELECT Id, Command, State, Attempts, Max_retries,
				       Created_at, Updated_at, Next_run_at, WorkerId
				FROM jobs
				WHERE State = ?
			`
		}

		var rows *sql.Rows
		if checkStateCmd == "" {
			rows, err = db.Query(query)
		} else {
			rows, err = db.Query(query, checkStateCmd)
		}

		if err != nil {
			fmt.Println("Query error:", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var Id, Command, State, Created_at, Updated_at string
			var Next_run_at sql.NullString
			var WorkerId sql.NullString
			var Attempts, Max_retries int

			err = rows.Scan(&Id, &Command, &State, &Attempts, &Max_retries,
				&Created_at, &Updated_at, &Next_run_at, &WorkerId)

			if err != nil {
				fmt.Println("Row error:", err)
				continue
			}

			worker := "none"
			if WorkerId.Valid && WorkerId.String != "" {
				worker = WorkerId.String
			}

			fmt.Printf(`
ID: %s
Command: %s
State: %s
Attempts: %d
Max Retries: %d
Next Run At: %s
Worker: %s
Created At: %s
Updated At: %s
`,
				Id, Command, State, Attempts, Max_retries,
				Next_run_at.String, worker, Created_at, Updated_at)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&checkStateCmd, "state", "s", "", "Filter jobs by state")
}
