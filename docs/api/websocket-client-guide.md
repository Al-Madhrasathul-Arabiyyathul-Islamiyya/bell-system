# WebSocket Client Integration Guide

Guide for native Windows client developers integrating with the Bell Schedule System WebSocket server.

## Quick Start

1. Authenticate via REST API: `POST /api/v1/auth/login` to get a JWT token
2. Connect to WebSocket: `ws://{host}:{port}/ws?token={jwt}&client_type=client&client_name=Main+Hall`
3. Listen for events, send heartbeats

## Connection

### Endpoint

```
ws://{host}:{port}/ws?token={jwt_token}&client_type={type}&client_name={name}
```

| Parameter     | Required | Values              | Description                    |
|---------------|----------|---------------------|--------------------------------|
| `token`       | Yes      | JWT string          | From `/api/v1/auth/login`         |
| `client_type` | Yes      | `admin` or `client` | `admin` requires admin role    |
| `client_name` | No       | Any string          | Human-readable display name    |

### Authentication

The JWT token must be valid and not expired. Obtain it via:

```
POST /api/v1/auth/login
Content-Type: application/json

{"username": "...", "password": "..."}
```

Response includes `token` field. Tokens expire after 180 minutes.

### Connection Response

On successful connection, you receive:

```json
{
  "type": "connection_acknowledged",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "connection_id": "uuid",
    "client_type": "client",
    "message": "Connection successful"
  }
}
```

Store `connection_id` for debugging purposes.

## Message Format

All messages use the same JSON envelope:

```json
{
  "type": "event_type",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {}
}
```

- `type` — Event identifier string
- `timestamp` — ISO 8601 UTC timestamp
- `payload` — Event-specific data (may be omitted for notification-only events)

## Server Events

### Schedule Events

#### `schedules_updated`
Sent when any schedule item is created, updated, or deleted. No payload — clients should re-fetch schedules via REST API (`GET /api/v1/schedule`).

#### `bell_triggered`
Sent when a bell is about to play. Client should play the referenced sound file.

```json
{
  "payload": {
    "scheduleItemId": "uuid",
    "soundId": "uuid",
    "name": "First Period"
  }
}
```

#### `bell_cancelled`
Sent when a scheduled bell has been cancelled (e.g., by admin override).

```json
{
  "payload": {
    "scheduleItemId": "uuid"
  }
}
```

### System Events

#### `system_state_changed`
Sent when the bell system is paused or resumed. When paused, clients should stop playing bells.

```json
{
  "payload": {
    "state": "active" | "paused"
  }
}
```

#### `audio_files_updated`
Sent when audio files are added, updated, or deleted. Clients should re-check audio file checksums via `GET /api/v1/audio/checksums` and re-download any changed files.

### Admin-Only Events

These are only sent to connections with `client_type=admin`.

#### `connected_clients`
Sent whenever any client connects or disconnects. Provides the full list of currently connected clients.

```json
{
  "payload": {
    "clients": [
      {
        "id": "connection-uuid",
        "ip": "192.168.1.100",
        "client_type": "client",
        "client_name": "Main Hall Display",
        "connected_since": "2024-03-13T14:20:30Z"
      }
    ]
  }
}
```

#### `system_log`
Real-time log entries for admin monitoring.

```json
{
  "payload": {
    "level": "info" | "warning" | "error",
    "message": "System message here",
    "source": "component_name"
  }
}
```

## Client Events

### `heartbeat`
Send periodically (every 30 seconds recommended) to keep the connection alive.

```json
{
  "type": "heartbeat",
  "timestamp": "2024-03-13T15:30:45Z"
}
```

### `register`
Optional — update your client name after connecting.

```json
{
  "type": "register",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "client_type": "client",
    "client_name": "Updated Name"
  }
}
```

## Reconnection Strategy

The server sends periodic pings (every 30 seconds). If you miss a pong, the server closes the connection.

Implement reconnection with exponential backoff:

1. On disconnect, wait 1 second, then reconnect
2. If that fails, wait 2 seconds
3. Double the wait each time, up to 30 seconds max
4. Reset backoff on successful connection
5. Re-authenticate if the token has expired (login again)

```
delay = min(2^attempt * 1000ms, 30000ms)
```

## Error Responses

Pre-upgrade errors return standard HTTP responses:

| Status | Code                 | Cause                            |
|--------|----------------------|----------------------------------|
| 401    | `missing_token`      | No `token` query parameter       |
| 401    | `invalid_token`      | Expired or malformed JWT         |
| 400    | `invalid_client_type`| Not `admin` or `client`          |
| 403    | `forbidden`          | Non-admin user using admin type  |

## Audio File Sync

When you receive `audio_files_updated`:

1. Fetch checksums: `GET /api/v1/audio/checksums`
2. Compare with locally cached checksums
3. Download changed files: `GET /api/v1/audio/{id}` (returns the file binary)
4. Cache the new checksum

This ensures clients always have the latest audio files without downloading everything on every update.

## Offline Resilience

Desktop clients should remain functional when disconnected from the server.

### Local Bell Triggering

- On connect (or reconnect), fetch the full schedule: `GET /api/v1/schedule` and current session: `GET /api/v1/sessions/current`
- Cache the schedule locally
- When disconnected, the client should trigger bells locally based on the cached schedule and system clock
- When reconnected, stop local triggers and resume server-driven mode

### Local Pause/Resume

- Track the system state (`active`/`paused`) locally
- When disconnected, allow local pause/resume via the client UI
- On reconnect, fetch the server state (`GET /api/v1/system/state`) and adopt it as the source of truth

### On-Demand Unscheduled Bells

- The client may support triggering bells manually outside the schedule (e.g., emergency bell)
- This is a client-only feature — the server does not need to be involved

### Reconnection Flow

1. Re-authenticate if the token has expired
2. Connect to WebSocket with `register` message
3. Fetch system state: `GET /api/v1/system/state`
4. Fetch current schedule: `GET /api/v1/sessions/current` + `GET /api/v1/schedule`
5. Fetch audio checksums and sync any changed files
6. Resume server-driven mode (stop local bell triggers)
