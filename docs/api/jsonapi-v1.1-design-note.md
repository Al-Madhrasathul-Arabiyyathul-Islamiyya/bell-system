# JSON:API v1.1 Design Note

## Purpose

This note proposes a target response and request contract for migrating the Bell System REST API toward JSON:API v1.1.

It is intentionally pragmatic:

- Core CRUD endpoints should follow JSON:API v1.1 closely.
- Authentication and binary download flows may remain specialized if forcing them into JSON:API creates unnecessary complexity.
- Pagination and standardized error responses should be introduced as part of the migration, since they do not exist consistently today.

## Goals

- Standardize REST response envelopes across the backend.
- Replace route-specific list wrappers with a reusable collection format.
- Replace the current ad hoc error contract with JSON:API `errors`.
- Add pagination for collection endpoints that can grow over time.
- Make OpenAPI a reliable contract for generated clients and QA.

## Non-Goals

- Redesigning WebSocket message formats.
- Changing domain concepts such as sessions, schedule items, audio files, and system state.
- Forcing token issuance or raw file streaming into awkward JSON:API shapes if a clearer exception is better.

## Current Baseline

The current backend uses a partial house style:

- Errors: `{ "error": { "code": "...", "message": "..." } }`
- Lists: `{ "total": n, "items": [...] }`
- Single resources: bare object bodies
- Action endpoints: `{ "success": true, "message": "..." }`
- Deletes: `204 No Content`
- Audio checksums: array of `{ id, type, checksum }` objects (structured but not enveloped)
- Audio download: binary response from the same route that may return metadata in some environments

This is serviceable, but it is not JSON:API-compliant and it is not a single consistent contract.

## Media Type

JSON:API endpoints should use:

```http
Content-Type: application/vnd.api+json
Accept: application/vnd.api+json
```

Binary content endpoints should keep their existing non-JSON media type.

## Resource Types

Suggested JSON:API resource types:

- `users`
- `sessions`
- `schedule-items`
- `audio-files`
- `system-state`

No synthetic resource types. Read models like `/schedule/current` and auth endpoints are documented exceptions (see Route-Specific Design Decisions).

## Top-Level Response Rules

Use these top-level members consistently:

- `data` for successful resource and collection responses
- `errors` for failures
- `meta` for counts, pagination, and operation metadata
- `links` for self, pagination, and related URLs
- `included` for compound documents with sideloaded relationships

Do not use:

- top-level `error`
- top-level `items`
- top-level `users`
- top-level `sessions`
- top-level `files`
- top-level `success`

## Standard Success Shapes

### Single Resource

```json
{
  "data": {
    "type": "users",
    "id": "uuid",
    "attributes": {
      "username": "admin",
      "role": "admin",
      "createdAt": "2026-03-20T10:00:00Z"
    },
    "links": {
      "self": "/api/v1/users/uuid"
    }
  }
}
```

### Collection (Unpaginated)

For small, bounded collections (sessions, schedule items) where pagination adds no value:

```json
{
  "data": [
    {
      "type": "sessions",
      "id": "uuid",
      "attributes": {
        "name": "Morning",
        "startTime": "08:00",
        "endTime": "12:00"
      }
    }
  ],
  "meta": {
    "total": 1
  },
  "links": {
    "self": "/api/v1/sessions"
  }
}
```

### Collection (Paginated)

For collections that can grow (audio files, users):

```json
{
  "data": [
    {
      "type": "audio-files",
      "id": "uuid",
      "attributes": {
        "name": "bell.mp3",
        "fileType": "bell",
        "checksum": "sha256..."
      }
    }
  ],
  "meta": {
    "total": 125,
    "page": {
      "number": 2,
      "size": 25,
      "pages": 5
    }
  },
  "links": {
    "self": "/api/v1/audio?page[number]=2&page[size]=25",
    "first": "/api/v1/audio?page[number]=1&page[size]=25",
    "prev": "/api/v1/audio?page[number]=1&page[size]=25",
    "next": "/api/v1/audio?page[number]=3&page[size]=25",
    "last": "/api/v1/audio?page[number]=5&page[size]=25"
  }
}
```

