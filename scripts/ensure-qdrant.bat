@echo off
echo Checking if Qdrant is running...

REM Try to connect to Qdrant
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:6333/collections' -Method GET -ErrorAction Stop; Write-Host 'Qdrant is already running!' -ForegroundColor Green; exit 0 } catch { Write-Host 'Qdrant is not running. Attempting to start...' -ForegroundColor Yellow }"

REM Check if another Qdrant instance might be running on a different port
powershell -Command "try { $dockerPs = docker ps --format '{{.Names}} {{.Ports}}' | Select-String -Pattern 'qdrant'; if ($dockerPs) { Write-Host 'Found existing Qdrant container: ' + $dockerPs -ForegroundColor Yellow; Write-Host 'Using existing Qdrant instance instead of starting a new one.' -ForegroundColor Green; exit 0 } } catch { }"

REM If we get here, Qdrant is not running
REM Check if Docker is installed
docker --version > nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Docker is not installed. Please install Docker to run Qdrant.
    echo You can download Docker from: https://www.docker.com/products/docker-desktop
    exit /b 1
)

REM Start Qdrant using Docker
echo Starting Qdrant using Docker...
docker run -d -p 6333:6333 -p 6334:6334 --name qdrant qdrant/qdrant

REM Check if Qdrant started successfully
timeout /t 5 /nobreak > nul
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:6333/collections' -Method GET -ErrorAction Stop; Write-Host 'Qdrant started successfully!' -ForegroundColor Green; exit 0 } catch { Write-Host 'Failed to start Qdrant.' -ForegroundColor Red; exit 1 }"

echo.
echo If Qdrant is running, the memory client will work seamlessly in the background.
echo You don't need to manually start or manage it - Cline/Roo will handle everything.
echo.