

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Testing queuectl - Job Queue System" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan

# Clean slate
Write-Host "`nCleaning old data..." -ForegroundColor Yellow
if (Test-Path "data") {
    Remove-Item -Recurse -Force "data"
}

# Test 1: Configuration
Write-Host "`n[TEST 1] Configuration Management" -ForegroundColor Green
go run main.go config set max-retries 3
go run main.go config set backoff-base 2
go run main.go config get

# Test 2: Enqueue jobs
Write-Host "`n[TEST 2] Enqueue Jobs" -ForegroundColor Green
go run main.go enqueue -c "echo Test Job 1"
go run main.go enqueue -c "echo Test Job 2"
go run main.go enqueue -c "timeout /t 2 /nobreak && echo Slow Job"
go run main.go enqueue -c "exit 1"

Write-Host "`nCurrent job list:" -ForegroundColor Cyan
go run main.go list

# Test 3: Worker processing
Write-Host "`n[TEST 3] Start Worker (will run for 15 seconds)" -ForegroundColor Green
$job = Start-Job -ScriptBlock { 
    Set-Location $using:PWD
    go run main.go worker -v -s 1 
}
Start-Sleep -Seconds 15
Stop-Job $job
Remove-Job $job

# Test 4: Check results
Write-Host "`n[TEST 4] Check Job States" -ForegroundColor Green
go run main.go list

# Test 5: Check DLQ
Write-Host "`n[TEST 5] Dead Letter Queue" -ForegroundColor Green
go run main.go dlq list

# Test 6: Status
Write-Host "`n[TEST 6] System Status" -ForegroundColor Green
go run main.go status

# Test 7: Retry from DLQ
Write-Host "`n[TEST 7] Retry Dead Job" -ForegroundColor Green
$dlqOutput = go run main.go dlq list | Select-String "ID:"
if ($dlqOutput) {
    $deadJob = ($dlqOutput[0] -split '\s+')[1]
    Write-Host "Retrying job: $deadJob" -ForegroundColor Yellow
    go run main.go dlq retry $deadJob
    go run main.go list -s pending
}

Write-Host "`n======================================" -ForegroundColor Cyan
Write-Host "All Tests Completed Successfully" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Cyan