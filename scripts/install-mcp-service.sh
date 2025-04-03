#!/bin/bash
# Script to install the Memory Client MCP service on Linux using systemd

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root or with sudo"
    exit 1
fi

# Set installation directories
INSTALL_DIR="/opt/memory-client-mcp"
CONFIG_DIR="/etc/memory-client-mcp"
SERVICE_NAME="memory-client-mcp"
VECTOR_SERVICE_NAME="vector-service"
EXECUTABLE_NAME="memory-client"

echo "Memory Client MCP Service Installer"
echo "==================================="
echo ""

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Get the current user (the one who ran sudo)
if [ -n "$SUDO_USER" ]; then
    CURRENT_USER="$SUDO_USER"
    HOME_DIR=$(eval echo ~$SUDO_USER)
else
    CURRENT_USER="$(whoami)"
    HOME_DIR="$HOME"
fi

# Check if vector service is running
if systemctl is-active --quiet "$VECTOR_SERVICE_NAME"; then
    echo "Vector service is already running, will skip vector service installation."
    SKIP_VECTOR=1
else
    echo "Vector service not detected, will include vector service installation."
    SKIP_VECTOR=0
fi

# Check if MCP service exists
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "MCP service already exists. Stopping and removing..."
    systemctl stop "$SERVICE_NAME"
    systemctl disable "$SERVICE_NAME"
    rm -f "/etc/systemd/system/$SERVICE_NAME.service"
    systemctl daemon-reload
    sleep 5
fi

# Create installation directories
echo "Creating installation directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/scripts"
mkdir -p "$INSTALL_DIR/vscode-memory-extension"
mkdir -p "$CONFIG_DIR"

# Build the executable
echo "Building memory-client binary..."
cd "$ROOT_DIR" || exit 1
go build -o "$INSTALL_DIR/$EXECUTABLE_NAME"

if [ $? -ne 0 ]; then
    echo "Failed to build memory-client."
    exit 1
fi

# Copy integration scripts
echo "Copying Windsurf integration script..."
cp -f "$ROOT_DIR/scripts/windsurf-memory-integration.js" "$INSTALL_DIR/scripts/"

echo "Copying VS Code extension..."
cp -rf "$ROOT_DIR/vscode-memory-extension/"* "$INSTALL_DIR/vscode-memory-extension/"

# Create default configuration
CONFIG_FILE="$CONFIG_DIR/config.json"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating default configuration..."
    cat > "$CONFIG_FILE" << EOF
{
  "database": {
    "type": "sqlite",
    "connection": "$CONFIG_DIR/memory.db"
  },
  "api": {
    "port": 10010,
    "host": "localhost"
  },
  "dashboard": {
    "port": 8081,
    "host": "localhost"
  },
  "tagging": {
    "defaultMode": "automatic",
    "bufferSize": 5,
    "categories": [
      "technical",
      "planning",
      "question",
      "feedback"
    ]
  }
}
EOF
fi

# Create the service file
echo "Creating systemd service file..."
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Memory Client MCP Service for conversation capture and tagging
After=network.target

[Service]
Type=simple
User=$CURRENT_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$EXECUTABLE_NAME mcp-server --config $CONFIG_FILE
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Create symlink to the executable
echo "Creating symlink to the executable..."
ln -sf "$INSTALL_DIR/$EXECUTABLE_NAME" "/usr/local/bin/$EXECUTABLE_NAME"

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable and start the service
echo "Enabling and starting the service..."
systemctl enable "$SERVICE_NAME.service"
systemctl start "$SERVICE_NAME.service"

# Check if service started successfully
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "Memory Client MCP service has been installed and started successfully."
else
    echo "Failed to start Memory Client MCP service. Please check the logs with 'journalctl -u $SERVICE_NAME'."
    exit 1
fi

echo ""
echo "Installation completed successfully!"
echo ""
echo "Memory Client MCP Service has been installed and started."
echo ""
echo "Conversation Capture Features:"
echo "- HTTP API running on port 10010"
echo "- Automatic message tagging and categorization"
echo "- VS Code extension available in $INSTALL_DIR/vscode-memory-extension"
echo "- Windsurf integration script available in $INSTALL_DIR/scripts"
echo ""
echo "To use the conversation capture client:"
echo "memory-client message -role=user -content=\"Your message\""
echo "memory-client tag -tag=\"your-tag\""
echo "memory-client tag-mode -mode=automatic"
echo ""
echo "To view the dashboard:"
echo "memory-client dashboard"
echo ""
echo "You can manage the service with the following commands:"
echo "  sudo systemctl start $SERVICE_NAME.service"
echo "  sudo systemctl stop $SERVICE_NAME.service"
echo "  sudo systemctl restart $SERVICE_NAME.service"
echo "  sudo systemctl status $SERVICE_NAME.service"
