-- ============================================================================
-- Table: exchange_rates
-- Purpose: Historical currency exchange rates for multi-currency support
-- ============================================================================

BEGIN;

CREATE TABLE exchange_rates (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                from_currency VARCHAR(3) NOT NULL,
                                to_currency VARCHAR(3) NOT NULL,
                                rate DECIMAL(15, 6) NOT NULL,
                                date DATE NOT NULL,
                                source VARCHAR(100) NOT NULL DEFAULT 'manual',
                                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

                                CONSTRAINT exchange_rates_from_currency_check
                                    CHECK (from_currency IN ('RSD', 'EUR')),
                                CONSTRAINT exchange_rates_to_currency_check
                                    CHECK (to_currency IN ('RSD', 'EUR')),
                                CONSTRAINT exchange_rates_rate_positive
                                    CHECK (rate > 0),
                                CONSTRAINT exchange_rates_currencies_different
                                    CHECK (from_currency != to_currency),
                                CONSTRAINT exchange_rates_unique_date
                                    UNIQUE (from_currency, to_currency, date)
);

CREATE INDEX idx_exchange_rates_date
    ON exchange_rates(date DESC);
CREATE INDEX idx_exchange_rates_currencies_date
    ON exchange_rates(from_currency, to_currency, date DESC);

COMMENT ON TABLE exchange_rates IS
    'Historical currency exchange rates for multi-currency support. Updated daily via external API.';
COMMENT ON COLUMN exchange_rates.from_currency IS
    'Source currency code. MVP: RSD or EUR only.';
COMMENT ON COLUMN exchange_rates.to_currency IS
    'Target currency code. MVP: RSD or EUR only.';
COMMENT ON COLUMN exchange_rates.rate IS
    'Exchange rate: 1 from_currency = rate * to_currency. Example: 1 EUR = 117.50 RSD. DECIMAL(15,6) for high precision.';
COMMENT ON COLUMN exchange_rates.date IS
    'Date for which this rate is valid. One rate per currency pair per day.';
COMMENT ON COLUMN exchange_rates.source IS
    'Source of exchange rate. Examples: "exchangerate-api.com", "fixer.io", "manual", "central-bank-serbia"';
COMMENT ON COLUMN exchange_rates.created_at IS
    'Timestamp when rate was inserted into database. Not the same as rate.date!';

COMMIT;