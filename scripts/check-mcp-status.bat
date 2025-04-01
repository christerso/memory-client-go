@echo off
:: Simple MCP Service Status Check Script
:: This script checks if the MCP service is running and responding

echo Checking Memory Client MCP Service status...

:: Check if the service exists and is running
sc query MemoryClientMCP | findstr "STATE" | findstr "RUNNING" >nul
if %errorLevel% equ 0 (
    echo Service status: RUNNING
) else (
    sc query MemoryClientMCP | findstr "STATE" | findstr "PAUSED" >nul
    if %errorLevel% equ 0 (
        echo Service status: PAUSED
    ) else (
        sc query MemoryClientMCP >nul 2>&1
        if %errorLevel% equ 0 (
            echo Service status: NOT RUNNING
        ) else (
            echo Service status: NOT INSTALLED
        )
    )
)

:: Check if the service is responding to HTTP requests
echo.
echo Checking HTTP response...
curl -s -o nul -w "HTTP Status: %%{http_code}" http://localhost:9580/status
echo.

:: Check service logs if they exist
set "LOG_DIR=%~dp0..\logs"
if exist "%LOG_DIR%\mcp_stderr.log" (
    echo.
    echo Recent errors from log:
    type "%LOG_DIR%\mcp_stderr.log" | findstr /v "^$" | tail -10
)

echo.
echo To restart the service, run restart-mcp-service.bat as administrator
echo.

pause
