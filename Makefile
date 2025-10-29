# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOTOOL=$(GOCMD) tool
GOLIST=$(GOCMD) list
GOPATH=$(shell $(GOCMD) env GOPATH)
GOAMD64?=v3
GO_VERSION=$(shell go version | cut -d' ' -f3)

# Build variables
APP_BINARY_NAME=release
APP_BUILD_DIR=dist
APP_MAIN_FILE=./cmd/release/main.go
VERSION?=$(shell git describe --tags --dirty --always 2>/dev/null || echo "v0.0.0-dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)"
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Test variables
TEST_COVERAGE_DIR=coverage
TEST_COVERAGE_FILE=$(TEST_COVERAGE_DIR)/coverage.out
TEST_COVERAGE_HTML=$(TEST_COVERAGE_DIR)/coverage.html

# Default target
.PHONY: all
all: clean qa build

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_BINARY_NAME)..."
	@mkdir -p $(APP_BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(APP_BUILD_DIR)/$(APP_BINARY_NAME) $(APP_MAIN_FILE)
	@echo "Build completed: $(APP_BUILD_DIR)/$(APP_BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all:
	@echo "Building $(APP_BINARY_NAME) for multiple platforms..."
	@mkdir -p $(APP_BUILD_DIR)
	$(eval PLATFORMS_LIST = $(subst $(comma), ,$(PLATFORMS)))
	@for platform in $(PLATFORMS_LIST); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		output_name=$(APP_BUILD_DIR)/$(APP_BINARY_NAME); \
		if [ $$os = "windows" ]; then output_name=$$output_name.exe; fi; \
		output_name=$$output_name-$$os-$$arch; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $$output_name $(APP_MAIN_FILE) || exit 1; \
	done
	@echo "All builds completed"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	@rm -rf $(APP_BUILD_DIR)
	@rm -rf $(TEST_COVERAGE_DIR)
	@echo "Clean complete"

# Run tests with coverage
.PHONY: test
test:
	@echo "Running tests with coverage..."
	@mkdir -p $(TEST_COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(TEST_COVERAGE_FILE) -covermode=atomic ./...
	@$(GOCMD) tool cover -html=$(TEST_COVERAGE_FILE) -o $(TEST_COVERAGE_HTML)
	@$(GOCMD) tool cover -func=$(TEST_COVERAGE_FILE)
	@echo "Tests completed. Coverage report: $(TEST_COVERAGE_HTML)"

# Generate test report for CI
.PHONY: test-ci
test-ci:
	@echo "Running tests for CI..."
	@mkdir -p $(TEST_COVERAGE_DIR)
	@$(GOTEST) -v -race -coverprofile=$(TEST_COVERAGE_FILE) -covermode=atomic ./... \
		2>&1 | tee test-output.log
	@$(GOCMD) tool cover -func=$(TEST_COVERAGE_FILE) | grep "total:" | awk '{print "coverage: " $$3 " of statements"}'

# Install dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

# Tidy go.mod and go.sum
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Dependencies tidied"

# Install the application to GOPATH/bin
.PHONY: install
install: build
	@echo "Installing $(APP_BINARY_NAME)..."
	@cp $(APP_BUILD_DIR)/$(APP_BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation complete: $(GOPATH)/bin/$(APP_BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Running $(APP_BINARY_NAME)..."
	@$(APP_BUILD_DIR)/$(APP_BINARY_NAME)

# Install development tools
.PHONY: tools-install
tools-install:
	@echo "Installing development tools..."
	@echo "Installing golangci-lint"
	@$(GOCMD) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

	@echo "Installing gofumpt"
	@$(GOCMD) install mvdan.cc/gofumpt@latest

	@echo "Installing fieldalignment"
	@$(GOCMD) install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

	@echo "Installing gci"
	@$(GOCMD) install github.com/daixiang0/gci@latest

	@echo "Installing govulncheck"
	@$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest

	@echo "All tools installed successfully!"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...
	@gofumpt -l -w . 2>/dev/null || true
	@gci write --skip-generated -s standard -s default -s "prefix(github.com/boynoiz/release-tool)" --skip-vendor ./... 2>/dev/null || true
	@echo "Formatting completed"

# Lint and fix code
.PHONY: lint
lint:
	@echo "Linting code..."
	@golangci-lint run | true
	@echo "Linting completed"

# Lint and fix code
.PHONY: lint-fix
lint-fix:
	@echo "Linting and fixing code..."
	@golangci-lint run --fix | true
	@echo "Linting and fixing completed"

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	@$(GOCMD) vet ./... | true
	@echo "Vetting completed"

# Check vulnerabilities
.PHONY: vuln-check
vuln-check:
	@echo "Go code vulnerabilities check..."
	@govulncheck ./... | true
	@echo "Go code vulnerabilities check completed"

# Fix struct field alignment
.PHONY: field-align
field-align:
	@echo "Fixing struct field alignment..."
	@fieldalignment -fix ./... | true
	@echo "Field alignment completed"

# Complete quality assurance
.PHONY: qa
qa: tidy fmt lint-fix vet vuln-check field-align test

# CI-specific quality checks (no fixing)
.PHONY: qa-ci
qa-ci: fmt-check lint vet vuln-check test-ci

# Check formatting (for CI)
.PHONY: fmt-check
.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l . | grep -v '.cache/' | grep -v 'vendor/')" ]; then \
		echo "Code is not formatted. Run 'make fmt' to fix."; \
		gofmt -l . | grep -v '.cache/' | grep -v 'vendor/'; \
		exit 1; \
	fi
	@echo "Code formatting is correct"

# Check git-chglog config
.PHONY: chglog-check
chglog-check:
	@if [ ! -d ".chglog" ]; then \
		echo "Error: .chglog directory not found. Run 'git-chglog --init' first."; \
		exit 1; \
	else \
		echo "✓ Git-chglog configuration found"; \
		ls -la .chglog/; \
	fi

# Generate changelog (standalone command)
.PHONY: changelog
changelog:
	@echo "Generating changelog..."
	@if command -v git-chglog > /dev/null 2>&1; then \
		git-chglog --output CHANGELOG.md; \
		echo "✓ Changelog updated: CHANGELOG.md"; \
	else \
		echo "git-chglog not found. Install it first:"; \
		echo "  brew install git-chglog  # or"; \
		echo "  go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest"; \
	fi

# Check if we're on main branch
.PHONY: check-main-branch
check-main-branch:
	@current_branch=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$current_branch" != "main" ]; then \
		echo "Error: You must be on the main branch to create a release"; \
		echo "Current branch: $$current_branch"; \
		echo "Switch to main: git checkout main"; \
		exit 1; \
	fi; \
	echo "✓ On main branch"

# Check if working directory is clean
.PHONY: check-clean-working-dir
check-clean-working-dir:
	@if ! git diff-index --quiet HEAD --; then \
		echo "Error: Working directory is not clean"; \
		echo "Please commit or stash your changes first"; \
		git status --short; \
		exit 1; \
	fi; \
	echo "✓ Working directory is clean"

# Check if git-chglog is available
.PHONY: check-git-chglog
check-git-chglog:
	@if ! command -v git-chglog > /dev/null 2>&1; then \
		echo "Error: git-chglog not found"; \
		echo "Install it first:"; \
		echo "  brew install git-chglog  # or"; \
		echo "  go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest"; \
		exit 1; \
	fi; \
	echo "✓ git-chglog is available"

# Get the latest tag or default to v0.0.0 if no tags exist
.PHONY: get-latest-tag
get-latest-tag:
	@latest_tag=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "$$latest_tag"

# Calculate next version based on release type
# Calculate next version based on release type
.PHONY: calculate-next-version
calculate-next-version:
	@if [ -z "$(RELEASE_TYPE)" ]; then \
		echo "$(RED)Error: RELEASE_TYPE not specified$(NC)"; \
		exit 1; \
	fi; \
	latest_tag=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "$(BLUE)Current version: $$latest_tag$(NC)" >&2; \
	\
	version_number=$$(echo $$latest_tag | sed 's/^v//'); \
	if [ "$$version_number" = "0.0.0" ] && [ ! $$(git tag | wc -l) -gt 0 ]; then \
		if [ "$(RELEASE_TYPE)" = "major" ]; then \
			next_version="v1.0.0"; \
		elif [ "$(RELEASE_TYPE)" = "minor" ]; then \
			next_version="v0.1.0"; \
		else \
			next_version="v0.0.1"; \
		fi; \
	else \
		major=$$(echo $$version_number | cut -d. -f1); \
		minor=$$(echo $$version_number | cut -d. -f2); \
		patch=$$(echo $$version_number | cut -d. -f3); \
		\
		if [ "$(RELEASE_TYPE)" = "major" ]; then \
			major=$$((major + 1)); \
			minor=0; \
			patch=0; \
		elif [ "$(RELEASE_TYPE)" = "minor" ]; then \
			minor=$$((minor + 1)); \
			patch=0; \
		else \
			patch=$$((patch + 1)); \
		fi; \
		next_version="v$$major.$$minor.$$patch"; \
	fi; \
	echo "$(GREEN)Next version: $$next_version$(NC)" >&2; \
	echo "$$next_version"

# Generate changelog with next version
.PHONY: generate-changelog-with-version
generate-changelog-with-version:
	@if [ -z "$(NEXT_VERSION)" ]; then \
		echo "Error: NEXT_VERSION not specified"; \
		exit 1; \
	fi; \
	echo "Generating changelog for $(NEXT_VERSION)..."; \
	git-chglog --next-tag $(NEXT_VERSION) -o CHANGELOG.md; \
	echo "✓ Changelog generated"; \
	echo ""; \
	echo "=== Changelog Preview ==="; \
	head -30 CHANGELOG.md; \
	echo "========================="

# Ask user for confirmation
.PHONY: confirm-release
confirm-release:
	@echo ""; \
	echo "Release Summary:"; \
	echo "  Version: $(NEXT_VERSION)"; \
	echo "  Type: $(RELEASE_TYPE)"; \
	echo "  Branch: main"; \
	echo ""; \
	echo "This will:"; \
	echo "  1. Commit the updated CHANGELOG.md"; \
	echo "  2. Create and push tag $(NEXT_VERSION)"; \
	echo "  3. Push changes to origin/main"; \
	echo ""; \
	read -p "Do you want to proceed? [y/N]: " confirm; \
	if [ "$$confirm" != "y" ] && [ "$$confirm" != "Y" ]; then \
		echo "Release cancelled"; \
		exit 1; \
	fi

# Commit and tag the release
.PHONY: commit-and-tag
commit-and-tag:
	@echo "Committing changelog..."; \
	git add CHANGELOG.md; \
	git commit -m "chore: update changelog for $(NEXT_VERSION)"; \
	echo "Creating tag $(NEXT_VERSION)..."; \
	git tag $(NEXT_VERSION); \
	echo "Pushing to origin..."; \
	git push origin main; \
	git push origin $(NEXT_VERSION); \
	echo "✅ Release $(NEXT_VERSION) completed successfully!"; \
	echo "GitLab CI/CD pipeline should now be running..."

# Main release targets
.PHONY: release-major
release-major: RELEASE_TYPE := major
release-major: check-main-branch check-clean-working-dir check-git-chglog chglog-check
	@echo "$(BLUE)Starting MAJOR release...$(NC)"
	@NEXT_VERSION=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=major 2>/dev/null | tail -1); \
	$(MAKE) --no-print-directory generate-changelog-with-version NEXT_VERSION=$$NEXT_VERSION; \
	$(MAKE) --no-print-directory confirm-release NEXT_VERSION=$$NEXT_VERSION RELEASE_TYPE=major; \
	$(MAKE) --no-print-directory commit-and-tag NEXT_VERSION=$$NEXT_VERSION

.PHONY: release-minor
release-minor: RELEASE_TYPE := minor
release-minor: check-main-branch check-clean-working-dir check-git-chglog chglog-check
	@echo "$(BLUE)Starting MINOR release...$(NC)"
	@NEXT_VERSION=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=minor 2>/dev/null | tail -1); \
	$(MAKE) --no-print-directory generate-changelog-with-version NEXT_VERSION=$$NEXT_VERSION; \
	$(MAKE) --no-print-directory confirm-release NEXT_VERSION=$$NEXT_VERSION RELEASE_TYPE=minor; \
	$(MAKE) --no-print-directory commit-and-tag NEXT_VERSION=$$NEXT_VERSION

.PHONY: release-fix release
release-fix release: RELEASE_TYPE := patch
release-fix release: check-main-branch check-clean-working-dir check-git-chglog chglog-check
	@echo "$(BLUE)Starting PATCH release...$(NC)"
	@NEXT_VERSION=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=patch 2>/dev/null | tail -1); \
	$(MAKE) --no-print-directory generate-changelog-with-version NEXT_VERSION=$$NEXT_VERSION; \
	$(MAKE) --no-print-directory confirm-release NEXT_VERSION=$$NEXT_VERSION RELEASE_TYPE=patch; \
	$(MAKE) --no-print-directory commit-and-tag NEXT_VERSION=$$NEXT_VERSION

