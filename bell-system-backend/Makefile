.PHONY: dev build test clean

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
	go test ./... -v

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
