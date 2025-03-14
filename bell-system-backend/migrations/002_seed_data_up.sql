-- Insert initial data for bell schedule system

-- Insert sessions
INSERT INTO Sessions (Name, StartTime, EndTime) VALUES
    ('Morning Session', '06:45', '12:10'),
    ('Afternoon Session', '12:15', '18:10');

-- Insert admin user with hashed password 'admin123' (you should change this in production)
-- bcrypt hash for 'admin123'
INSERT INTO Users (Username, PasswordHash, Role) VALUES
    ('admin', '$2a$10$xVR.FqM8kq8tKHXh9WWqIe3faG3F8bFY6VUxmk1cqWRqWGQM8ZmXi', 'admin'),
    ('morning_user', '$2a$10$xVR.FqM8kq8tKHXh9WWqIe3faG3F8bFY6VUxmk1cqWRqWGQM8ZmXi', 'morning_user'),
    ('afternoon_user', '$2a$10$xVR.FqM8kq8tKHXh9WWqIe3faG3F8bFY6VUxmk1cqWRqWGQM8ZmXi', 'afternoon_user');
