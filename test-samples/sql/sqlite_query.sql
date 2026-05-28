CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
INSERT INTO users (id, name) VALUES (1, 'Alice');
SELECT id, name FROM users WHERE id = 1 ORDER BY name;