# Dry run commands - Fixed version
.PHONY: release-major-dry release-minor-dry release-fix-dry release-dry
release-major-dry: RELEASE_TYPE := major
release-major-dry: check-main-branch check-git-chglog chglog-check
	@echo "$(YELLOW)DRY RUN: MAJOR release$(NC)"
	@NEXT_VERSION=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=major 2>/dev/null | tail -1); \
	echo "$(GREEN)Would create version: $$NEXT_VERSION$(NC)"; \
	$(MAKE) --no-print-directory generate-changelog-with-version NEXT_VERSION=$$NEXT_VERSION; \
	echo "$(YELLOW)This is a dry run - no changes were made$(NC)"

release-minor-dry: RELEASE_TYPE := minor
release-minor-dry: check-main-branch check-git-chglog chglog-check
	@echo "$(YELLOW)DRY RUN: MINOR release$(NC)"
	@NEXT_VERSION=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=minor 2>/dev/null | tail -1); \
	echo "$(GREEN)Would create version: $$NEXT_VERSION$(NC)"; \
	$(MAKE) --no-print-directory generate-changelog-with-version NEXT_VERSION=$$NEXT_VERSION; \
	echo "$(YELLOW)This is a dry run - no changes were made$(NC)"

release-fix-dry release-dry: RELEASE_TYPE := patch
release-fix-dry release-dry: check-main-branch check-git-chglog chglog-check
	@echo "$(YELLOW)DRY RUN: PATCH release$(NC)"
	@NEXT_VERSION=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=patch 2>/dev/null | tail -1); \
	echo "$(GREEN)Would create version: $$NEXT_VERSION$(NC)"; \
	$(MAKE) --no-print-directory generate-changelog-with-version NEXT_VERSION=$$NEXT_VERSION; \
	echo "$(YELLOW)This is a dry run - no changes were made$(NC)"

