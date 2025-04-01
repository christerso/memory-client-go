#!/bin/bash
# Script to stop the Memory Client MCP service on Mac/Linux

if [ -f /tmp/memory-client-mcp.pid ]; then
    PID=$(cat /tmp/memory-client-mcp.pid)
    if ps -p "$PID" > /dev/null; then
        echo "Stopping Memory Client MCP Service (PID: $PID)..."
        kill "$PID"
        rm /tmp/memory-client-mcp.pid
        echo "Service stopped."
    else
        echo "No running MCP service found with PID: $PID"
        rm /tmp/memory-client-mcp.pid
    fi
else
    echo "No PID file found. Service may not be running."
    # Try to find and kill the process by name
    PID=$(pgrep -f "go run main.go mcp" || true)
    if [ -n "$PID" ]; then
        echo "Found MCP service with PID: $PID. Stopping..."
        kill "$PID"
        echo "Service stopped."
    fi
fi
