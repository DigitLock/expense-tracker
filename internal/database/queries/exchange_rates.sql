-- name: GetExchangeRate :one
SELECT * FROM exchange_rates
WHERE from_currency = $1
  AND to_currency = $2
  AND date = $3;

-- name: GetLatestExchangeRate :one
SELECT * FROM exchange_rates
WHERE from_currency = $1
  AND to_currency = $2
  AND date <= $3
ORDER BY date DESC
LIMIT 1;

-- name: ListExchangeRatesByDate :many
SELECT * FROM exchange_rates
WHERE date = $1
ORDER BY from_currency, to_currency;

-- name: ListExchangeRatesHistory :many
SELECT * FROM exchange_rates
WHERE from_currency = $1
  AND to_currency = $2
  AND date >= $3
  AND date <= $4
ORDER BY date DESC;

-- name: CreateExchangeRate :one
INSERT INTO exchange_rates (
    id, from_currency, to_currency, rate, date
) VALUES (
             $1, $2, $3, $4, $5
         )
RETURNING *;

-- name: UpsertExchangeRate :one
INSERT INTO exchange_rates (
    id, from_currency, to_currency, rate, date
) VALUES (
             $1, $2, $3, $4, $5
         )
ON CONFLICT (from_currency, to_currency, date)
    DO UPDATE SET rate = EXCLUDED.rate
RETURNING *;