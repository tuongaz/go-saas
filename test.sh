#!/bin/bash

# Test script for gosaas-cli

echo "Testing gosaas-cli..."

# Create a temporary test directory
TEST_DIR=$(mktemp -d)
echo "Created test directory: $TEST_DIR"

# Navigate to the CLI directory and build
cd "$(dirname "$0")"
go build -o gosaas

# Move to test directory
cd "$TEST_DIR"

# Run the CLI tool
echo "Running gosaas new test-project..."
"$OLDPWD/gosaas" new test-project

# Check if the project was created successfully
if [ -d "$TEST_DIR/test-project" ]; then
    echo "✅ Project created successfully!"
    echo "Project structure:"
    find test-project -type f | sort
else
    echo "❌ Failed to create project!"
    exit 1
fi

# Clean up
echo "Cleaning up..."
rm -rf "$TEST_DIR"
echo "Test completed successfully!" 