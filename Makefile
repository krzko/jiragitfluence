.PHONY: build clean dev fmt help install lint run test

# Binary name
BINARY_NAME=jiragitfluence

# Main directory
CMD_DIR=./cmd/jiragitfluence

# Build directory
# BUILD_DIR=./build
BUILD_DIR=.

# Default target
.DEFAULT_GOAL := help

# Help target
help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	rm -rf ./tmp
	rm -f build-errors.log

dev: ## Run with hot-reload using air
	air

fmt: ## Format code
	go fmt ./...

install: ## Install dependencies
	go mod download
	go mod tidy

lint: ## Run linters
	golangci-lint run

run: ## Run the application
	go run $(CMD_DIR)

test: ## Run tests
	go test -v ./...

# Specific command targets
fetch: build ## Run fetch command
	$(BUILD_DIR)/$(BINARY_NAME) fetch

generate: build ## Run generate command
	$(BUILD_DIR)/$(BINARY_NAME) generate

publish: build ## Run publish command
	$(BUILD_DIR)/$(BINARY_NAME) publish
