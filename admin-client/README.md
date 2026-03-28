# Arabiyya Bell Schedule System - Admin Client

## Development

This app is managed with `pnpm`.

```bash
pnpm install
pnpm dev
pnpm build
pnpm preview
```

## Quality Checks

```bash
pnpm lint
pnpm format:check
pnpm build
pnpm check
```

The frontend now uses:

- `oxlint` for linting
- `oxfmt` for formatting
- `vue-tsc` for type checking

## Container Deployment

Build locally:

```bash
docker build -t bell-system-admin-client:test .
docker run --rm -p 8081:80 bell-system-admin-client:test
```

Run the published GHCR image locally with Compose:

```bash
docker compose up -d
```

The default Compose image is:

- `ghcr.io/al-madhrasathul-arabiyyathul-islamiyya/bell-system/admin:latest`

You can override it if needed:

```bash
$env:ADMIN_CLIENT_IMAGE="ghcr.io/al-madhrasathul-arabiyyathul-islamiyya/bell-system/admin:v1.0.0"
docker compose up -d
```

For hosted deployment with the backend behind Traefik, use the repo root deploy file:

```bash
docker compose -f ../docker-compose.deploy.yml up -d
```
