-- name: GetCategory :one
SELECT * FROM categories
WHERE id = $1 AND is_active = true;

-- name: GetCategoryIncludingInactive :one
SELECT * FROM categories
WHERE id = $1;

-- name: ListCategoriesByFamily :many
SELECT * FROM categories
WHERE family_id = $1 AND is_active = true
ORDER BY type, name;

-- name: ListAllCategoriesByFamily :many
SELECT * FROM categories
WHERE family_id = $1
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

-- name: SeedDefaultCategories :exec
-- Seeds the standard income/expense category tree for a new family.
-- Delegates to the create_default_categories() DB function (migration 009).
SELECT create_default_categories($1);

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
    is_active = $4,
    updated_at = NOW()
WHERE id = $1 AND family_id = $5
RETURNING *;

-- name: DeleteCategory :exec
UPDATE categories
SET is_active = false, updated_at = NOW()
WHERE id = $1;

-- name: FindInactiveCategoryByName :one
-- Find inactive category with matching name, type, and parent_id
SELECT id, family_id, name, type, parent_id, description, created_at, updated_at, is_active
FROM categories
WHERE family_id = sqlc.arg(family_id)
  AND LOWER(name) = LOWER(sqlc.arg(name))
  AND type = sqlc.arg(type)
  AND is_active = false
  AND (
    (sqlc.narg(parent_id)::uuid IS NULL AND parent_id IS NULL) OR
    (parent_id = sqlc.narg(parent_id))
    )
ORDER BY updated_at DESC
LIMIT 1;

-- name: RestoreCategory :one
-- Reactivate a soft-deleted category
UPDATE categories
SET is_active = true,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND is_active = false
RETURNING id, family_id, name, type, parent_id, description, created_at, updated_at, is_active;