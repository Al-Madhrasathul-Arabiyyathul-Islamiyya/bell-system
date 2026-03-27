# Post-v1 Roadmap

This folder captures follow-up work that sits beyond the current `v1.0`
baseline.

The admin client is now effectively feature-complete for the planned v1 scope,
so the next work should focus on operational maturity and feature expansion
rather than missing core screens.

## Documents

1. `logging-and-audit-roadmap.md`
   Structured logging improvements for Loki, audit trails, and system health
   telemetry.
2. `sessionless-schedule-items.md`
   Roadmap for schedule items that are not tied to Morning or Afternoon
   sessions, including scheduler and permission implications.

## Guiding Principles

- Keep post-v1 work additive and low-risk where possible.
- Favor explicit contracts in backend models and OpenAPI before expanding the
  admin client UX.
- Treat observability as a product capability, not just infrastructure.
- Keep role behavior aligned across backend authorization, frontend filtering,
  and scheduler execution semantics.
