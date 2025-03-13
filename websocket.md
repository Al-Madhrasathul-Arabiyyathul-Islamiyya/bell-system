# WebSocket Events Documentation

## Overview
WebSocket provides real-time communication between the server and client applications. The Bell Schedule System uses WebSockets for live updates to schedules, system state changes, and bell events.

## Connection

WebSocket endpoint: `ws://{base_url}/ws`

Authentication is required via query parameters:
- Token: `ws://{base_url}/ws?token={jwt_token}`
- Client type: `&client_type=admin|client`
- Client name (optional): `&client_name=Main Hall Display`

Upon connection, clients must identify their type and optionally provide a name for easier identification.

## Connection Management

### Initial Connection

When a client connects, it should immediately send a registration message:

```json
{
  "type": "register",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "client_type": "admin" | "client",
    "client_name": "Optional display name"
  }
}
```



### Server → Client Events

All events follow a standard format:
```json
{
  "type": "event_type",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {}
}
```

### connected_clients

    - Sent to admin connections when any client connects or disconnects
    - Provides a list of all connected clients for monitoring

```json
{
  "type": "connected_clients",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "clients": [
      {
        "id": "connection-uuid",
        "ip": "192.168.1.100",
        "client_type": "client",
        "client_name": "Main Hall Display",
        "connected_since": "2024-03-13T14:20:30Z"
      },
      {
        "id": "connection-uuid",
        "ip": "192.168.1.101",
        "client_type": "admin",
        "client_name": "Admin Panel",
        "connected_since": "2024-03-13T15:15:22Z"
      }
    ]
  }
}
```

### connection_acknowledged

    - Sent to a client immediately after successful connection and registration
    - Confirms connection and provides client's ID for reference

```json
{
  "type": "connection_acknowledged",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "connection_id": "connection-uuid",
    "client_type": "admin" | "client",
    "message": "Connection successful"
  }
}
```

### Schedule Events

1. schedules_updated

    - Sent when any schedule item is created, updated, or deleted
    - Clients should refresh their schedule data

```json
{
  "type": "schedules_updated",
  "timestamp": "2024-03-13T15:30:45Z"
}
```

2. bell_triggered

    - Sent when a bell is about to play
    - Allows client to prepare UI or handle playback

```json
{
  "type": "bell_triggered",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "scheduleItemId": "uuid",
    "soundId": "uuid",
    "name": "First Period"
  }
}
```

3. bell_cancelled

    - Sent when a scheduled bell has been cancelled

```json
{
  "type": "bell_cancelled",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "scheduleItemId": "uuid"
  }
}
```


### System State Events

1. system_state_changed

    - Sent when the bell system is paused or resumed

```json
{
  "type": "system_state_changed",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "state": "active" | "paused"
  }
}
```

2. audio_files_updated

    - Sent when any system audio file is added, updated, or deleted
    - Clients should refresh their audio file data and checksums

```json
{
  "type": "audio_files_updated",
  "timestamp": "2024-03-13T15:30:45Z"
}
```

### Admin Events

1. system_log

    - Admin-only event for real-time logging
    - Useful for monitoring system activity

```json
{
  "type": "system_log",
  "timestamp": "2024-03-13T15:30:45Z",
  "payload": {
    "level": "info" | "warning" | "error",
    "message": "System message here",
    "source": "component_name"
  }
}
```

The server will record the client's IP address, connection time, and provided details.

## Client → Server Events
Typically, most client actions are handled via REST API calls rather than WebSocket messages. However, for certain real-time actions, the following events can be used:

1. heartbeat
    - Sent periodically to keep the connection alive

```json
{
  "type": "heartbeat",
  "timestamp": "2024-03-13T15:30:45Z"
}
```

