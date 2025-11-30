#!/bin/bash
set -e

echo "======================================"
echo "Testing queuectl - Job Queue System"
echo "======================================"

# Clean slate
echo "
🧹 Cleaning old data..."
rm -rf data/

# Test 1: Configuration
echo "
✅ Test 1: Configuration Management"
go run main.go config set max-retries 3
go run main.go config set backoff-base 2
go run main.go config get

# Test 2: Enqueue jobs
echo "
✅ Test 2: Enqueue Jobs"
go run main.go enqueue -c "echo 'Test Job 1'"
go run main.go enqueue -c "echo 'Test Job 2'"
go run main.go enqueue -c "sleep 2 && echo 'Slow Job'"
go run main.go enqueue -c "exit 1"  # This will fail

echo "
📋 Current job list:"
go run main.go list

# Test 3: Worker processing
echo "
✅ Test 3: Start Worker (will run for 15 seconds)"
timeout 15 go run main.go worker -v -s 1 || true

# Test 4: Check results
echo "
✅ Test 4: Check Job States"
go run main.go list

# Test 5: Check DLQ
echo "
✅ Test 5: Dead Letter Queue"
go run main.go dlq list

# Test 6: Status
echo "
✅ Test 6: System Status"
go run main.go status

# Test 7: Retry from DLQ
echo "
✅ Test 7: Retry Dead Job"
DEAD_JOB=$(go run main.go dlq list | grep "ID:" | head -1 | awk '{print $2}')
if [ ! -z "$DEAD_JOB" ]; then
    echo "Retrying job: $DEAD_JOB"
    go run main.go dlq retry $DEAD_JOB
    go run main.go list -s pending
fi

echo "
======================================"
echo "✅ All Tests Completed Successfully"
echo "======================================"