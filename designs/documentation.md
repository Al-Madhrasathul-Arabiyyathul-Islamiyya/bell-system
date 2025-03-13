# Bell Schedule System Documentation

## System Overview
A bell schedule automation system for school use. The system handles automated bell schedules for morning and afternoon sessions, and provides the National Anthem and School Song on demand. The backend provides the API while a native Windows client handles the playback and user interface.

## Core Components

### 1. Schedule Management
- Two sessions: Morning and Afternoon
- Different schedules for each day of the week
- Real-time schedule updates via WebSocket
- Automatic bell playback at scheduled times
- Next bell countdown/indicator

### 2. System Audio Management (Admin Only)
- Upload and manage bell sounds
- Upload and manage National Anthem
- Upload and manage School Song
- Files stored in application public directory
- Metadata stored in database
- File checksums for synchronization

### 3. User Management
- Three user types:
  - Administrator (full system access)
  - Morning Session User (limited access)
  - Afternoon Session User (limited access)
- Admin interface requires authentication

## Access Control & Permissions

### Admin Interface
1. Administrator (Full Access)
   - Manage all schedule items (session-bound and session-independent)
   - Manage all system audio files
   - Full user management
   - No time restrictions on schedule creation

2. Session Users (Limited Access)
   - Can only manage schedules for their assigned session
   - Can only create schedules within their session's time frame
   - Can only select from bell sounds (not other system sounds)
   - Cannot create session-independent schedules

## Operational Features

### Schedule Management
1. Time Constraints
   - Session users: Can only create schedules within their session hours
   - Admin: No time restrictions

2. Bell System Control
   - Global pause/resume for scheduled bells
     - Useful during:
       - Examinations
       - Special events
       - Emergency situations
   - Individual bell cancellation
     - Allows stopping wrong/unneeded bells
     - Does not affect future scheduled bells

### Real-time Updates
- Schedule changes
- Bell system status (active/paused)
- Current playback status

## Technical Specifications

### Database Schema (SQL Server)
```sql
-- Users for admin interface
CREATE TABLE Users (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Username NVARCHAR(50) UNIQUE NOT NULL,
    PasswordHash NVARCHAR(255) NOT NULL,
    Role NVARCHAR(20) CHECK (Role IN ('admin', 'morning_user', 'afternoon_user')),
    CreatedAt DATETIME2 DEFAULT GETDATE()
)

-- Sessions (Morning/Afternoon)
CREATE TABLE Sessions (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Name NVARCHAR(50) UNIQUE NOT NULL,
    StartTime TIME NOT NULL,
    EndTime TIME NOT NULL
)

-- System Audio Files (bells, anthem, school song, other)
CREATE TABLE SystemAudioFiles (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Name NVARCHAR(255) NOT NULL,
    FilePath NVARCHAR(500) NOT NULL,
    FileType NVARCHAR(20) CHECK (FileType IN ('bell', 'anthem', 'school_song', 'other')),
    Checksum NVARCHAR(64) NOT NULL,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2
)

-- Schedule items
CREATE TABLE ScheduleItems (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    SessionId UNIQUEIDENTIFIER NULL FOREIGN KEY REFERENCES Sessions(Id),
    Name NVARCHAR(100) NOT NULL,
    Time TIME NOT NULL,
    SoundId UNIQUEIDENTIFIER FOREIGN KEY REFERENCES SystemAudioFiles(Id),
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2
)

-- Days that a schedule item applies to
CREATE TABLE ScheduleDays (
    ScheduleItemId UNIQUEIDENTIFIER FOREIGN KEY REFERENCES ScheduleItems(Id),
    DayOfWeek TINYINT CHECK (DayOfWeek BETWEEN 1 AND 7),
    PRIMARY KEY (ScheduleItemId, DayOfWeek)
)
```

### Schedule Management Details

- Schedule items can be:
   - Session-specific (Morning/Afternoon)
   - General (no session, e.g., school opening/closing)
- Each schedule item has:
   - Name (serves as description)
   - Time
   - Associated sound
   - Applicable days of the week

### System Audio Types

- bell: Bell sounds for schedule items
- anthem: National Anthem
- school_song: School Song
- other: Additional system sounds

### Admin Interface

- When creating schedule items:
   - Session selection is optional
   - Sound selection shows:
      - For users: Only bell sounds
      - For admin: Bell sounds and other system sounds

## Architecture

### Backend (Go):

- Lightweight HTTP server
- WebSocket for real-time updates
- SQL Server connection
- File system operations for system audio files
- API endpoints as specified in open-api.yml

### Client (Native Windows):

- Built with [Windows technology of choice, e.g., WPF, UWP, WinUI]
- Runs in kiosk mode with auto-login
- Handles schedule display and audio playback
- Connects to backend API for data
- Connects to WebSocket for real-time updates
- Local SQLite database for caching

### Configuration Requirements

1. System Setup:
   - SQL Server connection details
   - System audio directory path
   - WebSocket settings
   - HTTP server settings

2. Client Requirements:
   - Previous generation hardware (4GB RAM, 2 cores)
   - Network access to backend server
   - Network access to shared media folder
   - Audio output capability
   - Windows OS

## System States & Error Handling

### Critical States to Monitor

1. System Health

   - Database connection
   - File system access
   - WebSocket connection
   - Audio system status

2. Error Scenarios

   - Audio playback failures
   - Schedule conflicts
   - File access issues

### Logging Requirements

   - System startup/shutdown
   - Schedule changes
   - Bell playback events
   - User actions in admin panel
   - Error events

## Deployment & Environment

### Configuration Settings

   - Database connection string
   - System audio directory path
   - Allowed file types
   - Session times
   - WebSocket settings

### Backup Considerations

   - Database backup strategy
   - System audio files backup
   - Schedule export/import functionality
