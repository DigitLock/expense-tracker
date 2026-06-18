-- ============================================================================
-- Migration: 012 exchange_rates provenance
-- Purpose: Support syncing rates from the currency-rate-service:
--          add fetched_at (when the rate was fetched) and widen the currency
--          CHECKs to allow USD. `source` already exists (006) — not re-added.
--          The unique key (from_currency, to_currency, date) already exists
--          as exchange_rates_unique_date (006) — not re-created.
-- ============================================================================

BEGIN;

-- Provenance: when this rate was fetched from the upstream service (UTC).
ALTER TABLE exchange_rates ADD COLUMN fetched_at TIMESTAMPTZ;

COMMENT ON COLUMN exchange_rates.fetched_at IS
    'When this rate was fetched from the currency-rate-service (UTC). NULL for legacy/manual rows.';

-- Widen currency CHECKs to include USD; the sync writes USD->RSD rows.
ALTER TABLE exchange_rates DROP CONSTRAINT exchange_rates_from_currency_check;
ALTER TABLE exchange_rates ADD CONSTRAINT exchange_rates_from_currency_check
    CHECK (from_currency IN ('RSD', 'EUR', 'USD'));

ALTER TABLE exchange_rates DROP CONSTRAINT exchange_rates_to_currency_check;
ALTER TABLE exchange_rates ADD CONSTRAINT exchange_rates_to_currency_check
    CHECK (to_currency IN ('RSD', 'EUR', 'USD'));

COMMIT;
