.PHONY: help build install test lint fmt lint-only clean

# Docker image versions
GOLANGCI_LINT_VERSION := v2.6.2

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the terraform provider"
	@echo "  install   - Build and install the provider"
	@echo "  test      - Run tests"
	@echo "  lint      - Format code and run golangci-lint"
	@echo "  fmt       - Format code using golangci-lint"
	@echo "  lint-only - Run golangci-lint without formatting"
	@echo "  clean     - Clean build artifacts"

build:
	go build -o terraform-provider-garage

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hoose/garage/0.0.1/linux_amd64
	cp terraform-provider-garage ~/.terraform.d/plugins/registry.terraform.io/hoose/garage/0.0.1/linux_amd64/

test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./... || true

# Format code using golangci-lint formatters (faster than separate tools)
fmt:
	docker run --rm \
		-u "$(shell id -u):$(shell id -g)" \
		-e GOCACHE=/tmp/go-cache \
		-e GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache \
		-v "$(PWD):/app" \
		-v "$(HOME)/.cache:/home/cache" \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run --fix

# Run golangci-lint (formats first, then lints)
lint:
	docker run --rm \
		-u "$(shell id -u):$(shell id -g)" \
		-e GOCACHE=/tmp/go-cache \
		-e GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache \
		-v "$(PWD):/app" \
		-v "$(HOME)/.cache:/home/cache" \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run --fix

# Run only linting without formatting
lint-only:
	docker run --rm \
		-u "$(shell id -u):$(shell id -g)" \
		-e GOCACHE=/tmp/go-cache \
		-e GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache \
		-v "$(PWD):/app" \
		-v "$(HOME)/.cache:/home/cache" \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run

clean:
	go clean
	rm -f terraform-provider-garage coverage.txt


