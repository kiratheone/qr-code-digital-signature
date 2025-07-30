#!/bin/bash

# Simple test script to validate environment variable loading
# This script only tests the env loading logic without running deployment

set -e

echo "Testing environment variable loading..."

ENV_FILE=".env.prod"

# Test the environment loading logic
if [ -f "$ENV_FILE" ]; then
    echo "✅ Environment file found: $ENV_FILE"
    
    # Count total lines
    total_lines=$(wc -l < "$ENV_FILE")
    echo "📄 Total lines in env file: $total_lines"
    
    # Count comment lines
    comment_lines=$(grep -c '^#' "$ENV_FILE" || true)
    echo "💬 Comment lines: $comment_lines"
    
    # Count empty lines
    empty_lines=$(grep -c '^$' "$ENV_FILE" || true)
    echo "📝 Empty lines: $empty_lines"
    
    # Count valid environment variable lines
    valid_lines=$(grep -v '^#' "$ENV_FILE" | grep -v '^$' | grep -c '=' || true)
    echo "✅ Valid environment variables: $valid_lines"
    
    # Test the filtering logic
    echo ""
    echo "Testing environment variable filtering..."
    
    # Show first 5 valid environment variables that would be loaded
    echo "First 5 variables that would be loaded:"
    grep -v '^#' "$ENV_FILE" | grep -v '^$' | head -5 | while IFS= read -r line; do
        if [[ "$line" =~ ^[A-Za-z_][A-Za-z0-9_]*= ]]; then
            echo "  ✅ $line"
        else
            echo "  ❌ Invalid format: $line"
        fi
    done
    
    echo ""
    echo "✅ Environment variable loading test completed successfully!"
    
else
    echo "❌ Environment file $ENV_FILE not found"
    exit 1
fi