### Relationship-Heavy Resource

Example `schedule-item`:

```json
{
  "data": {
    "type": "schedule-items",
    "id": "uuid",
    "attributes": {
      "name": "First Bell",
      "time": "07:00",
      "days": [1, 2, 3, 4, 5],
      "createdAt": "2026-03-20T10:00:00Z",
      "updatedAt": "2026-03-20T10:10:00Z"
    },
    "relationships": {
      "session": {
        "data": { "type": "sessions", "id": "uuid" }
      },
      "sound": {
        "data": { "type": "audio-files", "id": "uuid" }
      }
    }
  }
}
```

### Compound Document (Sideloaded Relationships)

Use `included` to sideload related resources and avoid N+1 client requests. Clients opt in via `?include=session,sound`:

```json
{
  "data": {
    "type": "schedule-items",
    "id": "uuid",
    "attributes": {
      "name": "First Bell",
      "time": "07:00",
      "days": [1, 2, 3, 4, 5]
    },
    "relationships": {
      "session": {
        "data": { "type": "sessions", "id": "session-uuid" }
      },
      "sound": {
        "data": { "type": "audio-files", "id": "audio-uuid" }
      }
    }
  },
  "included": [
    {
      "type": "sessions",
      "id": "session-uuid",
      "attributes": {
        "name": "Morning",
        "startTime": "08:00",
        "endTime": "12:00"
      }
    },
    {
      "type": "audio-files",
      "id": "audio-uuid",
      "attributes": {
        "name": "bell.mp3",
        "fileType": "bell",
        "checksum": "sha256..."
      }
    }
  ]
}
```

Supported `include` parameters by endpoint:

- `GET /api/v1/schedule` and `GET /api/v1/schedule/{id}`: `session`, `sound`
- `GET /api/v1/schedule/current`: `sound` (session is already embedded)

### Delete

Keep:

- `204 No Content`

Do not return a JSON body for successful deletes.

## Standard Error Shape

JSON:API errors should use:

```json
{
  "errors": [
    {
      "status": "400",
      "code": "invalid_request",
      "title": "Invalid request",
      "detail": "invalid JSON body"
    }
  ]
}
```

Recommended fields:

- `status`: HTTP status code as a string
- `code`: stable internal machine-readable code
- `title`: short human-readable summary
- `detail`: detailed explanation
- `source.pointer`: for request-body validation problems
- `source.parameter`: for query-string validation problems

Status code boundary: use `400` for malformed or unparseable requests (bad JSON, missing required fields, invalid UUID format). Use `422` for requests that parse correctly but fail semantic validation (password too short, time range overlap, invalid enum value).

Validation example:

```json
{
  "errors": [
    {
      "status": "422",
      "code": "validation_error",
      "title": "Invalid attribute",
      "detail": "must be at least 8 characters",
      "source": {
        "pointer": "/data/attributes/password"
      }
    }
  ]
}
```

## Standard Request Shape

### Create / Update

Create and update requests should use a JSON:API document:

```json
{
  "data": {
    "type": "users",
    "attributes": {
      "username": "admin",
      "password": "securepass123",
      "role": "admin"
    }
  }
}
```

Relationship example:

```json
{
  "data": {
    "type": "schedule-items",
    "attributes": {
      "name": "First Bell",
      "time": "07:00",
      "days": [1, 2, 3, 4, 5]
    },
    "relationships": {
      "session": {
        "data": { "type": "sessions", "id": "uuid" }
      },
      "sound": {
        "data": { "type": "audio-files", "id": "uuid" }
      }
    }
  }
}
```

## Route-Specific Design Decisions

### `POST /api/v1/auth/login`

Two viable options:

1. Keep it outside strict JSON:API.
2. Model it as a synthetic `auth-tokens` resource.

Recommended option: keep it outside strict JSON:API.

Reason:

- Token issuance is not a persisted domain resource.
- Keeping login specialized simplifies both clients and OpenAPI.

If retained as a specialized endpoint, document it explicitly as a JSON:API exception.

### `POST /api/v1/auth/change-password`

Recommended:

