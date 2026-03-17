# Arabiyya Bell Schedule System

An automated school bell schedule and media playback system for Arabiyya School. Manages bell schedules, plays audio files on cue, and provides administrative control through a web interface and desktop client.

[![CI](https://github.com/arabiyya-edu/bell-system/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/arabiyya-edu/bell-system/actions/workflows/ci.yml)
[![Release](https://github.com/arabiyya-edu/bell-system/actions/workflows/release.yml/badge.svg)](https://github.com/arabiyya-edu/bell-system/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/arabiyya-edu/bell-system/branch/develop/graph/badge.svg)](https://codecov.io/gh/arabiyya-edu/bell-system)

## Architecture

This is a monorepo containing three components:

| Component | Tech | Description |
|-----------|------|-------------|
| [bell-system-backend](bell-system-backend/) | Go, Chi, SQL Server | REST API, scheduler, WebSocket server |
| [admin-client](admin-client/) | React, TypeScript, Vite | Admin web interface for schedule management |
| [desktop-client](desktop-client/) | .NET (WPF) | Desktop app for audio playback and display |

## Tech Stack

- **Backend**: Go 1.26+, go-chi/chi v5, spf13/viper (TOML config), go.uber.org/zap
- **Database**: SQL Server (microsoft/go-mssqldb)
- **Auth**: JWT (180min expiry)
- **Admin Client**: React 18, TypeScript, Vite
- **Desktop Client**: .NET, WPF

## Getting Started

### Prerequisites

- Go 1.26+
- SQL Server
- Node.js 18+ and Yarn (for admin client)
- .NET 9+ SDK (for desktop client)

### Clone

```bash
git clone https://github.com/arabiyya-edu/bell-system.git
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
├── admin-client/         # React admin dashboard
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
