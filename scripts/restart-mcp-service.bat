@echo off
:: Simple MCP Service Restart Script
:: This script stops any existing MCP service, rebuilds the executable, and starts a new service

echo Restarting Memory Client MCP Service...

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
echo Stopping any running MCP processes...
taskkill /f /im memory-client.exe >nul 2>&1

:: Stop and remove existing services
echo Stopping and removing existing MCP services...
sc stop MemoryClientMCP >nul 2>&1
sc delete MemoryClientMCP >nul 2>&1
sc stop MemoryClientMCPNew >nul 2>&1
sc delete MemoryClientMCPNew >nul 2>&1

:: Wait to ensure services are fully removed
echo Waiting for services to be fully removed...
timeout /t 10 /nobreak >nul

:: Build the executable - build only the main package, not test files
echo Building memory-client executable...
go build -o memory-client.exe ./cmd/memory-client
if %errorLevel% neq 0 (
    echo Failed to build the executable. Please check your Go installation.
    pause
    exit /b 1
)
echo Executable built successfully.

:: Create necessary directories
set "MEMORY_CLIENT_DIR=%APPDATA%\memory-client"
set "MCP_CONFIG_DIR=%APPDATA%\Roo-Code\MCP"

if not exist "%MEMORY_CLIENT_DIR%" mkdir "%MEMORY_CLIENT_DIR%"
if not exist "%MCP_CONFIG_DIR%" mkdir "%MCP_CONFIG_DIR%"

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
)

:: Create or update MCP settings
echo Setting up MCP configuration...
set "MCP_SETTINGS_PATH=%MCP_CONFIG_DIR%\mcp_settings.json"
set "EXE_PATH=%ROOT_DIR%\memory-client.exe"
set "EXE_PATH_FORMATTED=%EXE_PATH:\=/%"
set "ROOT_DIR_FORMATTED=%ROOT_DIR:\=/%"

(
    echo {
    echo     "executable": "%EXE_PATH_FORMATTED%",
    echo     "arguments": ["mcp"],
    echo     "workingDir": "%ROOT_DIR_FORMATTED%"
    echo }
) > "%MCP_SETTINGS_PATH%"

:: Ensure Qdrant is running
echo Checking if Qdrant is running...
call "%SCRIPT_DIR%ensure-qdrant.bat"

:: Install the service
set "SERVICE_NAME=MemoryClientMCP"
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

:: Set up logging
set "LOG_DIR=%ROOT_DIR%\logs"
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"
nssm set %SERVICE_NAME% AppStdout "%LOG_DIR%\mcp_stdout.log"
nssm set %SERVICE_NAME% AppStderr "%LOG_DIR%\mcp_stderr.log"
nssm set %SERVICE_NAME% AppRotateFiles 1
nssm set %SERVICE_NAME% AppRotateBytes 1048576

:: Start the service
echo Starting the service...
nssm start %SERVICE_NAME%
if %errorLevel% neq 0 (
    echo Failed to start service. Error code: %errorLevel%
    echo Checking logs for errors...
    if exist "%LOG_DIR%\mcp_stderr.log" type "%LOG_DIR%\mcp_stderr.log"
    pause
    exit /b 1
)

:: Verify the service is running
echo Verifying service status...
timeout /t 5 /nobreak >nul
nssm status %SERVICE_NAME% | findstr "SERVICE_RUNNING" >nul
if %errorLevel% neq 0 (
    echo Service is not running. Checking logs for errors...
    if exist "%LOG_DIR%\mcp_stderr.log" type "%LOG_DIR%\mcp_stderr.log"
    pause
    exit /b 1
) else (
    echo Service is running successfully!
)

echo.
echo Memory Client MCP service has been restarted with the latest executable.
echo The service will automatically start when Windows boots.
echo Service is running on port 8080.
echo To verify the service is working, visit: http://localhost:8080/status
echo.

pause
