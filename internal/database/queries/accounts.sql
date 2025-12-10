-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 AND is_active = true;

-- name: GetAccountIncludingInactive :one
SELECT * FROM accounts
WHERE id = $1;

-- name: ListAccountsByFamily :many
SELECT * FROM accounts
WHERE family_id = $1 AND is_active = true
ORDER BY name;

-- name: ListAllAccountsByFamily :many
SELECT * FROM accounts
WHERE family_id = $1
ORDER BY name;

-- name: ListAccountsByType :many
SELECT * FROM accounts
WHERE family_id = $1 AND type = $2 AND is_active = true
ORDER BY name;

-- name: CreateAccount :one
INSERT INTO accounts (
    id, family_id, name, type, currency, initial_balance, current_balance, description
) VALUES (
             $1, $2, $3, $4, $5, $6, $6, $7
         )
RETURNING *;

-- name: UpdateAccount :one
UPDATE accounts
SET
    name = COALESCE($3, name),
    description = $4,
    is_active = COALESCE($5, is_active),
    updated_at = NOW()
WHERE id = $1 AND family_id = $2
RETURNING *;

-- name: DeleteAccount :exec
UPDATE accounts
SET is_active = false, updated_at = NOW()
WHERE id = $1;

-- name: GetAccountBalance :one
SELECT current_balance, currency FROM accounts
WHERE id = $1 AND is_active = true;

-- name: GetTotalBalanceByFamily :one
SELECT
    COALESCE(SUM(current_balance), 0)::numeric as total_balance,
    COUNT(*) as account_count
FROM accounts
WHERE family_id = $1 AND is_active = true;