@echo off
:: This batch file launches the PowerShell installation script with administrator privileges

echo Installing Memory Client MCP Service...

:: Check for administrative privileges
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo This script requires administrator privileges.
    echo Right-click on this batch file and select "Run as administrator".
    pause
    exit /b 1
)

:: Get the directory where the batch file is located
set "SCRIPT_DIR=%~dp0"
set "PS_SCRIPT=%SCRIPT_DIR%install-mcp-service.ps1"

:: Run the PowerShell script with execution policy bypass
echo Running PowerShell installation script...
powershell.exe -ExecutionPolicy Bypass -File "%PS_SCRIPT%"

if %errorLevel% neq 0 (
    echo Installation failed. Please check the error messages above.
    pause
    exit /b 1
)

echo.
echo Installation completed successfully!
echo To verify the service is working, visit: http://localhost:9580/status
echo.

pause
