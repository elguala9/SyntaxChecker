UPDATE users SET age = age + 1 WHERE id = 5;
DELETE FROM logs WHERE created_at < '2024-01-01';
REPLACE INTO settings (k, v) VALUES ('theme', 'dark');
