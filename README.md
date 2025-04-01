# Memory Client for MCP

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Qdrant](https://img.shields.io/badge/Qdrant-Vector%20DB-FF4F8B?style=flat-square)](https://qdrant.tech)
[![MCP](https://img.shields.io/badge/MCP-Protocol-4B32C3?style=flat-square)](https://github.com/roo-cline/mcp)
[![Version](https://img.shields.io/badge/Version-1.2.0-success?style=flat-square)](https://github.com/roo-cline/memory-client-go)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

> A Go-based memory client for the Model Context Protocol (MCP) that provides persistent conversation storage and project context using Qdrant vector database. This client enables Cline/Roo to maintain conversation history and project knowledge across sessions **automatically and seamlessly in the background**.

## 📋 Table of Contents

- [Features](#-features)
- [How It Works](#-how-it-works)
- [Installation](#-installation)
- [Project Memory](#-project-memory)
- [Usage](#-usage)
- [Advanced Usage Examples](#-advanced-usage-examples)
- [Data Management](#-data-management)
- [MCP API Reference](#-mcp-api-reference)
- [Configuration](#-configuration)
- [Author](#-author)
- [License](#-license)

## ✨ Features

<table>
  <tr>
    <td width="50%">
      <h3>🧠 Persistent Memory</h3>
      <ul>
        <li>Maintains conversation history across sessions</li>
        <li>Seamless background operation</li>
        <li>Automatic startup and shutdown</li>
        <li>Persists across system restarts</li>
      </ul>
    </td>
    <td width="50%">
      <h3>📁 Project Context</h3>
      WARNING: This will delete ALL data from Qdrant!
      This action cannot be undone.
      Are you sure you want to continue? (y/N):
      <ul>
        <li>Indexes your project files</li>
        <li>Tracks file changes automatically</li>
        <li>Provides code-aware assistance</li>
        <li>Excludes binary and media files</li>
      </ul>
    </td>
  </tr>
  <tr>
    <td width="50%">
      <h3>🔍 Semantic Search</h3>
      <ul>
        <li>Vector-based similarity search</li>
        <li>Find related conversations</li>
        <li>Search across project files</li>
        <li>Context-aware results</li>
      </ul>
    </td>
    <td width="50%">
      <h3>🔌 Seamless Integration</h3>
      <ul>
        <li>Full MCP protocol support</li>
        <li>Works with existing Qdrant instances</li>
        <li>Automatic VSCode integration</li>
        <li>Cross-platform support</li>
      </ul>
    </td>
  </tr>
</table>

## 🔄 How It Works

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│             │     │             │     │             │
│  Cline/Roo  │◄───►│ Memory MCP  │◄───►│   Qdrant    │
│             │     │   Server    │     │ Vector DB   │
└─────────────┘     └─────────────┘     └─────────────┘
       ▲                   ▲                   ▲
       │                   │                   │
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  User Chat  │     │Project Files│     │ Persistent  │
│   History   │     │   Context   │     │   Storage   │
└─────────────┘     └─────────────┘     └─────────────┘
```

Once installed, the memory client:

1. **Starts automatically** when Cline/Roo needs it (via MCP protocol)
2. **Runs silently** in the background without requiring user interaction
3. **Stores conversations** in the Qdrant vector database
4. **Indexes project files** when you open a folder in Visual Studio Code
5. **Provides context** to Cline/Roo when needed
6. **Persists across restarts** thanks to permanent installation

## 📥 Installation

### Quick Installation (Recommended)

Simply run the provided installation script:

<table>
<tr>
<th>Windows (PowerShell)</th>
<th>Windows (Batch)</th>
<th>Linux/macOS</th>
</tr>
<tr>
<td>

```powershell
.\install.ps1
```

</td>
<td>

```batch
install.bat
```

</td>
<td>

```bash
chmod +x install.sh
./install.sh
```

</td>
</tr>
</table>

The installation script will:
- Build the memory client
- Install it to a persistent location
- Configure it to work with Cline/Roo
- Set up Qdrant to run at startup
- Detect existing Qdrant instances

For detailed installation instructions, see [INSTALL.md](INSTALL.md).

## 📁 Project Memory

The memory client can index your project files to provide context-aware assistance:

<table>
<tr>
<th>Command</th>
<th>Description</th>
</tr>
<tr>
<td>

```bash
memory-client index-project
```

</td>
<td>Index all files in the current directory</td>
</tr>
<tr>
<td>

```bash
memory-client index-project --tag "project-name"
```

</td>
<td>Index files with a specific tag for better organization and retrieval</td>
</tr>
<tr>
<td>

```bash
memory-client update-project
```

</td>
<td>Update only modified files in the project</td>
</tr>
<tr>
<td>

```bash
memory-client watch-project
```

</td>
<td>Continuously monitor for file changes</td>
</tr>
<tr>
<td>

```bash
memory-client watch-project --tag "project-name"
```

</td>
<td>Watch a project directory with a specific tag</td>
</tr>
</table>

### Automatic Project Indexing

When used with Cline/Roo, the memory client automatically indexes your project files when you open a folder in Visual Studio Code. This allows Cline/Roo to:

- Understand your project structure
- Reference specific files and code
- Provide more contextually relevant assistance
- Remember project details across sessions

The project memory excludes binary files, media files, and other non-text content to focus on code and documentation.

### Project File Tagging

The memory client supports tagging project files during indexing, which helps organize and categorize your codebase:

- **Organize by Project**: Tag files with project names to separate different codebases
- **Categorize by Purpose**: Use tags like "frontend", "backend", "tests", etc.
- **Improve Retrieval**: Tags help the memory client find the most relevant files for your queries

Example usage:

```bash
# Index a repository with a specific tag
memory-client index-project --path /path/to/repo --tag "my-project"

# Watch a repository with a specific tag
memory-client watch-project --path /path/to/repo --tag "my-project"
```

Tags are stored in the vector database along with the file content, making it easier to retrieve related files later.

## 🚀 Usage

<table>
<tr>
<th>Command</th>
<th>Description</th>
</tr>
<tr>
<td>

```bash
memory-client mcp
```

</td>
<td>Start as an MCP server (used by Cline/Roo)</td>
</tr>
<tr>
<td>

```bash
memory-client dashboard
```

</td>
<td>Start the memory dashboard</td>
</tr>
<tr>
<td>

```bash
memory-client history
```

</td>
<td>View conversation history</td>
</tr>
<tr>
<td>

```bash
memory-client search "query"
```

</td>
<td>Search for messages or files</td>
</tr>
<tr>
<td>

```bash
memory-client add user "message"
```

</td>
<td>Add a message to conversation history</td>
</tr>
<tr>
<td>

```bash
memory-client version
```

</td>
<td>Display version information</td>
</tr>
</table>

## 🔍 Advanced Usage Examples

### Conversation Management

<table>
<tr>
<th>Task</th>
<th>Command</th>
<th>Description</th>
</tr>
<tr>
<td>Add a user message</td>
<td>

```bash
memory-client add user "How do I implement a binary search tree?"
```

</td>
<td>Adds a user message to the conversation history</td>
</tr>
<tr>
<td>Add an assistant message</td>
<td>

```bash
memory-client add assistant "Here's how to implement a binary search tree in Go..."
```

</td>
<td>Adds an assistant message to the conversation history</td>
</tr>
<tr>
<td>View recent conversation</td>
<td>

```bash
memory-client history --limit 10
```

</td>
<td>Shows the 10 most recent messages in the conversation</td>
</tr>
<tr>
<td>Search conversations</td>
<td>

```bash
memory-client search "binary search tree" --limit 5
```

</td>
<td>Finds up to 5 messages related to binary search trees</td>
</tr>
<tr>
<td>Tag conversations</td>
<td>

```bash
memory-client tag --query "golang error handling" --tags "golang,errors,best-practices"
```

</td>
<td>Tags messages about Golang error handling for easier retrieval</td>
</tr>
<tr>
<td>Get messages by tag</td>
<td>

```bash
memory-client get-by-tag "golang"
```

</td>
<td>Retrieves all messages tagged with "golang"</td>
</tr>
</table>

### Project Management

<table>
<tr>
<th>Task</th>
<th>Command</th>
<th>Description</th>
</tr>
<tr>
<td>Index a specific project</td>
<td>

```bash
memory-client index-project --project "/path/to/project" --tag "my-project"
```

</td>
<td>Indexes all files in the specified directory with a tag</td>
</tr>
<tr>
<td>Watch multiple projects</td>
<td>

```bash
# In separate terminals:
memory-client watch-project --project "/path/to/project1" --tag "project1"
memory-client watch-project --project "/path/to/project2" --tag "project2"
```

</td>
<td>Watches multiple projects simultaneously, each with its own tag</td>
</tr>
<tr>
<td>Search within a project</td>
<td>

```bash
memory-client search "database connection" --tag "my-project"
```

</td>
<td>Searches only within files tagged with "my-project"</td>
</tr>
<tr>
<td>Batch index with filters</td>
<td>

```bash
# Use the provided scripts for more advanced indexing:
./scripts/index-project.sh --project "/path/to/project" --tag "my-project" --max-file-size 2048
```

</td>
<td>Indexes a project with custom file size limits and batch processing</td>
</tr>
</table>

### System Management

<table>
<tr>
<th>Task</th>
<th>Command</th>
<th>Description</th>
</tr>
<tr>
<td>Add to PATH</td>
<td>

```bash
# Windows PowerShell:
./scripts/add-to-path.ps1

# Linux/macOS:
./scripts/add-to-path.sh
```

</td>
<td>Adds memory-client to your PATH for easier access from any terminal</td>
</tr>
<tr>
<td>Run as system service</td>
<td>

```bash
# Windows (as Administrator):
./scripts/install-mcp-service.ps1

# Linux (as root):
sudo ./scripts/install-mcp-service.sh

# macOS:
./scripts/install-mcp-service-mac.sh
```

</td>
<td>Installs memory-client as a system service that starts automatically</td>
</tr>
<tr>
<td>Check memory stats</td>
<td>

```bash
memory-client stats
```

</td>
<td>Shows statistics about memory usage, including vector counts and storage</td>
</tr>
<tr>
<td>Run dashboard</td>
<td>

```bash
memory-client dashboard --port 8082
```

</td>
<td>Runs the memory dashboard on a custom port</td>
</tr>
</table>

### Data Cleanup

<table>
<tr>
<th>Task</th>
<th>Command</th>
<th>Description</th>
</tr>
<tr>
<td>Clear recent messages</td>
<td>

```bash
memory-client clear day
```

</td>
<td>Deletes all messages from the current day</td>
</tr>
<tr>
<td>Clear older messages</td>
<td>

```bash
memory-client clear week
```

</td>
<td>Deletes all messages from the current week (Monday to now)</td>
</tr>
<tr>
<td>Delete specific message</td>
<td>

```bash
# First find the message ID:
memory-client search "query to find message"

# Then delete by ID:
memory-client delete-message "message-id-12345"
```

</td>
<td>Deletes a specific message by its ID</td>
</tr>
<tr>
<td>Reset everything</td>
<td>

```bash
memory-client purge
```

</td>
<td>Completely purges all data from Qdrant (requires confirmation)</td>
</tr>
</table>

## 🧹 Data Management

The memory client provides several commands to manage your data and maintain your Qdrant database:

<table>
<tr>
<th>Command</th>
<th>Description</th>
</tr>
<tr>
<td>

```bash
memory-client purge
```

</td>
<td>Completely purge all data from Qdrant (requires confirmation)</td>
</tr>
<tr>
<td>

```bash
memory-client clear day
```

</td>
<td>Delete all messages from the current day</td>
</tr>
<tr>
<td>

```bash
memory-client clear week
```

</td>
<td>Delete all messages from the current week (Monday to now)</td>
</tr>
<tr>
<td>

```bash
memory-client clear month
```

</td>
<td>Delete all messages from the current month (1st to now)</td>
</tr>
<tr>
<td>

```bash
memory-client clear range --from 2025-01-01 --to 2025-01-31
```

</td>
<td>Delete messages within a specific date range (YYYY-MM-DD format)</td>
</tr>
</table>

These commands help you manage your conversation history and maintain your database size. The `purge` command is useful for completely resetting your database, while the `clear` commands allow for more targeted data cleanup.

## 🔌 MCP API Reference

The Memory Client implements the Model Context Protocol (MCP) and exposes the following tools and resources to MCP clients:

### Tools

| Tool Name | Description | Required Parameters | Optional Parameters |
|-----------|-------------|---------------------|---------------------|
| `add_message` | Add a message to the conversation history | `role` (user/assistant/system), `content` | None |
| `get_conversation_history` | Retrieve the conversation history | None | `limit` |
| `search_similar_messages` | Search for messages similar to a query | `query` | `limit` |
| `index_project` | Index files in a project directory | `path` | `verbose` |
| `update_project` | Update modified files in a project directory | `path` | `verbose` |
| `search_project_files` | Search for files in the project | `query` | `limit` |
| `get_memory_stats` | Get statistics about memory usage | None | None |
| `delete_message` | Delete a message from the conversation history by ID | `id` | None |
| `delete_all_messages` | Delete all messages from the conversation history | None | None |
| `delete_project_file` | Delete a project file by path | `path` | None |
| `delete_all_project_files` | Delete all project files | None | None |
| `tag_messages` | Add tags to messages matching a query | `query`, `tags` | `limit` |
| `summarize_and_tag_messages` | Summarize and tag messages matching a query | `query`, `summary`, `tags` | `limit` |
| `get_messages_by_tag` | Retrieve messages with a specific tag | `tag` | `limit` |

### Resources

| Resource URI | Name | Description |
|--------------|------|-------------|
| `memory:///conversation_history` | Conversation History | Complete history of the conversation |
| `memory:///project_files` | Project Files | Source code and other files from the current project |

### Tool Examples

#### Adding a Message

```json
{
  "id": "request-123",
  "type": "tool_call",
  "data": {
    "name": "add_message",
    "arguments": {
      "role": "user",
      "content": "How do I implement a binary search tree in Go?"
    }
  }
}
```

#### Searching Project Files

```json
{
  "id": "request-456",
  "type": "tool_call",
  "data": {
    "name": "search_project_files",
    "arguments": {
      "query": "binary search tree implementation",
      "limit": 5
    }
  }
}
```

#### Tagging Messages

```json
{
  "id": "request-789",
  "type": "tool_call",
  "data": {
    "name": "tag_messages",
    "arguments": {
      "query": "binary search tree",
      "tags": ["data-structures", "algorithms", "go"],
      "limit": 10
    }
  }
}
```

### Resource Access Examples

#### Accessing Conversation History

```json
{
  "id": "request-abc",
  "type": "resource_access",
  "data": {
    "uri": "memory:///conversation_history"
  }
}
```

#### Accessing Project Files

```json
{
  "id": "request-def",
  "type": "resource_access",
  "data": {
    "uri": "memory:///project_files"
  }
}
```

## ⚙️ Configuration

The client can be configured using a `config.yaml` file:

```yaml
# Qdrant server URL
QDRANT_URL: "http://localhost:6333"

# Collection name for storing conversation memory
COLLECTION_NAME: "conversation_memory"

# Size of embedding vectors
EMBEDDING_SIZE: 384
```

Configuration locations:
- Windows: `%APPDATA%\memory-client\config.yaml`
- Linux/macOS: `~/.config/memory-client/config.yaml`

## MCP Service Management

The Memory Client MCP service provides persistent conversation storage for Windsurf IDE. Several scripts are available to help manage the service:

### Windows Scripts

- **restart-mcp-service.bat**: Stops, rebuilds, and restarts the MCP service with the latest code
- **check-mcp-status.bat**: Checks if the service is running and responding correctly
- **fix-mcp-service.bat**: Fixes issues with the service by checking version and reinstalling if necessary
- **install-mcp-service.bat**: Installs the MCP service as a Windows service (requires NSSM)
- **uninstall-mcp-service.bat**: Properly removes the MCP service

### Mac/Linux Scripts

- **check-mcp-status.sh**: Checks if the service is running and responding correctly
- **install-mcp-service.sh**: Installs the MCP service as a systemd service (Linux)
- **install-mcp-service-mac.sh**: Installs the MCP service as a launchd service (macOS)

See the [scripts/README.md](scripts/README.md) for detailed information on all available scripts.

### Service Troubleshooting

If the MCP service is not working correctly:

1. Check the service status: `scripts/check-mcp-status.bat` (Windows) or `scripts/check-mcp-status.sh` (Mac/Linux)
2. If the service is in a PAUSED state, use `scripts/fix-mcp-service.bat` to repair it
3. After making code changes, use `scripts/restart-mcp-service.bat` to rebuild and restart the service
4. Check the logs in the `logs` directory for error messages

## 👤 Author

**Christer Söderlund** - *Lead Developer*

## 📄 License

MIT