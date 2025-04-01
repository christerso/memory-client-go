Write-Host "Starting Memory Client MCP Service..." -ForegroundColor Green
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$mainPath = Join-Path -Path (Split-Path -Parent $scriptPath) -ChildPath "main.go"

# Start the MCP service in the background
Start-Process -FilePath "go" -ArgumentList "run", $mainPath, "mcp" -WindowStyle Hidden

Write-Host "MCP Service started on port 8080." -ForegroundColor Green
Write-Host "You can now start your editor." -ForegroundColor Green