# Show current version and what the next versions would be - Fixed version
.PHONY: version-info
version-info: check-git-chglog
	@echo "$(BLUE)Version Information:$(NC)"; \
	latest_tag=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "$(GREEN)Current version: $$latest_tag$(NC)"; \
	echo ""; \
	echo "$(YELLOW)Next versions would be:$(NC)"; \
	next_patch=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=patch 2>/dev/null | tail -1); \
	next_minor=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=minor 2>/dev/null | tail -1); \
	next_major=$$($(MAKE) --no-print-directory calculate-next-version RELEASE_TYPE=major 2>/dev/null | tail -1); \
	echo "  Patch (fix): $$next_patch"; \
	echo "  Minor:       $$next_minor"; \
	echo "  Major:       $$next_major"

# Docker build (optional)
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t release-tool:$(VERSION) .
	@echo "Docker image built: release-tool:$(VERSION)"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all            - Clean, QA, and build the application"
	@echo "  build          - Build the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests with coverage"
	@echo "  test-ci        - Run tests for CI (with coverage output)"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  deps           - Download dependencies"
	@echo "  tidy           - Tidy go.mod and go.sum"
	@echo "  install        - Install to GOPATH/bin"
	@echo "  run            - Run the application"
	@echo ""
	@echo "Development tools:"
	@echo "  tools-install  - Install development tools"
	@echo "  fmt            - Format code"
	@echo "  fmt-check      - Check code formatting (CI)"
	@echo "  lint           - Lint code"
	@echo "  lint-fix       - Lint and fix code"
	@echo "  vet            - Vet code"
	@echo "  vuln-check     - Check for vulnerabilities"
	@echo "  field-align    - Fix struct field alignment"
	@echo "  qa             - Run all quality checks (with fixes)"
	@echo "  qa-ci          - Run all quality checks (CI mode)"
	@echo ""
	@echo "Release Management Commands:"
	@echo ""
	@echo "Main Commands:"
	@echo "  make release-major    - Create a major release (x.0.0)"
	@echo "  make release-minor    - Create a minor release (x.y.0)"
	@echo "  make release-fix      - Create a patch release (x.y.z)"
	@echo "  make release          - Alias for release-fix"
	@echo ""
	@echo "Dry Run Commands:"
	@echo "  make release-major-dry - Preview major release"
	@echo "  make release-minor-dry - Preview minor release"
	@echo "  make release-fix-dry   - Preview patch release"
	@echo "  make release-dry       - Alias for release-fix-dry"
	@echo ""
	@echo "Info Commands:"
	@echo "  make version-info      - Show current and next version info"
	@echo "  make chglog-check      - Check git-chglog configuration"
	@echo ""
	@echo "Requirements:"
	@echo "  - Must be on main branch"
	@echo "  - Working directory must be clean"
	@echo "  - git-chglog must be installed"
	@echo "  - .chglog configuration must exist"
	@echo "Docker:"
	@echo "  docker-build   - Build Docker image"

# Make sure these aren't treated as files
.PHONY: all build clean test test-ci build-all deps tidy install run tools-install fmt fmt-check lint lint-fix vet vuln-check field-align qa qa-ci changelog chglog-check docker-build help
