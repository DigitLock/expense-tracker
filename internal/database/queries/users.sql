-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 AND is_active = true;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND is_active = true;

-- name: ListUsersByFamily :many
SELECT * FROM users
WHERE family_id = $1 AND is_active = true
ORDER BY name;

-- name: CreateUser :one
INSERT INTO users (
    id, family_id, email, name, password_hash
) VALUES (
             $1, $2, $3, $4, $5
         )
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET
    name = $2,
    email = $3,
    updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET
    password_hash = $2,
    updated_at = NOW()
WHERE id = $1 AND is_active = true;

-- name: DeleteUser :exec
UPDATE users
SET is_active = false, updated_at = NOW()
WHERE id = $1;