Create migration files for database tables
Implement basic API handlers for the core functionality
Add authentication middleware and JWT handling

Configuration (internal/config)

Load environment-specific configurations
Define configuration structures for:

Server settings (port, environment, etc.)
Database connection
Audio file storage paths
Session times
JWT settings
WebSocket settings



2. Database Layer (internal/database)

Connection setup and management
Migration utilities
Transaction handling
Repository implementations for:

Users
Sessions
SystemAudioFiles
ScheduleItems
ScheduleDays



3. Models (internal/models)

Data structures matching the database schema
JSON serialization tags
Validation logic
User model with password hashing

4. HTTP API Handlers (internal/handlers)

REST API routes implementation based on OpenAPI spec
Request/response structures
Input validation
Error responses

Specific handler groups:

Authentication handlers
Schedule management
Audio file management
System state management
User management

5. Middleware (internal/middleware)

Authentication middleware (JWT validation)
Role-based authorization
Request logging
Error handling
CORS support

6. WebSocket (internal/websocket)

Connection management
Client registration and tracking
Event broadcasting
Message handling based on websocket.md spec

7. Scheduler (internal/scheduler)

Bell schedule management
Timer implementation for bell triggering
Schedule recalculation on changes
Bell cancellation functionality

8. Storage (internal/storage)

Audio file management
File upload handling
Checksum calculation
File system operations

9. Application Service (internal/app)

Central coordination
Server initialization and routing
Dependency wiring
Graceful shutdown