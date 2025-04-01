# Script to add memory-client binary to user PATH

Write-Host "Adding memory-client to your PATH..." -ForegroundColor Cyan

# Get the memory client bin directory
$memoryClientBinDir = Join-Path $env:APPDATA "memory-client\bin"

# Check if the directory exists
if (-not (Test-Path $memoryClientBinDir)) {
    Write-Host "Memory client bin directory not found at: $memoryClientBinDir" -ForegroundColor Red
    Write-Host "Please run the installation script first." -ForegroundColor Yellow
    exit 1
}

# Get current user PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")

# Check if the directory is already in PATH
if ($currentPath -split ";" -contains $memoryClientBinDir) {
    Write-Host "Memory client is already in your PATH." -ForegroundColor Green
    exit 0
}

# Add to PATH
$newPath = $currentPath + ";" + $memoryClientBinDir
[Environment]::SetEnvironmentVariable("PATH", $newPath, "User")

# Verify it was added
$updatedPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($updatedPath -split ";" -contains $memoryClientBinDir) {
    Write-Host "Successfully added memory-client to your PATH!" -ForegroundColor Green
    Write-Host "You can now run 'memory-client' from any command prompt." -ForegroundColor Green
    Write-Host "Note: You'll need to open a new command prompt for the changes to take effect." -ForegroundColor Yellow
} else {
    Write-Host "Failed to add memory-client to your PATH." -ForegroundColor Red
}