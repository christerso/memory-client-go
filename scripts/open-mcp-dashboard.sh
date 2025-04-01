#!/bin/bash
# Script to open the MCP dashboards
# This script opens both the MCP service status page and optionally the full dashboard

echo "Opening MCP dashboards..."

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$SCRIPT_DIR/.."
cd "$ROOT_DIR"

# First check if the MCP service is running
if ! curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo "MCP service does not appear to be running."
    echo "Please start the MCP service first."
    exit 1
fi

# Open the MCP service status page
echo "Opening MCP service status page..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    open http://localhost:8080
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    xdg-open http://localhost:8080 &> /dev/null || \
    sensible-browser http://localhost:8080 &> /dev/null || \
    x-www-browser http://localhost:8080 &> /dev/null || \
    gnome-open http://localhost:8080 &> /dev/null || \
    echo "Could not open browser automatically. Please open http://localhost:8080 manually."
else
    echo "Could not detect OS type. Please open http://localhost:8080 in your browser manually."
fi

# Ask if user wants to start the full dashboard
echo ""
echo "The MCP service status page has been opened in your browser."
echo ""
read -p "Do you want to start the full dashboard with dark/light mode? (y/N): " start_full

if [[ "$start_full" == "y" || "$start_full" == "Y" ]]; then
    echo ""
    echo "Starting full dashboard..."
    
    # Start the dashboard in a new terminal
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        osascript -e "tell application \"Terminal\" to do script \"cd '$ROOT_DIR' && ./memory-client dashboard\""
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux - try different terminal emulators
        gnome-terminal -- bash -c "cd '$ROOT_DIR' && ./memory-client dashboard; exec bash" 2>/dev/null || \
        xterm -e "cd '$ROOT_DIR' && ./memory-client dashboard; exec bash" 2>/dev/null || \
        konsole -e "cd '$ROOT_DIR' && ./memory-client dashboard; exec bash" 2>/dev/null || \
        x-terminal-emulator -e "cd '$ROOT_DIR' && ./memory-client dashboard; exec bash" 2>/dev/null || \
        echo "Could not open terminal automatically. Please run 'memory-client dashboard' manually."
    else
        echo "Could not detect OS type. Please run 'memory-client dashboard' manually."
    fi
    
    # Wait a moment for the dashboard to start
    sleep 3
    
    # Open the full dashboard in the browser
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        open http://localhost:8081
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        xdg-open http://localhost:8081 &> /dev/null || \
        sensible-browser http://localhost:8081 &> /dev/null || \
        x-www-browser http://localhost:8081 &> /dev/null || \
        gnome-open http://localhost:8081 &> /dev/null || \
        echo "Could not open browser automatically. Please open http://localhost:8081 manually."
    else
        echo "Could not detect OS type. Please open http://localhost:8081 in your browser manually."
    fi
    
    echo "Full dashboard started and opened in your browser."
else
    echo "Full dashboard not started."
fi

echo ""
echo "Done!"
