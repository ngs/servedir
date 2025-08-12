.PHONY: build test clean run install lint fmt coverage help

# Variables
BINARY_NAME=servedir
GO=go
GOFLAGS=-v
MAIN_PATH=.

# Default target
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

test: ## Run tests
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out ./...

coverage: test ## Generate coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Remove build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GO) clean

run: ## Run the application (default port 8000)
	$(GO) run $(MAIN_PATH) .

install: ## Install the binary to GOPATH/bin
	$(GO) install $(GOFLAGS) $(MAIN_PATH)

lint: ## Run golangci-lint
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

fmt: ## Format code
	$(GO) fmt ./...
	$(GO) mod tidy

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

update-deps: ## Update dependencies
	$(GO) get -u ./...
	$(GO) mod tidy

release-dry-run: ## Test release with goreleaser
	@which goreleaser > /dev/null || (echo "goreleaser not found. Install it from https://goreleaser.com/install/" && exit 1)
	goreleaser release --snapshot --clean --skip=publish

# Development targets
dev: ## Run with hot reload (requires air)
	@which air > /dev/null || (echo "air not found. Install it with: go install github.com/air-verse/air@latest" && exit 1)
	air

benchmark: ## Run benchmarks
	$(GO) test -bench=. -benchmem ./...