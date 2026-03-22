# Admin Client Documentation

This folder defines the target implementation for the `admin-client` frontend.

## Recommended Document Order

1. `implementation-spec.md`
   Main frontend blueprint: stack, packages, routing, state, UX, folder structure, and delivery order.
2. `api-integration-spec.md`
   REST and WebSocket integration details, JSON:API handling rules, and known backend contract gaps.
3. `theme.css`
   DaisyUI theme tokens already prepared for the admin client.

## Current Project State

- `admin-client/` now has the initial app shell, routing, persisted theme/auth placeholders, and the base Tailwind + DaisyUI pipeline.
- Backend API contracts live in `docs/api/openapi/` and are partially aligned with the Go implementation.
- The frontend should be implemented as a Vue 3 SPA using VueUse helpers and DaisyUI theming.

## High-Level Decisions

- Use Vue 3 + TypeScript + Vite.
- Use `vue-router` for route-level layout and guards.
- Use Pinia for auth and small client-only state.
- Use TanStack Vue Query for API cache, loading, invalidation, and optimistic refresh.
- Use VueUse for storage, color mode, media queries, debouncing, intervals, document title, and WebSocket handling.
- Use Regle with Zod schemas for form validation instead of `vee-validate`.
- Use Tailwind CSS v4 + DaisyUI, and import the theme tokens from this folder.
- Treat the backend as JSON:API-first, but explicitly handle the exception endpoints:
  - `POST /auth/login`
  - `POST /auth/change-password`
  - `POST /auth/logout`
  - `GET /schedule/current`
  - `GET /audio/{id}/content`
  - WebSocket `/ws`

## Expected Deliverables

- Authentication flow
- Dashboard
- Users CRUD
- Sessions CRUD
- Schedule CRUD
- Audio library management
- System controls
- WebSocket-driven live refresh
- Production-ready container build for `admin-client/`
