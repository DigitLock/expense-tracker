-- ============================================================================
-- Migration: 009 create default categories function
-- Description: Function to create default income/expense categories for a family
-- Date: 2025-12-14
-- ============================================================================

BEGIN;

-- Function to create default categories for a new family
CREATE OR REPLACE FUNCTION create_default_categories(p_family_id UUID)
    RETURNS VOID
    LANGUAGE plpgsql
AS $$
DECLARE
    -- Parent category IDs (will be set after creation)
    v_income_parent_id UUID;
    v_expense_parent_id UUID;
BEGIN
    -- ========================================
    -- INCOME CATEGORIES
    -- ========================================

    -- Parent: Salary & Wages
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Salary & Wages', 'income', NULL, true)
    RETURNING id INTO v_income_parent_id;

    -- Children under Salary & Wages
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Monthly Salary', 'income', v_income_parent_id, true),
        (p_family_id, 'Bonus', 'income', v_income_parent_id, true),
        (p_family_id, 'Overtime', 'income', v_income_parent_id, true);

    -- Parent: Business Income
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Business Income', 'income', NULL, true)
    RETURNING id INTO v_income_parent_id;

    -- Children under Business Income
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Freelance', 'income', v_income_parent_id, true),
        (p_family_id, 'Consulting', 'income', v_income_parent_id, true);

    -- Parent: Investments
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Investments', 'income', NULL, true)
    RETURNING id INTO v_income_parent_id;

    -- Children under Investments
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Dividends', 'income', v_income_parent_id, true),
        (p_family_id, 'Interest', 'income', v_income_parent_id, true);

    -- Standalone income categories
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Gifts Received', 'income', NULL, true),
        (p_family_id, 'Other Income', 'income', NULL, true);

    -- ========================================
    -- EXPENSE CATEGORIES
    -- ========================================

    -- Parent: Food & Dining
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Food & Dining', 'expense', NULL, true)
    RETURNING id INTO v_expense_parent_id;

    -- Children under Food & Dining
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Groceries', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Restaurants', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Food Delivery', 'expense', v_expense_parent_id, true);

    -- Parent: Transportation
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Transportation', 'expense', NULL, true)
    RETURNING id INTO v_expense_parent_id;

    -- Children under Transportation
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Gas', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Public Transport', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Parking', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Car Maintenance', 'expense', v_expense_parent_id, true);

    -- Parent: Housing
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Housing', 'expense', NULL, true)
    RETURNING id INTO v_expense_parent_id;

    -- Children under Housing
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Rent', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Utilities', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Internet & Phone', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Home Maintenance', 'expense', v_expense_parent_id, true);

    -- Parent: Shopping
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Shopping', 'expense', NULL, true)
    RETURNING id INTO v_expense_parent_id;

    -- Children under Shopping
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Clothing', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Electronics', 'expense', v_expense_parent_id, true),
        (p_family_id, 'General Shopping', 'expense', v_expense_parent_id, true);

    -- Parent: Entertainment
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Entertainment', 'expense', NULL, true)
    RETURNING id INTO v_expense_parent_id;

    -- Children under Entertainment
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Movies & Shows', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Sports & Hobbies', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Subscriptions', 'expense', v_expense_parent_id, true);

    -- Parent: Health & Fitness
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES (p_family_id, 'Health & Fitness', 'expense', NULL, true)
    RETURNING id INTO v_expense_parent_id;

    -- Children under Health & Fitness
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Medical', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Gym & Sports', 'expense', v_expense_parent_id, true),
        (p_family_id, 'Pharmacy', 'expense', v_expense_parent_id, true);

    -- Standalone expense categories
    INSERT INTO categories (family_id, name, type, parent_id, is_active)
    VALUES
        (p_family_id, 'Education', 'expense', NULL, true),
        (p_family_id, 'Gifts & Donations', 'expense', NULL, true),
        (p_family_id, 'Travel', 'expense', NULL, true),
        (p_family_id, 'Insurance', 'expense', NULL, true),
        (p_family_id, 'Other Expenses', 'expense', NULL, true);

    RAISE NOTICE 'Default categories created for family: %', p_family_id;
END;
$$;

-- Add comment
COMMENT ON FUNCTION create_default_categories(UUID) IS
    'Creates a comprehensive set of default income and expense categories for a family. Called automatically during user registration.';

COMMIT;