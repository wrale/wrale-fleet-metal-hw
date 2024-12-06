# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Binary output
BINARY_NAME=wrale-fleet-metal-hw
BUILD_DIR=build

# Test flags
TESTFLAGS=-v -race
COVERFLAGS=-coverprofile=coverage.out
SIMFLAGS=-tags=simulation
HWFLAGS=-tags=hardware

# Linting
GOLINT=golangci-lint
LINTFLAGS=run --timeout=5m

# Simulation environment
SIM_DIR=/tmp/wrale-sim

.PHONY: all test test-unit test-sim test-hw test-all clean lint fmt deps build coverage sim-start sim-stop help

all: clean lint test build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v

# Individual test targets
test-unit:
	$(GOTEST) $(TESTFLAGS) ./...

test-sim: sim-start
	$(GOTEST) $(TESTFLAGS) $(SIMFLAGS) ./...
	@make sim-stop

test-hw:
	$(GOTEST) $(TESTFLAGS) $(HWFLAGS) ./...

# Main test targets
test: test-unit test-sim  # CI-safe tests
	@echo "CI test suite completed successfully"

test-all: test test-hw  # All tests including hardware
	@echo "All tests completed successfully"

coverage:
	$(GOTEST) $(TESTFLAGS) $(COVERFLAGS) ./...
	go tool cover -html=coverage.out

lint:
	$(GOLINT) $(LINTFLAGS)

fmt:
	$(GOFMT) ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

deps:
	$(GOMOD) download
	$(GOMOD) tidy

sim-start:
	mkdir -p $(SIM_DIR)/{gpio,power,thermal,secure}
	@echo "Simulation environment created at $(SIM_DIR)"

sim-stop:
	rm -rf $(SIM_DIR)
	@echo "Simulation environment cleaned up"

verify: fmt lint test coverage

# Development tooling setup
dev-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

help:
	@echo "Make targets:"
	@echo " all          - Clean, lint, test, and build"
	@echo " build        - Build binary"
	@echo " test         - Run CI-safe tests (unit + simulation)"
	@echo " test-all     - Run ALL tests (including hardware tests)"
	@echo " test-unit    - Run unit tests only"
	@echo " test-sim     - Run simulation tests"
	@echo " test-hw      - Run hardware tests (requires physical hardware)"
	@echo " coverage     - Run tests with coverage report"
	@echo " lint         - Run linter"
	@echo " fmt          - Format code"
	@echo " clean        - Remove build artifacts"
	@echo " deps         - Download and tidy dependencies"
	@echo " sim-start    - Create simulation environment"
	@echo " sim-stop     - Clean up simulation environment"
	@echo " verify       - Format, lint, test, and generate coverage"
	@echo " dev-tools    - Install development tools"

# Default target
.DEFAULT_GOAL := help