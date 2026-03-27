# Loki-Friendly Logging, Audit, and Health Roadmap

## Goal

Improve operational visibility after `v1.0` by making the backend and
supporting services easier to observe in Loki-compatible pipelines, while also
adding first-class audit and health-oriented logs.

## Scope

- Loki-friendly structured logging
- Audit trails for sensitive user and system actions
- System health logs for scheduler/runtime monitoring
- Clear frontend and API consequences where relevant

## Why This Is Post-v1

The current system already has functional structured logging and websocket
events, but it is not yet shaped as an operational observability product.

The immediate product value after v1 is:

- easier production troubleshooting
- traceable operator actions
- better incident review and accountability
- clearer runtime signals for scheduling, websocket, and audio failures

## Current Baseline

- Backend uses structured logging via the logger package.
- Some websocket and scheduler actions already emit logs.
- The admin client surfaces realtime state, but does not expose audit history
  or health history as a first-class feature.
- There is no defined audit schema or documented Loki label strategy.

## Workstream 1: Loki-Friendly Structured Logging

### Target Outcome

Every operationally important backend log line should be easy to ingest, query,
and correlate in Loki.

### Requirements

- Keep logs structured JSON, not free-form strings.
- Normalize keys across subsystems:
  - `component`
  - `operation`
  - `user_id`
  - `username`
  - `role`
  - `schedule_item_id`
  - `session_id`
  - `sound_id`
  - `client_id`
  - `client_type`
  - `request_id`
  - `outcome`
- Keep high-cardinality values out of labels unless intentionally chosen.
- Define a small stable label set for Loki queries, for example:
  - `service=bell-system-backend`
  - `component=scheduler|websocket|http|audio|auth`
  - `level=info|warn|error`
  - `environment`

### Backend Changes

- Standardize logger fields across handlers, repositories, scheduler, and
  websocket hub.
- Add request correlation fields in HTTP middleware.
- Ensure websocket-related logs carry:
  - client type
  - client name
  - authenticated user identity where available
- Make scheduler logs explicit about:
  - session loaded
  - number of timers scheduled
  - skipped bells
  - reload triggers

### Deliverables

- logging field conventions document
- backend logging refactor
- optional promtail/Loki example configuration

## Workstream 2: Audit Trails

### Target Outcome

Sensitive user and system actions should be traceable as durable audit events,
not just transient logs.

### Audit-Relevant Actions

- login success and failure
- password changes
- user CRUD
- session CRUD
- schedule CRUD
- audio upload/update/delete
- system pause/resume
- cancel-next-bell

### Audit Event Shape

Recommended audit event fields:

- `event_id`
- `timestamp`
- `actor_user_id`
- `actor_username`
- `actor_role`
- `action`
- `resource_type`
- `resource_id`
- `summary`
- `before`
- `after`
- `source_ip`
- `request_id`

### Storage Options

1. Database-backed audit table
   Best for durable history and admin-client viewing.
2. Log-only audit stream
   Faster to add, but weaker for productized history access.
3. Hybrid
   Store important audit records in the database and emit them to logs too.

Recommended direction: hybrid.

### Admin Client Impact

Post-v1, the admin client could add:

- an audit trail page
- recent actions card on the dashboard
- filters by user, resource type, and action

## Workstream 3: System Health Logs

### Target Outcome

Operators should be able to diagnose scheduler and delivery issues quickly from
logs and eventually from dashboards.

### Health Signals To Emit

- scheduler loaded with no active session
- scheduler loaded N timers for session X
- bell triggered
- bell cancelled
- system paused/resumed
- websocket hub client count changes
- failed websocket authentication
- failed audio file lookup or playback dispatch
- schedule reload cause:
  - schedule update
  - system state change
  - startup
  - periodic reload

### Recommended Severity Split

- `info`
  - normal lifecycle and operator actions
- `warn`
  - unexpected but recoverable conditions
- `error`
  - failed persistence, scheduling, websocket delivery, or auth checks

## Suggested Delivery Order

1. Define shared field names and Loki label strategy.
2. Refactor backend logs to match the shared schema.
3. Add durable audit trail storage for sensitive actions.
4. Expose audit history in the admin client.
5. Add dashboards/alerts around health logs.

## Open Questions

- Which fields should become Loki labels versus JSON fields only?
- Should audit trails be immutable at the database layer?
- Should the admin client expose audit history to all authenticated users or
  admin only?
- Do we want health metrics in Prometheus as a separate follow-up from logs?
