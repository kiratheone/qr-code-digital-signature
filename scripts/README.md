# Scripts Directory

This directory contains utility scripts for the Digital Signature System project.

## Available Scripts

### `simulate-ci.sh`

A comprehensive CI simulation script that replicates your GitHub Actions workflow locally.

**Purpose:** Debug failing backend and frontend tests by running the complete CI pipeline on your local machine.

**Usage:**
```bash
# Run full CI simulation
./scripts/simulate-ci.sh

# Run specific test suites
./scripts/simulate-ci.sh --only-backend
./scripts/simulate-ci.sh --only-frontend

# Skip time-consuming tests
./scripts/simulate-ci.sh --skip-e2e
./scripts/simulate-ci.sh --skip-docker

# Skip dependency checks (if already verified)
./scripts/simulate-ci.sh --skip-deps
```

**Features:**
- ✅ Dependency verification (Go, Node.js, Docker, PostgreSQL)
- 🐘 Automatic PostgreSQL setup via Docker
- 🧪 Backend unit tests, benchmarks, and coverage
- ⚛️ Frontend linting, type checking, Jest tests, and builds
- 🔗 Integration tests
- 🌐 End-to-end tests with full server setup
- 🐳 Docker image builds and compose validation
- 🎨 Colored output with clear success/failure indicators

**Requirements:**
- Go 1.22+
- Node.js 20+
- Docker (optional, for PostgreSQL and image builds)
- PostgreSQL client tools (optional, will use Docker if not available)

### `setup-secrets.sh`

Script for setting up secrets and environment variables.

## Documentation

For detailed troubleshooting of CI failures, see [../docs/CI_TROUBLESHOOTING.md](../docs/CI_TROUBLESHOOTING.md).