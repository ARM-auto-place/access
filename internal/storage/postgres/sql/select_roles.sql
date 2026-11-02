SELECT user_id, role
FROM users_role
WHERE ($1::uuid IS NULL OR user_id > $1::uuid)
ORDER BY user_id
LIMIT $2;