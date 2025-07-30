# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=digital-signature-system
BINARY_UNIX=$(BINARY_NAME)_unix

# Build directories (excluding standalone-tests)
BUILD_DIRS=./cmd/... ./internal/...

.PHONY: all build clean test deps run

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

test:
	$(GOTEST) -v $(BUILD_DIRS)

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out $(BUILD_DIRS)
	$(GOCMD) tool cover -html=coverage.out

# Check if code compiles without running tests
compile-check:
	$(GOCMD) build $(BUILD_DIRS)

deps:
	$(GOMOD) download
	$(GOMOD) tidy

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/main.go
	./$(BINARY_NAME)

# Standalone tests (run individually)
test-audit:
	$(GOCMD) run standalone-tests/test_audit_standalone.go

test-monitoring:
	$(GOCMD) run standalone-tests/test_audit_monitoring.go

test-e2e:
	$(GOCMD) run standalone-tests/test_e2e.go

test-load:
	$(GOCMD) run standalone-tests/test_load.go

# Linting and formatting
fmt:
	$(GOCMD) fmt $(BUILD_DIRS)

vet:
	$(GOCMD) vet $(BUILD_DIRS)

lint: fmt vet

# Development helpers
dev-build: deps lint build

dev-test: deps lint test

# Docker helpers
docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run -p 8000:8000 $(BINARY_NAME)