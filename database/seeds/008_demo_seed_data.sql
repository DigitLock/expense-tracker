-- IMPORTANT: This is DEMO data for presentation/testing purposes
-- For production personal use, create separate seed data with your real info
--
--
-- 1. FAMILY
INSERT INTO families (id, name, base_currency)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Demo Family', 'RSD');


-- 2. USERS
-- Demo User 1 (Primary)
-- Email: demo@example.com
-- Password: Demo123! (bcrypt hash below)
INSERT INTO users (id, family_id, email, password_hash, name)
VALUES
    ('550e8400-e29b-41d4-a716-446655440001',
     '550e8400-e29b-41d4-a716-446655440000',
     'demo@example.com',
     '$$2a$12$3JswJWa4H4aNb6ORGl2bM.SOCwlhga0kpaofi9NNtYR36I7RLEJqW',
     'Demo User');

-- Demo User 2 (Family Member)
-- Email: demo2@example.com
-- Password: Demo123!
INSERT INTO users (id, family_id, email, password_hash, name)
VALUES
    ('550e8400-e29b-41d4-a716-446655440002',
     '550e8400-e29b-41d4-a716-446655440000',
     'demo2@example.com',
     '$$2a$12$3JswJWa4H4aNb6ORGl2bM.SOCwlhga0kpaofi9NNtYR36I7RLEJqW',
     'Demo Family Member');


-- 3. ACCOUNTS (Realistic balances)
-- Cash RSD
INSERT INTO accounts (id, family_id, name, type, currency, initial_balance, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440010',
     '550e8400-e29b-41d4-a716-446655440000',
     'Cash Wallet RSD',
     'cash',
     'RSD',
     15000.00,
     'Physical cash in wallet');

-- Checking Account
INSERT INTO accounts (id, family_id, name, type, currency, initial_balance, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440011',
     '550e8400-e29b-41d4-a716-446655440000',
     'Bank Card RSD',
     'checking',
     'RSD',
     120000.00,
     'Primary bank account');

-- Cash EUR
INSERT INTO accounts (id, family_id, name, type, currency, initial_balance, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440012',
     '550e8400-e29b-41d4-a716-446655440000',
     'Cash Wallet EUR',
     'cash',
     'EUR',
     300.00,
     'Euro cash at home');

-- Savings Account
INSERT INTO accounts (id, family_id, name, type, currency, initial_balance, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440013',
     '550e8400-e29b-41d4-a716-446655440000',
     'Savings Account',
     'savings',
     'RSD',
     500000.00,
     'Long-term savings');


-- 4. CATEGORIES - EXPENSE
-- Parent: Food
INSERT INTO categories (id, family_id, name, type, parent_id, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440020',
     '550e8400-e29b-41d4-a716-446655440000',
     'Food & Dining',
     'expense',
     NULL,
     'All food-related expenses');

-- Children of Food
INSERT INTO categories (family_id, name, type, parent_id, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Groceries', 'expense',
     '550e8400-e29b-41d4-a716-446655440020', 'Supermarket shopping'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Restaurants', 'expense',
     '550e8400-e29b-41d4-a716-446655440020', 'Dining out'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Food Delivery', 'expense',
     '550e8400-e29b-41d4-a716-446655440020', 'Delivery services');

-- Parent: Transport
INSERT INTO categories (id, family_id, name, type, parent_id, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440021',
     '550e8400-e29b-41d4-a716-446655440000',
     'Transport',
     'expense',
     NULL,
     'Transportation expenses');

-- Children of Transport
INSERT INTO categories (family_id, name, type, parent_id, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Gas', 'expense',
     '550e8400-e29b-41d4-a716-446655440021', 'Fuel for car'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Public Transport', 'expense',
     '550e8400-e29b-41d4-a716-446655440021', 'Bus, taxi, etc'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Parking', 'expense',
     '550e8400-e29b-41d4-a716-446655440021', 'Parking fees');

-- Parent: Housing
INSERT INTO categories (id, family_id, name, type, parent_id, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440022',
     '550e8400-e29b-41d4-a716-446655440000',
     'Housing',
     'expense',
     NULL,
     'Home-related expenses');

