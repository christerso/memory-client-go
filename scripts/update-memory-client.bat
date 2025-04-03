@echo off
echo Updating Memory Client...

REM Build the updated Memory Client
echo Building updated Memory Client...
cd /d "%~dp0.."
go build -o "memory-client-go.exe" ./cmd/memory-client

echo Memory Client updated successfully!
echo.
echo Please restart the Memory Client application to apply the changes.
