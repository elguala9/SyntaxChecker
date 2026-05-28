SELECT a.id, a.email
FROM accounts AS a
WHERE a.balance > 100.0
GROUP ORDER BY a.email;
