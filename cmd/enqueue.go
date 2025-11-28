package cmd

import (
	"fmt"
	"os"
	"time"
	"encoding/json"
	"github.com/spf13/cobra"
)

// enqueueCmd represents the enqueue command
var enqueueCmd = &cobra.Command{
	Use:   "enqueue",
	Short: "used for adding jobs inside the queue",
	Long: `The enqueue command helps to add job specification inside 
	which would help the workers to pick jobs based on a static job map
	which consists of essential details related to the job specification`,
	Run: func(cmd *cobra.Command, args []string) {

		var jobId = 0
		var jobAttempts = 0
		var job_maxRetries = 0
		var job jobSpec;
		jobId += 1
		job.Id = jobId
		job.Command = "command not found"
		job.State = "pending" 
		job.Attempts = jobAttempts
		job.Max_retries = job_maxRetries
		job.Created_at = time.Now().String()[0:19]
		job.Updated_at = time.Now().String()[0:19]


		jsonData,err := json.Marshal(job)

		if err != nil{
			fmt.Println("Error encoding json file")
		}

		err = os.WriteFile("data.json",jsonData,0644)

		if err != nil{
			fmt.Println("error creating json file")
			return 
		}

		fmt.Println(jobId)
	},
}

func init() {
	rootCmd.AddCommand(enqueueCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// enqueueCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// enqueueCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
