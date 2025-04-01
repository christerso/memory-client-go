@echo off
:: Memory Client MCP Service Uninstallation Script
:: This batch file completely removes the Memory Client MCP service

echo Uninstalling Memory Client MCP Service...

:: Check for administrative privileges
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo This script requires administrator privileges.
    echo Right-click on this batch file and select "Run as administrator".
    pause
    exit /b 1
)

:: Kill any running MCP server processes
echo Checking for running MCP server processes...
tasklist /fi "imagename eq memory-client-go.exe" | find "memory-client-go.exe" >nul
if %errorLevel% equ 0 (
    echo Stopping running MCP server processes...
    taskkill /f /im memory-client-go.exe >nul 2>&1
    timeout /t 2 /nobreak >nul
    echo MCP server processes stopped.
)

:: Check for both possible service names
set "OLD_SERVICE_NAME=MemoryClientMCP"
set "NEW_SERVICE_NAME=MemoryClientMCPService"

:: Stop and remove the old service if it exists
sc query %OLD_SERVICE_NAME% >nul 2>&1
if %errorLevel% equ 0 (
    echo Stopping and removing %OLD_SERVICE_NAME% service...
    net stop %OLD_SERVICE_NAME% >nul 2>&1
    sc delete %OLD_SERVICE_NAME% >nul 2>&1
    echo Service removal initiated. Waiting for completion...
    timeout /t 10 /nobreak >nul
    echo %OLD_SERVICE_NAME% service removed.
)

:: Stop and remove the new service if it exists
sc query %NEW_SERVICE_NAME% >nul 2>&1
if %errorLevel% equ 0 (
    echo Stopping and removing %NEW_SERVICE_NAME% service...
    net stop %NEW_SERVICE_NAME% >nul 2>&1
    sc delete %NEW_SERVICE_NAME% >nul 2>&1
    echo Service removal initiated. Waiting for completion...
    timeout /t 10 /nobreak >nul
    echo %NEW_SERVICE_NAME% service removed.
)

:: Use NSSM to make sure the services are completely removed
where nssm >nul 2>&1
if %errorLevel% equ 0 (
    echo Using NSSM to ensure complete service removal...
    nssm remove %OLD_SERVICE_NAME% confirm >nul 2>&1
    nssm remove %NEW_SERVICE_NAME% confirm >nul 2>&1
    timeout /t 5 /nobreak >nul
)

echo.
echo Memory Client MCP services have been completely removed.
echo.

pause
