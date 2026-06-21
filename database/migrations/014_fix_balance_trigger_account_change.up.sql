-- ============================================================================
-- Migration: 014 fix balance trigger on account change
-- Purpose: When a transaction's account_id changes on UPDATE, recalculate BOTH
--          the old and the new account. The migration-010 trigger recalculated
--          only NEW.account_id, leaving the old account's balance stale.
--
-- Only the function body changes; the existing trigger
-- (trigger_transactions_update_balance, AFTER INSERT OR UPDATE OR DELETE) is
-- left in place. A one-time recalc at the end heals any accumulated drift.
-- ============================================================================

BEGIN;

-- Single-account recalc helper: initial_balance + SUM(active income - expense).
CREATE OR REPLACE FUNCTION recalc_account_balance(p_account_id UUID)
    RETURNS VOID AS $$
BEGIN
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
                                                         WHERE t.account_id = p_account_id
                                                           AND t.is_active = true
                                                     ), 0)
    WHERE id = p_account_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION recalc_account_balance(UUID) IS
    'Recalculates a single account current_balance = initial_balance + SUM(active income - expense) in native currency.';

-- Trigger function: recalc every affected account.
CREATE OR REPLACE FUNCTION update_account_balance()
    RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        PERFORM recalc_account_balance(NEW.account_id);
    ELSIF TG_OP = 'DELETE' THEN
        PERFORM recalc_account_balance(OLD.account_id);
    ELSE -- UPDATE
        PERFORM recalc_account_balance(NEW.account_id);
        IF OLD.account_id IS DISTINCT FROM NEW.account_id THEN
            PERFORM recalc_account_balance(OLD.account_id);
        END IF;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_account_balance() IS
    'Recalculates affected account balances on transaction INSERT/UPDATE/DELETE. On UPDATE that moves a transaction between accounts, both old and new accounts are recalculated.';

-- One-time heal: recalculate every account to clear any stale balances left by
-- the previous trigger behavior.
UPDATE accounts a
SET current_balance = a.initial_balance + COALESCE((
                                                       SELECT SUM(
                                                                      CASE
                                                                          WHEN t.type = 'income' THEN t.amount
                                                                          WHEN t.type = 'expense' THEN -t.amount
                                                                          ELSE 0
                                                                          END
                                                              )
                                                       FROM transactions t
                                                       WHERE t.account_id = a.id
                                                         AND t.is_active = true
                                                   ), 0);

COMMIT;
