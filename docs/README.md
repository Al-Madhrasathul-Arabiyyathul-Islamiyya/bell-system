# Bell Schedule System — Documentation

## Directory Structure

```
docs/
├── api/                    # API specifications
│   ├── openapi/            # Modular OpenAPI 3.0 spec (JSON:API v1.1)
│   │   ├── openapi.yaml    # Root spec with path references
│   │   ├── parameters.yaml # Reusable query parameters
│   │   ├── responses.yaml  # Reusable error responses
│   │   ├── paths/          # Endpoint definitions per resource
│   │   └── schemas/        # JSON:API schemas, requests, exceptions
│   ├── jsonapi-v1.1-design-note.md  # JSON:API v1.1 API contract and migration notes
│   ├── websocket-protocol.md  # WebSocket events and message formats
│   └── websocket-client-guide.md  # Client integration guide for WebSocket
├── architecture/           # System design documents
│   ├── system-overview.md  # Full system documentation and requirements
│   └── implementation-plan.md  # Backend package structure and build plan
├── database/               # Database documentation
│   └── schema-setup.sql    # Full SQL Server schema + seed data
└── design-assets/          # UI mockups and design files
    ├── bell-system-sketches.ai
    ├── bg.ai / bg.png
    ├── page-1.png / page-2.png
    ├── main.html
    └── tailwindcss.theme.js
```
