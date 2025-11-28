package cmd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
	"github.com/spf13/cobra"
)

var userCommand string

var enqueueCmd = &cobra.Command{
	Use:   "enqueue",
	Short: "used for adding jobs inside the queue",
	Long: `The enqueue command helps to add job specification inside 
	which would help the workers to pick jobs based on a static job map
	which consists of essential details related to the job specification`,
	Run: func(cmd *cobra.Command, args []string) {

		// Create random ID
		var charset = "abcdefghijklmnopqrstuvwqyz1234567890"
		var jobId = ""

		for i := 0; i < 8; i++ {
			jobId += string(charset[rand.Intn(len(charset))])
		}

		// Other job fields
		var jobAttempts = 0
		var job_maxRetries = 0

		// Create the job
		var job jobSpec
		job.Id = jobId

		// Use user-provided command OR fallback
		if userCommand == "" {
			job.Command = "command not found"
		} else {
			job.Command = userCommand
		}

		job.State = "pending"
		job.Attempts = jobAttempts
		job.Max_retries = job_maxRetries
		job.Created_at = time.Now().String()[0:19]
		job.Updated_at = time.Now().String()[0:19]

		// Convert to JSON
		jsonData, err := json.Marshal(job)
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
			return
		}

		// Open file in append mode
		file, err := os.OpenFile(
			"data.jsonl",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		
		if err != nil {
			fmt.Println("Error opening JSON file:", err)
			return
		}
		defer file.Close()

		// Write JSON line
		_, err = file.Write(append(jsonData, '\n'))
		if err != nil {
			fmt.Println("Error writing to JSON file:", err)
			return
		}

		fmt.Println("Job added successfully!")
	},
}

func init() {
	rootCmd.AddCommand(enqueueCmd)

	// Add the --command flag (simple and clean)
	enqueueCmd.Flags().StringVarP(&userCommand, "command", "c", "no command found", "Command for the job")
}
