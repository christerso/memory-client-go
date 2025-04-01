@echo off
:: Memory Client MCP Service Fix Script
:: This script fixes issues with the MCP service by properly stopping, removing and reinstalling it

echo Fixing Memory Client MCP Service...

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

:: Check if the service exists and check its executable version
set "SERVICE_NAME=MemoryClientMCP"
set "NEW_SERVICE_NAME=MemoryClientMCPNew"
set "NEED_REINSTALL=0"

:: Check if either service exists
sc query %SERVICE_NAME% >nul 2>&1
if %errorLevel% equ 0 (
    set "CURRENT_SERVICE=%SERVICE_NAME%"
    echo Found existing service: %SERVICE_NAME%
    goto :CHECK_SERVICE
)

sc query %NEW_SERVICE_NAME% >nul 2>&1
if %errorLevel% equ 0 (
    set "CURRENT_SERVICE=%NEW_SERVICE_NAME%"
    echo Found existing service: %NEW_SERVICE_NAME%
    goto :CHECK_SERVICE
) else (
    echo No existing MCP service found. Will install new service.
    set "NEED_REINSTALL=1"
    goto :KILL_PROCESSES
)

:CHECK_SERVICE
:: Get the path of the executable used by the service
for /f "tokens=*" %%a in ('nssm get %CURRENT_SERVICE% Application 2^>nul') do set "SERVICE_EXE=%%a"
echo Current service executable: %SERVICE_EXE%

