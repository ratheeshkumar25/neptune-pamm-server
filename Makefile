# ─────────────────────────────────────────────────────────────────────────────
# Neptune-Pamm-Server — Makefile
# ─────────────────────────────────────────────────────────────────────────────

# Binary / build
BINARY      := neptune-pamm-server
MAIN        := ./cmd/main.go
BIN_DIR     := bin
BIN         := $(BIN_DIR)/$(BINARY)

# Docker
IMAGE       := neptune-pamm-server
TAG         := latest

# Version metadata (injected via ldflags)
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS     := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)

# Tools
GO          := go

# Load .env for local runs if present
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

.DEFAULT_GOAL := help

# ─── Help ────────────────────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

# ─── Dependencies ────────────────────────────────────────────────────────────
.PHONY: tidy
tidy: ## Sync and tidy go.mod / go.sum
	$(GO) mod tidy

.PHONY: deps
deps: ## Download modules
	$(GO) mod download

# ─── Build / Run ─────────────────────────────────────────────────────────────
.PHONY: build
build: ## Compile the binary into ./bin
	@mkdir -p $(BIN_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) $(MAIN)

.PHONY: run
run: ## Run the server locally
	$(GO) run $(MAIN)

.PHONY: clean
clean: ## Remove build artifacts
	@rm -rf $(BIN_DIR) coverage.out
	$(GO) clean

# ─── Quality ─────────────────────────────────────────────────────────────────
.PHONY: fmt
fmt: ## Format the code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint (must be installed)
	@command -v golangci-lint >/dev/null 2>&1 \
		&& golangci-lint run ./... \
		|| echo "golangci-lint not installed: https://golangci-lint.run/usage/install/"

.PHONY: test
test: ## Run tests
	$(GO) test ./... -race -count=1

.PHONY: cover
cover: ## Run tests with coverage report
	$(GO) test ./... -race -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

.PHONY: check
check: fmt vet lint test ## Run fmt, vet, lint and tests

# ─── Docker ──────────────────────────────────────────────────────────────────
.PHONY: docker
docker: ## Build the container image
	docker build -t $(IMAGE):$(TAG) --build-arg VERSION=$(VERSION) .

.PHONY: docker-run
docker-run: ## Run the container image with .env
	docker run --rm --env-file .env -p 8080:8080 $(IMAGE):$(TAG)
