-- Create tables for bell schedule system

-- Users for admin interface
CREATE TABLE Users (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Username NVARCHAR(50) UNIQUE NOT NULL,
    PasswordHash NVARCHAR(255) NOT NULL,
    Role NVARCHAR(20) CHECK (Role IN ('admin', 'morning_user', 'afternoon_user')),
    CreatedAt DATETIME2 DEFAULT GETDATE()
);

-- Sessions (Morning/Afternoon)
CREATE TABLE Sessions (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Name NVARCHAR(50) UNIQUE NOT NULL,
    StartTime TIME NOT NULL,
    EndTime TIME NOT NULL
);

-- System Audio Files (bells, anthem, school song, other)
CREATE TABLE SystemAudioFiles (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Name NVARCHAR(255) NOT NULL,
    FilePath NVARCHAR(500) NOT NULL,
    FileType NVARCHAR(20) CHECK (FileType IN ('bell', 'anthem', 'school_song', 'other')),
    Checksum NVARCHAR(64) NOT NULL,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2
);

-- Schedule items
CREATE TABLE ScheduleItems (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    SessionId UNIQUEIDENTIFIER NULL FOREIGN KEY REFERENCES Sessions(Id),
    Name NVARCHAR(100) NOT NULL,
    Time TIME NOT NULL,
    SoundId UNIQUEIDENTIFIER FOREIGN KEY REFERENCES SystemAudioFiles(Id),
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2
);

-- Days that a schedule item applies to
CREATE TABLE ScheduleDays (
    ScheduleItemId UNIQUEIDENTIFIER FOREIGN KEY REFERENCES ScheduleItems(Id),
    DayOfWeek TINYINT CHECK (DayOfWeek BETWEEN 1 AND 7),
    PRIMARY KEY (ScheduleItemId, DayOfWeek)
);
