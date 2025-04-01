@echo off
setlocal enabledelayedexpansion

echo Memory Client MCP Service Updater
echo ================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo This script requires administrator privileges.
    echo Please run this script as an administrator.
    exit /b 1
)

set "INSTALL_DIR=%ProgramFiles%\MemoryClientMCP"
set "OLD_SERVICE_NAME=MemoryClientMCP"
set "SERVICE_NAME=MemoryClientMCPService"
set "EXECUTABLE_NAME=memory-client.exe"
set "CONFIG_DIR=%APPDATA%\MemoryClientMCP"
set "CONFIG_FILE=%CONFIG_DIR%\config.json"
set "VECTOR_SERVICE_NAME=VectorService"
set "BACKUP_DIR=%CONFIG_DIR%\backup_%date:~-4,4%%date:~-7,2%%date:~-10,2%_%time:~0,2%%time:~3,2%%time:~6,2%"
set "BACKUP_DIR=%BACKUP_DIR: =0%"
set "REPO_DIR=%~dp0.."

echo Checking for existing services...

REM Check for the old service name first
sc query "%OLD_SERVICE_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo Old service format detected: %OLD_SERVICE_NAME%
    echo Stopping and removing old service...
    sc stop "%OLD_SERVICE_NAME%" >nul 2>&1
    timeout /t 5 /nobreak >nul
    sc delete "%OLD_SERVICE_NAME%" >nul 2>&1
    timeout /t 10 /nobreak >nul
    echo Old service removed.
)

REM Check for the current service name
sc query "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo Current service detected: %SERVICE_NAME%
    echo Stopping and removing current service...
    sc stop "%SERVICE_NAME%" >nul 2>&1
    timeout /t 5 /nobreak >nul
    sc delete "%SERVICE_NAME%" >nul 2>&1
    timeout /t 10 /nobreak >nul
    echo Current service removed.
)

REM Check if vector service is running
sc query "%VECTOR_SERVICE_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo Vector service is already running, will skip vector service installation.
    set "SKIP_VECTOR=1"
) else (
    echo Vector service not detected, will include vector service installation.
    set "SKIP_VECTOR=0"
)

REM Backup existing configuration if it exists
if exist "%CONFIG_FILE%" (
    echo Backing up existing configuration...
    if not exist "%BACKUP_DIR%" mkdir "%BACKUP_DIR%"
    copy "%CONFIG_FILE%" "%BACKUP_DIR%\config.json.bak" >nul
    echo Configuration backed up to %BACKUP_DIR%\config.json.bak
)

echo Creating installation directory...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

echo Creating configuration directory...
if not exist "%CONFIG_DIR%" mkdir "%CONFIG_DIR%"

echo Building latest version of memory-client...
cd /d "%REPO_DIR%"
go build -o "%INSTALL_DIR%\%EXECUTABLE_NAME%" "%REPO_DIR%\main.go"
if %errorlevel% neq 0 (
    echo Failed to build memory-client.
    exit /b 1
)

echo Copying Windsurf integration script...
copy /Y "%REPO_DIR%\scripts\windsurf-memory-integration.js" "%INSTALL_DIR%\windsurf-memory-integration.js" >nul

echo Copying VS Code extension...
if not exist "%INSTALL_DIR%\vscode-extension" mkdir "%INSTALL_DIR%\vscode-extension"
xcopy /E /Y /Q "%REPO_DIR%\vscode-memory-extension\*" "%INSTALL_DIR%\vscode-extension\" >nul

echo Creating or updating configuration...
if not exist "%CONFIG_DIR%\config.json" (
    echo Creating new configuration file...
    (
        echo {
        echo   "database": {
        echo     "type": "sqlite",
        echo     "connection": "%CONFIG_DIR:\=\\%\\memory.db"
        echo   },
        echo   "api": {
        echo     "port": 10010,
        echo     "host": "localhost"
        echo   },
        echo   "dashboard": {
        echo     "port": 9581,
        echo     "host": "localhost"
        echo   },
        echo   "tagging": {
        echo     "defaultMode": "automatic",
        echo     "bufferSize": 5,
        echo     "categories": [
        echo       "technical",
        echo       "planning",
        echo       "question",
        echo       "feedback"
        echo     ]
        echo   }
    ) > "%CONFIG_DIR%\config.json"
)

echo Installing MCP service...
:: Use NSSM for service installation
nssm install %SERVICE_NAME% "%INSTALL_DIR%\%EXECUTABLE_NAME%" mcp
if %errorlevel% neq 0 (
    echo Failed to install service. Error code: %errorlevel%
    exit /b 1
)

echo Setting service description...
nssm set %SERVICE_NAME% DisplayName "Memory Client MCP Service"
nssm set %SERVICE_NAME% Description "Memory Client MCP Service for persistent conversation storage"
nssm set %SERVICE_NAME% Start SERVICE_AUTO_START
nssm set %SERVICE_NAME% AppDirectory "%INSTALL_DIR%"

:: Set up logging
set "LOG_DIR=%INSTALL_DIR%\logs"
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"
nssm set %SERVICE_NAME% AppStdout "%LOG_DIR%\mcp_stdout.log"
nssm set %SERVICE_NAME% AppStderr "%LOG_DIR%\mcp_stderr.log"
nssm set %SERVICE_NAME% AppRotateFiles 1
nssm set %SERVICE_NAME% AppRotateBytes 1048576

echo Starting the service...
nssm start %SERVICE_NAME%
if %errorlevel% neq 0 (
    echo Failed to start service. Error code: %errorlevel%
    echo Checking logs for errors...
    if exist "%LOG_DIR%\mcp_stderr.log" type "%LOG_DIR%\mcp_stderr.log"
    exit /b 1
)

:: Verify the service is running
echo Verifying service status...
timeout /t 5 /nobreak >nul
nssm status %SERVICE_NAME% | findstr "SERVICE_RUNNING" >nul
if %errorlevel% neq 0 (
    echo Service is not running. Checking logs for errors...
    if exist "%LOG_DIR%\mcp_stderr.log" type "%LOG_DIR%\mcp_stderr.log"
    exit /b 1
) else (
    echo Service is running successfully!
)

echo.
echo Update completed successfully!
echo.
echo Memory Client MCP Service has been updated and started with the new tagging features.
echo.
echo Conversation Capture Features:
echo - HTTP API running on port 10010
echo - Automatic message tagging and categorization
echo - VS Code extension available in %INSTALL_DIR%\vscode-extension
echo - Windsurf integration script available in %INSTALL_DIR%\windsurf-memory-integration.js
echo.
echo To use the conversation capture client:
echo memory-client message -role=user -content="Your message"
echo memory-client tag -tag="your-tag"
echo memory-client tag-mode -mode=automatic
echo.
echo To view the dashboard:
echo memory-client dashboard
echo.
echo Status Page: http://localhost:9580/status
echo Dashboard: http://localhost:9581 (run 'memory-client dashboard')
echo.

endlocal
