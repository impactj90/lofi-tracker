# Basic variables
CLI_NAME := lofi-tracker
CLI_MAIN := ./cmd/tracker

DAEMON_NAME := lofi-daemon
DAEMON_MAIN := ./cmd/daemon

VERSION ?= $(shell cat version)
BIN_DIR := $(CURDIR)/bin
TMP_BIN_DIR := /tmp/bin
GO := go

# For prettier output
Q = @
M = $(shell printf "\033[34;1m->\033[0m")

.PHONY: build
build: build-cli build-daemon ## Build both CLI and daemon

# Build the binary
.PHONY: build-cli
build-cli: $(BIN_DIR) ; $(info $(M) building $(CLI_NAME)...) ## Build the CLI binary
	$(Q)CGO_ENABLED=1 $(GO) build \
		-ldflags '-X main.Version=$(VERSION) -s -w' \
		-o $(BIN_DIR)/$(CLI_NAME) $(CLI_MAIN)

# Build the binary
.PHONY: build-daemon
build-daemon: $(BIN_DIR) ; $(info $(M) building $(DAEMON_NAME)...) ## Build the daemon binary
	$(Q)CGO_ENABLED=1 $(GO) build \
		-ldflags '-X main.Version=$(VERSION) -s -w' \
		-o $(BIN_DIR)/$(DAEMON_NAME) $(DAEMON_MAIN)

# Create directories
$(BIN_DIR):
	$(Q)mkdir -p $@

$(TMP_BIN_DIR):
	$(Q)mkdir -p $@

# Clean build artifacts
.PHONY: clean
clean: ; $(info $(M) cleaning...) ## Clean up build artifacts
	$(Q)rm -rf $(BIN_DIR) $(TMP_BIN_DIR)

# Run the application
.PHONY: run-cli
run-cli: build-cli ; $(info $(M) running cli...) ## Run CLI Manually
	$(Q)./$(BIN_DIR)/$(CLI_NAME)

# Run the application
.PHONY: run-daemon
run-daemon: build-daemon ; $(info $(M) running daemon...) ## Run CLI Manually
	$(Q)./$(BIN_DIR)/$(DAEMON_NAME)


# Run tests
.PHONY: test
test: ; $(info $(M) running tests...) ## Run all tests with race detection
	$(Q)$(GO) test -race ./...

# Linting
.PHONY: lint
lint: lint-install ; $(info $(M) running linter...) ## Run golangci-lint
	$(Q)golangci-lint run --timeout 10m0s --sort-results --fix

.PHONY: lint-install
lint-install: ; $(info $(M) installing golangci-lint...) ## Install golangci-lint
	$(Q)$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

# Help target
.PHONY: help
help: ## Show this help
	$(Q)grep -hE '^[ a-zA-Z0-9_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
