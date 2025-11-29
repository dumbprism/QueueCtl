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
var workerCount int

func randomWorkerID() string {
	charset := "abcdefghijklmnopqrstuvwxyz1234567890"
	id := ""
	for i := 0; i < 8; i++ {
		id += string(charset[rand.Intn(len(charset))])
	}
	return id
}

func stopAllWorkers() {
	db, err := sql.Open("sqlite", "data/queue.db")
	if err != nil {
		fmt.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
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

func startWorker(database *sql.DB) {

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	workerId := randomWorkerID()
	now := time.Now().Format("2006-01-02 15:04:05")

	_, err := database.Exec(`
		INSERT INTO workers (WorkerId, Started_at, Last_heartbeat)
		VALUES (?, ?, ?)
	`, workerId, now, now)

	if err != nil {
		fmt.Println("Error registering worker:", err)
		return
	}

	fmt.Println("Worker started:", workerId)

	defer func() {
		database.Exec(`DELETE FROM workers WHERE WorkerId=?`, workerId)
		fmt.Println("Worker stopped:", workerId)
	}()

	// Load configurable values
	var cfgMaxRetries int = 3
	var cfgBackoffBase int = 2

	row := database.QueryRow(`SELECT Value FROM config WHERE Key='max-retries'`)
	row.Scan(&cfgMaxRetries)

	row = database.QueryRow(`SELECT Value FROM config WHERE Key='backoff-base'`)
	row.Scan(&cfgBackoffBase)

	jobCount := 0

	for {

		var stopFlag string
		database.QueryRow(`
			SELECT Value FROM control WHERE Key='stop'
		`).Scan(&stopFlag)

		if stopFlag == "true" {
			fmt.Println("Stop signal received → Worker exiting:", workerId)
			return
		}

		database.Exec(`
			UPDATE workers SET Last_heartbeat=? WHERE WorkerId=?
		`, time.Now().Format("2006-01-02 15:04:05"), workerId)

		if workerLimit > 0 && jobCount >= workerLimit {
			return
		}

		// Fetch a pending job ready for execution
		row := database.QueryRow(`
			SELECT Id, Command, State, Attempts, Max_retries, Created_at, Updated_at
			FROM jobs
			WHERE State='pending'
			AND Next_run_at <= CURRENT_TIMESTAMP
			ORDER BY Created_at
			LIMIT 1
		`)

		var Id, Command, State, CreatedAt, UpdatedAt string
		var Attempts, MaxRetries int

		err := row.Scan(&Id, &Command, &State, &Attempts, &MaxRetries, &CreatedAt, &UpdatedAt)
		if err != nil {
			time.Sleep(time.Duration(workerSleep) * time.Second)
			continue
		}

		// Assign job to worker
		database.Exec(`
			UPDATE jobs 
			SET State='running', WorkerId=?, Updated_at=?
			WHERE Id=?
		`, workerId, time.Now().Format("2006-01-02 15:04:05"), Id)

		var execCmd *exec.Cmd
		if runtime.GOOS == "windows" {
			execCmd = exec.Command("cmd", "/C", Command)
		} else {
			execCmd = exec.Command("bash", "-c", Command)
		}

		err = execCmd.Run()

		if err != nil {

			Attempts++

			// Select config value instead of DB-stored max retries
			MaxRetries = cfgMaxRetries

			if Attempts > MaxRetries {

				database.Exec(`
					UPDATE jobs
					SET State='dead', Attempts=?, WorkerId=NULL, Updated_at=?
					WHERE Id=?
				`, Attempts, time.Now().Format("2006-01-02 15:04:05"), Id)

			} else {

				backoff := time.Duration(cfgBackoffBase<<Attempts) * time.Second
				nextRun := time.Now().Add(backoff).Format("2006-01-02 15:04:05")

				database.Exec(`
					UPDATE jobs
					SET State='pending',
						Attempts=?,
						WorkerId=NULL,
						Next_run_at=?,
						Updated_at=?
					WHERE Id=?
				`, Attempts, nextRun, time.Now().Format("2006-01-02 15:04:05"), Id)
			}

			continue
		}

		// Success
		database.Exec(`
			UPDATE jobs 
			SET State='completed', WorkerId=NULL, Updated_at=?
			WHERE Id=?
		`, time.Now().Format("2006-01-02 15:04:05"), Id)

		jobCount++
	}
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run queue workers",
	Run: func(cmd *cobra.Command, args []string) {

		if workerStop {
			stopAllWorkers()
			return
		}

		os.MkdirAll("data", 0755)

		database, err := sql.Open("sqlite", "data/queue.db")
		if err != nil {
			fmt.Println("DB error:", err)
			return
		}

		database.Exec("PRAGMA journal_mode=WAL;")
		database.Exec("PRAGMA busy_timeout=5000;")

		database.Exec(`
			INSERT INTO control (Key, Value)
			VALUES ('stop', 'false')
			ON CONFLICT(Key) DO UPDATE SET Value='false'
		`)

		for i := 0; i < workerCount; i++ {
			go startWorker(database)
		}

		select {}
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)

	workerCmd.Flags().IntVarP(&workerLimit, "limit", "l", 0, "Maximum number of jobs a worker processes")
	workerCmd.Flags().IntVarP(&workerSleep, "sleep", "s", 3, "Sleep time when idle")
	workerCmd.Flags().BoolVarP(&workerVerbose, "verbose", "v", false, "Verbose logs")
	workerCmd.Flags().BoolVarP(&workerStop, "stop", "", false, "Stop all workers immediately")
	workerCmd.Flags().IntVarP(&workerCount, "count", "c", 1, "Number of workers to spawn")
}
