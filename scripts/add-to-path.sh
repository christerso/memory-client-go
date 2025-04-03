#!/bin/bash
# Script to add memory-client binary to user PATH

echo "Adding memory-client to your PATH..."

# Get the memory client bin directory
MEMORY_CLIENT_BIN_DIR="$HOME/.local/bin"

# Check if the directory exists
if [ ! -d "$MEMORY_CLIENT_BIN_DIR" ]; then
    echo "Memory client bin directory not found at: $MEMORY_CLIENT_BIN_DIR"
    echo "Please run the installation script first."
    exit 1
fi

# Check if memory-client exists in the bin directory
if [ ! -f "$MEMORY_CLIENT_BIN_DIR/memory-client" ]; then
    echo "Memory client executable not found at: $MEMORY_CLIENT_BIN_DIR/memory-client"
    echo "Please run the installation script first."
    exit 1
fi

# Determine which shell config file to use
SHELL_CONFIG=""
if [ -f "$HOME/.bashrc" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
elif [ -f "$HOME/.bash_profile" ]; then
    SHELL_CONFIG="$HOME/.bash_profile"
elif [ -f "$HOME/.zshrc" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
else
    echo "Could not find a shell configuration file (.bashrc, .bash_profile, or .zshrc)"
    echo "Please add the following line to your shell configuration file manually:"
    echo "export PATH=\"\$PATH:$MEMORY_CLIENT_BIN_DIR\""
    exit 1
fi

# Check if the directory is already in PATH
if echo $PATH | grep -q "$MEMORY_CLIENT_BIN_DIR"; then
    echo "Memory client is already in your PATH."
    exit 0
fi

# Add to PATH in shell config
echo "" >> "$SHELL_CONFIG"
echo "# Added by memory-client installation" >> "$SHELL_CONFIG"
echo "export PATH=\"\$PATH:$MEMORY_CLIENT_BIN_DIR\"" >> "$SHELL_CONFIG"

echo "Successfully added memory-client to your PATH!"
echo "You can now run 'memory-client' from any terminal."
echo "Note: You'll need to restart your terminal or run 'source $SHELL_CONFIG' for the changes to take effect."
