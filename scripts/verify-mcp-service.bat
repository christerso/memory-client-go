@echo off
:: MCP Service Verification Script
:: This script verifies that the MCP service is running and accessible

echo Verifying Memory Client MCP Service...

:: Check if the service is running
echo Checking service status...
sc query MemoryClientMCP >nul 2>&1
if %errorLevel% equ 0 (
    echo Service MemoryClientMCP exists.
    sc query MemoryClientMCP | findstr "RUNNING"
    if %errorLevel% equ 0 (
        echo Service is RUNNING.
    ) else (
        echo Service exists but is NOT RUNNING.
        echo Attempting to start the service...
        net start MemoryClientMCP
        if %errorLevel% neq 0 (
            echo Failed to start service. Please check the service configuration.
        ) else (
            echo Service started successfully.
        )
    )
) else (
    echo Service MemoryClientMCP does not exist. Checking alternative service name...
    sc query MemoryClientMCPService >nul 2>&1
    if %errorLevel% equ 0 (
        echo Service MemoryClientMCPService exists.
        sc query MemoryClientMCPService | findstr "RUNNING"
        if %errorLevel% equ 0 (
            echo Service is RUNNING.
        ) else (
            echo Service exists but is NOT RUNNING.
            echo Attempting to start the service...
            net start MemoryClientMCPService
            if %errorLevel% neq 0 (
                echo Failed to start service. Please check the service configuration.
            ) else (
                echo Service started successfully.
            )
        )
    ) else (
        echo No MCP service found. Please install the service first.
        exit /b 1
    )
)

:: Check if the HTTP server is accessible
echo.
echo Checking HTTP server accessibility...
curl -s -o nul -w "%%{http_code}" http://localhost:9580/status >temp.txt
set /p STATUS_CODE=<temp.txt
del temp.txt

if "%STATUS_CODE%" == "200" (
    echo MCP HTTP server is accessible at http://localhost:9580/status
) else (
    echo MCP HTTP server is NOT accessible at http://localhost:9580/status
    echo Status code: %STATUS_CODE%
    echo Please check if the server is running and the port is correct.
)

:: Display MCP configuration
echo.
echo Current MCP configuration:
type "%APPDATA%\Roo-Code\MCP\mcp_settings.json"

:: Display Windsurf configuration
echo.
echo Windsurf MCP configuration:
type "%~dp0windsurf-mcp-config.json"
echo.
echo To configure Windsurf to use the MCP service:
echo 1. Open Windsurf
echo 2. Go to Settings
echo 3. Navigate to the "Memory" or "MCP" section
echo 4. Set the MCP server URL to: http://localhost:9580
echo 5. Save the settings

echo.
echo Verification complete.
pause
