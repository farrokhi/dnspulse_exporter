# SPDX-License-Identifier: BSD-2-Clause
# Copyright (c) 2026 Babak Farrokhi

# Binary name
BINARY_NAME=dnspulse_exporter

# Build variables
VERSION?=$(shell grep -E 'version.*=.*"[0-9]' dnspulse_exporter.go | head -1 | sed 's/.*"\([^"]*\)".*/\1/')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Installation paths
PREFIX?=/usr/local
BINDIR?=$(PREFIX)/bin
SYSCONFDIR?=/etc
SYSTEMDDIR?=/etc/systemd/system

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m

.PHONY: all build test test-race test-coverage fmt vet lint clean install uninstall help

# Default target
all: fmt vet test build

## help: Display this help message
help:
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

##@ Development

## build: Build the binary
build:
	@echo "$(COLOR_YELLOW)Building $(BINARY_NAME)...$(COLOR_RESET)"
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

## build-all: Build for multiple platforms
build-all:
	@echo "$(COLOR_YELLOW)Building for multiple platforms...$(COLOR_RESET)"
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64
	GOOS=freebsd GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-freebsd-amd64
	@echo "$(COLOR_GREEN)Multi-platform build complete$(COLOR_RESET)"

## test: Run tests
test:
	@echo "$(COLOR_YELLOW)Running tests...$(COLOR_RESET)"
	$(GOTEST) -v ./...

## test-race: Run tests with race detector
test-race:
	@echo "$(COLOR_YELLOW)Running tests with race detector...$(COLOR_RESET)"
	$(GOTEST) -v -race ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(COLOR_YELLOW)Running tests with coverage...$(COLOR_RESET)"
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "$(COLOR_GREEN)Coverage report generated: coverage.html$(COLOR_RESET)"

## fmt: Format code
fmt:
	@echo "$(COLOR_YELLOW)Formatting code...$(COLOR_RESET)"
	@$(GOFMT) -s -w .
	@echo "$(COLOR_GREEN)Code formatted$(COLOR_RESET)"

## fmt-check: Check if code is formatted
fmt-check:
	@echo "$(COLOR_YELLOW)Checking code formatting...$(COLOR_RESET)"
	@if [ -n "$$($(GOFMT) -s -l .)" ]; then \
		echo "$(COLOR_BOLD)The following files are not formatted:$(COLOR_RESET)"; \
		$(GOFMT) -s -l .; \
		exit 1; \
	fi
	@echo "$(COLOR_GREEN)All files are properly formatted$(COLOR_RESET)"

## vet: Run go vet
vet:
	@echo "$(COLOR_YELLOW)Running go vet...$(COLOR_RESET)"
	$(GOVET) ./...
	@echo "$(COLOR_GREEN)go vet passed$(COLOR_RESET)"

## lint: Run staticcheck (install if needed)
lint:
	@echo "$(COLOR_YELLOW)Running staticcheck...$(COLOR_RESET)"
	@which staticcheck > /dev/null || (echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...
	@echo "$(COLOR_GREEN)Linting passed$(COLOR_RESET)"

## mod-tidy: Tidy go modules
mod-tidy:
	@echo "$(COLOR_YELLOW)Tidying go modules...$(COLOR_RESET)"
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)Modules tidied$(COLOR_RESET)"

## mod-verify: Verify dependencies
mod-verify:
	@echo "$(COLOR_YELLOW)Verifying dependencies...$(COLOR_RESET)"
	$(GOMOD) verify
	@echo "$(COLOR_GREEN)Dependencies verified$(COLOR_RESET)"

##@ Deployment

## install: Install binary and config files
install: build
	@echo "$(COLOR_YELLOW)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(BINARY_NAME) $(DESTDIR)$(BINDIR)/$(BINARY_NAME)
	install -d $(DESTDIR)$(SYSCONFDIR)
	install -m 644 dnspulse.yml $(DESTDIR)$(SYSCONFDIR)/dnspulse.yml
	@if [ -d systemd ]; then \
		install -d $(DESTDIR)$(SYSTEMDDIR); \
		install -m 644 systemd/dnspulse.service $(DESTDIR)$(SYSTEMDDIR)/dnspulse.service; \
		echo "$(COLOR_GREEN)Systemd service installed$(COLOR_RESET)"; \
	fi
	@echo "$(COLOR_GREEN)Installation complete$(COLOR_RESET)"

## uninstall: Uninstall binary and config files
uninstall:
	@echo "$(COLOR_YELLOW)Uninstalling $(BINARY_NAME)...$(COLOR_RESET)"
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY_NAME)
	rm -f $(DESTDIR)$(SYSTEMDDIR)/dnspulse.service
	@echo "$(COLOR_GREEN)Uninstallation complete$(COLOR_RESET)"
	@echo "Note: Config file at $(DESTDIR)$(SYSCONFDIR)/dnspulse.yml was not removed"

##@ Maintenance

## clean: Clean build artifacts
clean:
	@echo "$(COLOR_YELLOW)Cleaning...$(COLOR_RESET)"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.txt coverage.html
	@echo "$(COLOR_GREEN)Clean complete$(COLOR_RESET)"

## ci: Run all CI checks locally
ci: fmt-check vet test-race
	@echo "$(COLOR_GREEN)All CI checks passed$(COLOR_RESET)"