- keep route shape
- keep plain JSON request body: `{ "oldPassword": "...", "newPassword": "..." }`
- return `204 No Content` on success
- return JSON:API `errors` on failure

A JSON:API request body would require a `type` and `attributes` wrapper for a non-resource action, which adds complexity for no benefit. Error responses should still use the standard JSON:API `errors` format.

### `POST /api/v1/auth/logout`

Recommended:

- return `204 No Content`

Do not keep a success body unless logout must return metadata for a client workflow.

### `GET /api/v1/schedule/current`

This is a read model, not a normal resource collection. It returns the active session with its schedule items.

Keep it as a documented JSON:API exception. A synthetic `current-schedules` resource type would not map to any database entity, would need its own serializer, and would confuse clients about what is a real resource versus a view. The real resources (`sessions`, `schedule-items`) already have proper CRUD routes.

The response should still use JSON:API error format on failure. The success shape can remain specialized or use a loose `data` envelope wrapping the existing structure.

### `GET /api/v1/audio/{id}`

Do not combine JSON metadata and binary file content in one route.

Recommended split:

- `GET /api/v1/audio/{id}` returns JSON:API metadata
- `GET /api/v1/audio/{id}/content` streams the actual binary

Metadata example:

```json
{
  "data": {
    "type": "audio-files",
    "id": "uuid",
    "attributes": {
      "name": "bell.mp3",
      "fileType": "bell",
      "checksum": "sha256..."
    },
    "links": {
      "self": "/api/v1/audio/uuid",
      "content": "/api/v1/audio/uuid/content"
    }
  }
}
```

### `GET /api/v1/audio/checksums`

Options:

1. Remove it and rely on `GET /api/v1/audio?fields[audio-files]=checksum` (sparse fieldsets).
2. Keep it as a specialized optimization endpoint.

Recommended:

- keep it for client sync performance if needed
- document it as a specialized endpoint if it remains outside full JSON:API semantics

The current implementation returns an array of `{ id, type, checksum }` objects (not a bare array). Wrap it in a `data` envelope for consistency:

```json
{
  "data": [
    { "type": "audio-files", "id": "uuid", "attributes": { "fileType": "bell", "checksum": "sha256..." } }
  ]
}
```

## Pagination

Pagination is not implemented in the backend today.

Introduce pagination for collection endpoints that can realistically grow:

- `GET /api/v1/audio`
- `GET /api/v1/users`

Do not paginate inherently small, bounded collections:

- `GET /api/v1/sessions` (a school has 3-5 sessions)
- `GET /api/v1/schedule` (typically 20-30 items)

These return full collections with a `meta.total` count but no page links. If requirements change, pagination can be added later without breaking clients (adding `page` query parameters is backwards-compatible).

Recommended query parameters:

- `page[number]`
- `page[size]`

Recommended collection response metadata:

```json
{
  "meta": {
    "total": 125,
    "page": {
      "number": 2,
      "size": 25,
      "pages": 5
    }
  }
}
```

Recommended pagination links:

- `self`
- `first`
- `prev`
- `next`
- `last`

Example:

```json
{
  "links": {
    "self": "/api/v1/audio?page[number]=2&page[size]=25",
    "first": "/api/v1/audio?page[number]=1&page[size]=25",
    "prev": "/api/v1/audio?page[number]=1&page[size]=25",
    "next": "/api/v1/audio?page[number]=3&page[size]=25",
    "last": "/api/v1/audio?page[number]=5&page[size]=25"
  }
}
```

Add sorting and filtering alongside pagination where relevant.

Sorting:

- `sort=name` (ascending)
- `sort=-createdAt` (descending, prefix `-`)
- `sort=type,-createdAt` (multi-field: primary ascending, secondary descending)

Sortable fields by endpoint:

- `/api/v1/audio`: `name`, `fileType`, `createdAt`
- `/api/v1/users`: `username`, `role`, `createdAt`
- `/api/v1/schedule`: `name`, `time`
- `/api/v1/sessions`: `name`, `startTime`

Filtering:

- `filter[role]=admin`
- `filter[fileType]=bell`
- `filter[sessionId]=uuid`
- `filter[day]=1`

