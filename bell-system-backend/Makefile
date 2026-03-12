.PHONY: dev build test test-integration clean lint check setup-hooks

# Application settings
APP_NAME=bell-schedule-system
BUILD_DIR=./bin

# Development
dev:
	air

# Build for production
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

# Run tests
test:
	go test ./... -v -coverpkg=arabiyya.edu.mv/bell-system-backend/internal/...

# Run integration tests (requires Docker)
test-integration:
	go test -tags integration -v -count=1 -timeout 300s ./tests/integration/

# Clean built files
clean:
	rm -rf $(BUILD_DIR)
	rm -rf ./tmp

# Run database migrations
migrate-up:
	go run ./cmd/migrate

# Rollback migrations
migrate-down:
	go run ./cmd/migrate down

# Dependencies
deps:
	go mod tidy
	go mod vendor

# Setup git hooks (repo-wide, run from any directory)
setup-hooks:
	git config core.hooksPath scripts/hooks

# Lint: check formatting and run static analysis
lint:
	gofmt -l .
	go vet ./...

# Full check: lint, tidy, test, and build
check: lint
	go mod tidy
	go test ./... -count=1 -coverpkg=arabiyya.edu.mv/bell-system-backend/internal/...
	go build ./cmd/server
