-- ============================================================================
-- Rollback: Restore original account balance trigger using amount_base
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

COMMIT;
