package cmd

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var workerLimit int
var workerSleep int
var workerVerbose bool
var workerStop bool

func randomWorkerID() string {
	charset := "abcdefghijklmnopqrstuvwxyz1234567890"
	id := ""
	for i := 0; i < 8; i++ {
		id += string(charset[rand.Intn(len(charset))])
	}
	return id
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run a worker to process pending jobs",
	Long:  "This worker picks pending jobs, executes their commands, updates job state, and can be gracefully stopped.",
	Run: func(cmd *cobra.Command, args []string) {

		// If user only wants to stop all workers:
		if workerStop {
			stopAllWorkers()
			return
		}

		os.MkdirAll("data", 0755)

		database, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("DB connection error:", err)
			return
		}
		defer database.Close()

		// ----------------------------------------------------
		// 1. Register worker
		// ----------------------------------------------------
		workerId := randomWorkerID()
		now := time.Now().Format("2006-01-02 15:04:05")

		_, err = database.Exec(`
			INSERT INTO workers (WorkerId, Started_at, Last_heartbeat)
			VALUES (?, ?, ?)
		`, workerId, now, now)

		if err != nil {
			fmt.Println("Error registering worker:", err)
			return
		}

		fmt.Println("Worker started with ID:", workerId)

		defer func() {
			database.Exec(`DELETE FROM workers WHERE WorkerId=?`, workerId)
			fmt.Println("Worker stopped:", workerId)
		}()

		jobCount := 0

		for {
			// ----------------------------------------------------
			// 2. Check STOP FLAG
			// ----------------------------------------------------
			var stopValue string
			database.QueryRow(`
				SELECT Value FROM control WHERE Key='stop'
			`).Scan(&stopValue)

			if stopValue == "true" {
				fmt.Println("Stop signal received → Worker exiting:", workerId)
				return
			}

			// Update heartbeat
			database.Exec(`
				UPDATE workers SET Last_heartbeat=? WHERE WorkerId=?
			`, time.Now().Format("2006-01-02 15:04:05"), workerId)

			// Limit reached
			if workerLimit > 0 && jobCount >= workerLimit {
				fmt.Println("Worker finished processing assigned jobs")
				return
			}

			// ----------------------------------------------------
			// 3. Fetch pending job
			// ----------------------------------------------------
			row := database.QueryRow(`
				SELECT Id, Command, State, Attempts, Max_retries, Created_at, Updated_at
				FROM jobs WHERE State='pending'
				ORDER BY Created_at LIMIT 1
			`)

			var Id, Command, State, Created, Updated string
			var Attempts, MaxRetries int

			err := row.Scan(&Id, &Command, &State, &Attempts, &MaxRetries, &Created, &Updated)

			// No jobs
			if err != nil {
				if workerVerbose {
					fmt.Println("No pending jobs. Sleeping...")
				}
				time.Sleep(time.Duration(workerSleep) * time.Second)
				continue
			}

			if workerVerbose {
				fmt.Println("Worker picked job:", Id)
			}

			database.Exec(`
				UPDATE jobs SET State='running', Updated_at=? WHERE Id=?
			`, time.Now().Format("2006-01-02 15:04:05"), Id)

			// ----------------------------------------------------
			// 4. Execute command
			// ----------------------------------------------------
			var execCmd *exec.Cmd

			if runtime.GOOS == "windows" {
				execCmd = exec.Command("cmd", "/C", Command)
			} else {
				execCmd = exec.Command("bash", "-c", Command)
			}

			err = execCmd.Run()

			// ----------------------------------------------------
			// 5. Handle failure
			// ----------------------------------------------------
			if err != nil {
				Attempts++

				if Attempts >= MaxRetries {
					database.Exec(`
						UPDATE jobs SET State='failed', Attempts=?, Updated_at=? WHERE Id=?
					`, Attempts, time.Now().Format("2006-01-02 15:04:05"), Id)
				} else {
					database.Exec(`
						UPDATE jobs SET State='pending', Attempts=?, Updated_at=? WHERE Id=?
					`, Attempts, time.Now().Format("2006-01-02 15:04:05"), Id)
				}

				continue
			}

			// ----------------------------------------------------
			// 6. Success
			// ----------------------------------------------------
			database.Exec(`
				UPDATE jobs SET State='completed', Updated_at=? WHERE Id=?
			`, time.Now().Format("2006-01-02 15:04:05"), Id)

			if workerVerbose {
				fmt.Println("Completed job:", Id)
			}

			jobCount++
		}
	},
}

// ----------------------------------------------------
// STOP logic function
// ----------------------------------------------------
func stopAllWorkers() {
	database, err := sql.Open("sqlite", "data/queue.db")
	if err != nil {
		fmt.Println("DB error:", err)
		return
	}
	defer database.Close()

	_, err = database.Exec(`
		INSERT INTO control (Key, Value)
		VALUES ('stop', 'true')
		ON CONFLICT(Key) DO UPDATE SET Value='true'
	`)
	if err != nil {
		fmt.Println("Error sending stop signal:", err)
		return
	}

	fmt.Println("Stop signal sent to all workers.")
}

func init() {
	rootCmd.AddCommand(workerCmd)

	workerCmd.Flags().IntVarP(&workerLimit, "limit", "l", 0, "Max number of jobs to process")
	workerCmd.Flags().IntVarP(&workerSleep, "sleep", "s", 3, "Seconds to sleep when idle")
	workerCmd.Flags().BoolVarP(&workerVerbose, "verbose", "v", false, "Enable verbose logging")

	// The STOP FLAG
	workerCmd.Flags().BoolVarP(&workerStop, "stop", "", false, "Stop all running workers")
}
