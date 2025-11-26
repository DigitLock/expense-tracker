-- name: GetFamily :one
SELECT * FROM families
WHERE id = $1 AND is_active = true;

-- name: GetFamilyByName :one
SELECT * FROM families
WHERE name = $1 AND is_active = true;

-- name: ListFamilies :many
SELECT * FROM families
WHERE is_active = true
ORDER BY name;

-- name: CreateFamily :one
INSERT INTO families (
    id, name, base_currency
) VALUES (
             $1, $2, $3
         )
RETURNING *;

-- name: UpdateFamily :one
UPDATE families
SET
    name = $2,
    base_currency = $3,
    updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: DeleteFamily :exec
UPDATE families
SET is_active = false, updated_at = NOW()
WHERE id = $1;