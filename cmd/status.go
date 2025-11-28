package cmd

import (
	"fmt"
	_ "modernc.org/sqlite"
	"database/sql"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:`,
	Run: func(cmd *cobra.Command, args []string) {
		database,err := sql.Open("sqlite","data/queue.db")

		if err != nil{
			fmt.Println("There is an error in the database : ",err)
		}
		defer database.Close();

		rows,err := database.Query(`SELECT Id,State FROM jobs`)

		if err!=nil{
			fmt.Println("row not recorded : ",err)
		}

		for rows.Next(){
			var Id string
			var State string
			
			err = rows.Scan(&Id,&State)

			if err !=nil{
				fmt.Println("There seems to be an error : ",err)
			}

			fmt.Printf("\nID:%s\nState:%s",Id,State)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
