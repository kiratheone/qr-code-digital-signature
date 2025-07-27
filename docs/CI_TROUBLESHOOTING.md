# CI Troubleshooting Guide

This guide helps you debug failing backend and frontend tests in your CI pipeline.

## Quick Start

Test your changes locally by running the individual test commands:

```bash
# Backend tests
cd backend && go test -v ./...

# Frontend tests
cd frontend && npm test

# End-to-end tests
cd e2e && npm test

# Docker tests
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

## Common Issues and Solutions

### Backend Test Failures

#### 1. Database Connection Issues
**Symptoms:** Tests fail with database connection errors
**Solutions:**
- Ensure PostgreSQL is running: `pg_isready -h localhost -p 5432 -U postgres`
- Check if the test database exists: `PGPASSWORD=password psql -h localhost -U postgres -l`
- Create test database manually: `PGPASSWORD=password createdb -h localhost -U postgres digital_signature_test`

#### 2. Go Module Issues
**Symptoms:** `go mod download` fails or missing dependencies
**Solutions:**
```bash
cd backend
go mod tidy
go mod download
```

#### 3. Test Environment Variables
**Symptoms:** Tests fail due to missing environment variables
**Check these variables are set:**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=digital_signature_test
export DB_SSLMODE=disable
export PRIVATE_KEY=test-private-key
export PUBLIC_KEY=test-public-key
export JWT_SECRET=test-jwt-secret
```

#### 4. Race Condition Failures
**Symptoms:** Tests pass individually but fail with `-race` flag
**Solutions:**
- Run tests without race detection first: `go test -v ./...`
- Fix race conditions in code by using proper synchronization
- Check for shared global variables

### Frontend Test Failures

#### 1. Node.js/npm Issues
**Symptoms:** `npm ci` fails or version conflicts
**Solutions:**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

#### 2. Jest Configuration Issues
**Symptoms:** Jest tests fail to run or find test files
**Check:**
- `jest.config.js` exists and is properly configured
- `jest.setup.js` exists
- Test files follow naming convention (`*.test.js`, `*.spec.js`)

#### 3. TypeScript Compilation Errors
**Symptoms:** `npm run type-check` fails
**Solutions:**
```bash
cd frontend
npx tsc --noEmit --listFiles  # See what files are being checked
npm run type-check 2>&1 | head -20  # See first 20 errors
```

#### 4. ESLint Failures
**Symptoms:** `npm run lint` fails
**Solutions:**
```bash
cd frontend
npm run lint -- --fix  # Auto-fix issues
npm run lint -- --max-warnings 0  # Treat warnings as errors
```

#### 5. Build Failures
**Symptoms:** `npm run build` fails
**Common causes:**
- TypeScript errors
- Missing environment variables
- Import/export issues
- Next.js configuration problems

### Integration Test Failures

#### 1. Server Startup Issues
**Symptoms:** Backend server fails to start for integration tests
**Check:**
- Port 8000 is available: `lsof -i :8000`
- Database is accessible
- All required environment variables are set

#### 2. API Endpoint Issues
**Symptoms:** Integration tests can't reach API endpoints
**Solutions:**
```bash
# Test if server is responding
curl http://localhost:8000/health

# Check server logs
./main 2>&1 | tee server.log
```

### End-to-End Test Failures

#### 1. Frontend/Backend Communication
**Symptoms:** E2E tests fail due to API communication issues
**Check:**
- Both servers are running
- CORS is properly configured
- API URLs are correct (`NEXT_PUBLIC_API_URL=http://localhost:8000`)

#### 2. Database State Issues
**Symptoms:** E2E tests fail due to database state
**Solutions:**
- Reset database between test runs
- Use separate test database
- Implement proper test data cleanup

### Docker Build Failures

#### 1. Docker Image Build Issues
**Symptoms:** Docker build fails for backend or frontend
**Common solutions:**
```bash
# Build with verbose output
docker build --no-cache -t digital-signature-backend:test ./backend

# Check Dockerfile syntax
docker build --dry-run -t test ./backend
```

#### 2. Docker Compose Issues
**Symptoms:** `docker-compose config` fails
**Solutions:**
- Validate YAML syntax
- Check environment variable references
- Ensure all referenced files exist

## Debugging Commands

### Backend Debugging
```bash
# Run specific test package
cd backend
go test -v ./internal/domain/...

# Run with verbose output and no cache
go test -v -count=1 ./...

# Run benchmarks only
go test -bench=. -run=^$ ./...

# Check test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend Debugging
```bash
# Run tests in watch mode
cd frontend
npm run test:watch

# Run specific test file
npm test -- --testPathPattern=components

# Debug test with verbose output
npm test -- --verbose

# Check bundle analysis
npm run build
npx @next/bundle-analyzer
```

### Database Debugging
```bash
# Connect to test database
PGPASSWORD=password psql -h localhost -U postgres digital_signature_test

# Check database tables
\dt

# Reset test database
PGPASSWORD=password dropdb -h localhost -U postgres digital_signature_test
PGPASSWORD=password createdb -h localhost -U postgres digital_signature_test
```

## Performance Tips

1. **Parallel Testing:** Use `-p` flag for Go tests when safe
2. **Cache Dependencies:** Ensure npm and Go module caches are working
3. **Skip Heavy Tests:** Use build tags to skip slow tests during development
4. **Database Optimization:** Use in-memory database for unit tests

## Getting Help

If you're still having issues:

1. Run the simulation script with verbose output
2. Check the specific error messages
3. Look at the GitHub Actions logs for comparison
4. Test individual components in isolation

## Environment Setup Verification

Run this quick verification:

```bash
# Check all required tools
go version
node --version
npm --version
docker --version
psql --version

# Check database connectivity
pg_isready -h localhost -p 5432 -U postgres

# Check if ports are available
lsof -i :8000  # Backend port
lsof -i :3000  # Frontend port
lsof -i :5432  # PostgreSQL port
```