:: Check if we need to rebuild (if service exe doesn't match our path)
set "EXE_PATH=%ROOT_DIR%\memory-client-go.exe"
if "%SERVICE_EXE%" neq "%EXE_PATH%" (
    echo Service is using a different executable path. Need to reinstall.
    set "NEED_REINSTALL=1"
    goto :STOP_SERVICE
)

:: Check if service is paused
sc query %CURRENT_SERVICE% | findstr "PAUSED" >nul
if %errorLevel% equ 0 (
    echo Service is PAUSED. Will restart it.
    set "NEED_REINSTALL=1"
    goto :STOP_SERVICE
)

:: Check if the executable has been modified since the service was started
:: Get service start time (approximate)
for /f "tokens=*" %%a in ('sc queryex %CURRENT_SERVICE% ^| findstr "STATE" ^| find /c "RUNNING"') do set "IS_RUNNING=%%a"
if "%IS_RUNNING%"=="0" (
    echo Service is not running. Need to restart.
    set "NEED_REINSTALL=1"
    goto :STOP_SERVICE
)

:: Compare file modification time with current time
for /f "tokens=*" %%a in ('dir /a:-d /tw "%EXE_PATH%" ^| findstr /r /c:"memory-client-go\.exe"') do set "FILE_INFO=%%a"
echo Executable file info: %FILE_INFO%

echo Building a new executable to compare with the running one...
go build -o memory-client-go.exe.new
if %errorLevel% neq 0 (
    echo Failed to build the executable. Please check your Go installation.
    del memory-client-go.exe.new 2>nul
    pause
    exit /b 1
)

:: Compare file sizes as a simple version check
for /f "tokens=3" %%a in ('dir /a:-d "%EXE_PATH%" ^| findstr /r /c:"memory-client-go\.exe"') do set "OLD_SIZE=%%a"
for /f "tokens=3" %%a in ('dir /a:-d "memory-client-go.exe.new" ^| findstr /r /c:"memory-client-go\.exe\.new"') do set "NEW_SIZE=%%a"

if "%OLD_SIZE%" neq "%NEW_SIZE%" (
    echo Executable has changed size. Need to update the service.
    set "NEED_REINSTALL=1"
    move /y memory-client-go.exe.new memory-client-go.exe >nul
) else (
    echo Executable appears to be up to date.
    del memory-client-go.exe.new
    
    :: Check if service is running properly
    curl -s -o nul -w "%%{http_code}" http://localhost:8080/status >temp.txt 2>nul
    set /p STATUS_CODE=<temp.txt
    del temp.txt 2>nul
    
    if "%STATUS_CODE%"=="200" (
        echo MCP service is running correctly at http://localhost:8080/status
        echo No need to reinstall or restart.
        goto :END
    ) else (
        echo MCP service is not responding correctly. Need to restart.
        set "NEED_REINSTALL=1"
    )
)

:STOP_SERVICE
:: If we need to reinstall, stop the service
if "%NEED_REINSTALL%"=="1" (
    echo Stopping service %CURRENT_SERVICE%...
    sc stop %CURRENT_SERVICE% >nul 2>&1
    timeout /t 5 /nobreak >nul
    
    echo Removing service %CURRENT_SERVICE%...
    sc delete %CURRENT_SERVICE% >nul 2>&1
    
    :: Wait for service to be fully removed
    echo Waiting for service to be fully removed...
    timeout /t 10 /nobreak >nul
    echo Service removed.
)

:KILL_PROCESSES
:: Kill any running MCP server processes
echo Checking for running MCP server processes...
tasklist /fi "imagename eq memory-client-go.exe" | find "memory-client-go.exe" >nul
if %errorLevel% equ 0 (
    echo Stopping running MCP server processes...
    taskkill /f /im memory-client-go.exe >nul 2>&1
    timeout /t 2 /nobreak >nul
    echo MCP server processes stopped.
)

:: If no need to reinstall, we're done
if "%NEED_REINSTALL%"=="0" goto :END

:: Check if NSSM is installed
where nssm >nul 2>&1
if %errorLevel% neq 0 (
    echo NSSM (Non-Sucking Service Manager) is not installed or not in PATH.
    echo Please download and install NSSM from https://nssm.cc/download
    pause
    exit /b 1
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

:: Test the executable directly first to ensure it works
echo Testing the executable directly...
start /b cmd /c "%EXE_PATH%" mcp
timeout /t 5 /nobreak >nul
taskkill /f /im memory-client-go.exe >nul 2>&1
echo Executable test complete.

:: Install the service with a different name to avoid conflicts
set "NEW_SERVICE_NAME=MemoryClientMCPNew"
echo Installing Memory Client MCP as a Windows service with name %NEW_SERVICE_NAME%...
nssm install %NEW_SERVICE_NAME% "%EXE_PATH%" mcp
if %errorLevel% neq 0 (
    echo Failed to install service. Error code: %errorLevel%
    pause
    exit /b 1
)

echo Configuring service properties...
nssm set %NEW_SERVICE_NAME% DisplayName "Memory Client MCP Service"
nssm set %NEW_SERVICE_NAME% Description "Memory Client MCP Service for persistent conversation storage"
nssm set %NEW_SERVICE_NAME% Start SERVICE_AUTO_START
nssm set %NEW_SERVICE_NAME% AppDirectory "%ROOT_DIR%"

:: Set stdout and stderr logs
echo Setting up service logs...
set "LOG_DIR=%ROOT_DIR%\logs"
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"
nssm set %NEW_SERVICE_NAME% AppStdout "%LOG_DIR%\mcp_stdout.log"
nssm set %NEW_SERVICE_NAME% AppStderr "%LOG_DIR%\mcp_stderr.log"
nssm set %NEW_SERVICE_NAME% AppRotateFiles 1
nssm set %NEW_SERVICE_NAME% AppRotateBytes 1048576

:: Start the service
echo Starting the service...
nssm start %NEW_SERVICE_NAME%
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
nssm status %NEW_SERVICE_NAME% | findstr "SERVICE_RUNNING" >nul
if %errorLevel% neq 0 (
    echo Service is not running. Checking logs for errors...
    if exist "%LOG_DIR%\mcp_stderr.log" type "%LOG_DIR%\mcp_stderr.log"
    pause
    exit /b 1
) else (
    echo Service is running successfully!
)

:END
echo.
echo Memory Client MCP service has been checked and fixed if needed.
echo The service will automatically start when Windows boots.
echo Service is running on port 8080.
echo To verify the service is working, visit: http://localhost:8080/status
echo.

pause
