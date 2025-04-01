#!/bin/bash
# Script to install the Memory Client MCP service on Linux using systemd

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root or with sudo"
    exit 1
fi

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Get the current user (the one who ran sudo)
if [ -n "$SUDO_USER" ]; then
    CURRENT_USER="$SUDO_USER"
else
    CURRENT_USER="$(whoami)"
fi

echo "Installing Memory Client MCP service..."

# Create a binary by building the Go code
echo "Building memory-client-go binary..."
cd "$ROOT_DIR" || exit 1
go build -o memory-client-go

# Copy the binary to /usr/local/bin
echo "Copying binary to /usr/local/bin..."
cp memory-client-go /usr/local/bin/

# Create the service file
echo "Creating systemd service file..."
SERVICE_FILE="/etc/systemd/system/memory-client-mcp.service"

cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Memory Client MCP Service
After=network.target

[Service]
Type=simple
User=$CURRENT_USER
WorkingDirectory=$ROOT_DIR
ExecStart=/usr/local/bin/memory-client-go mcp
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable and start the service
echo "Enabling and starting the service..."
systemctl enable memory-client-mcp.service
systemctl start memory-client-mcp.service

echo "Memory Client MCP service has been installed and started."
echo "Service status:"
systemctl status memory-client-mcp.service

echo ""
echo "You can manage the service with the following commands:"
echo "  sudo systemctl start memory-client-mcp.service"
echo "  sudo systemctl stop memory-client-mcp.service"
echo "  sudo systemctl restart memory-client-mcp.service"
echo "  sudo systemctl status memory-client-mcp.service"
