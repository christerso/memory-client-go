# Memory Client MCP Service Scripts

This directory contains scripts to help you run the Memory Client MCP service before starting your editor.

## Available Scripts

### Windows Scripts

#### start-mcp-service.bat
Windows batch script to start the MCP service in the background.

```
scripts\start-mcp-service.bat
```

#### start-mcp-service.ps1
PowerShell script to start the MCP service in the background.

```powershell
.\scripts\start-mcp-service.ps1
```

#### install-mcp-service.ps1
PowerShell script to install the MCP service as a Windows service (requires Administrator privileges and NSSM).

```powershell
# Run as Administrator
.\scripts\install-mcp-service.ps1
```

### Mac/Linux Scripts

#### start-mcp-service.sh
Shell script to start the MCP service in the background on Mac or Linux.

```bash
# Make the script executable first
chmod +x scripts/start-mcp-service.sh
./scripts/start-mcp-service.sh
```

#### stop-mcp-service.sh
Shell script to stop the MCP service on Mac or Linux.

```bash
# Make the script executable first
chmod +x scripts/stop-mcp-service.sh
./scripts/stop-mcp-service.sh
```

#### install-mcp-service.sh
Shell script to install the MCP service as a systemd service on Linux (requires root privileges).

```bash
# Make the script executable first
chmod +x scripts/install-mcp-service.sh
sudo ./scripts/install-mcp-service.sh
```

#### install-mcp-service-mac.sh
Shell script to install the MCP service as a launchd service on macOS.

```bash
# Make the script executable first
chmod +x scripts/install-mcp-service-mac.sh
./scripts/install-mcp-service-mac.sh
```

## Running the MCP Service Directly

You can also run the MCP service directly using:

```
go run main.go mcp
```

The MCP service runs on port 8080 by default.

## Using with Editors

1. Start the MCP service using one of the methods above
2. Start your editor (VS Code, Visual Studio, etc.)
3. The memory service will be available for any editor or tool that connects to it

## Checking Service Status

You can check if the MCP service is running by visiting:
http://localhost:8080/status

## Dashboard

You can also view the memory dashboard by running:

```
go run main.go dashboard
```

The dashboard runs on port 8081 by default.
