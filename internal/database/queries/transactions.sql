-- name: GetTransaction :one
SELECT * FROM transactions
WHERE id = $1 AND is_active = true;

-- name: ListTransactionsByFamily :many
SELECT * FROM transactions
WHERE family_id = $1 AND is_active = true
ORDER BY transaction_date DESC, created_at DESC;

-- name: ListTransactionsByAccount :many
SELECT * FROM transactions
WHERE account_id = $1 AND is_active = true
ORDER BY transaction_date DESC, created_at DESC;

-- name: ListTransactionsByCategory :many
SELECT * FROM transactions
WHERE category_id = $1 AND is_active = true
ORDER BY transaction_date DESC, created_at DESC;

-- name: ListTransactionsByDateRange :many
SELECT * FROM transactions
WHERE family_id = $1
  AND transaction_date >= $2
  AND transaction_date <= $3
  AND is_active = true
ORDER BY transaction_date DESC, created_at DESC;

-- name: ListTransactionsPaginated :many
SELECT * FROM transactions
WHERE family_id = $1 AND is_active = true
ORDER BY transaction_date DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTransaction :one
INSERT INTO transactions (
    id, family_id, account_id, category_id, type,
    amount, currency, amount_base, description, transaction_date, created_by
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
         )
RETURNING *;

-- name: UpdateTransaction :one
UPDATE transactions
SET
    category_id = $2,
    amount = $3,
    currency = $4,
    amount_base = $5,
    description = $6,
    transaction_date = $7,
    updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: DeleteTransaction :exec
UPDATE transactions
SET is_active = false, updated_at = NOW()
WHERE id = $1;

-- name: GetTransactionsSummaryByType :many
SELECT
    type,
    COUNT(*) as count,
    COALESCE(SUM(amount_base), 0)::numeric as total
FROM transactions
WHERE family_id = $1
  AND transaction_date >= $2
  AND transaction_date <= $3
  AND is_active = true
GROUP BY type;

-- name: GetTransactionsSummaryByCategory :many
SELECT
    category_id,
    COUNT(*) as count,
    COALESCE(SUM(amount_base), 0)::numeric as total
FROM transactions
WHERE family_id = $1
  AND type = $2
  AND transaction_date >= $3
  AND transaction_date <= $4
  AND is_active = true
GROUP BY category_id
ORDER BY total DESC;