# Sessionless Schedule Items

## Goal

Support schedule items that are not tied to Morning or Afternoon sessions,
while keeping CRUD permissions and scheduler behavior explicit and predictable.

## Desired Outcome

- Schedule items may exist with `session = null`.
- Admin, morning, and afternoon users can perform CRUD operations on those
  sessionless items.
- The scheduler can load and ring those items correctly.
- `/schedule/current` and the dashboard can represent them cleanly.

## Current State

The current system is only partially ready for this.

### What Already Works

- Backend model allows nullable session IDs.
- Schedule create/update API schemas document nullable `session`.
- Repository CRUD can persist and return sessionless schedule items.
- Frontend integration notes already acknowledge nullable `session`.

### What Does Not Yet Work

- The runtime scheduler only loads items for the current active session.
- `/schedule/current` only returns the current session and that session's items.
- The admin client schedule form currently assumes a chosen session in normal
  UX.
- Role scoping in the admin client is session-centric, not sessionless-aware.

## Why This Matters

Sessionless items are useful for bells that should ring regardless of Morning
or Afternoon session boundaries, for example:

- campus-wide reminders
- special one-off routine bells
- general transition bells not owned by a named session

## Product Rules

### CRUD Permissions

- `admin`
  - can create, edit, delete, and view all schedule items
  - can assign either a named session or no session
- `morning_user`
  - can manage Morning session items
  - can also manage sessionless items
- `afternoon_user`
  - can manage Afternoon session items
  - can also manage sessionless items

### Execution Rules

- sessionless items should be evaluated independently of the current active
  session
- they still obey:
  - day-of-week filters
  - system active/paused state
- if a named session is active, session-scoped and sessionless items may both
  be scheduled for that day

## Backend Work Required

## 1. Clarify Query and Runtime Semantics

The current scheduler path is session-first:

1. find current session
2. load that session's schedule items
3. create timers

This needs to change to include sessionless items for the current day.

### Recommended Direction

Add a repository method such as:

- `GetRunnableSchedules(ctx, currentSessionID *uuid.UUID, day int)`

Behavior:

- always include sessionless items for the current day
- include session-bound items for the active session when one exists
- order by time

## 2. Update `/schedule/current`

The current read model should be extended so the dashboard and other clients can
reason about both categories.

Possible response direction:

- keep `session` for the active session
- include items from:
  - active session
  - sessionless schedule items
- optionally annotate each item with scope:
  - `scope: "session"`
  - `scope: "global"`

## 3. Authorization Rules

The backend should make role behavior explicit rather than relying on frontend
filtering.

Suggested policy:

- `admin`: all schedule items
- `morning_user`: Morning session items + sessionless items
- `afternoon_user`: Afternoon session items + sessionless items

That applies to:

- list
- create
- update
- delete

## 4. Tests

Add tests for:

- create/update schedule item with `session = null`
- list sessionless items
- scheduler includes sessionless items
- `/schedule/current` includes sessionless items
- role authorization for Morning/Afternoon users on sessionless items

Integration coverage should confirm actual ring scheduling behavior, not only
CRUD persistence.

## Admin Client Work Required

## 1. Schedule Form

The create/edit form should allow:

- named session selection
- explicit no-session option

Recommended UX:

- a select with:
  - `No session`
  - all permitted named sessions

## 2. Schedule Filtering

The schedule page should support:

- filtering by named session
- filtering to sessionless items
- optionally a combined `All permitted items` view

## 3. Role-Aware Views

- `admin` sees all items
- `morning_user` sees Morning + sessionless
- `afternoon_user` sees Afternoon + sessionless

This should match backend authorization exactly.

## 4. Dashboard and Topbar

Upcoming bell summaries should be able to show sessionless items naturally.

If useful, the UI can display a small badge such as:

- `Sessionless`
- `Morning`
- `Afternoon`

## OpenAPI and Documentation Changes

When implemented, update:

- schedule request/response schemas
- schedule current read-model docs
- admin-client integration docs
- any role/authorization documentation

## Suggested Delivery Order

1. Define the desired runtime semantics for sessionless items.
2. Implement backend repository and scheduler support.
3. Update `/schedule/current` contract.
4. Add backend authorization and tests.
5. Update admin client form, filters, and dashboard rendering.
6. Update OpenAPI and project docs.

## Recommendation

Treat sessionless schedule items as a deliberate second scheduling scope, not as
an edge case. The implementation should make their behavior explicit in both
code and API contracts, otherwise the feature will remain half-supported and
confusing for operators.
