#!/bin/bash
set -e

echo "Installing Memory Client for MCP..."

# Build the memory client
go build -o memory-client
if [ $? -ne 0 ]; then
    echo "Failed to build memory client"
    exit 1
fi

# Create config directories if they don't exist
mkdir -p ~/.config/cline/
mkdir -p ~/.config/memory-client/

# Create persistent directory
PERSISTENT_DIR=~/.local/bin
mkdir -p $PERSISTENT_DIR

# Copy the memory client executable to a persistent location
echo "Copying memory-client to persistent location..."
cp memory-client "$PERSISTENT_DIR/memory-client"
if [ $? -ne 0 ]; then
    echo "Failed to copy memory-client"
    exit 1
fi
chmod +x "$PERSISTENT_DIR/memory-client"

# Copy MCP settings
echo "Updating MCP settings..."
cp mcp_settings.json ~/.config/cline/mcp_settings.json
if [ $? -ne 0 ]; then
    echo "Failed to copy MCP settings"
    exit 1
fi

# Update path in settings
sed -i "s|c:/Users/christer/Desktop/memory-client-go/memory-client.exe|$PERSISTENT_DIR/memory-client|g" ~/.config/cline/mcp_settings.json

# Copy config
echo "Copying configuration file..."
cp config.yaml ~/.config/memory-client/config.yaml

# Set up Qdrant to run at startup
echo "Setting up Qdrant to run at startup..."
ENSURE_QDRANT_PATH="$PERSISTENT_DIR/ensure-qdrant.sh"
cp ensure-qdrant.sh "$ENSURE_QDRANT_PATH"
chmod +x "$ENSURE_QDRANT_PATH"

# Add to user's crontab to run at startup
(crontab -l 2>/dev/null || echo "") | grep -v "$ENSURE_QDRANT_PATH" | { cat; echo "@reboot $ENSURE_QDRANT_PATH"; } | crontab -

# Check if Qdrant is running
echo "Checking if Qdrant is running..."
if curl -s http://localhost:6333/collections > /dev/null; then
    echo -e "\e[32mQdrant is running!\e[0m"
else
    echo -e "\e[33mQdrant is not running on default port. Checking for other instances...\e[0m"
    
    # Check if another Qdrant instance might be running on a different port
    if docker ps | grep -q qdrant; then
        echo -e "\e[33mFound existing Qdrant container:\e[0m"
        docker ps | grep qdrant
        echo -e "\e[32mWill use existing Qdrant instance.\e[0m"
    else
        echo -e "\e[33mNo running Qdrant instances found.\e[0m"
        echo -e "\e[32mQdrant will be started automatically when needed.\e[0m"
    fi
fi

# Set up Qdrant startup
echo "Setting up Qdrant startup..."
chmod +x ensure-qdrant.sh
./ensure-qdrant.sh

echo ""
echo "Installation complete!"
echo ""
echo "The memory client is now set up to run automatically in the background."
echo "You don't need to manually start it - Cline/Roo will launch it when needed."
echo ""
echo "Next steps:"
echo "1. Ensure Qdrant is running (default: http://localhost:6333)"
echo "2. Restart Cline/Roo to load the new MCP server"
echo ""
echo "To verify it's working:"
echo "1. Have a conversation in Cline/Roo"
echo "2. Close and restart Cline/Roo"
echo "3. Ask about something from your previous conversation"
echo ""
echo "You can also manually check your conversation history with:"
echo "  $PERSISTENT_DIR/memory-client history"
echo ""