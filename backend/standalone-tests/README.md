# Standalone Tests

This directory contains standalone test files that have their own `main` functions and are designed to be run individually, not as part of the regular test suite.

## Available Tests

- `test_audit_standalone.go` - Standalone audit service testing
- `test_audit_monitoring.go` - Audit and monitoring integration tests  
- `test_e2e.go` - End-to-end testing scenarios
- `test_load.go` - Load testing utilities

## Running Standalone Tests

Use the Makefile commands from the backend directory:

```bash
# Run individual standalone tests
make test-audit
make test-monitoring  
make test-e2e
make test-load
```

Or run them directly:

```bash
go run standalone-tests/test_audit_standalone.go
go run standalone-tests/test_audit_monitoring.go
go run standalone-tests/test_e2e.go
go run standalone-tests/test_load.go
```

## Note

These files are excluded from the regular build process (`go build ./...`) to avoid conflicts with multiple `main` function declarations.