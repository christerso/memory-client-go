#!/bin/bash
# Script to start the Memory Client MCP service on Mac/Linux

echo "Starting Memory Client MCP Service..."

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Start the MCP service in the background
cd "$ROOT_DIR" || exit 1
nohup go run main.go mcp > /tmp/memory-client-mcp.log 2>&1 &

# Save the process ID
echo $! > /tmp/memory-client-mcp.pid

echo "MCP Service started on port 8080 (PID: $(cat /tmp/memory-client-mcp.pid))."
echo "Log file: /tmp/memory-client-mcp.log"
echo "You can now start your editor."
