# queuectl - Job Queue Management System

A job queue system built with Go and SQLite. It helps you manage background tasks with automatic retries, failure handling, and multiple workers.


## What is This?

This is a command-line tool that helps you run background jobs (commands) with automatic retry logic. If a job fails, it will retry a few times before giving up. You can also view failed jobs and retry them manually.

Think of it like a to-do list for your computer:
- Add tasks (jobs) to the queue
- Workers pick up tasks and execute them
- Failed tasks are retried automatically
- Permanently failed tasks go to a "dead letter queue" for review

---

## Features

- **Saves jobs to database**: All jobs are stored in SQLite, so they survive even if you restart the program
- **Automatic retries**: Failed jobs are retried automatically with increasing wait times
- **Multiple workers**: Run multiple workers at the same time to process jobs faster
- **Dead letter queue**: View jobs that failed too many times and retry them manually
- **Easy configuration**: Set how many times to retry and how long to wait between retries
- **Real-time monitoring**: Check which jobs are running and which workers are active

---

## Setup Instructions

### What You Need

- Go version 1.24 or higher
- That's it! SQLite is included automatically

### How to Run

1. **Download the code**:
```bash
git clone <repository-url>
cd queuectl
```

2. **Install dependencies**:
```bash
go mod download
```

3. **Build the program**:
```bash
go build -o queuectl
```

4. **Or run directly without building**:
```bash
go run main.go <command>
```

### First Time Setup

The program creates a `data/` folder and database automatically on first run. You just need to set some basic configuration:

```bash
# Set how many times to retry failed jobs (default is 3)
./queuectl config set max-retries 3

# Set the backoff time in seconds (default is 2)
./queuectl config set backoff-base 2
```

---

## Usage Examples

### 1. Setting Up Configuration

**Set how many times to retry**:
```bash
$ queuectl config set max-retries 5
Configuration updated: max-retries = 5
Jobs will retry up to 5 times before moving to DLQ

$ queuectl config set backoff-base 2
Configuration updated: backoff-base = 2
Retry delays will be: 2s, 4s, 8s, 16s...
```

**View your settings**:
```bash
$ queuectl config get

===== CONFIGURATION =====
max-retries  = 5
  Maximum retry attempts before moving to DLQ
backoff-base = 2
  Retry delays: 2s, 4s, 8s, 16s...
=========================
```

### 2. Adding Jobs

**Add a simple job**:
```bash
$ queuectl enqueue -c "echo Hello World"
Job added successfully with ID: kj5p8v9y
```

**Add a job that calls an API**:
```bash
$ queuectl enqueue -c "curl https://api.example.com/webhook"
Job added successfully with ID: abc123xy
```

### 3. Viewing Jobs

**See all jobs**:
```bash
$ queuectl list

ID: kj5p8v9y
Command: echo Hello World
State: completed
Attempts: 0
Max Retries: 3
Created At: 2025-11-30 15:30:22
Updated At: 2025-11-30 15:30:25
```

**Filter by state**:
```bash
$ queuectl list -s pending     # Show waiting jobs
$ queuectl list -s processing  # Show running jobs
$ queuectl list -s failed      # Show failed jobs
$ queuectl list -s dead        # Show permanently failed jobs
```

### 4. Running Workers

**Start one worker**:
```bash
$ queuectl worker
Worker started: wmg14l71
```

**Start multiple workers with detailed logs**:
```bash
$ queuectl worker -v -c 3 -s 1

Worker started: abc123xy
Worker started: def456uv
Worker started: ghi789st

[abc123xy] Picked up NEW job: job001 (Command: echo test)
[abc123xy] Job job001 state: processing
[abc123xy] Job job001 state: completed

[def456uv] Picked up NEW job: job002 (Command: exit 1)
[def456uv] Job job002 FAILED (Attempt 1/3): exit status 1
[def456uv] Job job002 state: failed - Will retry in 2s
[def456uv] Job job002 state: pending (scheduled for retry)
```

