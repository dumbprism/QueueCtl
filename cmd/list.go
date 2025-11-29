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

		database, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("DB error:", err)
			return
		}
		defer database.Close()

		var query string

		if checkStateCmd == "" {
			query = "SELECT * FROM jobs"
		} else {
			query = "SELECT * FROM jobs WHERE State = ?"
		}

		var rows *sql.Rows
		if checkStateCmd == "" {
			rows, err = database.Query(query)
		} else {
			rows, err = database.Query(query, checkStateCmd)
		}

		if err != nil {
			fmt.Println("Query error:", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var Id, Command, State, Created_at, Updated_at string
			var Attempts, Max_retries int

			err = rows.Scan(&Id, &Command, &State, &Attempts, &Max_retries, &Created_at, &Updated_at)
			if err != nil {
				fmt.Println("Row error:", err)
				continue
			}

			fmt.Printf("\nID: %s\nCommand: %s\nState: %s\nAttempts: %d\nMax Retries: %d\nCreated At: %s\nUpdated At: %s\n",
				Id, Command, State, Attempts, Max_retries, Created_at, Updated_at)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&checkStateCmd, "state", "s", "", "Filter jobs by state")
}
