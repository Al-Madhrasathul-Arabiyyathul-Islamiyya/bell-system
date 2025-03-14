-- Remove seed data
DELETE FROM Users WHERE Username IN ('admin', 'morning_user', 'afternoon_user');
DELETE FROM Sessions WHERE Name IN ('Morning Session', 'Afternoon Session');