**Worker options**:
- `-v`: Show detailed logs
- `-c N`: Number of workers (default: 1)
- `-s N`: How often to check for jobs in seconds (default: 3)
- `-l N`: Maximum jobs per worker
- `--stop`: Stop all running workers

**Stop all workers**:
```bash
$ queuectl worker --stop
Stop signal sent to all workers.
```

### 5. Checking System Status

**See what's happening**:
```bash
$ queuectl status

===== JOB STATES =====
ID: job001   State: completed
ID: job002   State: failed
ID: job003   State: processing

===== ACTIVE WORKERS =====
Worker: abc123xy   Heartbeat: 2025-11-30 15:35:22
Worker: def456uv   Heartbeat: 2025-11-30 15:35:23
===========================
```

### 6. Dead Letter Queue (Failed Jobs)

**See permanently failed jobs**:
```bash
$ queuectl dlq list

===== DEAD LETTER QUEUE =====

ID: xuya6a8w
Command: curl https://failing-api.com
Attempts: 3
Max_retries: 3
Updated: 2025-11-30 15:40:12

==============================
```

**Retry a failed job**:
```bash
$ queuectl dlq retry xuya6a8w
Job xuya6a8w has been requeued with max_retries = 3
```

---

## How It Works

### Job States

A job goes through different states during its lifetime:

```
pending → processing → completed (success!)
                   → failed → pending (retry after waiting)
                           → dead (too many failures)
```

**What each state means**:

- **pending**: Job is waiting for a worker to pick it up
- **processing**: A worker is currently running the job
- **completed**: Job finished successfully
- **failed**: Job failed but will be retried
- **dead**: Job failed too many times, moved to dead letter queue

### How Jobs Are Stored

All job information is saved in a SQLite database file at `data/queue.db`. This means:
- Jobs are saved even if you close the program
- You can restart the system and jobs will still be there
- No external database needed

The database has four tables:
- **jobs**: Stores all job information (command, state, attempts, etc.)
- **workers**: Tracks active workers
- **config**: Stores your settings (max retries, backoff time)
- **control**: Used to signal workers to stop

### How Workers Process Jobs

Workers follow this process:

1. Register themselves in the database
2. Look for pending jobs every few seconds
3. Pick up a job and change its state to "processing"
4. Run the command
5. If successful, mark as "completed"
6. If failed, increase attempt count and retry later
7. If too many failures, mark as "dead"

### Retry Logic

When a job fails, it waits before retrying. The wait time increases each time:

```
First retry:  2 seconds
Second retry: 4 seconds
Third retry:  8 seconds
Fourth retry: 16 seconds
```

This is called "exponential backoff". It gives failing services time to recover.

**Example**: If `max-retries = 3` and `backoff-base = 2`:
- Job fails → wait 2s → retry
- Job fails again → wait 4s → retry
- Job fails again → wait 8s → retry
- Job fails again → moved to dead letter queue

### Multiple Workers

You can run multiple workers at the same time. They will:
- Share the job queue
- Not pick up the same job (SQLite handles this)
- Process jobs in parallel
- Each worker has a unique ID

---

## Design Decisions

### Why SQLite?

I chose SQLite for the database because:
- **No setup needed**: It's just a file, no server to install
- **Simple to use**: Works well for single-machine use
- **Good enough**: Handles moderate workloads fine
- **ACID transactions**: Keeps data consistent

**Limitations**:
- Not suitable if you need multiple machines
- Write speed is limited compared to bigger databases
- Best for <1000 jobs per second

### Why Polling Instead of Real-time?

Workers check for new jobs every few seconds instead of being notified immediately.

**Why I chose this**:
- Much simpler to implement
- No extra infrastructure needed
- Workers can start/stop independently

**Downside**:
- Small delay (up to 3 seconds by default)
- Database gets checked even when idle

### Why Exponential Backoff?

When a job fails, the wait time doubles each retry.

**Why I chose this**:
- Prevents hammering a failing service
- Gives external systems time to recover
- Standard practice in industry

