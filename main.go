/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
"database/sql"
"fmt"
"os"
"queuectl/internal/db"
"queuectl/cmd"
_"modernc.org/sqlite"
)
func main() {
    os.MkdirAll("data", 0755)

    database, err := sql.Open("sqlite", "data/queue.db")
    if err != nil {
        fmt.Println("DB open error:", err)
        return
    }

    // RUN MIGRATIONS
    err = db.Migrate(database)
    if err != nil {
        fmt.Println("Migration error:", err)
        return
    }

    // Pass control to Cobra root command
    cmd.Execute()
}
