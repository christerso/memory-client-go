#!/bin/bash
# Simple MCP Service Status Check Script for Mac/Linux
# This script checks if the MCP service is running and responding

echo "Checking Memory Client MCP Service status..."

# Check if the service is running using systemctl or launchctl
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    if launchctl list | grep -q "com.christerso.memory-client-mcp"; then
        echo "Service status: RUNNING"
    else
        echo "Service status: NOT RUNNING"
    fi
elif [[ -f /bin/systemctl ]] || [[ -f /usr/bin/systemctl ]]; then
    # Linux with systemd
    if systemctl is-active --quiet memory-client-mcp.service; then
        echo "Service status: RUNNING"
    else
        if systemctl list-unit-files | grep -q "memory-client-mcp.service"; then
            echo "Service status: NOT RUNNING"
        else
            echo "Service status: NOT INSTALLED"
        fi
    fi
else
    # Linux without systemd - check for running process
    if pgrep -f "memory-client mcp" > /dev/null; then
        echo "Process status: RUNNING"
    else
        echo "Process status: NOT RUNNING"
    fi
fi

# Check if the service is responding to HTTP requests
echo ""
echo "Checking HTTP response..."
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/status)
echo "HTTP Status: $HTTP_STATUS"

# Check service logs if they exist
LOG_DIR="$(dirname "$(realpath "$0")")/../logs"
if [[ -f "$LOG_DIR/mcp_stderr.log" ]]; then
    echo ""
    echo "Recent errors from log:"
    tail -10 "$LOG_DIR/mcp_stderr.log" | grep -v "^$"
fi

echo ""
echo "To restart the service:"
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Run: sudo ./restart-mcp-service.sh"
else
    echo "Run: sudo systemctl restart memory-client-mcp.service"
fi
echo ""
