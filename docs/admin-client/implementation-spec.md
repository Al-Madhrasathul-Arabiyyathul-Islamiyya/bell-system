# Admin Client Implementation Spec

## Goal

Build `admin-client/` as a Vue 3 admin SPA for the Bell Schedule System backend. The frontend should cover the existing API surface, respect the current backend behavior, and avoid hand-rolled plumbing when VueUse or established Vue libraries already solve the problem.

## Primary Constraints

- Framework: Vue 3 with TypeScript.
- Styling: Tailwind CSS + DaisyUI.
- Theme source: `docs/admin-client/theme.css`.
- Helper preference: prefer VueUse composables over custom utility code where they fit.
- API style: JSON:API for most CRUD routes, with a small set of plain JSON or binary exceptions.
- Deployment target: static SPA bundle served by nginx, matching the existing `admin-client/Dockerfile`.

## Package Decisions

### Runtime

- `vue`
- `vue-router`
- `pinia`
- `@tanstack/vue-query`
- `@vueuse/core`
- `@vueuse/router`
- `ofetch`
- `zod`
- `vee-validate`
- `@vee-validate/zod`
- `@iconify/vue`

### Build and Styling

- `@vitejs/plugin-vue`
- `tailwindcss`
- `@tailwindcss/vite`
- `daisyui`

### Quality and Testing

- `vitest`
- `@vue/test-utils`
- `happy-dom`
- `playwright`
- `openapi-typescript`
- `eslint`
- `eslint-plugin-vue`
- `@typescript-eslint/parser`
- `@typescript-eslint/eslint-plugin`
- `prettier`

## Why These Packages

- `vue-router`: route guards, nested layouts, and URL-driven filtering.
- `pinia`: auth token, current user, app shell state, and websocket status are small global concerns.
- `@tanstack/vue-query`: server state should not be stored in Pinia.
- `@vueuse/core`: use `useStorage`, `useColorMode`, `useBreakpoints`, `useDocumentTitle`, `useDebounceFn`, `watchDebounced`, `useIntervalFn`, `useWebSocket`, and `useTimeAgo`.
- `ofetch`: lighter than Axios and sufficient for JSON:API, login JSON, and multipart upload.
- `zod` + `vee-validate`: runtime-safe forms without duplicating validation logic.
- `openapi-typescript`: generate API-facing types from `docs/api/openapi/openapi.yaml`, then wrap them with JSON:API adapters.

## Explicit Frontend Decisions

1. Use Vue Query for all REST resources.
2. Use Pinia only for auth, UI preferences, and ephemeral shell state.
3. Keep JSON:API parsing in a single adapter layer under `src/lib/jsonapi/`.
4. Do not couple components directly to raw API documents.
5. Normalize API responses into view models in composables or service functions.
6. Use route-level data loading via Vue Query, not `onMounted` fetch code inside pages.
7. Use optimistic refresh and invalidation after mutations instead of manual in-place list patching unless latency becomes a problem.
8. Use DaisyUI components first, then add project-specific wrappers for repeated patterns.
9. Use modal dialogs for destructive confirmations and small forms; use dedicated pages/drawers for larger edit flows.
10. Implement WebSocket as enhancement, not as the only source of truth. REST remains authoritative.

## Recommended Install Commands

```bash
cd admin-client
yarn add vue-router pinia @tanstack/vue-query @vueuse/core @vueuse/router ofetch zod vee-validate @vee-validate/zod @iconify/vue
yarn add -D tailwindcss @tailwindcss/vite daisyui vitest @vue/test-utils happy-dom playwright openapi-typescript eslint eslint-plugin-vue @typescript-eslint/parser @typescript-eslint/eslint-plugin prettier
```

## App Structure

```text
admin-client/
├── public/
├── src/
│   ├── app/
│   │   ├── AppShell.vue
│   │   ├── router.ts
│   │   ├── query-client.ts
│   │   └── providers.ts
│   ├── assets/
│   ├── components/
│   │   ├── app/
│   │   ├── data-display/
│   │   ├── forms/
│   │   └── feedback/
│   ├── features/
│   │   ├── auth/
│   │   ├── dashboard/
│   │   ├── users/
│   │   ├── sessions/
│   │   ├── schedule/
│   │   ├── audio/
│   │   └── system/
│   ├── lib/
│   │   ├── api/
│   │   ├── jsonapi/
│   │   ├── env/
│   │   ├── forms/
│   │   └── utils/
│   ├── stores/
│   ├── styles/
│   ├── pages/
│   └── main.ts
├── index.html
└── vite.config.ts
```

## Routes and Screens

### Public

- `/login`
  - Username/password form
  - Redirect authenticated users to `/dashboard`

### Protected

- `/`
  - Redirect to `/dashboard`
- `/dashboard`
  - System state card
  - Current session summary
  - Today schedule summary
  - Connected clients panel from WebSocket
  - Quick actions: pause/resume, cancel next bell
