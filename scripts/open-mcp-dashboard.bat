@echo off
:: Script to open the MCP dashboards
:: This script opens both the MCP service status page and optionally the full dashboard

echo Opening MCP dashboards...

:: Get the directory where the batch file is located
set "SCRIPT_DIR=%~dp0"
set "ROOT_DIR=%SCRIPT_DIR%.."
cd /d "%ROOT_DIR%"

:: First check if the MCP service is running
curl -s http://localhost:9580 >nul 2>&1
if %errorLevel% neq 0 (
    echo MCP service does not appear to be running.
    echo Please start the MCP service first.
    pause
    exit /b 1
)

:: Open the MCP service status page
echo Opening MCP service status page...
start http://localhost:9580

:: Ask if user wants to start the full dashboard
echo.
echo The MCP service status page has been opened in your browser.
echo.
set /p start_full="Do you want to start the full dashboard with dark/light mode? (y/N): "

if /i "%start_full%"=="y" (
    echo.
    echo Starting full dashboard...
    start cmd /k "cd /d "%ROOT_DIR%" && memory-client dashboard"
    
    :: Wait a moment for the dashboard to start
    timeout /t 3 /nobreak >nul
    
    :: Open the full dashboard in the browser
    start http://localhost:9581
    echo Full dashboard started and opened in your browser.
) else (
    echo Full dashboard not started.
)

echo.
echo Done!
