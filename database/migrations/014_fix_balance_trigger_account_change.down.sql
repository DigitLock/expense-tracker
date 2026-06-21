-- ============================================================================
-- Rollback: 014 fix balance trigger on account change
-- Restores the migration-010 function body (recalculates only one account) and
-- drops the recalc helper. The trigger attachment is left untouched.
-- ============================================================================

BEGIN;

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
                                                                            WHEN t.type = 'income' THEN t.amount
                                                                            WHEN t.type = 'expense' THEN -t.amount
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
    'Recalculates account.current_balance using transaction amount in native currency. amount_base is used only for cross-currency reports.';

DROP FUNCTION IF EXISTS recalc_account_balance(UUID);

COMMIT;
