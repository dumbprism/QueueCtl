package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type jobSpec struct {
	Id          string `json:"id"`
	Command     string `json:"command"`
	State       string `json:"state"`
	Attempts    int    `json:"attempts"`
	Max_retries int    `json:"max_retries"`
	Created_at  string `json:"created_at"`
	Updated_at  string `json:"updated_at"`
}

var rootCmd = &cobra.Command{
	Use:   "queuectl",
	Short: "A brief description of your application",
	Long:  `long description`,

	Run: func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
