/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		os.MkdirAll("data",0755)

		database,err := sql.Open("sqlite","data/queue.db")

		if err !=nil{
			fmt.Println("facing an error :",err)
		}

		defer database.Close()

		row,err := database.Query("SELECT Id,Command,State,Attempts,Max_retries,Created_at,Updated_at FROM jobs")

		if err != nil{
			fmt.Println("error : ",err)
		}
		defer row.Close()

		

		for row.Next(){
			var Id string
			var Command string
			var State string
			var Attempts int
			var Max_retries int
			var Created_at string
			var Updated_at string

			err = row.Scan(&Id,&Command,&State,&Attempts,&Max_retries,&Created_at,&Updated_at)

			if err !=nil{
				fmt.Println("Error reading row : ",err)
				continue
			}

			fmt.Printf("\nID: %s\nCommand: %s\nState: %s\nAttempts: %d\nMax Retries: %d\nCreated At: %s\nUpdated At: %s\n",Id, Command, State, Attempts, Max_retries, Created_at, Updated_at)

		}
		


	},
}

func init() {
	rootCmd.AddCommand(listCmd)

}
