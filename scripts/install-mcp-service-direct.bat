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

:: Also check if the service is already running and remove it properly
set "SERVICE_NAME=MemoryClientMCP"
sc query %SERVICE_NAME% >nul 2>&1
if %errorLevel% equ 0 (
    echo Stopping and removing existing MCP service...
    net stop %SERVICE_NAME% >nul 2>&1
    sc delete %SERVICE_NAME% >nul 2>&1
    
    :: Wait for service to be fully removed
    echo Waiting for service to be fully removed...
    timeout /t 10 /nobreak >nul
    echo MCP service removed.
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

:: Install the service using NSSM
echo Installing MCP service using NSSM...
set "NSSM_PATH=%ROOT_DIR%\scripts\nssm.exe"

:: Check if NSSM exists
if not exist "%NSSM_PATH%" (
    echo NSSM not found at %NSSM_PATH%
    echo Downloading NSSM...
    
    :: Create a PowerShell script to download NSSM
    echo $url = 'https://nssm.cc/release/nssm-2.24.zip' > "%TEMP%\download_nssm.ps1"
    echo $output = '%TEMP%\nssm.zip' >> "%TEMP%\download_nssm.ps1"
    echo $extractPath = '%TEMP%\nssm' >> "%TEMP%\download_nssm.ps1"
    echo [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 >> "%TEMP%\download_nssm.ps1"
    echo Invoke-WebRequest -Uri $url -OutFile $output >> "%TEMP%\download_nssm.ps1"
    echo Expand-Archive -Path $output -DestinationPath $extractPath -Force >> "%TEMP%\download_nssm.ps1"
    echo $nssmExe = Get-ChildItem -Path $extractPath -Recurse -Filter 'nssm.exe' ^| Where-Object {$_.FullName -like '*win64*'} ^| Select-Object -First 1 >> "%TEMP%\download_nssm.ps1"
    echo Copy-Item -Path $nssmExe.FullName -Destination '%NSSM_PATH%' -Force >> "%TEMP%\download_nssm.ps1"
    
    :: Run the PowerShell script
    powershell -ExecutionPolicy Bypass -File "%TEMP%\download_nssm.ps1"
    
    :: Check if download was successful
    if not exist "%NSSM_PATH%" (
        echo Failed to download NSSM. Please download it manually from https://nssm.cc/
        echo and place it in the scripts directory as nssm.exe.
        pause
        exit /b 1
    )
    
    echo NSSM downloaded successfully.
)

:: Create logs directory if it doesn't exist
if not exist "%ROOT_DIR%\logs" (
    mkdir "%ROOT_DIR%\logs"
)

:: Use a different service name to avoid conflicts
set "SERVICE_NAME=MemoryClientMCPService"

:: Install the service
echo Installing service %SERVICE_NAME%...
"%NSSM_PATH%" install %SERVICE_NAME% "%ROOT_DIR%\memory-client-go.exe" "mcp"
"%NSSM_PATH%" set %SERVICE_NAME% AppDirectory "%ROOT_DIR%"
"%NSSM_PATH%" set %SERVICE_NAME% DisplayName "Memory Client MCP Service"
"%NSSM_PATH%" set %SERVICE_NAME% Description "Memory Client MCP Service for persistent conversation storage"
"%NSSM_PATH%" set %SERVICE_NAME% AppStdout "%ROOT_DIR%\logs\mcp_service_stdout.log"
"%NSSM_PATH%" set %SERVICE_NAME% AppStderr "%ROOT_DIR%\logs\mcp_service_stderr.log"
"%NSSM_PATH%" set %SERVICE_NAME% AppRotateFiles 1
"%NSSM_PATH%" set %SERVICE_NAME% AppRotateBytes 1048576
"%NSSM_PATH%" set %SERVICE_NAME% Start SERVICE_AUTO_START

:: Start the service
echo Starting service %SERVICE_NAME%...
net start %SERVICE_NAME%
if %errorLevel% neq 0 (
    echo Failed to start the service. Please check the logs.
    pause
    exit /b 1
)

:: Wait a moment for the service to fully start
echo Waiting for service to fully start...
timeout /t 5 /nobreak >nul

:: Verify the service is running
echo Verifying service status...
sc query %SERVICE_NAME% | find "RUNNING" >nul
if %errorLevel% neq 0 (
    echo Service is not running. Please check the logs.
    pause
    exit /b 1
)

:: Check if the MCP server is responding
echo Checking if MCP server is responding...
curl -s http://localhost:8080 >nul 2>&1
if %errorLevel% neq 0 (
    echo MCP server is not responding. Please check the logs.
    pause
    exit /b 1
)

echo MCP service installed and running successfully!

:: Open the MCP dashboard in the browser
echo Opening MCP dashboard in your browser...
start http://localhost:8080

:: Ask if user wants to start the full dashboard
echo.
echo The MCP service status page has been opened in your browser.
echo.
set /p start_full="Do you want to start the full dashboard with dark/light mode? (y/N): "

if /i "%start_full%"=="y" (
    echo.
    echo Starting full dashboard...
    start cmd /k "cd /d "%ROOT_DIR%" && memory-client-go.exe dashboard"
    
    :: Wait a moment for the dashboard to start
    timeout /t 3 /nobreak >nul
    
    :: Open the full dashboard in the browser
    start http://localhost:8081
    echo Full dashboard started and opened in your browser.
) else (
    echo Full dashboard not started.
)

echo.
echo Done!
pause
