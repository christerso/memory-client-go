# Memory Client Conversation Tagging and Categorization

This document explains how to use the conversation tagging and automatic categorization features of the Memory Client.

## Overview

The Memory Client now includes powerful features for tagging and categorizing conversations:

1. **Manual Tagging**: Explicitly tag conversations with custom labels
2. **Automatic Categorization**: Automatically analyze and categorize conversations based on content
3. **Tagging Modes**: Switch between automatic and manual tagging as needed
4. **Multi-Platform Integration**: Capture and tag conversations from VS Code, Windsurf, or any environment

## Installation

The tagging features are included in the standard Memory Client installation. Follow the installation instructions for your platform:

### Windows

```batch
# Run as Administrator
scripts\install-mcp-service-direct.bat
```

### Linux

```bash
# Run with sudo
sudo ./scripts/install-mcp-service.sh
```

## Conversation Tagging

### Setting a Conversation Tag

You can set a tag for the current conversation using any of these methods:

#### Command Line

```bash
# Set a conversation tag
memory-client tag -tag="project-planning"

# Get the current tag
memory-client tag -mode=get-tag
```

#### VS Code Extension

Use the VS Code extension commands:
- `Memory Client: Set Conversation Tag`
- `Memory Client: Get Current Tag`

#### Windsurf Integration

Click on the tag display in the bottom right corner of the Windsurf interface to set a tag.

### Tagging Modes

The Memory Client supports two tagging modes:

1. **Automatic Mode** (default): Messages are automatically analyzed and categorized
2. **Manual Mode**: Only explicitly set tags are applied to messages

#### Switching Between Modes

##### Command Line

```bash
# Switch to manual mode
memory-client tag-mode -mode=manual

# Switch back to automatic mode
memory-client tag-mode -mode=automatic

# Check current mode
memory-client tag-mode -mode=get-mode
```

##### Windsurf Integration

Click on the "Mode" indicator in the bottom right corner to toggle between automatic and manual modes.

## Automatic Categorization

When in automatic mode, the Memory Client analyzes conversations and applies category tags based on content:

### Default Categories

1. **Technical**: Code discussions, debugging, implementation details
   - Keywords: code, function, error, bug, implementation, class, variable, etc.

2. **Planning**: Project planning, roadmaps, timelines
   - Keywords: plan, schedule, timeline, milestone, feature, requirement, etc.

3. **Question**: General questions and inquiries
   - Keywords: how, what, why, when, question, help, etc.

4. **Feedback**: Feedback, reviews, suggestions
   - Keywords: feedback, review, suggestion, improve, better, etc.

### How It Works

1. Messages are buffered (5 by default) before analysis
2. When the buffer is full, messages are analyzed as a group
3. Appropriate category tags are applied based on keyword matching
4. Messages with the same conversation tag are grouped together

### Configuration

The tagging behavior can be configured in the config file:

```json
{
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
```

## Integration with IDEs

### VS Code Extension

The VS Code extension automatically captures conversations and sends them to the Memory Client. It adds commands to the command palette:

- `Memory Client: Capture Message`
- `Memory Client: Set Conversation Tag`
- `Memory Client: Get Current Tag`
- `Memory Client: Toggle Tagging Mode`

### Windsurf Integration

The Windsurf integration script adds UI elements to the Windsurf interface:

- **Tag Display**: Shows the current conversation tag (click to change)
- **Mode Display**: Shows the current tagging mode (click to toggle)
- **Status Indicator**: Shows the connection status to the Memory Client

To install the Windsurf integration:

1. Copy `scripts/windsurf-memory-integration.js` to your Windsurf scripts directory
2. Include the script in your Windsurf environment

## API Reference

The Memory Client exposes HTTP API endpoints for external clients:

- `POST /api/message`: Add a message to memory
  ```json
  {
    "role": "user",
    "content": "Message content"
  }
  ```

- `POST /api/set-conversation-tag`: Set the current conversation tag
  ```json
  {
    "tag": "project-planning"
  }
  ```

- `GET /api/get-conversation-tag`: Get the current conversation tag

- `POST /api/set-tagging-mode`: Set the tagging mode
  ```json
  {
    "mode": "automatic" // or "manual"
  }
  ```

- `GET /api/get-tagging-mode`: Get the current tagging mode

## Best Practices

1. **Use Descriptive Tags**: Choose clear, descriptive tags for conversations (e.g., "auth-feature-planning" instead of "meeting1")

2. **Switch to Manual Mode** when you want to maintain consistent tagging for a specific project or topic

3. **Switch to Automatic Mode** when you want the system to help categorize varied conversations

4. **Review Tags Periodically**: Check the applied tags to ensure they accurately represent your conversations

## Troubleshooting

### Common Issues

1. **Messages Not Being Tagged**:
   - Check if the Memory Client service is running
   - Verify the tagging mode (automatic vs. manual)
   - Ensure the buffer size hasn't been set too high

2. **Cannot Connect to Memory Client**:
   - Check if the service is running (`systemctl status memory-client-mcp` on Linux, Services app on Windows)
   - Verify the API port (default: 10010) is not blocked by a firewall
   - Check the logs for any errors

3. **Tags Not Appearing in Dashboard**:
   - Refresh the dashboard
   - Check if the messages were successfully added to memory
   - Verify the database connection in the config file

### Logs

Check the logs for more detailed information:

- **Windows**: Check the service logs in the Event Viewer
- **Linux**: Use `journalctl -u memory-client-mcp`
