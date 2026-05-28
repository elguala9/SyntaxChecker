CREATE TABLE accounts (
  id      SERIAL PRIMARY KEY,
  email   TEXT NOT NULL UNIQUE,
  balance NUMERIC(12,2) DEFAULT 0.0
);

SELECT a.id, a.email
FROM accounts AS a
WHERE a.balance > 100.0
ORDER BY a.email;
