-- ============================================================================
-- Table: accounts
-- Purpose: Financial accounts for cash and bank accounts
-- ============================================================================

BEGIN;

CREATE TABLE accounts (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          family_id UUID NOT NULL,
                          name VARCHAR(255) NOT NULL,
                          type VARCHAR(50) NOT NULL,
                          currency VARCHAR(3) NOT NULL,
                          initial_balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
                          current_balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
                          description TEXT,
                          created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          is_active BOOLEAN NOT NULL DEFAULT true,

                          CONSTRAINT fk_accounts_family
                              FOREIGN KEY (family_id)
                                  REFERENCES families(id)
                                  ON DELETE CASCADE,

                          CONSTRAINT accounts_type_check
                              CHECK (type IN ('cash', 'checking', 'savings')),

                          CONSTRAINT accounts_currency_check
                              CHECK (currency IN ('RSD', 'EUR')),

                          CONSTRAINT accounts_name_length
                              CHECK (LENGTH(name) >= 1)
);

CREATE INDEX idx_accounts_family
    ON accounts(family_id);
CREATE INDEX idx_accounts_type
    ON accounts(type);
CREATE INDEX idx_accounts_family_active
    ON accounts(family_id, is_active)
    WHERE is_active = true;
CREATE INDEX idx_accounts_currency
    ON accounts(currency);

COMMENT ON TABLE accounts IS
    'Financial accounts for cash and bank accounts. Each account belongs to a family and has a type and currency.';
COMMENT ON COLUMN accounts.id IS
    'UUID primary key. Generated automatically.';
COMMENT ON COLUMN accounts.family_id IS
    'Foreign key to families. ON DELETE CASCADE.';
COMMENT ON COLUMN accounts.name IS
    'User-defined account name. Examples: "My Wallet", "Salary Card", "Vacation Savings"';
COMMENT ON COLUMN accounts.type IS
    'Account type: cash (physical cash), checking (current/debit account), savings (savings account/deposit)';
COMMENT ON COLUMN accounts.currency IS
    'Account currency. MVP: RSD or EUR only. Post-MVP: any currency.';
COMMENT ON COLUMN accounts.initial_balance IS
    'Starting balance when account was created. Does NOT change after creation. Used for balance calculations.';
COMMENT ON COLUMN accounts.current_balance IS
    'Current account balance. AUTOMATICALLY calculated by trigger based on transactions. Formula: initial_balance + SUM(income) - SUM(expense). DO NOT update manually!';
COMMENT ON COLUMN accounts.description IS
    'Optional description/notes about the account';
COMMENT ON COLUMN accounts.created_at IS
    'Timestamp when account was created. Set automatically.';
COMMENT ON COLUMN accounts.updated_at IS
    'Timestamp of last update. Updated automatically by trigger.';
COMMENT ON COLUMN accounts.is_active IS
    'Soft delete flag. true = active account, false = closed account (data preserved for history)';

CREATE TRIGGER trigger_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TRIGGER trigger_accounts_updated_at ON accounts IS
    'Updates updated_at timestamp automatically on every UPDATE';

COMMIT;