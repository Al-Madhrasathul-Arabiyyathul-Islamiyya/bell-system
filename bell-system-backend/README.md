# Bell System Backend

REST API server for the Arabiyya Bell Schedule System. Handles schedule management, audio file storage, user authentication, and real-time communication with desktop clients via WebSocket.

[![CI](https://github.com/arabiyya-edu/bell-system/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/arabiyya-edu/bell-system/actions/workflows/ci.yml)
[![Release](https://github.com/arabiyya-edu/bell-system/actions/workflows/release.yml/badge.svg)](https://github.com/arabiyya-edu/bell-system/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/arabiyya-edu/bell-system/branch/develop/graph/badge.svg)](https://codecov.io/gh/arabiyya-edu/bell-system)

## Prerequisites

- Go 1.26+
- SQL Server

## Quick Start

```bash
# Configure
cp config.example.toml config.toml  # edit with your DB credentials

# Run migrations
make migrate-up

# Start with hot reload
make dev

# Or build and run
make build
./bin/bell-schedule-system
```

## Project Structure

```
bell-system-backend/
├── cmd/
│   ├── server/          # Application entrypoint
│   └── migrate/         # Migration runner
├── config/              # Configuration loading (Viper/TOML)
├── internal/
│   ├── database/        # DB connection and repository implementations
│   ├── handlers/        # HTTP handlers and middleware
│   ├── helpers/         # Utility functions
│   ├── models/          # Domain models
│   └── router/          # Chi router setup
├── migrations/          # SQL migration files
├── pkg/
│   ├── errors/          # Custom error types
│   └── logger/          # Zap logger setup
├── scripts/             # Build, deploy, dev, test scripts
├── tests/
│   ├── handlers/        # Handler integration tests
│   ├── repositories/    # Repository tests (sqlmock)
│   ├── unit/            # Unit tests (config, errors, helpers, logger, models)
│   ├── mocks/           # Shared mock implementations
│   └── testutil/        # Shared test helpers
├── api/                 # API specifications
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Testing

Tests are organized by layer:

- **`tests/handlers/`** — HTTP handler tests with mock services
- **`tests/repositories/`** — Database repository tests with go-sqlmock
- **`tests/unit/`** — Pure unit tests for config, errors, helpers, logger, models
- **`tests/mocks/`** — Shared mock structs used across test packages

### Running Tests

```bash
# Run all tests
make test

# Run all tests with verbose output
go test ./... -v

# Run a specific test package
go test ./tests/handlers/... -v
go test ./tests/repositories/... -v
go test ./tests/unit/... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Full check (lint + tidy + test + build)
make check
```

## API Documentation

See [docs/api/](../docs/api/) for OpenAPI specifications.

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make dev` | Start with hot reload (air) |
| `make build` | Build production binary |
| `make test` | Run all tests |
| `make lint` | Check formatting and run go vet |
| `make check` | Full check: lint, tidy, test, build |
| `make clean` | Remove build artifacts |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback migrations |
| `make deps` | Tidy and vendor dependencies |
| `make setup-hooks` | Configure git hooks |

## Development

### Git Hooks

Git hooks live at the repo root (`scripts/hooks/`) and are shared across all components. They detect which files changed and only run relevant checks.

```bash
# From anywhere in the repo
./scripts/hooks/setup.sh    # Bash
./scripts/hooks/setup.ps1   # PowerShell

# Or from bell-system-backend/
make setup-hooks
```

**Pre-commit** (backend): formatting (`gofmt`), static analysis (`go vet`), module tidiness (`go mod tidy`).

**Pre-push** (backend): all tests pass, binary builds successfully.
