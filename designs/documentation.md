# Bell Schedule System Documentation

## System Overview
A combined bell schedule automation system and media player for school use. The system handles automated bell schedules for morning and afternoon sessions, plays the National Anthem and School Song on demand, and includes a media player for general audio/video playback.

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

### 3. Media Player
- Reads directly from network share folder
- Supports common formats:
  - Audio: MP3, WAV, FLAC
  - Video: MP4, AVI, MKV, WEBM (H264/H265)
- Features:
  - Session playlist (temporary, not saved)
  - Repeat/shuffle options
  - Master volume control
  - Basic playback controls (play, pause, next, previous)

### 4. User Management
- Three user types:
  - Administrator (full system access)
  - Morning Session User (view only)
  - Afternoon Session User (view only)
- Client frontend unrestricted
- Admin interface requires authentication

## Access Control & Permissions

### Client Interface
1. Schedule Display
   - Only shows current session's schedule during active hours
   - No authentication required
   - Real-time updates via WebSocket

2. Audio Control Permissions
   - Cancel ongoing bell playback
   - Global bell schedule pause/resume
   - Full media player control

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
- File system operations for:
  - System audio files in app directory
  - Media files from network share

### Frontend (Next.js):

- Admin Interface:
  - Schedule management
  - System audio management
  - User authentication

- Client Interface:
  - Schedule display
  - Audio controls
  - Media player
  - Real-time updates

### Configuration Requirements
1. System Setup:
    - SQL Server connection details
    - Network share path
    - System audio directory path
    - WebSocket settings
    - HTTP server settings

2. Client Requirements:
   - Previous generation hardware (4GB RAM, 2 cores)
   - Network access to shared folder
   - Modern web browser
   - Audio output capability

## UI/UX Specifications

### Theme and Styling
[We should document the Tailwind theme configuration here once you share it]

### Layout Guidelines
1. Client Interface
   - Minimalist, focused design
   - Large, easily readable schedule display
   - Prominent next bell countdown
   - Clear audio controls
   - Media player area
   - Status indicators for:
     - Current session
     - Bell system status (active/paused)
     - Currently playing media

2. Admin Interface
   - Dashboard layout
   - Clear navigation between:
     - Schedule management
     - Audio file management
     - User management
   - Form validations and feedback

### Responsive Design
- Optimize for:
  - Desktop primary use
  - Large display screens
  - Minimum supported resolution: [Need to specify]

## System States & Error Handling

### Critical States to Monitor
1. System Health
   - Database connection
   - File system access
   - WebSocket connection
   - Audio system status

2. Error Scenarios
   - Network share unavailable
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
- Network share path
- Public directory path
- Allowed file types
- Session times
- WebSocket settings

### Backup Considerations
- Database backup strategy
- System audio files backup
- Schedule export/import functionality

## UI Theme Specifications

### Color Palette

#### Primary Colors
- Green (`#189543`): Main brand color, used for primary actions and key UI elements
- Red (`#DA231C`): Used for warnings, errors, and critical actions
- Accent (`#F5A623`): Used for highlighting and secondary actions

#### Secondary Colors
- Light Green (`#E8F5E9`): Background for success states and green-themed areas
- Light Red (`#FFEBEE`): Background for error states and warnings
- Gold (`#FFF3E0`): Background for accent elements

#### Neutral Colors
- Background (`#F8FAFC`): Main application background
- Surface (`#F1F5F9`): Cards and elevated surfaces
- Text (`#0F172A`): Primary text color

#### Dark Theme Colors
Primary:
- Green (`#1FB954`)
- Red (`#E53935`)
- Accent (`#FFB74D`)

Secondary:
- Dark Green (`#1B5E20`)
- Dark Red (`#B71C1C`)
- Dark Gold (`#FF8F00`)

Neutral:
- Background (`#121212`)
- Surface (`#1E1E1E`)
- Text (`#FFFFFF`)

### Typography

#### Font Families
1. Primary Font (Sans-Serif)
   - Font: Sofia Sans
   - Usage: Main UI text, buttons, inputs
   - Features: `"cv11", "ss01"`
   - Optical Size: 32

2. Secondary Font (Serif)
   - Font: Playfair Display
   - Usage: Headings, titles
   - Features: `"cv11", "ss01"`
   - Optical Size: 32

3. Monospace Font
   - Font: Geist Mono
   - Usage: Code, technical text

### Usage Guidelines

#### Client Interface
1. Header/Navigation
   - Background: neutral-surface
   - Text: neutral-text
   - Active items: primary-green

2. Schedule Display
   - Background: neutral-background
   - Current item: primary-green with secondary-green background
   - Text: neutral-text
   - Time indicators: primary-accent

3. Media Controls
   - Primary buttons: primary-green
   - Secondary buttons: neutral-surface with neutral-text
   - Progress bars: primary-accent
   - Volume control: primary-green

#### Admin Interface
1. Navigation
   - Background: neutral-surface
   - Active items: primary-green
   - Text: neutral-text

2. Forms
   - Input backgrounds: neutral-background
   - Borders: neutral-surface
   - Submit buttons: primary-green
   - Cancel buttons: primary-red
   - Validation: primary-red for errors

3. Tables/Lists
   - Headers: neutral-surface
   - Row hover: secondary-green
   - Active row: primary-green with secondary-green

4. Status Indicators
   - Success: primary-green
   - Error: primary-red
   - Warning: primary-accent
   - Info: neutral-text

### Dark Mode Considerations
- All color values automatically switch to their dark theme variants
- Maintain contrast ratios for accessibility
- Ensure readability of all text elements
- Preserve hierarchical relationships in the UI
