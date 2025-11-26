-- name: GetCategory :one
SELECT * FROM categories
WHERE id = $1 AND is_active = true;

-- name: ListCategoriesByFamily :many
SELECT * FROM categories
WHERE family_id = $1 AND is_active = true
ORDER BY type, name;

-- name: ListCategoriesByType :many
SELECT * FROM categories
WHERE family_id = $1 AND type = $2 AND is_active = true
ORDER BY name;

-- name: ListRootCategories :many
SELECT * FROM categories
WHERE family_id = $1 AND parent_id IS NULL AND is_active = true
ORDER BY type, name;

-- name: ListChildCategories :many
SELECT * FROM categories
WHERE parent_id = $1 AND is_active = true
ORDER BY name;

-- name: CreateCategory :one
INSERT INTO categories (
    id, family_id, name, type, parent_id
) VALUES (
             $1, $2, $3, $4, $5
         )
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET
    name = $2,
    parent_id = $3,
    updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: DeleteCategory :exec
UPDATE categories
SET is_active = false, updated_at = NOW()
WHERE id = $1;