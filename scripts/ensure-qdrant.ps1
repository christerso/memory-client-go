# Memory Client for MCP - PowerShell Script to Ensure Qdrant is Running

Write-Host "Checking if Qdrant is running..." -ForegroundColor Cyan

# Try to connect to Qdrant
try {
    $response = Invoke-WebRequest -Uri "http://localhost:6333/collections" -Method GET -ErrorAction Stop
    Write-Host "Qdrant is already running!" -ForegroundColor Green
    exit 0
}
catch {
    Write-Host "Qdrant is not running. Attempting to start..." -ForegroundColor Yellow
}

# Check if another Qdrant instance might be running on a different port
try {
    $dockerPs = docker ps --format "{{.Names}} {{.Ports}}" | Select-String -Pattern "qdrant"
    if ($dockerPs) {
        Write-Host "Found existing Qdrant container: $dockerPs" -ForegroundColor Yellow
        Write-Host "Using existing Qdrant instance instead of starting a new one." -ForegroundColor Green
        exit 0
    }
}
catch {
    # Continue if docker command fails
}

# Check if Docker is installed
try {
    $dockerVersion = docker --version
    Write-Host "Docker is installed: $dockerVersion" -ForegroundColor Green
}
catch {
    Write-Host "Docker is not installed. Please install Docker to run Qdrant." -ForegroundColor Red
    Write-Host "You can download Docker from: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

# Check if Qdrant container exists
$qdrantContainer = docker ps -a --filter "name=qdrant" --format "{{.Names}}"
if ($qdrantContainer -eq "qdrant") {
    # Container exists, check if it's running
    $qdrantRunning = docker ps --filter "name=qdrant" --format "{{.Names}}"
    if ($qdrantRunning -eq "qdrant") {
        Write-Host "Qdrant container is already running!" -ForegroundColor Green
        exit 0
    }
    else {
        # Container exists but not running, start it
        Write-Host "Starting existing Qdrant container..." -ForegroundColor Yellow
        docker start qdrant
    }
}
else {
    # Container doesn't exist, create and start it
    Write-Host "Creating and starting Qdrant container..." -ForegroundColor Yellow
    docker run -d -p 6333:6333 -p 6334:6334 --name qdrant qdrant/qdrant
}

# Wait for Qdrant to start
Write-Host "Waiting for Qdrant to start..." -ForegroundColor Yellow
$maxRetries = 10
$retryCount = 0
$success = $false

while ($retryCount -lt $maxRetries -and -not $success) {
    Start-Sleep -Seconds 2
    $retryCount++
    
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:6333/collections" -Method GET -ErrorAction Stop
        $success = $true
        Write-Host "Qdrant started successfully!" -ForegroundColor Green
    } 
    catch {
        Write-Host "Waiting for Qdrant to become available (attempt $retryCount of $maxRetries)..." -ForegroundColor Yellow
    }
}

if (-not $success) {
    Write-Host "Failed to start Qdrant after $maxRetries attempts." -ForegroundColor Red
    exit 1
}

Write-Host "`nQdrant is now running and ready to use with the memory client." -ForegroundColor Green
Write-Host "The memory client will work seamlessly in the background."
Write-Host "You don't need to manually start or manage it - Cline/Roo will handle everything."