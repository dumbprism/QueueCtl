package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var enqueueCmd = &cobra.Command{
	Use:   "enqueue",
	Short: "used for adding jobs inside the queue",
	Long: `The enqueue command helps to add job specification inside 
	which would help the workers to pick jobs based on a static job map
	which consists of essential details related to the job specification`,
	Run: func(cmd *cobra.Command, args []string) {

		// Temporary ID (your code still sets this to 1 every time)
		var jobId = 0
		var jobAttempts = 0
		var job_maxRetries = 0

		// Create the job
		var job jobSpec
		jobId += 1
		job.Id = jobId
		job.Command = "command not found"
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

		// Open file in append mode (this is the FIX)
		file, err := os.OpenFile(
			"data.json",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			fmt.Println("Error opening JSON file:", err)
			return
		}
		defer file.Close()

		// Write the JSON + newline (this is the FIX)
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
}
