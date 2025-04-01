#!/bin/bash

echo "Checking if Qdrant is running..."

# Try to connect to Qdrant
if curl -s http://localhost:6333/collections > /dev/null; then
    echo -e "\e[32mQdrant is already running!\e[0m"
    exit 0
else
    echo -e "\e[33mQdrant is not running. Attempting to start...\e[0m"
fi

# Check if another Qdrant instance might be running on a different port
if docker ps | grep -q qdrant; then
    echo -e "\e[33mFound existing Qdrant container:\e[0m"
    docker ps | grep qdrant
    echo -e "\e[32mUsing existing Qdrant instance instead of starting a new one.\e[0m"
    exit 0
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker to run Qdrant."
    echo "You can install Docker by following the instructions at: https://docs.docker.com/get-docker/"
    exit 1
fi

# Start Qdrant using Docker
echo "Starting Qdrant using Docker..."
docker run -d -p 6333:6333 -p 6334:6334 --name qdrant qdrant/qdrant

# Check if Qdrant started successfully
echo "Waiting for Qdrant to start..."
sleep 5
if curl -s http://localhost:6333/collections > /dev/null; then
    echo -e "\e[32mQdrant started successfully!\e[0m"
else
    echo -e "\e[31mFailed to start Qdrant.\e[0m"
    exit 1
fi

echo ""
echo "If Qdrant is running, the memory client will work seamlessly in the background."
echo "You don't need to manually start or manage it - Cline/Roo will handle everything."
echo ""