# Memory Client Installation Guide

This guide will help you install and set up the Memory Client to work seamlessly in the background with Cline/Roo, automatically storing and retrieving your conversation history without requiring any manual intervention.

## Prerequisites

- [Docker](https://www.docker.com/products/docker-desktop) (for running Qdrant)
  - The installation scripts will automatically start Qdrant in Docker if it's not already running
  - Default Qdrant URL: http://localhost:6333

## Installation Steps

### Windows

#### Using PowerShell (Recommended)

1. **Run the PowerShell installation script**:
   ```powershell
   .\install.ps1
   ```
   
   This script will:
   - Build the memory client
   - Create necessary directories
   - Copy the MCP settings to the Roo configuration directory
   - Set up the memory client to work with Cline/Roo
   - Check if Qdrant is running

2. **Verify installation**:
   ```powershell
   .\memory-client.exe history
   ```
   
   You should see your conversation history (if any).

#### Using Batch File (Alternative)

1. **Run the batch installation script**:
   ```
   install.bat
   ```
   
   This script performs the same operations as the PowerShell script.

2. **Verify installation**:
   ```
   memory-client.exe history
   ```

### Linux/macOS

1. **Build the memory client**:
   ```
   go build -o memory-client
   ```

2. **Make the installation script executable**:
   ```
   chmod +x install.sh
   ```

3. **Run the installation script**:
   ```
   ./install.sh
   ```
   
   This script will:
   - Build the memory client
   - Create necessary directories
   - Copy the MCP settings to the Cline configuration directory
   - Set up the memory client to work with Cline/Roo

4. **Verify installation**:
   ```
   ./memory-client history
   ```
   
   You should see your conversation history (if any).

## How It Works

Once installed, the memory client will:

1. **Automatically start** when Cline/Roo needs it (via MCP protocol)
2. **Silently run in the background** without requiring any user interaction
3. **Store all conversations** in the Qdrant vector database
4. **Index project files** when you open a folder in Visual Studio Code
5. **Provide conversation history and project context** to Cline/Roo when needed
6. **Persist across system restarts** thanks to permanent installation

The installation process:
1. Copies the memory client to a persistent location (`%APPDATA%\memory-client\bin` on Windows, `~/.local/bin` on Linux/macOS)
2. Sets up Qdrant to run automatically at system startup
3. Configures Cline/Roo to use the persistent memory client location

You don't need to manually start or manage the memory client - it will be launched automatically by Cline/Roo when needed and will shut down when no longer required.

## Project Memory

The memory client can index your project files to provide context-aware assistance. This happens automatically when you open a folder in Visual Studio Code, but you can also manually index projects:

```bash
# Windows
memory-client.exe index-project --project "C:\path\to\project"

# Linux/macOS
./memory-client index-project --project "/path/to/project"
```

### What Gets Indexed

The memory client indexes:
- Source code files
- Configuration files
- Documentation
- Text files

It automatically excludes:
- Binary files
- Media files (images, videos, etc.)
- Large files (>1MB)
- Hidden files and directories

### Keeping Project Memory Updated

The memory client provides several ways to keep your project memory up to date:

1. **Manual Updates**:
   ```bash
   # Update modified files in the current directory
   memory-client update-project
   
   # Update modified files in a specific directory
   memory-client update-project /path/to/project
   ```

2. **Automatic Watching**:
   ```bash
   # Watch the current directory for changes
   memory-client watch-project
   
   # Watch a specific directory for changes
   memory-client watch-project /path/to/project
   ```
   
   The watch command will:
   - Initially index all project files
   - Check for changes every 5 seconds
   - Update only modified files
   - Show progress information
   - Continue running until you press Ctrl+C

3. **Integration with VS Code**:
   When used with Cline/Roo, the memory client automatically detects when you open a folder in VS Code and indexes the project files.

### Vectorization Progress

When indexing or updating project files, the memory client shows progress information:

```
Indexing project directory: /path/to/project
Found 120 files to index
Indexing progress: 25% (30/120 files)
Indexing progress: 50% (60/120 files)
Indexing progress: 75% (90/120 files)
Indexing progress: 100% (120/120 files)
Indexing complete: 120 files indexed
```

This helps you understand how the vectorization is progressing, especially for large projects.

## Verifying It's Working

To verify the memory client is working correctly:

1. **Start a new conversation** in Cline/Roo
2. **Have a brief conversation** with the AI
3. **Close Cline/Roo completely**
4. **Restart Cline/Roo**
5. **Ask about something from your previous conversation**

The AI should be able to recall information from your previous conversation, demonstrating that the memory client is working correctly.

## Qdrant Integration

The memory client works with Qdrant in a flexible way:

1. **Automatic Detection**: During installation, the scripts will:
   - Check if Qdrant is already running on the default port (6333)
   - Look for existing Qdrant containers running on any port
   - Use any existing Qdrant instance if found
   - Set up automatic startup only if no existing instance is found

2. **Multiple Qdrant Instances**: If you're already running Qdrant for other purposes:
   - The memory client will use your existing instance
   - No new instance will be started
   - Your existing data will be preserved

3. **Automatic Startup**: If no existing Qdrant is found:
   - A startup script is created to run Qdrant at system boot
   - Qdrant will be started automatically when needed
   - The memory client will connect to it seamlessly

## Troubleshooting

If you encounter any issues:

1. **Ensure Qdrant is running**:
   ```powershell
   # Windows (PowerShell)
   .\ensure-qdrant.ps1
   
   # Windows (Batch)
   ensure-qdrant.bat
   
   # Linux/macOS
   ./ensure-qdrant.sh
   ```
   
   These scripts will:
   - Check if Qdrant is running on any port
   - Use existing Qdrant instances if found
   - Start Qdrant using Docker only if no instance is found
   - Wait for Qdrant to become available
   - Provide status information

2. **Verify MCP settings**:
   Check that the `mcp_settings.json` file in your Cline/Roo configuration directory contains the correct path to the memory client executable.

3. **Run in verbose mode**:
   Edit the `mcp_settings.json` file to add the `--verbose` flag to the arguments array.

4. **Check logs**:
   Look for any error messages in the Cline/Roo logs.

## Advanced Configuration

You can customize the memory client by editing the `config.yaml` file:

```yaml
# Qdrant server URL
QDRANT_URL: "http://localhost:6333"

# Collection name for storing conversation memory
COLLECTION_NAME: "conversation_memory"

# Size of embedding vectors
EMBEDDING_SIZE: 384
```

This file should be placed in:
- Windows: `%APPDATA%\memory-client\config.yaml`
- Linux/macOS: `~/.config/memory-client/config.yaml`