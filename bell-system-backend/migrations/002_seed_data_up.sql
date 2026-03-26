-- Insert initial data for bell schedule system

-- Insert sessions
INSERT INTO Sessions (Name, StartTime, EndTime) VALUES
    ('Morning Session', '06:45', '12:10'),
    ('Afternoon Session', '12:15', '18:10');

-- Insert admin user with hashed password 'admin123' (you should change this in production)
-- bcrypt hash for 'admin123'
INSERT INTO Users (Username, PasswordHash, Role) VALUES
    ('admin', '$2a$10$Leuj4M2jyvk18OOWo75/dOVyGYBjHztfnzpN.spyKuruap1NPDwRy', 'admin'),
    ('morning_user', '$2a$10$Leuj4M2jyvk18OOWo75/dOVyGYBjHztfnzpN.spyKuruap1NPDwRy', 'morning_user'),
    ('afternoon_user', '$2a$10$Leuj4M2jyvk18OOWo75/dOVyGYBjHztfnzpN.spyKuruap1NPDwRy', 'afternoon_user');
