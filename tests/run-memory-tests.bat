@echo off
echo Running Windsurf Memory Integration Tests
echo ========================================

REM Check if Node.js is installed
where node >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Node.js is required but not installed.
    echo Please install Node.js from https://nodejs.org/
    exit /b 1
)

REM Check if npm is installed
where npm >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo npm is required but not installed.
    echo Please install Node.js from https://nodejs.org/
    exit /b 1
)

REM Install dependencies
echo Installing test dependencies...
cd /d "%~dp0.."
call npm install

REM Run the tests
echo Running tests...
call npx jest tests/windsurf-memory-integration.test.js --verbose

REM Test the actual API connection
echo.
echo Testing connection to Memory Client API...
curl -s -o nul -w "%%{http_code}" http://localhost:10010/api/get-conversation-tag
if %ERRORLEVEL% neq 0 (
    echo Memory Client API is not running or not accessible.
    echo Please start the Memory Client API service first.
) else (
    echo Memory Client API is accessible.
    echo.
    echo Manual Test Instructions:
    echo 1. Open Windsurf and ensure the memory integration script is loaded
    echo 2. Type some messages in the chat
    echo 3. Check if the messages are being captured and sent to the Memory Client
    echo 4. Try setting tags and toggling tagging mode using the UI elements
)

echo.
echo Tests completed.
