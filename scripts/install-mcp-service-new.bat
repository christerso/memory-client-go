@echo off
:: Memory Client MCP Service Direct Installation Script
:: This batch file directly installs the Memory Client MCP service without requiring PowerShell

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
set "ROOT_DIR=%SCRIPT_DIR%.."
cd /d "%ROOT_DIR%"

:: Kill any running MCP server processes
echo Checking for running MCP server processes...
tasklist /fi "imagename eq memory-client-go.exe" | find "memory-client-go.exe" >nul
if %errorLevel% equ 0 (
    echo Stopping running MCP server processes...
    taskkill /f /im memory-client-go.exe >nul 2>&1
    timeout /t 2 /nobreak >nul
    echo MCP server processes stopped.
)

:: Use a different service name to avoid conflicts with the old one
set "SERVICE_NAME=MemoryClientMCPService"

:: Check if our new service already exists
sc query %SERVICE_NAME% >nul 2>&1
if %errorLevel% equ 0 (
    echo Stopping and removing existing %SERVICE_NAME% service...
    net stop %SERVICE_NAME% >nul 2>&1
    sc delete %SERVICE_NAME% >nul 2>&1
    
    :: Wait for service to be fully removed
    echo Waiting for service to be fully removed...
    timeout /t 10 /nobreak >nul
    echo Service removed.
)

:: Always build the executable to ensure it's up to date
echo Building memory-client-go.exe...
go build -o memory-client-go.exe
if %errorLevel% neq 0 (
    echo Failed to build the executable. Please check your Go installation.
    pause
    exit /b 1
)
echo Executable built successfully.

:: Create necessary directories
set "MEMORY_CLIENT_DIR=%APPDATA%\memory-client"
set "MCP_CONFIG_DIR=%APPDATA%\Roo-Code\MCP"

if not exist "%MEMORY_CLIENT_DIR%" (
    echo Creating memory-client directory: %MEMORY_CLIENT_DIR%
    mkdir "%MEMORY_CLIENT_DIR%"
)

if not exist "%MCP_CONFIG_DIR%" (
    echo Creating MCP configuration directory: %MCP_CONFIG_DIR%
    mkdir "%MCP_CONFIG_DIR%"
)

:: Create config.yaml if it doesn't exist
set "CONFIG_PATH=%MEMORY_CLIENT_DIR%\config.yaml"
if not exist "%CONFIG_PATH%" (
    echo Creating memory-client configuration file...
    (
        echo # Qdrant server URL
        echo QDRANT_URL: "http://localhost:6333"
        echo.
        echo # Collection name for storing conversation memory
        echo COLLECTION_NAME: "conversation_memory"
        echo.
        echo # Size of embedding vectors
        echo EMBEDDING_SIZE: 384
    ) > "%CONFIG_PATH%"
    echo Configuration file created at: %CONFIG_PATH%
)

:: Create or update MCP settings
echo Setting up MCP configuration...
set "MCP_SETTINGS_PATH=%MCP_CONFIG_DIR%\mcp_settings.json"
set "EXE_PATH=%ROOT_DIR%\memory-client-go.exe"
set "EXE_PATH_FORMATTED=%EXE_PATH:\=/%"
set "ROOT_DIR_FORMATTED=%ROOT_DIR:\=/%"

(
    echo {
    echo     "executable": "%EXE_PATH_FORMATTED%",
    echo     "arguments": ["mcp"],
    echo     "workingDir": "%ROOT_DIR_FORMATTED%"
    echo }
) > "%MCP_SETTINGS_PATH%"
echo MCP configuration has been set up at: %MCP_SETTINGS_PATH%

:: Ensure Qdrant is running
echo Checking if Qdrant is running...
call "%SCRIPT_DIR%ensure-qdrant.bat"

:: Check if NSSM is installed
where nssm >nul 2>&1
if %errorLevel% neq 0 (
    echo NSSM (Non-Sucking Service Manager) is not installed or not in PATH.
    echo Please download and install NSSM from https://nssm.cc/download
    pause
    exit /b 1
)

:: Try to forcefully remove the old service if it exists
echo Checking for old MemoryClientMCP service...
sc query MemoryClientMCP >nul 2>&1
if %errorLevel% equ 0 (
    echo Attempting to forcefully remove old MemoryClientMCP service...
    sc stop MemoryClientMCP >nul 2>&1
    sc delete MemoryClientMCP >nul 2>&1
    timeout /t 5 /nobreak >nul
)

:: Install the service with the new name
echo Installing Memory Client MCP as a Windows service...
nssm install %SERVICE_NAME% "%EXE_PATH%" mcp
if %errorLevel% neq 0 (
    echo Failed to install service. Error code: %errorLevel%
    pause
    exit /b 1
)

echo Configuring service properties...
nssm set %SERVICE_NAME% DisplayName "Memory Client MCP Service"
nssm set %SERVICE_NAME% Description "Memory Client MCP Service for persistent conversation storage"
nssm set %SERVICE_NAME% Start SERVICE_AUTO_START
nssm set %SERVICE_NAME% AppDirectory "%ROOT_DIR%"

:: Start the service
echo Starting the service...
net start %SERVICE_NAME%
if %errorLevel% neq 0 (
    echo Failed to start service. Error code: %errorLevel%
    pause
    exit /b 1
)

echo.
echo Memory Client MCP service has been installed and started.
echo The service will automatically start when Windows boots.
echo Service is running on port 9580.
echo To verify the service is working, visit: http://localhost:9580/status
echo.

pause
