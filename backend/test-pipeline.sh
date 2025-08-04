#!/bin/bash

# Simple Testing Pipeline for Digital Signature System
# Focuses on business logic testing with 70% coverage target

echo "=== Digital Signature System Test Pipeline ==="
echo "Running focused test suite for business-critical components..."

# Test business logic services (high priority)
echo "1. Testing business logic services..."
go test -v -cover ./internal/domain/services/...

# Test cryptographic operations (high priority)
echo "2. Testing cryptographic operations..."
go test -v -cover ./internal/infrastructure/crypto/...

# Test PDF processing (high priority)
echo "3. Testing PDF processing..."
go test -v -cover ./internal/infrastructure/pdf/...

# Test validation and logging (medium priority)
echo "4. Testing validation and logging..."
go test -v -cover ./internal/infrastructure/validation/...
go test -v -cover ./internal/infrastructure/logging/...

# Test database layer (medium priority)
echo "5. Testing database layer..."
go test -v -cover ./internal/infrastructure/database/...

echo "=== Test Pipeline Complete ==="
echo "Focus: Business logic, cryptographic operations, and document processing"
echo "Target: 70% coverage on critical components"