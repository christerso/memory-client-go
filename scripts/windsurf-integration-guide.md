# Windsurf Memory Integration Setup Guide

This guide will help you set up the Memory Client integration with Windsurf.

## Automatic Setup

The easiest way to set up the integration is to use the provided installation script:

1. Run the `update-mcp-service.bat` script as administrator
2. The script will:
   - Install the Memory Client MCP service
   - Copy the Windsurf integration scripts to the installation directory
   - Configure everything for you

## Manual Windsurf Integration

If you need to manually integrate with Windsurf, follow these steps:

### Method 1: Using the Browser Console

1. Open Windsurf in your browser
2. Open the browser's developer console (F12 or Ctrl+Shift+I)
3. Copy and paste the following code into the console:

```javascript
// Load the Windsurf Memory Integration script
const script = document.createElement('script');
script.src = 'C:/Program Files/MemoryClientMCP/windsurf-memory-integration.js';
document.head.appendChild(script);
```

### Method 2: Using Browser Extensions

You can use browser extensions like "User JavaScript and CSS" to automatically inject the script:

1. Install a user script manager extension like "Tampermonkey" or "User JavaScript and CSS"
2. Create a new script with the following content:

```javascript
// ==UserScript==
// @name         Windsurf Memory Integration
// @namespace    http://tampermonkey.net/
// @version      1.0
// @description  Integrate Windsurf with Memory Client
// @author       You
// @match        *://your-windsurf-url.com/*
// @grant        none
// ==/UserScript==

(function() {
    'use strict';
    
    // Load the integration script
    const script = document.createElement('script');
    script.src = 'C:/Program Files/MemoryClientMCP/windsurf-memory-integration.js';
    document.head.appendChild(script);
})();
```

### Method 3: Using the Windsurf Loader

1. Copy the `windsurf-loader.js` script to your Windsurf installation directory
2. Edit the script to set the correct path to the integration script
3. Load the loader script using one of the methods above

## Verifying the Integration

After setting up the integration, you should see:

1. A tag display in the bottom right corner of Windsurf
2. A mode display next to it
3. A status indicator showing the connection status

If you don't see these elements, check:

1. Browser console for errors
2. That the Memory Client MCP service is running
3. That the integration script path is correct

## Using the Integration

Once integrated, you can:

1. Type messages in Windsurf and they will be automatically captured
2. Click on the tag display to set a conversation tag
3. Click on the mode display to toggle between automatic and manual tagging

## Troubleshooting

If the integration is not working:

1. Check if the Memory Client MCP service is running
2. Verify the API is accessible at http://localhost:10010
3. Check browser console for JavaScript errors
4. Try reloading Windsurf after ensuring the service is running
