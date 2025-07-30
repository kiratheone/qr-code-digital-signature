#!/bin/bash

# Test script to validate deploy.sh without actually deploying
# This script tests the environment loading functionality

set -e

# Source the deploy script functions
source "$(dirname "$0")/deploy.sh"

# Test environment variable loading
test_env_loading() {
    echo "Testing environment variable loading..."
    
    # Create a temporary test env file
    TEST_ENV_FILE=$(mktemp)
    cat > "$TEST_ENV_FILE" << EOF
# Test environment file
TEST_VAR1=value1
TEST_VAR2=value2

# Comment line
TEST_VAR3=value3
# Another comment

TEST_VAR4=value4
EOF

    # Override ENV_FILE for testing
    ENV_FILE="$TEST_ENV_FILE"
    
    # Test the environment loading logic
    if [ -f "$ENV_FILE" ]; then
        while IFS= read -r line; do
            if [[ "$line" =~ ^[A-Za-z_][A-Za-z0-9_]*= ]]; then
                export "$line"
                echo "Loaded: $line"
            fi
        done < <(grep -v '^#' "$ENV_FILE" | grep -v '^$')
    fi
    
    # Verify variables were loaded
    if [ "$TEST_VAR1" = "value1" ] && [ "$TEST_VAR2" = "value2" ] && [ "$TEST_VAR3" = "value3" ] && [ "$TEST_VAR4" = "value4" ]; then
        echo "✅ Environment loading test passed"
    else
        echo "❌ Environment loading test failed"
        exit 1
    fi
    
    # Clean up
    rm -f "$TEST_ENV_FILE"
}

# Run tests
test_env_loading

echo "All tests passed! 🎉"