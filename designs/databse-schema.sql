-- Create Database
CREATE DATABASE BellScheduleDB;
GO

-- Create Login
CREATE LOGIN bell_schedule_user WITH PASSWORD = 'YourStrongPassword123!';
GO

USE BellScheduleDB;
GO

-- Create User and assign to database
CREATE USER bell_schedule_user FOR LOGIN bell_schedule_user;
GO

-- Grant permissions
ALTER ROLE db_datareader ADD MEMBER bell_schedule_user;
ALTER ROLE db_datawriter ADD MEMBER bell_schedule_user;
GO

-- Grant execute permissions for specific operations
GRANT CREATE TABLE TO bell_schedule_user;
GRANT ALTER ON SCHEMA::dbo TO bell_schedule_user;
GO

-- Create Tables
CREATE TABLE Users (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Username NVARCHAR(50) UNIQUE NOT NULL,
    PasswordHash NVARCHAR(255) NOT NULL,
    Role NVARCHAR(20) CHECK (Role IN ('admin', 'morning_user', 'afternoon_user')),
    CreatedAt DATETIME2 DEFAULT GETDATE()
);

CREATE TABLE Sessions (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Name NVARCHAR(50) UNIQUE NOT NULL,
    StartTime TIME NOT NULL,
    EndTime TIME NOT NULL
);

CREATE TABLE SystemAudioFiles (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Name NVARCHAR(255) NOT NULL,
    FilePath NVARCHAR(500) NOT NULL,
    FileType NVARCHAR(20) CHECK (FileType IN ('bell', 'anthem', 'school_song', 'other')),
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2
);

CREATE TABLE ScheduleItems (
    Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    SessionId UNIQUEIDENTIFIER NULL FOREIGN KEY REFERENCES Sessions(Id),
    Name NVARCHAR(100) NOT NULL,
    Time TIME NOT NULL,
    SoundId UNIQUEIDENTIFIER FOREIGN KEY REFERENCES SystemAudioFiles(Id),
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2
);

CREATE TABLE ScheduleDays (
    ScheduleItemId UNIQUEIDENTIFIER FOREIGN KEY REFERENCES ScheduleItems(Id),
    DayOfWeek TINYINT CHECK (DayOfWeek BETWEEN 1 AND 7),
    PRIMARY KEY (ScheduleItemId, DayOfWeek)
);

-- Insert test data
INSERT INTO Sessions (Name, StartTime, EndTime) VALUES
    ('Morning Session', '07:00', '12:00'),
    ('Afternoon Session', '12:00', '17:00');

-- Insert admin user with hashed password 'admin123' (you should change this in production)
INSERT INTO Users (Username, PasswordHash, Role) VALUES
    ('admin', '$2a$10$xVR.FqM8kq8tKHXh9WWqIe3faG3F8bFY6VUxmk1cqWRqWGQM8ZmXi', 'admin'),
    ('morning_user', '$2a$10$xVR.FqM8kq8tKHXh9WWqIe3faG3F8bFY6VUxmk1cqWRqWGQM8ZmXi', 'morning_user'),
    ('afternoon_user', '$2a$10$xVR.FqM8kq8tKHXh9WWqIe3faG3F8bFY6VUxmk1cqWRqWGQM8ZmXi', 'afternoon_user');
GO

-- Grant permissions to specific tables
GRANT SELECT, INSERT, UPDATE, DELETE ON Users TO bell_schedule_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON Sessions TO bell_schedule_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON SystemAudioFiles TO bell_schedule_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ScheduleItems TO bell_schedule_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ScheduleDays TO bell_schedule_user;
GO
