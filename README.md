# Arabiyya Bell Schedule System

An automated school bell schedule and media playback system for Arabiyya School. Manages bell schedules, plays audio files on cue, and provides administrative control through a web interface and desktop client.

[![CI](https://github.com/Al-Madhrasathul-Arabiyyathul-Islamiyya/bell-system/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/Al-Madhrasathul-Arabiyyathul-Islamiyya/bell-system/actions/workflows/ci.yml)
[![Release](https://github.com/Al-Madhrasathul-Arabiyyathul-Islamiyya/bell-system/actions/workflows/release.yml/badge.svg)](https://github.com/Al-Madhrasathul-Arabiyyathul-Islamiyya/bell-system/actions/workflows/release.yml)
[![codecov](https://codecov.io/github/Al-Madhrasathul-Arabiyyathul-Islamiyya/bell-system/graph/badge.svg?token=0G7WAL0STC)](https://codecov.io/github/Al-Madhrasathul-Arabiyyathul-Islamiyya/bell-system)

## Architecture

This is a monorepo containing three components:

| Component | Tech | Description |
|-----------|------|-------------|
| [bell-system-backend](bell-system-backend/) | Go, Chi, SQL Server | REST API, scheduler, WebSocket server |
| [admin-client](admin-client/) | Vue 3, TypeScript, Vite, pnpm | Admin web interface for schedule management |
| [desktop-client](desktop-client/) | .NET (WPF) | Desktop app for audio playback and display |

## Tech Stack

- **Backend**: Go 1.26+, go-chi/chi v5, spf13/viper (TOML config), go.uber.org/zap
- **Database**: SQL Server (microsoft/go-mssqldb)
- **Auth**: JWT (180min expiry)
- **Admin Client**: Vue 3, TypeScript, Vite, pnpm
- **Desktop Client**: .NET, WPF

## Getting Started

### Prerequisites

- Go 1.26+
- SQL Server
- Node.js 22+ and pnpm (for admin client)
- .NET 9+ SDK (for desktop client)

### Clone

```bash
git clone https://github.com/Al-Madhrasathul-Arabiyyathul-Islamiyya/bell-system.git
cd bell-system
```

### Backend

```bash
cd bell-system-backend
cp config.example.toml config.toml  # edit with your DB credentials
make dev                              # start with hot reload
```

See [bell-system-backend/README.md](bell-system-backend/README.md) for full setup instructions.

## Project Structure

```
bell-system/
├── .github/workflows/    # CI/CD pipelines
├── scripts/hooks/        # Git hooks (repo-wide)
├── bell-system-backend/  # Go REST API + scheduler
├── admin-client/         # Vue admin dashboard
├── desktop-client/       # .NET WPF desktop app
└── docs/                 # Shared documentation
    ├── api/              # OpenAPI specs
    ├── architecture/     # System design docs
    ├── database/         # Schema and migrations
    └── design-assets/    # UI mockups and assets
```

## Development

### Git Workflow

- Branch from `develop` for all features and fixes
- PRs to `develop` run lint + test + build checks
- PRs to `master` additionally require 90% test coverage
- Use signed commits (`git commit -S`)

### Git Hooks Setup

```bash
# From anywhere in the repo
./scripts/hooks/setup.sh    # Bash
./scripts/hooks/setup.ps1   # PowerShell
```

Hooks are repo-wide and detect which components changed. Only runs checks relevant to the modified files (e.g. Go checks for backend changes, future ESLint for admin-client).

## Release Versioning

Releases are tag-driven through [release.yml](.github/workflows/release.yml).

Tag formats:

- backend:
  - `backend/v1.0.0`
- admin-client:
  - `admin/v1.0.0`
- desktop-client:
  - `desktop/v1.0.0`

The workflow extracts the version directly from the pushed tag and uses it for:

- the GitHub Release name
- published artifact naming
- container image tags where applicable

Examples:

- pushing `backend/v1.0.0` creates a backend release and publishes:
  - `ghcr.io/<repo>/backend:v1.0.0`
  - `ghcr.io/<repo>/backend:latest`
- pushing `admin/v1.0.0` creates an admin-client release and publishes:
  - `ghcr.io/<repo>/admin:v1.0.0`
  - `ghcr.io/<repo>/admin:latest`
- pushing `desktop/v1.0.0` creates a desktop-client release with the packaged MSIX artifact

Important:

- the release version does not currently come from component-local manifest files such as `admin-client/package.json`
- the git tag is the source of truth for release versioning
