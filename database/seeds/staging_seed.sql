-- ============================================================================
-- Staging seed (v0.4.x) — idempotent.
--
-- Purpose: populate the isolated *staging* database with the same demo login as
-- 008_demo_seed_data.sql (so QA can sign in with demo@example.com / Demo123!),
-- plus a realistic 60-day transaction history for reports/QA.
--
-- Re-runnable: re-applying this file DELETEs this family's transactions,
-- accounts and categories (in FK order) and rebuilds them. It never DROPs and
-- never touches other families. Family + users are upserted (ON CONFLICT DO
-- NOTHING) so the login UUIDs stay stable across re-seeds.
--
-- Conventions enforced by the schema (see migrations 003/005/009):
--   * transactions.amount and amount_base are ALWAYS POSITIVE (CHECK > 0);
--     income vs expense is encoded by `type`. The balance trigger adds
--     amount_base for income and subtracts it for expense.
--   * accounts.current_balance is maintained by the trigger — never set here.
--   * categories come from create_default_categories(); never hand-rolled.
--
-- NOTE: password_hash is the real bcrypt hash of "Demo123!" written with a
--       single "$" so it stores correctly under a plain `psql < file` apply
--       (no $-expansion). 008 uses a doubled "$$" for its expanding pipeline.
-- ============================================================================

BEGIN;

-- Audit actor for the audit-log trigger (mirrors 008).
SET LOCAL app.current_user_id = '550e8400-e29b-41d4-a716-446655440001';

-- ----------------------------------------------------------------------------
-- 1. Family + users (stable UUIDs, upserted — same as 008)
-- ----------------------------------------------------------------------------
INSERT INTO families (id, name, base_currency)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'Demo Family', 'RSD')
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, family_id, email, password_hash, name)
VALUES
    ('550e8400-e29b-41d4-a716-446655440001',
     '550e8400-e29b-41d4-a716-446655440000',
     'demo@example.com',
     '$2a$12$3JswJWa4H4aNb6ORGl2bM.SOCwlhga0kpaofi9NNtYR36I7RLEJqW',
     'Demo User'),
    ('550e8400-e29b-41d4-a716-446655440002',
     '550e8400-e29b-41d4-a716-446655440000',
     'demo2@example.com',
     '$2a$12$3JswJWa4H4aNb6ORGl2bM.SOCwlhga0kpaofi9NNtYR36I7RLEJqW',
     'Demo Family Member')
ON CONFLICT (id) DO NOTHING;

-- ----------------------------------------------------------------------------
-- 2. Idempotent cleanup of this family's data (FK order: tx -> accounts -> cats)
-- ----------------------------------------------------------------------------
DELETE FROM transactions WHERE family_id = '550e8400-e29b-41d4-a716-446655440000';
DELETE FROM accounts     WHERE family_id = '550e8400-e29b-41d4-a716-446655440000';
DELETE FROM categories   WHERE family_id = '550e8400-e29b-41d4-a716-446655440000';

-- ----------------------------------------------------------------------------
-- 3. Categories — built by the canonical function (migration 009), not by hand
-- ----------------------------------------------------------------------------
SELECT create_default_categories('550e8400-e29b-41d4-a716-446655440000');

-- ----------------------------------------------------------------------------
-- 4. Accounts — one RSD, one EUR. Balances are computed by the trigger;
--    initial_balance = 0, current_balance left to its DEFAULT/trigger.
-- ----------------------------------------------------------------------------
INSERT INTO accounts (id, family_id, name, type, currency, initial_balance, description)
VALUES
    ('550e8400-e29b-41d4-a716-4466554400a1',
     '550e8400-e29b-41d4-a716-446655440000',
     'Staging Bank RSD', 'checking', 'RSD', 0.00, 'Primary staging account (RSD)'),
    ('550e8400-e29b-41d4-a716-4466554400a2',
     '550e8400-e29b-41d4-a716-446655440000',
     'Staging Cash EUR', 'cash', 'EUR', 0.00, 'Euro cash (staging)');

-- ----------------------------------------------------------------------------
-- 5. Transactions — ~177 rows over the last 60 days, dynamic dates from now().
--    Category ids are resolved by name from the rows created in step 3.
--    amount_base == amount for RSD; for EUR it is amount * 117.50 (fixed demo
--    rate, kept > 0 as required by the CHECK constraint).
-- ----------------------------------------------------------------------------

-- INCOME: Monthly salary (3 paydays, ~30 days apart) -> RSD account
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Monthly Salary' AND type = 'income' LIMIT 1),
       'income', v.amt, 'RSD', v.amt, v.descr, v.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (VALUES
    (156000.00::numeric, 'Monthly salary', (now() - interval '58 days')::date),
    (158000.00::numeric, 'Monthly salary', (now() - interval '28 days')::date),
    (160000.00::numeric, 'Monthly salary', (now() - interval '1 day')::date)
) AS v(amt, descr, dt);

-- INCOME: Freelance (3 random) -> RSD account
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Freelance' AND type = 'income' LIMIT 1),
       'income', g.amt, 'RSD', g.amt, 'Freelance project payment', g.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (
    SELECT round((18000 + random() * 27000)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 3)
) g;

