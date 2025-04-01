# Memory Client for MCP - PowerShell Installation Script

Write-Host "Installing Memory Client for MCP..." -ForegroundColor Cyan

# Build the memory client
Write-Host "Building memory client..." -ForegroundColor Yellow
go build -o memory-client.exe
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build memory client" -ForegroundColor Red
    exit 1
}

# Create necessary directories
$mcpDir = Join-Path $env:APPDATA "Roo-Code\MCP"
$memoryClientDir = Join-Path $env:APPDATA "memory-client"

if (-not (Test-Path $mcpDir)) {
    Write-Host "Creating MCP directory: $mcpDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $mcpDir -Force | Out-Null
}

if (-not (Test-Path $memoryClientDir)) {
    Write-Host "Creating memory-client directory: $memoryClientDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $memoryClientDir -Force | Out-Null
}

# Get the current directory
$currentDir = (Get-Location).Path

# Update the path in mcp_settings.json
Write-Host "Updating MCP settings with correct path..." -ForegroundColor Yellow
$mcpSettings = Get-Content -Path "mcp_settings.json" -Raw
$mcpSettings = $mcpSettings -replace 'c:/Users/christer/Desktop/memory-client-go/memory-client.exe', ($currentDir + '/memory-client.exe' -replace '\\', '/')
Set-Content -Path "mcp_settings.json" -Value $mcpSettings

# Copy the memory client executable to a persistent location
$persistentDir = Join-Path $env:APPDATA "memory-client\bin"
if (-not (Test-Path $persistentDir)) {
    Write-Host "Creating persistent directory: $persistentDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $persistentDir -Force | Out-Null
}

Write-Host "Copying memory-client.exe to persistent location..." -ForegroundColor Yellow
Copy-Item -Path "memory-client.exe" -Destination (Join-Path $persistentDir "memory-client.exe") -Force
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to copy memory-client.exe" -ForegroundColor Red
    exit 1
}

# Update MCP settings to use the persistent location
Write-Host "Updating MCP settings with persistent path..." -ForegroundColor Yellow
$persistentPath = Join-Path $persistentDir "memory-client.exe" -replace '\\', '/'
$mcpSettings = Get-Content -Path "mcp_settings.json" -Raw
$mcpSettings = $mcpSettings -replace 'C:/Users/christer/Desktop/memory-client-go/memory-client.exe', $persistentPath
Set-Content -Path "mcp_settings.json" -Value $mcpSettings

# Copy MCP settings
Write-Host "Copying MCP settings to Roo directory..." -ForegroundColor Yellow
Copy-Item -Path "mcp_settings.json" -Destination (Join-Path $mcpDir "mcp_settings.json") -Force
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to copy MCP settings" -ForegroundColor Red
    exit 1
}

# Copy config file
Write-Host "Copying configuration file..." -ForegroundColor Yellow
Copy-Item -Path "config.yaml" -Destination (Join-Path $memoryClientDir "config.yaml") -Force
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to copy configuration file" -ForegroundColor Red
    exit 1
}

# Create a shortcut to run Qdrant at startup
$startupFolder = [System.IO.Path]::Combine($env:APPDATA, "Microsoft\Windows\Start Menu\Programs\Startup")
$shortcutPath = [System.IO.Path]::Combine($startupFolder, "Ensure-Qdrant.lnk")
$ensureQdrantPath = Join-Path $currentDir "ensure-qdrant.ps1"

Write-Host "Creating startup shortcut for Qdrant..." -ForegroundColor Yellow
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut($shortcutPath)
$Shortcut.TargetPath = "powershell.exe"
$Shortcut.Arguments = "-ExecutionPolicy Bypass -WindowStyle Hidden -File `"$ensureQdrantPath`""
$Shortcut.WorkingDirectory = $currentDir
$Shortcut.Description = "Ensure Qdrant is running for Memory Client"
$Shortcut.Save()

Write-Host "Startup shortcut created at: $shortcutPath" -ForegroundColor Green

# Check if Qdrant is running
Write-Host "Checking if Qdrant is running..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:6333/collections" -Method GET -ErrorAction Stop
    Write-Host "Qdrant is running!" -ForegroundColor Green
}
catch {
    Write-Host "Qdrant is not running on default port. Checking for other instances..." -ForegroundColor Yellow
    
    # Check if another Qdrant instance might be running on a different port
    try {
        $dockerPs = docker ps --format "{{.Names}} {{.Ports}}" | Select-String -Pattern "qdrant"
        if ($dockerPs) {
            Write-Host "Found existing Qdrant container: $dockerPs" -ForegroundColor Yellow
            Write-Host "Will use existing Qdrant instance." -ForegroundColor Green
        } else {
            Write-Host "No running Qdrant instances found." -ForegroundColor Yellow
            Write-Host "Qdrant will be started automatically when needed." -ForegroundColor Green
        }
    }
    catch {
        Write-Host "Could not check for Docker containers." -ForegroundColor Yellow
        Write-Host "Qdrant will be started automatically when needed." -ForegroundColor Green
    }
}

Write-Host "`nInstallation complete!" -ForegroundColor Green
Write-Host "`nThe memory client is now set up to run automatically in the background."
Write-Host "You don't need to manually start it - Cline/Roo will launch it when needed."
Write-Host "`nNext steps:"
Write-Host "1. Ensure Qdrant is running (default: http://localhost:6333)"
Write-Host "2. Restart Cline/Roo to load the new MCP server"
Write-Host "`nTo verify it's working:"
Write-Host "1. Have a conversation in Cline/Roo"
Write-Host "2. Close and restart Cline/Roo"
Write-Host "3. Ask about something from your previous conversation"
Write-Host "`nYou can also manually check your conversation history with:"
Write-Host "  .\memory-client.exe history"
Write-Host "`nTo index a project directory:"
Write-Host "  .\memory-client.exe index-project --project C:\path\to\project"
Write-Host "`nTo watch a project directory for changes:"
Write-Host "  .\memory-client.exe watch-project --project C:\path\to\project"

Write-Host "`nPress any key to continue..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")