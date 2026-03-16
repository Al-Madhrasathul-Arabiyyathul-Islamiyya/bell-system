Create migration files for database tables
Implement basic API handlers for the core functionality
Add authentication middleware and JWT handling

## Package Structure

### `cmd/server/`
Entry point — starts the HTTP server, wires dependencies.

### `cmd/migrate/`
Migration runner — applies SQL migrations up/down.

### `config/`
TOML configuration via Viper — server, database, JWT, audio paths, WebSocket settings.

### `internal/database/`
Repositories for data access:
- Users
- Sessions
- ScheduleItems
- ScheduleDays
- SystemAudioFiles
- SystemState

### `internal/models/`
Data structures matching the database schema — JSON tags, validation logic.

### `internal/handlers/`
HTTP handlers for all API endpoints, plus auth middleware:
- Authentication (login, logout, change-password)
- Schedule management (CRUD + current)
- Audio file management (CRUD + checksums + upload/download)
- Session management (CRUD + current)
- System state management (get/set state, cancel next bell)
- User management (CRUD)

### `internal/router/`
Chi route registration — mounts all handler groups under `/api/v1`.

### `internal/websocket/`
Real-time communication:
- Hub — connection registry, broadcast
- Client — per-connection read/write pumps
- Handler — upgrade HTTP to WebSocket, authenticate via query params
- Notifier — bridges handler events to WebSocket broadcasts

### `internal/scheduler/`
Bell timer engine:
- Schedules timers for upcoming bells
- Reloads on schedule/state changes via reload notifier
- Respects system state (active/paused)

### `internal/services/`
Business logic services:
- Token service (JWT generation/validation)
- Password hasher (bcrypt)
- File storage (audio file I/O, checksum calculation)

### `internal/helpers/`
Shared utilities (e.g., schedule day helpers).

### `pkg/logger/`
Zap wrapper — structured logging.

### `pkg/errors/`
Application error types.
