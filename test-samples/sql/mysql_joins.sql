SELECT u.id, u.name, o.total
FROM users u
INNER JOIN orders o ON o.user_id = u.id
LEFT JOIN payments p ON p.order_id = o.id
WHERE o.total > 100
GROUP BY u.id, u.name, o.total
HAVING COUNT(p.id) > 0
ORDER BY o.total DESC
LIMIT 10;