**Example**: If an API is down, we don't want to call it 100 times per second. We try, wait 2s, try again, wait 4s, etc.

### What Could Be Better?

Some things I didn't implement to keep it simple:

1. **Job Timeouts**: Jobs run forever until they finish. A hanging job will block a worker.

2. **Job Priority**: All jobs are equal. You can't mark some jobs as more important.

3. **Per-Job Configuration**: All jobs use the same retry settings. You can't set different retries for different jobs.

4. **Distributed System**: Only works on one machine. Can't spread workers across multiple computers.

5. **Job Dependencies**: Can't say "run job B only after job A finishes".

---

## Testing

### Quick Test

Here's a simple way to test if everything works:

```bash
# Step 1: Clean start
rm -rf data/

# Step 2: Configure
go run main.go config set max-retries 3
go run main.go config set backoff-base 2

# Step 3: Add jobs
go run main.go enqueue -c "echo Success"
go run main.go enqueue -c "exit 1"

# Step 4: Run worker for 15 seconds
timeout 15 go run main.go worker -v -s 1

# Step 5: Check results
go run main.go list
go run main.go dlq list
```

### Automated Test Script

For Windows PowerShell, save this as `test.ps1`:

```powershell
Write-Host "Testing queuectl..." -ForegroundColor Cyan

# Clean and setup
Remove-Item -Recurse -Force data -ErrorAction SilentlyContinue
go run main.go config set max-retries 3
go run main.go config set backoff-base 2

# Add jobs
go run main.go enqueue -c "echo Test Job"
go run main.go enqueue -c "exit 1"

# Run worker
$job = Start-Job -ScriptBlock { 
    Set-Location $using:PWD
    go run main.go worker -v -s 1 
}
Start-Sleep -Seconds 15
Stop-Job $job
Remove-Job $job

# Check results
go run main.go status
go run main.go dlq list

Write-Host "Tests completed!" -ForegroundColor Green
```

Run it:
```powershell
.\test.ps1
```

### What to Check

After running tests, verify:

- Configuration is saved correctly
- Jobs are added with unique IDs
- Workers pick up and run jobs
- Successful jobs show "completed" state
- Failed jobs retry automatically
- Jobs move to "dead" state after max retries
- Dead jobs appear in DLQ
- You can retry dead jobs manually

### Testing Retries

To see retry behavior clearly:

```bash
# Set up
go run main.go config set max-retries 3
go run main.go config set backoff-base 2

# Add failing job
go run main.go enqueue -c "exit 1"

# Watch it retry (with logs)
go run main.go worker -v -s 1
```

You should see:
```
Job xxx FAILED (Attempt 1/3)
Will retry in 2s
... 2 seconds later ...
RETRYING job xxx (Attempt 2)
Job xxx FAILED (Attempt 2/3)
Will retry in 4s
... 4 seconds later ...
RETRYING job xxx (Attempt 3)
Job xxx FAILED (Attempt 3/3)
Job xxx state: DEAD
```

### Testing Multiple Workers

```bash
# Terminal 1: Start workers
go run main.go worker -v -c 3 -s 1

# Terminal 2: Add many jobs
for i in {1..10}; do
    go run main.go enqueue -c "echo Job $i && sleep 1"
done

# Terminal 3: Watch status
go run main.go status
```

You should see multiple workers processing jobs in parallel.

---

## Project Structure

```
queuectl/
├── cmd/                    # All commands
│   ├── config.go          # Config management
│   ├── dlq.go             # Dead letter queue
│   ├── enqueue.go         # Add jobs
│   ├── list.go            # View jobs
│   ├── root.go            # Main command
│   ├── status.go          # System status
│   └── worker.go          # Worker logic
├── internal/db/           # Database code
│   ├── connect.go         # Database connection
│   └── migrate.go         # Create tables
├── data/                  # Created automatically
│   └── queue.db           # SQLite database
├── main.go                # Program entry point
└── README.md              # This file
```

---
