@echo off
setlocal enabledelayedexpansion

echo Memory Client MCP Service Installer
echo ===================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo This script requires administrator privileges.
    echo Please run this script as an administrator.
    exit /b 1
)

set "INSTALL_DIR=%ProgramFiles%\MemoryClientMCP"
set "SERVICE_NAME=MemoryClientMCPService"
set "EXECUTABLE_NAME=memory-client.exe"
set "CONFIG_DIR=%APPDATA%\MemoryClientMCP"
set "CONFIG_FILE=%CONFIG_DIR%\config.json"
set "VECTOR_SERVICE_NAME=VectorService"

echo Checking for existing services...

REM Check if vector service is running
sc query "%VECTOR_SERVICE_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo Vector service is already running, will skip vector service installation.
    set "SKIP_VECTOR=1"
) else (
    echo Vector service not detected, will include vector service installation.
    set "SKIP_VECTOR=0"
)

REM Check if MCP service exists
sc query "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo MCP service already exists. Stopping and removing...
    sc stop "%SERVICE_NAME%" >nul 2>&1
    timeout /t 5 /nobreak >nul
    sc delete "%SERVICE_NAME%" >nul 2>&1
    timeout /t 10 /nobreak >nul
)

echo Creating installation directory...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

echo Creating configuration directory...
if not exist "%CONFIG_DIR%" mkdir "%CONFIG_DIR%"

echo Building latest version of memory-client...
cd /d "%~dp0.."
go build -o "%INSTALL_DIR%\%EXECUTABLE_NAME%" .

if %errorlevel% neq 0 (
    echo Failed to build memory-client.
    exit /b 1
)

echo Copying Windsurf integration script...
if not exist "%INSTALL_DIR%\scripts" mkdir "%INSTALL_DIR%\scripts"
copy /Y "%~dp0..\scripts\windsurf-memory-integration.js" "%INSTALL_DIR%\scripts\" >nul

echo Copying VS Code extension...
if not exist "%INSTALL_DIR%\vscode-memory-extension" mkdir "%INSTALL_DIR%\vscode-memory-extension"
xcopy /E /Y "%~dp0..\vscode-memory-extension\*" "%INSTALL_DIR%\vscode-memory-extension\" >nul

echo Creating default configuration...
if not exist "%CONFIG_FILE%" (
    echo {
    echo   "database": {
    echo     "type": "sqlite",
    echo     "connection": "%CONFIG_DIR%\memory.db"
    echo   },
    echo   "api": {
    echo     "port": 10010,
    echo     "host": "localhost"
    echo   },
    echo   "dashboard": {
    echo     "port": 8081,
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
    echo } > "%CONFIG_FILE%"
)

echo Installing MCP service...
sc create "%SERVICE_NAME%" binPath= "\"%INSTALL_DIR%\%EXECUTABLE_NAME%\" mcp-server --config \"%CONFIG_FILE%\"" start= auto DisplayName= "Memory Client MCP Service"
if %errorlevel% neq 0 (
    echo Failed to create MCP service.
    exit /b 1
)

echo Setting service description...
sc description "%SERVICE_NAME%" "Memory Client MCP service for conversation capture and tagging"

echo Starting MCP service...
sc start "%SERVICE_NAME%"
if %errorlevel% neq 0 (
    echo Failed to start MCP service.
    exit /b 1
)

echo.
echo Installation completed successfully!
echo.
echo Memory Client MCP Service has been installed and started.
echo.
echo Conversation Capture Features:
echo - HTTP API running on port 10010
echo - Automatic message tagging and categorization
echo - VS Code extension available in %INSTALL_DIR%\vscode-memory-extension
echo - Windsurf integration script available in %INSTALL_DIR%\scripts
echo.
echo To use the conversation capture client:
echo memory-client message -role=user -content="Your message"
echo memory-client tag -tag="your-tag"
echo memory-client tag-mode -mode=automatic
echo.
echo To view the dashboard:
echo memory-client dashboard
echo.

endlocal