Sparse fieldsets (JSON:API `fields` parameter):

- `fields[audio-files]=name,checksum` (return only specified attributes)
- `fields[users]=username,role`

## Error Coverage Requirements

The current backend already emits many errors, but OpenAPI does not document them consistently.

Every route family should document applicable shared errors.

### Common Error Statuses

- `400 Bad Request`
- `401 Unauthorized`
- `403 Forbidden`
- `404 Not Found`
- `409 Conflict` (already implemented for duplicate usernames and session conflicts)
- `415 Unsupported Media Type` (must be implemented as part of migration — JSON:API requires `application/vnd.api+json` content negotiation, which the backend does not currently enforce)
- `422 Unprocessable Entity`
- `500 Internal Server Error`

### Recommended Coverage by Route Family

#### Authentication

- `400` malformed request
- `401` invalid credentials or missing auth
- `404` user not found where applicable
- `500` internal failure

#### Users

- `400` malformed request or invalid ID
- `401` unauthenticated
- `403` forbidden
- `404` missing user
- `409` duplicate username
- `422` semantic validation
- `500` internal failure

#### Sessions

- `400` malformed request or invalid ID
- `401` unauthenticated
- `403` forbidden
- `404` missing session
- `409` session conflict
- `422` semantic validation
- `500` internal failure

#### Schedule

- `400` malformed request or invalid ID
- `401` unauthenticated
- `403` forbidden
- `404` missing item
- `409` future schedule conflict if introduced
- `422` semantic validation
- `500` internal failure

#### Audio

- `400` malformed request or invalid ID
- `401` unauthenticated
- `403` forbidden
- `404` missing file
- `415` wrong media type
- `422` semantic validation
- `500` internal failure

#### System

- `400` malformed request
- `401` unauthenticated
- `403` forbidden
- `404` no pending bell or missing state resource
- `500` internal failure

## OpenAPI Design Requirements

When the OpenAPI document is updated, add shared reusable schemas for:

- JSON:API resource identifier
- JSON:API links object
- JSON:API error object
- single-resource document
- collection document
- pagination metadata

Also:

- switch JSON API endpoints to `application/vnd.api+json`
- document exceptions explicitly for auth and binary content endpoints
- stop using route-specific list wrappers in the spec

## Migration Strategy

### Versioning

No versioning needed. The app is not deployed and nothing depends on it. All changes are made in-place under `/api/v1`.

### Recommended Order

1. **Add JSON:API error middleware.** A single `jsonapi.ErrorMiddleware` intercepts `writeError` calls and re-envelopes them into JSON:API `errors` format. This is one middleware addition rather than touching every handler individually.
2. **Add `application/vnd.api+json` content negotiation middleware.** Validate `Content-Type` and `Accept` headers on JSON:API endpoints. Return `415` for wrong media types. Exception endpoints (auth, binary) skip this middleware.
3. Standardize collection responses with `data`, `meta`, and pagination links (for audio and users).
4. Standardize single-resource responses with `data` and `attributes`.
5. Move foreign keys to `relationships` and implement `?include=` for compound documents.
6. Split audio metadata from binary content (`/audio/{id}` vs `/audio/{id}/content`).
7. Update OpenAPI to reflect the new contract.

## Pragmatic Recommendation

Use a hybrid contract:

- **Full JSON:API v1.1**: CRUD resources (`users`, `sessions`, `schedule-items`, `audio-files`, `system-state`)
- **Documented exceptions** (plain JSON request/response, JSON:API errors on failure):
  - `POST /auth/login` — token issuance is not a persisted resource
  - `POST /auth/change-password` — action endpoint, no resource type
  - `POST /auth/logout` — returns `204`, no body
  - `GET /schedule/current` — read model, not a CRUD resource
  - `GET /audio/checksums` — optimization endpoint for client sync
- **Non-JSON endpoints**: `GET /audio/{id}/content` — binary streaming

All endpoints (including exceptions) should use JSON:API `errors` format for error responses. This gives the backend a coherent API design without forcing unnatural JSON:API modeling onto action endpoints or raw media transfer.
