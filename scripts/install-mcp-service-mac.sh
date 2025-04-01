#!/bin/bash
# Script to install the Memory Client MCP service on macOS using launchd

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Installing Memory Client MCP service for macOS..."

# Create a binary by building the Go code
echo "Building memory-client-go binary..."
cd "$ROOT_DIR" || exit 1
go build -o memory-client-go

# Create the launchd plist file
PLIST_FILE="$HOME/Library/LaunchAgents/com.christerso.memory-client-mcp.plist"
PLIST_DIR="$HOME/Library/LaunchAgents"

# Create LaunchAgents directory if it doesn't exist
mkdir -p "$PLIST_DIR"

echo "Creating launchd plist file..."

cat > "$PLIST_FILE" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.christerso.memory-client-mcp</string>
    <key>ProgramArguments</key>
    <array>
        <string>${ROOT_DIR}/memory-client-go</string>
        <string>mcp</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>${ROOT_DIR}</string>
    <key>StandardOutPath</key>
    <string>${HOME}/Library/Logs/memory-client-mcp.log</string>
    <key>StandardErrorPath</key>
    <string>${HOME}/Library/Logs/memory-client-mcp.log</string>
</dict>
</plist>
EOF

# Load the service
echo "Loading the service..."
launchctl load "$PLIST_FILE"

echo "Memory Client MCP service has been installed and started."
echo "The service will automatically start when you log in."
echo "Log file: $HOME/Library/Logs/memory-client-mcp.log"

echo ""
echo "You can manage the service with the following commands:"
echo "  launchctl load $PLIST_FILE"
echo "  launchctl unload $PLIST_FILE"
echo "  launchctl start com.christerso.memory-client-mcp"
echo "  launchctl stop com.christerso.memory-client-mcp"
