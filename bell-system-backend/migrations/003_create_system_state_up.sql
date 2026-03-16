CREATE TABLE SystemState (
    [Key] NVARCHAR(50) PRIMARY KEY,
    Value NVARCHAR(255) NOT NULL,
    UpdatedAt DATETIME2 NOT NULL DEFAULT GETDATE()
);
INSERT INTO SystemState ([Key], Value, UpdatedAt) VALUES ('system_state', 'active', GETDATE());
