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

# Create necessary directories
CONFIG_DIR="$HOME/.config/memory-client"
MCP_CONFIG_DIR="$HOME/.config/cline"
mkdir -p "$CONFIG_DIR"
mkdir -p "$MCP_CONFIG_DIR"

# Create the config.yaml file if it doesn't exist
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    echo "Creating memory-client configuration file..."
    cat > "$CONFIG_DIR/config.yaml" << EOF
# Qdrant server URL
QDRANT_URL: "http://localhost:6333"

# Collection name for storing conversation memory
COLLECTION_NAME: "conversation_memory"

# Size of embedding vectors
EMBEDDING_SIZE: 384
EOF
fi

# Create or update MCP settings
echo "Setting up MCP configuration..."
MCP_SETTINGS_FILE="$MCP_CONFIG_DIR/mcp_settings.json"
cat > "$MCP_SETTINGS_FILE" << EOF
{
    "executable": "${ROOT_DIR}/memory-client-go",
    "arguments": ["mcp"],
    "workingDir": "${ROOT_DIR}"
}
EOF

echo "MCP configuration has been set up at: $MCP_SETTINGS_FILE"

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

# Ensure Qdrant is running
echo "Checking if Qdrant is running..."
"$SCRIPT_DIR/ensure-qdrant.sh"

# Load the service
echo "Loading the service..."
launchctl unload "$PLIST_FILE" 2>/dev/null || true
launchctl load "$PLIST_FILE"

echo "Memory Client MCP service has been installed and started."
echo "The service will automatically start when you log in."
echo "You can check the service status with: launchctl list | grep memory-client"
echo "Logs are available at: $HOME/Library/Logs/memory-client-mcp.log"
echo ""
echo "To verify the service is working, visit: http://localhost:8080/status"