- `/schedule`
  - Table or timeline grouped by session
  - Filters: session, day, search, include session-independent bells
  - Create/edit/delete schedule item
- `/audio`
  - Paginated library
  - Filter by file type
  - Upload audio file
  - Rename / retype / delete
  - Optional preview by using `/audio/{id}/content`
- `/sessions`
  - Sessions list
  - Create/edit/delete session
- `/users`
  - Paginated user management
  - Create/edit/delete user
  - Change role
  - Reset password through edit flow
- `/system`
  - System state controls
  - Current websocket connection status
  - Change-password form for current user

## Navigation Layout

- Left sidebar on desktop, drawer on mobile.
- Top app bar with:
  - current user
  - websocket indicator
  - theme toggle
  - logout
- Use DaisyUI drawer, navbar, menu, stat, alert, table, badge, modal, join, and toast patterns.

## Auth Spec

- Login endpoint returns plain JSON, not JSON:API.
- Persist JWT and current user with `useStorage`.
- Add a router guard that checks token presence.
- Add an API fetch interceptor:
  - include `Authorization: Bearer <token>` when present
  - send `Content-Type: application/vnd.api+json` for JSON:API writes
  - do not send JSON:API content type for login or multipart upload
- On `401`, clear auth state and redirect to `/login`.

## Feature Modules

### Auth

- `useAuthStore`
- `useLoginMutation`
- `useLogoutMutation`
- `useChangePasswordMutation`
- `requireAuth` route guard

### Dashboard

- combine:
  - `GET /system/state`
  - `GET /sessions/current`
  - `GET /schedule/current`
  - WebSocket live status

### Users

- server-side pagination
- role filter
- create/update forms with Zod schemas

### Sessions

- unpaginated list
- time-range editing
- conflict errors shown inline when backend returns `409`

### Schedule

- list with `include=session,sound`
- day filter using backend `filter[day]`
- session filter using backend `filter[sessionId]`
- form supports nullable `session`
- show sound and session names via included resources or preloaded option lists

### Audio

- paginated list
- multipart upload
- optimistic refresh on upload/update/delete
- preview action using metadata route plus binary route

### System

- pause/resume action via `POST /system/state`
- cancel-next-bell action via `POST /system/cancel-next-bell`
- display last updated timestamp

## Composable Strategy

Prefer feature composables over page-local logic:

- `useUsersQuery`
- `useCreateUserMutation`
- `useSessionsQuery`
- `useScheduleItemsQuery`
- `useAudioFilesQuery`
- `useSystemStateQuery`
- `useBellSystemSocket`

VueUse helpers to use directly:

- `useStorage` for JWT, theme, and last-used filters
- `useColorMode` for DaisyUI theme switching
- `useBreakpoints` for responsive drawer behavior
- `useDocumentTitle` per route
- `useDebounceFn` or `watchDebounced` for search/filter sync
- `useIntervalFn` for fallback polling on dashboard if socket is down
- `useWebSocket` for `/ws`

## UI and Theme Spec

- Import `docs/admin-client/theme.css` into the Tailwind/DaisyUI pipeline and keep the theme names `light` and `dark`.
- Default to the light theme.
- Use the existing green/red/gold palette as the product identity.
- Avoid Vite starter styling entirely.
- Add a small set of local CSS utilities only for app-shell layout and branded surfaces.

## Table and Form Patterns

- Use reusable data-table wrappers with:
  - loading state
  - empty state
  - error state
  - pagination controls where needed
- Use reusable form sections with:
  - field label
  - helper text
  - inline error
  - submit bar
- Show backend validation and conflict errors in a banner plus field-level mapping where possible.

## Error Handling

- Centralize JSON:API error parsing.
- Standardize an app-level error shape:
  - `status`
  - `code`
  - `title`
  - `detail`
- Use toasts for transient success.
- Use inline alerts for failed page data or form submissions.
- For `404` on `cancel-next-bell`, show "No pending bell to cancel" instead of generic failure.

## Environment Variables

Use Vite env vars:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_BASE_URL=ws://localhost:8080/ws
VITE_APP_NAME=Bell System Admin
```

## Delivery Order

1. Replace starter app with app shell, router, Pinia, Vue Query, Tailwind, and DaisyUI.
2. Implement auth and protected routing.
3. Implement shared API layer and JSON:API adapters.
4. Implement dashboard.
5. Implement sessions and users.
6. Implement schedule management.
7. Implement audio management.
8. Add WebSocket live refresh.
9. Add tests, production polish, and container verification.

## Definition of Done

- All admin CRUD flows work against the current backend.
- JSON:API and plain JSON exception routes are both handled correctly.
- WebSocket reconnection works and degrades gracefully.
- Mobile and desktop layouts are usable.
- Theme toggle persists.
- Forms show actionable validation messages.
- Unit tests cover adapters, stores, and major composables.
- Playwright covers login plus one CRUD path per major feature.
