# Memory Client for MCP

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Qdrant](https://img.shields.io/badge/Qdrant-Vector%20DB-FF4F8B?style=flat-square)](https://qdrant.tech)
[![MCP](https://img.shields.io/badge/MCP-Protocol-4B32C3?style=flat-square)](https://github.com/roo-cline/mcp)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

> A Go-based memory client for the Model Context Protocol (MCP) that provides persistent conversation storage and project context using Qdrant vector database. This client enables Cline/Roo to maintain conversation history and project knowledge across sessions **automatically and seamlessly in the background**.

## 📋 Table of Contents

- [Features](#-features)
- [How It Works](#-how-it-works)
- [Installation](#-installation)
- [Project Memory](#-project-memory)
- [Usage](#-usage)
- [MCP API Reference](#-mcp-api-reference)
- [Configuration](#-configuration)
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
memory-client search-project "query"
```

</td>
<td>Search for content in project files</td>
</tr>
</table>

### Automatic Project Indexing

When used with Cline/Roo, the memory client automatically indexes your project files when you open a folder in Visual Studio Code. This allows Cline/Roo to:

- Understand your project structure
- Reference specific files and code
- Provide more contextually relevant assistance
- Remember project details across sessions

The project memory excludes binary files, media files, and other non-text content to focus on code and documentation.

## 🛠️ Usage

### Command Line Interface

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
memory-client serve
```

</td>
<td>Start the memory server in daemon mode</td>
</tr>
<tr>
<td>

```bash
memory-client add user "message"
```

</td>
<td>Add a user message to memory</td>
</tr>
<tr>
<td>

```bash
memory-client history
```

</td>
<td>Show conversation history</td>
</tr>
<tr>
<td>

```bash
memory-client search "query"
```

</td>
<td>Search conversation memory</td>
</tr>
</table>

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

## 📄 License

MIT