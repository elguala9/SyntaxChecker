WITH recent_orders AS (
  SELECT user_id, COUNT(*) AS n
  FROM orders
  WHERE created_at > now() - interval '30 days'
  GROUP BY user_id
)
SELECT u.id, u.email, r.n
FROM users u
JOIN recent_orders r ON r.user_id = u.id
WHERE r.n > 5;
