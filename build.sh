#!/bin/bash
set -e

# Build the application
echo "Building Groupie Tracker..."
go build -o app cmd/srv/main.go

# Make executable
chmod +x app

echo "Build complete!"