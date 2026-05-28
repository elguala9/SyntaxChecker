SELECT
  data->>'name'        AS name,
  data#>>'{address,city}' AS city
FROM events
WHERE data @> '{"active": true}'::jsonb
  AND (data->'tags') ? 'urgent';
