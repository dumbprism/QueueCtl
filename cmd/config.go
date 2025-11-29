package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage queuectl configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		key := args[0]
		value := args[1]

		allowedKeys := map[string]bool{
			"max-retries":  true,
			"backoff-base": true,
		}

		if !allowedKeys[key] {
			fmt.Println("Invalid key. Allowed keys:")
			fmt.Println("  max-retries")
			fmt.Println("  backoff-base")
			return
		}

		db, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("DB error:", err)
			return
		}
		defer db.Close()

		_, err = db.Exec(`
            INSERT INTO config (Key, Value)
            VALUES (?, ?)
            ON CONFLICT(Key) DO UPDATE SET Value=excluded.Value
        `, key, value)

		if err != nil {
			fmt.Println("Error saving configuration:", err)
			return
		}

		fmt.Printf("Configuration updated: %s = %s\n", key, value)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show all configuration values",
	Run: func(cmd *cobra.Command, args []string) {

		db, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("DB error:", err)
			return
		}
		defer db.Close()

		rows, err := db.Query(`SELECT Key, Value FROM config`)
		if err != nil {
			fmt.Println("Error fetching config:", err)
			return
		}
		defer rows.Close()

		fmt.Println("\nCurrent Configuration:")
		fmt.Println("----------------------")

		empty := true
		for rows.Next() {
			empty = false
			var key, value string
			rows.Scan(&key, &value)
			fmt.Printf("%s = %s\n", key, value)
		}

		if empty {
			fmt.Println("No configuration set yet.")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
}