-- Children of Housing
INSERT INTO categories (family_id, name, type, parent_id, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Rent', 'expense',
     '550e8400-e29b-41d4-a716-446655440022', 'Monthly rent'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Utilities', 'expense',
     '550e8400-e29b-41d4-a716-446655440022', 'Electricity, water, gas'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Internet & Phone', 'expense',
     '550e8400-e29b-41d4-a716-446655440022', 'Communication services');

-- Standalone categories
INSERT INTO categories (family_id, name, type, parent_id, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Entertainment', 'expense',
     NULL, 'Movies, games, hobbies'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Shopping', 'expense',
     NULL, 'Clothes, electronics, etc'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Health & Fitness', 'expense',
     NULL, 'Medical, pharmacy, gym'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Education', 'expense',
     NULL, 'Courses, books, training');


-- 5. CATEGORIES - INCOME
INSERT INTO categories (family_id, name, type, parent_id, description)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Salary', 'income',
     NULL, 'Monthly salary'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Freelance', 'income',
     NULL, 'Freelance projects'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Investments', 'income',
     NULL, 'Investment returns'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Other Income', 'income',
     NULL, 'Misc income');


-- 6. SAMPLE TRANSACTIONS (последний месяц)
-- Session user for audit trail
SET LOCAL app.current_user_id = '550e8400-e29b-41d4-a716-446655440001';

-- INCOME: Salary
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440011',  -- Bank Card id
    (SELECT id FROM categories WHERE name = 'Salary'),
    'income',
    150000.00, 'RSD', 150000.00,
    'November salary',
    CURRENT_DATE - INTERVAL '25 days',
    '550e8400-e29b-41d4-a716-446655440001'
);

-- EXPENSE: Rent
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440011',  -- Bank Card id
    (SELECT id FROM categories WHERE name = 'Rent'),
    'expense',
    45000.00, 'RSD', 45000.00,
    'November rent payment',
    CURRENT_DATE - INTERVAL '24 days',
    '550e8400-e29b-41d4-a716-446655440001'
);

-- EXPENSE: Groceries (multiple transactions)
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440010',  -- Cash id
     (SELECT id FROM categories WHERE name = 'Groceries'),
     'expense',
     3500.00, 'RSD', 3500.00,
     'Weekly groceries at Maxi',
     CURRENT_DATE - INTERVAL '20 days',
     '550e8400-e29b-41d4-a716-446655440001'),

    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440010',
     (SELECT id FROM categories WHERE name = 'Groceries'),
     'expense',
     4200.00, 'RSD', 4200.00,
     'Groceries at IDEA',
     CURRENT_DATE - INTERVAL '13 days',
     '550e8400-e29b-41d4-a716-446655440002'),

    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440010',
     (SELECT id FROM categories WHERE name = 'Groceries'),
     'expense',
     3800.00, 'RSD', 3800.00,
     'Weekly groceries',
     CURRENT_DATE - INTERVAL '6 days',
     '550e8400-e29b-41d4-a716-446655440001');

-- EXPENSE: Restaurants
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440011',
     (SELECT id FROM categories WHERE name = 'Restaurants'),
     'expense',
     4500.00, 'RSD', 4500.00,
     'Dinner at restaurant',
     CURRENT_DATE - INTERVAL '15 days',
     '550e8400-e29b-41d4-a716-446655440001'),

    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440011',
     (SELECT id FROM categories WHERE name = 'Restaurants'),
     'expense',
     2800.00, 'RSD', 2800.00,
     'Lunch with friends',
     CURRENT_DATE - INTERVAL '8 days',
     '550e8400-e29b-41d4-a716-446655440002');

-- EXPENSE: Transport (Gas)
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440011',
     (SELECT id FROM categories WHERE name = 'Gas'),
     'expense',
     5000.00, 'RSD', 5000.00,
     'Gas station',
     CURRENT_DATE - INTERVAL '18 days',
     '550e8400-e29b-41d4-a716-446655440001'),

    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440011',
     (SELECT id FROM categories WHERE name = 'Gas'),
     'expense',
     4800.00, 'RSD', 4800.00,
     'Gas refill',
     CURRENT_DATE - INTERVAL '4 days',
     '550e8400-e29b-41d4-a716-446655440001');

