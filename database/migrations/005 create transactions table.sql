-- ============================================================================
-- Table: transactions
-- Purpose: Core financial transactions (income and expenses)
-- ============================================================================

BEGIN;

CREATE TABLE transactions (
                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              family_id UUID NOT NULL,
                              account_id UUID NOT NULL,
                              category_id UUID NOT NULL,
                              type VARCHAR(50) NOT NULL,
                              amount DECIMAL(15, 2) NOT NULL,
                              currency VARCHAR(3) NOT NULL,
                              amount_base DECIMAL(15, 2) NOT NULL,
                              description TEXT,
                              transaction_date DATE NOT NULL,
                              created_by UUID NOT NULL,
                              created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                              updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                              is_active BOOLEAN NOT NULL DEFAULT true,

                              CONSTRAINT fk_transactions_family
                                  FOREIGN KEY (family_id)
                                      REFERENCES families(id)
                                      ON DELETE CASCADE,
                              CONSTRAINT fk_transactions_account
                                  FOREIGN KEY (account_id)
                                      REFERENCES accounts(id)
                                      ON DELETE RESTRICT,
                              CONSTRAINT fk_transactions_category
                                  FOREIGN KEY (category_id)
                                      REFERENCES categories(id)
                                      ON DELETE RESTRICT,
                              CONSTRAINT fk_transactions_user
                                  FOREIGN KEY (created_by)
                                      REFERENCES users(id)
                                      ON DELETE RESTRICT,
                              CONSTRAINT transactions_type_check
                                  CHECK (type IN ('income', 'expense')),
                              CONSTRAINT transactions_currency_check
                                  CHECK (currency IN ('RSD', 'EUR')),
                              CONSTRAINT transactions_amount_positive
                                  CHECK (amount > 0),
                              CONSTRAINT transactions_amount_base_positive
                                  CHECK (amount_base > 0),
                              CONSTRAINT transactions_date_not_future
                                  CHECK (transaction_date <= CURRENT_DATE)
);

CREATE INDEX idx_transactions_family
    ON transactions(family_id);
CREATE INDEX idx_transactions_account
    ON transactions(account_id);
CREATE INDEX idx_transactions_category
    ON transactions(category_id);
CREATE INDEX idx_transactions_date
    ON transactions(transaction_date DESC);
CREATE INDEX idx_transactions_type
    ON transactions(type);
CREATE INDEX idx_transactions_created_by
    ON transactions(created_by);
CREATE INDEX idx_transactions_active
    ON transactions(family_id, is_active)
    WHERE is_active = true;
CREATE INDEX idx_transactions_family_date_type
    ON transactions(family_id, transaction_date DESC, type);

COMMENT ON TABLE transactions IS
    'Core financial transactions (income and expenses). Automatic balance calculation via trigger.';
COMMENT ON COLUMN transactions.amount IS
    'Transaction amount in original currency. Always positive (type determines income/expense).';
COMMENT ON COLUMN transactions.currency IS
    'Original transaction currency (RSD or EUR in MVP)';
COMMENT ON COLUMN transactions.amount_base IS
    'Amount converted to base currency (RSD) using exchange rate at transaction date. Used for reports and balance calculations.';
COMMENT ON COLUMN transactions.transaction_date IS
    'Date when transaction occurred (not when it was recorded). Cannot be in future.';
COMMENT ON COLUMN transactions.created_by IS
    'User who created this transaction. For audit trail.';

CREATE TRIGGER trigger_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE FUNCTION update_account_balance()
    RETURNS TRIGGER AS $$
DECLARE
    affected_account_id UUID;
BEGIN
    IF TG_OP = 'DELETE' THEN
        affected_account_id := OLD.account_id;
    ELSE
        affected_account_id := NEW.account_id;
    END IF;

    UPDATE accounts
    SET current_balance = initial_balance + COALESCE((
                                                         SELECT SUM(
                                                                        CASE
                                                                            WHEN t.type = 'income' THEN t.amount_base
                                                                            WHEN t.type = 'expense' THEN -t.amount_base
                                                                            ELSE 0
                                                                            END
                                                                )
                                                         FROM transactions t
                                                         WHERE t.account_id = affected_account_id
                                                           AND t.is_active = true
                                                     ), 0)
    WHERE id = affected_account_id;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
COMMENT ON FUNCTION update_account_balance() IS
    'Recalculates account.current_balance based on all active transactions. Called by trigger on transactions INSERT/UPDATE/DELETE.';

CREATE TRIGGER trigger_transactions_update_balance
    AFTER INSERT OR UPDATE OR DELETE ON transactions
    FOR EACH ROW
EXECUTE FUNCTION update_account_balance();

COMMENT ON TRIGGER trigger_transactions_update_balance ON transactions IS
    'Automatically recalculates account balance when transaction is inserted/updated/deleted';

COMMIT;