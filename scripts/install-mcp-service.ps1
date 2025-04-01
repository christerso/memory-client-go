# This script installs the Memory Client MCP service as a Windows service
# Requires Administrator privileges

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "This script requires Administrator privileges. Please run as Administrator." -ForegroundColor Red
    exit 1
}

# Get the script directory and main.go path
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootPath = Split-Path -Parent $scriptPath
$exePath = Join-Path -Path $rootPath -ChildPath "memory-client-go.exe"

# Build the executable if it doesn't exist
if (-not (Test-Path $exePath)) {
    Write-Host "Building memory-client-go.exe..." -ForegroundColor Yellow
    Set-Location $rootPath
    & go build -o memory-client-go.exe
    if (-not $?) {
        Write-Host "Failed to build the executable. Please check your Go installation." -ForegroundColor Red
        exit 1
    }
}

# Create a service using NSSM (Non-Sucking Service Manager)
# You need to install NSSM first: https://nssm.cc/download
$nssmPath = "nssm.exe"
try {
    # Check if NSSM is installed
    $null = Get-Command $nssmPath -ErrorAction Stop
} catch {
    Write-Host "NSSM (Non-Sucking Service Manager) is not installed or not in PATH." -ForegroundColor Red
    Write-Host "Please download and install NSSM from https://nssm.cc/download" -ForegroundColor Red
    exit 1
}

$serviceName = "MemoryClientMCP"

# Check if service already exists
$serviceExists = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
if ($serviceExists) {
    Write-Host "Service '$serviceName' already exists. Removing it first..." -ForegroundColor Yellow
    & $nssmPath remove $serviceName confirm
}

# Install the service
Write-Host "Installing Memory Client MCP as a Windows service..." -ForegroundColor Green
& $nssmPath install $serviceName $exePath "mcp"
& $nssmPath set $serviceName DisplayName "Memory Client MCP Service"
& $nssmPath set $serviceName Description "Memory Client MCP Service for persistent conversation storage"
& $nssmPath set $serviceName Start SERVICE_AUTO_START

# Start the service
Write-Host "Starting the service..." -ForegroundColor Green
Start-Service -Name $serviceName

Write-Host "Memory Client MCP service has been installed and started." -ForegroundColor Green
Write-Host "The service will automatically start when Windows boots." -ForegroundColor Green
Write-Host "Service is running on port 8080." -ForegroundColor Green
