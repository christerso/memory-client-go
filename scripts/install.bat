@echo off
echo Installing Memory Client for MCP...

REM Build the memory client
go build -o memory-client.exe
if %ERRORLEVEL% neq 0 (
    echo Failed to build memory client
    exit /b 1
)

REM Create necessary directories
if not exist "%APPDATA%\Roo-Code\MCP" (
    mkdir "%APPDATA%\Roo-Code\MCP"
)
if not exist "%APPDATA%\memory-client" (
    mkdir "%APPDATA%\memory-client"
)

REM Get the current directory
set CURRENT_DIR=%CD%

REM Update the path in mcp_settings.json
echo Updating MCP settings with correct path...
powershell -Command "(Get-Content mcp_settings.json) -replace 'c:/Users/christer/Desktop/memory-client-go/memory-client.exe', '%CURRENT_DIR:\=/%/memory-client.exe' | Set-Content mcp_settings.json"

REM Create persistent directory
set PERSISTENT_DIR=%APPDATA%\memory-client\bin
if not exist "%PERSISTENT_DIR%" (
    echo Creating persistent directory: %PERSISTENT_DIR%
    mkdir "%PERSISTENT_DIR%"
)

REM Copy the memory client executable to a persistent location
echo Copying memory-client.exe to persistent location...
copy /Y memory-client.exe "%PERSISTENT_DIR%\memory-client.exe"
if %ERRORLEVEL% neq 0 (
    echo Failed to copy memory-client.exe
    exit /b 1
)

REM Update MCP settings to use the persistent location
echo Updating MCP settings with persistent path...
powershell -Command "(Get-Content mcp_settings.json) -replace 'C:/Users/christer/Desktop/memory-client-go/memory-client.exe', '%PERSISTENT_DIR:\=/%/memory-client.exe' | Set-Content mcp_settings.json"

REM Copy MCP settings
echo Copying MCP settings to Roo directory...
copy /Y mcp_settings.json "%APPDATA%\Roo-Code\MCP\mcp_settings.json"
if %ERRORLEVEL% neq 0 (
    echo Failed to copy MCP settings
    exit /b 1
)

REM Copy config file
echo Copying configuration file...
copy /Y config.yaml "%APPDATA%\memory-client\config.yaml"
if %ERRORLEVEL% neq 0 (
    echo Failed to copy configuration file
    exit /b 1
)

REM Create a shortcut to run Qdrant at startup
echo Creating startup shortcut for Qdrant...
set STARTUP_FOLDER=%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup
set SHORTCUT_PATH=%STARTUP_FOLDER%\Ensure-Qdrant.lnk
set ENSURE_QDRANT_PATH=%CD%\ensure-qdrant.bat

powershell -Command "$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%SHORTCUT_PATH%'); $Shortcut.TargetPath = 'cmd.exe'; $Shortcut.Arguments = '/c \"%ENSURE_QDRANT_PATH%\"'; $Shortcut.WorkingDirectory = '%CD%'; $Shortcut.Description = 'Ensure Qdrant is running for Memory Client'; $Shortcut.Save()"

echo Startup shortcut created at: %SHORTCUT_PATH%

REM Check if Qdrant is running
echo Checking if Qdrant is running...
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:6333/collections' -Method GET -ErrorAction Stop; Write-Host 'Qdrant is running!' -ForegroundColor Green } catch { Write-Host 'Qdrant is not running on default port. Checking for other instances...' -ForegroundColor Yellow; try { $dockerPs = docker ps --format '{{.Names}} {{.Ports}}' | Select-String -Pattern 'qdrant'; if ($dockerPs) { Write-Host ('Found existing Qdrant container: ' + $dockerPs) -ForegroundColor Yellow; Write-Host 'Will use existing Qdrant instance.' -ForegroundColor Green } else { Write-Host 'No running Qdrant instances found.' -ForegroundColor Yellow; Write-Host 'Qdrant will be started automatically when needed.' -ForegroundColor Green } } catch { Write-Host 'Could not check for Docker containers.' -ForegroundColor Yellow; Write-Host 'Qdrant will be started automatically when needed.' -ForegroundColor Green } }"

REM Ensure Qdrant is running
echo Setting up Qdrant startup...
call ensure-qdrant.bat

echo.
echo Installation complete!
echo.
echo The memory client is now set up to run automatically in the background.
echo You don't need to manually start it - Cline/Roo will launch it when needed.
echo.
echo Next steps:
echo 1. Ensure Qdrant is running (default: http://localhost:6333)
echo 2. Restart Cline/Roo to load the new MCP server
echo.
echo To verify it's working:
echo 1. Have a conversation in Cline/Roo
echo 2. Close and restart Cline/Roo
echo 3. Ask about something from your previous conversation
echo.
echo You can also manually check your conversation history with:
echo   memory-client.exe history
echo.

pause