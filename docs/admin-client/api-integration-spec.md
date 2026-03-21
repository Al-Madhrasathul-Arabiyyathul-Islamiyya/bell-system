# Admin Client API Integration Spec

## Base URLs

- REST base: `/api/v1`
- WebSocket base: `/ws`

Important: the WebSocket endpoint is not under `/api/v1`.

## Auth and Transport Rules

### Authenticated REST

- Send `Authorization: Bearer <jwt>`

### JSON:API write endpoints

- Send `Content-Type: application/vnd.api+json`
- Expect `application/vnd.api+json` responses

### Exception endpoints

- `POST /auth/login`: plain JSON request and plain JSON response
- `POST /auth/change-password`: plain JSON request, `204` response
- `POST /auth/logout`: no body, `204` response
- `POST /audio`: multipart upload
- `GET /schedule/current`: plain JSON response
- `GET /audio/{id}/content`: binary response

## Resource Summary

### Users

- `GET /users`
  - auth required
  - admin only
  - paginated
  - filters: `filter[role]`
  - sort: `sort=username,role,createdAt`
- `POST /users`
  - JSON:API create request
- `GET /users/{id}`
- `PUT /users/{id}`
- `DELETE /users/{id}`

### Sessions

- `GET /sessions`
  - auth required
  - unpaginated
- `GET /sessions/current`
  - public
- `POST /sessions`
- `PUT /sessions/{id}`
- `DELETE /sessions/{id}`

### Schedule

- `GET /schedule`
  - auth required
  - unpaginated
  - filters:
    - `filter[sessionId]`
    - `filter[day]`
  - include:
    - `include=session,sound`
- `GET /schedule/current`
  - public
  - plain JSON read model
- `POST /schedule`
- `PUT /schedule/{id}`
- `DELETE /schedule/{id}`

### Audio

- `GET /audio`
  - auth required
  - paginated
  - filters: `filter[fileType]`
  - sparse fields: `fields[audio-files]=...`
- `POST /audio`
  - auth required
  - multipart form:
    - `file`
    - `name`
    - `type`
- `GET /audio/checksums`
  - public
- `GET /audio/{id}`
  - public metadata route
- `GET /audio/{id}/content`
  - public binary route
- `PUT /audio/{id}`
- `DELETE /audio/{id}`

### System

- `GET /system/state`
  - public
- `POST /system/state`
  - auth required
  - JSON:API body
- `POST /system/cancel-next-bell`
  - auth required
  - no request body
  - returns `204` or `404`

## JSON:API Shape Notes

### Collection document

Expect:

```json
{
  "data": [],
  "meta": {},
  "links": {}
}
```

### Single resource document

Expect:

```json
{
  "data": {
    "type": "users",
    "id": "uuid",
    "attributes": {}
  }
}
```

### Relationships

Schedule items use JSON:API relationships:

- `session`
- `sound`

Nullable session is represented as:

```json
{
  "relationships": {
    "session": {
      "data": null
    }
  }
}
```

## Suggested Frontend Models

The UI should map raw API documents into simple frontend models.

### UserVm

```ts
type UserVm = {
  id: string
  username: string
  role: 'admin' | 'morning_user' | 'afternoon_user'
  createdAt: string
}
```

### SessionVm

```ts
type SessionVm = {
  id: string
  name: string
  startTime: string
  endTime: string
}
```

### ScheduleItemVm

```ts
type ScheduleItemVm = {
  id: string
  name: string
  time: string
  days: number[]
  createdAt: string
  updatedAt: string
  sessionId: string | null
  soundId: string
  session?: SessionVm
  sound?: AudioFileVm
}
```

### AudioFileVm

```ts
type AudioFileVm = {
  id: string
  name: string
  fileType: 'bell' | 'anthem' | 'school_song' | 'other'
  checksum: string
  createdAt: string
  updatedAt: string
  contentUrl: string
}
```

### SystemStateVm

```ts
type SystemStateVm = {
  state: 'active' | 'paused'
  lastUpdated: string
}
```

## WebSocket Contract

Connect with:

```text
ws://host:port/ws?token=<jwt>&client_type=admin&client_name=<name>
```

Admin frontend should use `client_type=admin`.

### Events to Handle

- `connection_acknowledged`
- `connected_clients`
- `schedules_updated`
- `audio_files_updated`
- `system_state_changed`
- `system_log`
- `bell_triggered`
- `bell_cancelled`

### Frontend Reaction Rules

- `schedules_updated`
  - invalidate schedule queries
  - refresh dashboard current schedule card
- `audio_files_updated`
  - invalidate audio list queries
- `system_state_changed`
  - patch or invalidate system-state query
- `connected_clients`
  - replace in-memory connected-clients list
- `system_log`
  - append to admin activity feed if implemented

## Query Key Plan

- `['auth', 'me']`
- `['users', page, size, role, sort]`
- `['user', id]`
- `['sessions', sort]`
- `['session', id]`
- `['sessions', 'current']`
- `['schedule', sessionId, day, include, sort]`
- `['schedule', id, include]`
- `['schedule', 'current']`
- `['audio', page, size, fileType, sort, fields]`
- `['audio', id]`
- `['audio', 'checksums']`
- `['system', 'state']`
- `['ws', 'connected-clients']`

## Known Contract Gaps and Mismatches

These should be documented in the frontend code as assumptions or follow-up tasks.

1. Product docs describe role-scoped schedule and audio permissions for session users, but the current backend only enforces admin-only access on `/users`. Other authenticated routes currently accept any authenticated role.
2. `docs/api/websocket-client-guide.md` still mentions downloading audio from `GET /audio/{id}` in one section, but the current OpenAPI and backend implementation use `GET /audio/{id}/content` for the binary.
3. `GET /schedule/current` is documented in OpenAPI as possibly `404`, but the current handler returns `200` with `{ session: null, items: [] }` when no session is active.
4. WebSocket pre-upgrade errors use a legacy `{ "error": ... }` envelope, not JSON:API `errors`.
5. `POST /system/cancel-next-bell` is behind auth and JSON:API middleware, but has no request body and should be called with an empty body.

## Frontend Guardrails

- Build the API layer to tolerate documented exceptions instead of forcing a single parser path.
- Use adapter functions per endpoint family.
- Keep OpenAPI-generated types isolated from component code.
- Log websocket errors with enough context to debug token expiry, invalid client type, and reconnect loops.
