@echo off
echo Starting Memory Client MCP Service...
start /b "" "go" run %~dp0\..\main.go mcp
echo MCP Service started on port 9580.
echo You can now start your editor.
