# Memory Client MCP Service Scripts

This directory contains scripts to help you manage the Memory Client MCP service.

## Available Scripts

### Windows Scripts

#### MCP Service Management

#### restart-mcp-service.bat
Windows batch script to stop, rebuild, and restart the MCP service. Use this after code changes to ensure the service is running the latest version.

```batch
# Run as Administrator
scripts\restart-mcp-service.bat
```

#### check-mcp-status.bat
Windows batch script to check if the MCP service is running and responding.

```batch
scripts\check-mcp-status.bat
```

#### fix-mcp-service.bat
Windows batch script to fix issues with the MCP service by checking version, stopping, removing, and reinstalling it if necessary.

```batch
# Run as Administrator
scripts\fix-mcp-service.bat
```

#### uninstall-mcp-service.bat
Windows batch script to stop and remove the MCP service.

```batch
# Run as Administrator
scripts\uninstall-mcp-service.bat
```

#### install-mcp-service.bat
Windows batch script to install the MCP service as a Windows service (requires Administrator privileges and NSSM).

```batch
# Run as Administrator
scripts\install-mcp-service.bat
```

#### start-mcp-service.bat
Windows batch script to start the MCP service in the background.

```batch
scripts\start-mcp-service.bat
```

#### verify-mcp-service.bat
Windows batch script to verify the MCP service is installed and running correctly.

```batch
scripts\verify-mcp-service.bat
```

#### ensure-qdrant.bat
Windows batch script to ensure Qdrant is running, which is required for the MCP service.

```batch
scripts\ensure-qdrant.bat
```

#### PowerShell Scripts

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

#### ensure-qdrant.ps1
PowerShell script to ensure Qdrant is running.

```powershell
.\scripts\ensure-qdrant.ps1
```

#### add-to-path.ps1
PowerShell script to add the memory-client to your PATH for easier command-line access.

```powershell
.\scripts\add-to-path.ps1
```

#### index-project.ps1
PowerShell script to index a project directory with customizable options.

```powershell
# Basic usage
.\scripts\index-project.ps1 -ProjectPath "C:\path\to\project" -Tag "my-project"

# Advanced usage with custom batch size and max file size
.\scripts\index-project.ps1 -ProjectPath "C:\path\to\project" -Tag "my-project" -BatchSize 100 -MaxFileSizeKB 2048
```

### Mac/Linux Scripts

#### check-mcp-status.sh
Shell script to check if the MCP service is running and responding on macOS or Linux.

```bash
./scripts/check-mcp-status.sh
```

#### start-mcp-service.sh
Shell script to start the MCP service in the background on macOS or Linux.

```bash
./scripts/start-mcp-service.sh
```

#### stop-mcp-service.sh
Shell script to stop the MCP service on macOS or Linux.

```bash
./scripts/stop-mcp-service.sh
```

#### install-mcp-service.sh
Shell script to install the MCP service as a system service on macOS or Linux.

```bash
# On macOS
sudo ./scripts/install-mcp-service-mac.sh

# On Linux
sudo ./scripts/install-mcp-service.sh
```

#### ensure-qdrant.sh
Shell script to ensure Qdrant is running on macOS or Linux.

```bash
./scripts/ensure-qdrant.sh
```

#### add-to-path.sh
Shell script to add the memory-client to your PATH on macOS or Linux.

```bash
./scripts/add-to-path.sh
```

#### index-project.sh
Shell script to index a project directory on macOS or Linux.

```bash
# Basic usage
./scripts/index-project.sh -p /path/to/project -t my-project

# Advanced usage
./scripts/index-project.sh -p /path/to/project -t my-project -b 100 -m 2048
```

## Dashboard Scripts

### Windows

- **open-mcp-dashboard.bat**: Opens the MCP service status page in your browser and optionally starts the full dashboard with dark/light mode

### Mac/Linux

- **open-mcp-dashboard.sh**: Opens the MCP service status page in your browser and optionally starts the full dashboard with dark/light mode

## Configuration Files

#### windsurf-mcp-config.json
Configuration file for Windsurf to connect to the MCP service.

#### memory-client-mcp.service
Systemd service file for Linux installations.

## Troubleshooting

If you encounter issues with the MCP service:

1. Check the service status using `check-mcp-status.bat` or `check-mcp-status.sh`
2. Look at the logs in the `logs` directory
3. Try restarting the service with `restart-mcp-service.bat` or by running `sudo systemctl restart memory-client-mcp.service` on Linux
4. If the service is in a PAUSED state, use `fix-mcp-service.bat` to repair it

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
