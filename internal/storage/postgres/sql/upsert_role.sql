INSERT INTO users_role (user_id, role)
VALUES ($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET role = EXCLUDED.role;