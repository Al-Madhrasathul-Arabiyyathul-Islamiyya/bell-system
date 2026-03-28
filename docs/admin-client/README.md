# Admin Client Documentation

This folder describes the implemented `admin-client` frontend and the supporting decisions behind it.

## Recommended Document Order

1. `implementation-spec.md`
   Main frontend blueprint: stack, packages, routing, state, UX, folder structure, and delivery order.
2. `api-integration-spec.md`
   REST and WebSocket integration details, JSON:API handling rules, and known backend contract gaps.
3. `theme.css`
   DaisyUI theme tokens already prepared for the admin client.

## Current Project State

- `admin-client/` is feature-complete for the current v1 scope.
- The app includes:
  - login/auth persistence
  - dashboard
  - sessions CRUD
  - schedule CRUD
  - audio management with preview
  - users CRUD
  - system controls
  - WebSocket-driven live status updates
- `/users` is gated to the `admin` role.
- Schedule visibility is role-aware for `admin`, `morning_user`, and `afternoon_user`.
- The frontend uses Vue 3, VueUse, Pinia, TanStack Vue Query, DaisyUI, Regle, Zod, `oxlint`, and `oxfmt`.
- The published container image is released to GHCR under:
  - `ghcr.io/al-madhrasathul-arabiyyathul-islamiyya/bell-system/admin:<tag>`

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

## Deployment Notes

- Local image build uses [Dockerfile](/D:/Stuff/projects/bell-system/admin-client/Dockerfile).
- Local Compose usage uses [docker-compose.yml](/D:/Stuff/projects/bell-system/admin-client/docker-compose.yml) and defaults to the GHCR `latest` image.
- Hosted multi-service deployment uses [docker-compose.deploy.yml](/D:/Stuff/projects/bell-system/docker-compose.deploy.yml).
- Versioned releases are driven by git tags in the form `admin/v*`.