-- EXPENSE: Utilities
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440011',
    (SELECT id FROM categories WHERE name = 'Utilities'),
    'expense',
    8500.00, 'RSD', 8500.00,
    'Electricity + Water + Gas',
    CURRENT_DATE - INTERVAL '22 days',
    '550e8400-e29b-41d4-a716-446655440001'
);

-- EXPENSE: Internet & Phone
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440011',
    (SELECT id FROM categories WHERE name = 'Internet & Phone'),
    'expense',
    3200.00, 'RSD', 3200.00,
    'Monthly internet + mobile',
    CURRENT_DATE - INTERVAL '23 days',
    '550e8400-e29b-41d4-a716-446655440001'
);

-- EXPENSE: Entertainment
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440010',
     (SELECT id FROM categories WHERE name = 'Entertainment'),
     'expense',
     1500.00, 'RSD', 1500.00,
     'Cinema tickets',
     CURRENT_DATE - INTERVAL '12 days',
     '550e8400-e29b-41d4-a716-446655440002'),

    ('550e8400-e29b-41d4-a716-446655440000',
     '550e8400-e29b-41d4-a716-446655440011',
     (SELECT id FROM categories WHERE name = 'Entertainment'),
     'expense',
     2500.00, 'RSD', 2500.00,
     'Steam games',
     CURRENT_DATE - INTERVAL '7 days',
     '550e8400-e29b-41d4-a716-446655440001');

-- EXPENSE: Shopping
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440011',
    (SELECT id FROM categories WHERE name = 'Shopping'),
    'expense',
    12000.00, 'RSD', 12000.00,
    'New jacket',
    CURRENT_DATE - INTERVAL '10 days',
    '550e8400-e29b-41d4-a716-446655440002'
);

-- EXPENSE: EUR transaction (multi-currency)
-- Курс: 1 EUR = 117.50 RSD
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440012',  -- EUR account id
    (SELECT id FROM categories WHERE name = 'Shopping'),
    'expense',
    50.00, 'EUR', 5875.00,  -- 50 * 117.50 = 5875
    'Online purchase in EUR',
    CURRENT_DATE - INTERVAL '5 days',
    '550e8400-e29b-41d4-a716-446655440001'
);

-- INCOME: Freelance
INSERT INTO transactions (
    family_id, account_id, category_id, type,
    amount, currency, amount_base,
    description, transaction_date, created_by
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440011',
    (SELECT id FROM categories WHERE name = 'Freelance'),
    'income',
    25000.00, 'RSD', 25000.00,
    'Freelance project payment',
    CURRENT_DATE - INTERVAL '9 days',
    '550e8400-e29b-41d4-a716-446655440001'
);
--
--
-- ============================================================================
-- DEMO CREDENTIALS
-- ============================================================================
--
-- For presentation/testing, provide these credentials:
-- Primary User:
--   Email: demo@example.com
--   Password: Demo123!
--
-- Secondary User:
--   Email: demo2@example.com
--   Password: Demo123!
--
-- Both users belong to "Demo Family" and can see the same data.
--
--
-- ============================================================================
-- NOTES
-- ============================================================================
-- 1. This seed creates realistic demo data showing:
--    ✅ Multi-user family
--    ✅ Multiple accounts (cash, checking, savings)
--    ✅ Multi-currency (RSD and EUR)
--    ✅ Hierarchical categories
--    ✅ Sample transactions (income & expense)
--    ✅ Automatic balance calculation
--    ✅ Audit trail
--
-- 2. Balances will be automatically calculated by triggers:
--    - Cash Wallet RSD: 15000 - 11500 = 3500
--    - Bank Card: 120000 + 175000 - 100300 = 194700
--    - Cash EUR: 300 - 50 = 250
--    - Savings: 500000 (no transactions)
--
-- 3. For production personal use:
--    - Create separate seed file with your real data
--    - Use environment variables for sensitive info
--    - Don't commit personal data to Git