-- INCOME: Consulting in EUR (2 payments) -> EUR account; amount_base = amount * 117.50
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a2'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Consulting' AND type = 'income' LIMIT 1),
       'income', v.amt, 'EUR', round(v.amt * 117.50, 2), 'Consulting payment (EUR)', v.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (VALUES
    (1800.00::numeric, (now() - interval '40 days')::date),
    (2200.00::numeric, (now() - interval '10 days')::date)
) AS v(amt, dt);

-- EXPENSE: Groceries (45) -> RSD; created_by mixed across the two users
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Groceries' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'RSD', g.amt, 'Groceries', g.dt,
       CASE WHEN random() < 0.35 THEN '550e8400-e29b-41d4-a716-446655440002'::uuid
            ELSE '550e8400-e29b-41d4-a716-446655440001'::uuid END
FROM (
    SELECT round((800 + random() * 5200)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 45)
) g;

-- EXPENSE: Restaurants (22) -> RSD; created_by mixed
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Restaurants' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'RSD', g.amt, 'Dining out', g.dt,
       CASE WHEN random() < 0.35 THEN '550e8400-e29b-41d4-a716-446655440002'::uuid
            ELSE '550e8400-e29b-41d4-a716-446655440001'::uuid END
FROM (
    SELECT round((1200 + random() * 5800)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 22)
) g;

-- EXPENSE: Food Delivery (18) -> RSD
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Food Delivery' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'RSD', g.amt, 'Food delivery', g.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (
    SELECT round((900 + random() * 2600)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 18)
) g;

-- EXPENSE: Public Transport (25) -> RSD
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Public Transport' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'RSD', g.amt, 'Bus / taxi', g.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (
    SELECT round((90 + random() * 310)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 25)
) g;

-- EXPENSE: Gas (10) -> RSD
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Gas' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'RSD', g.amt, 'Fuel', g.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (
    SELECT round((3000 + random() * 4000)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 10)
) g;

-- EXPENSE: Subscriptions (8) -> RSD
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Subscriptions' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'RSD', g.amt, 'Streaming / SaaS subscription', g.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (
    SELECT round((500 + random() * 2000)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 8)
) g;

-- EXPENSE: Movies & Shows (12) -> RSD
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Movies & Shows' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'RSD', g.amt, 'Cinema / entertainment', g.dt,
       CASE WHEN random() < 0.4 THEN '550e8400-e29b-41d4-a716-446655440002'::uuid
            ELSE '550e8400-e29b-41d4-a716-446655440001'::uuid END
FROM (
    SELECT round((600 + random() * 2400)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 12)
) g;

-- EXPENSE: Pharmacy (8) -> RSD
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Pharmacy' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'RSD', g.amt, 'Pharmacy', g.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (
    SELECT round((400 + random() * 3600)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 8)
) g;

-- EXPENSE: Utilities (2 fixed) -> RSD
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Utilities' AND type = 'expense' LIMIT 1),
       'expense', v.amt, 'RSD', v.amt, 'Electricity + water + gas', v.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (VALUES
    (8400.00::numeric, (now() - interval '28 days')::date),
    (9100.00::numeric, (now() - interval '4 days')::date)
) AS v(amt, dt);

-- EXPENSE: Internet & Phone (2 fixed) -> RSD
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a1'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Internet & Phone' AND type = 'expense' LIMIT 1),
       'expense', v.amt, 'RSD', v.amt, 'Internet + mobile', v.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (VALUES
    (3200.00::numeric, (now() - interval '27 days')::date),
    (3300.00::numeric, (now() - interval '3 days')::date)
) AS v(amt, dt);

-- EXPENSE: General Shopping (12) -> EUR account; amount_base = amount * 117.50
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a2'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'General Shopping' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'EUR', round(g.amt * 117.50, 2), 'Online shopping (EUR)', g.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (
    SELECT round((10 + random() * 140)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 12)
) g;

-- EXPENSE: Electronics (8) -> EUR account
INSERT INTO transactions (family_id, account_id, category_id, type, amount, currency, amount_base, description, transaction_date, created_by)
SELECT '550e8400-e29b-41d4-a716-446655440000'::uuid,
       '550e8400-e29b-41d4-a716-4466554400a2'::uuid,
       (SELECT id FROM categories WHERE family_id = '550e8400-e29b-41d4-a716-446655440000' AND name = 'Electronics' AND type = 'expense' LIMIT 1),
       'expense', g.amt, 'EUR', round(g.amt * 117.50, 2), 'Electronics (EUR)', g.dt,
       '550e8400-e29b-41d4-a716-446655440001'::uuid
FROM (
    SELECT round((30 + random() * 370)::numeric, 2) AS amt,
           (now() - ((random() * 60)::int) * interval '1 day')::date AS dt
    FROM generate_series(1, 8)
) g;

COMMIT;

-- ============================================================================
-- STAGING CREDENTIALS (same as demo): demo@example.com / Demo123!
-- Account balances are computed automatically by the transactions trigger.
-- ============================================================================
