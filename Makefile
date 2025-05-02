# Basic variables
BINARY_NAME := core
VERSION ?= $(shell cat version)
BIN_DIR := $(CURDIR)/bin
TMP_BIN_DIR := /tmp/bin
MAIN_FILE := ./cmd/tracker
GO := go

# For prettier output
Q = @
M = $(shell printf "\033[34;1m->\033[0m")

# Build the binary
.PHONY: build
build: $(BIN_DIR) ; $(info $(M) building $(BINARY_NAME)...) ## Build the core binary
	$(Q)CGO_ENABLED=1 $(GO) build \
		-ldflags '-X main.Version=$(VERSION) -s -w' \
		-o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_FILE)

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
.PHONY: run
run: build ; $(info $(M) running...) ## Run the application
	$(Q)./$(BIN_DIR)/$(BINARY_NAME)

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

# Build specifically for air hot-reload
.PHONY: build-air
build-air: $(TMP_BIN_DIR) ; $(info $(M) building for air hot-reload...) @ ## Build for air hot-reload
	$Q CGO_ENABLED=0 $(GO) build \
		-ldflags '-X main.Version=$(VERSION) -s -w' \
		-o $(TMP_BIN_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Run with air for hot-reload
.PHONY: run/live
run/live: build-air ; $(info $(M) running with air hot-reload...) @ ## Run with air hot-reload
	$Q go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build-air" \
		--build.bin "$(TMP_BIN_DIR)/$(BINARY_NAME)" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true"

# Help target
.PHONY: help
help: ## Show this help
	$(Q)grep -hE '^[ a-zA-Z0-9_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
