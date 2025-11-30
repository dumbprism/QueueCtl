package cmd

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage queuectl configuration",
	Long:  "Configure retry behavior and backoff settings for the queue system",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set configuration values for the queue system.

Available keys:
  max-retries   - Maximum number of retry attempts before job moves to DLQ (default: 3)
  backoff-base  - Base multiplier for exponential backoff in seconds (default: 2)
                  Retry delays: backoff-base^attempt (e.g., 2s, 4s, 8s, 16s...)`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		key := args[0]
		value := args[1]

		allowedKeys := map[string]bool{
			"max-retries":  true,
			"backoff-base": true,
		}

		if !allowedKeys[key] {
			fmt.Println("❌ Invalid key. Allowed keys:")
			fmt.Println("  max-retries   - Maximum retry attempts (e.g., 3)")
			fmt.Println("  backoff-base  - Exponential backoff base in seconds (e.g., 2)")
			return
		}

		// Validate numeric values
		numValue, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("❌ Error: value must be a number, got: %s\n", value)
			return
		}

		// Validate ranges
		if key == "max-retries" && numValue < 0 {
			fmt.Println("❌ Error: max-retries must be >= 0")
			return
		}

		if key == "backoff-base" && numValue < 1 {
			fmt.Println("❌ Error: backoff-base must be >= 1")
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

		fmt.Printf("✅ Configuration updated: %s = %s\n", key, value)

		// Show what this means
		if key == "max-retries" {
			fmt.Printf("   Jobs will retry up to %d times before moving to DLQ\n", numValue)
		} else if key == "backoff-base" {
			fmt.Printf("   Retry delays will be: %ds, %ds, %ds, %ds...\n",
				numValue, numValue*2, numValue*4, numValue*8)
		}
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

		rows, err := db.Query(`SELECT Key, Value FROM config ORDER BY Key`)
		if err != nil {
			fmt.Println("Error fetching config:", err)
			return
		}
		defer rows.Close()

		fmt.Println("\n===== CONFIGURATION =====")

		configMap := make(map[string]string)
		for rows.Next() {
			var key, value string
			rows.Scan(&key, &value)
			configMap[key] = value
		}

		if len(configMap) == 0 {
			fmt.Println("No configuration set yet. Using defaults:")
			fmt.Println("  max-retries  = 3")
			fmt.Println("  backoff-base = 2")
		} else {
			// Show max-retries
			if val, ok := configMap["max-retries"]; ok {
				fmt.Printf("max-retries  = %s\n", val)
				fmt.Println("  → Maximum retry attempts before moving to DLQ")
			} else {
				fmt.Println("max-retries  = 3 (default)")
			}

			// Show backoff-base
			if val, ok := configMap["backoff-base"]; ok {
				fmt.Printf("backoff-base = %s\n", val)
				numVal, _ := strconv.Atoi(val)
				fmt.Printf("  → Retry delays: %ds, %ds, %ds, %ds...\n",
					numVal, numVal*2, numVal*4, numVal*8)
			} else {
				fmt.Println("backoff-base = 2 (default)")
				fmt.Println("  → Retry delays: 2s, 4s, 8s, 16s...")
			}
		}

		fmt.Println("=========================\n")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
}
