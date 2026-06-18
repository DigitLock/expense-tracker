-- ============================================================================
-- Rollback: 012 exchange_rates provenance
-- Reverse order of the up migration. USD rows must be removed before the
-- narrow {RSD, EUR} CHECK can be restored. fetched_at is dropped last.
-- ============================================================================

BEGIN;

-- Remove USD rows so the narrow CHECK can be re-applied.
DELETE FROM exchange_rates WHERE from_currency = 'USD' OR to_currency = 'USD';

ALTER TABLE exchange_rates DROP CONSTRAINT exchange_rates_to_currency_check;
ALTER TABLE exchange_rates ADD CONSTRAINT exchange_rates_to_currency_check
    CHECK (to_currency IN ('RSD', 'EUR'));

ALTER TABLE exchange_rates DROP CONSTRAINT exchange_rates_from_currency_check;
ALTER TABLE exchange_rates ADD CONSTRAINT exchange_rates_from_currency_check
    CHECK (from_currency IN ('RSD', 'EUR'));

ALTER TABLE exchange_rates DROP COLUMN fetched_at;

COMMIT;
