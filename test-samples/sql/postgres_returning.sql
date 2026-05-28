INSERT INTO users (email, name)
VALUES ('a@example.com', 'Alice')
RETURNING id, created_at;

UPDATE users SET name = 'Bob' WHERE id = 1 RETURNING id;
