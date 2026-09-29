# Database Migrations

This directory contains SQL migration files for the Expense Tracker database schema.

## Migration Naming Convention

Migrations follow the pattern: `{version} {description}.sql` and `{version} {description} - drop.sql`

- `version`: Sequential number (001, 002, etc.)
- `description`: Brief description of what the migration does
- `.sql`: Forward migration (apply changes)
- ` - drop.sql`: Rollback migration (revert changes)

## Migration Files

### Schema (001–014; 008 intentionally absent — the demo seed lives in `database/seeds/`)

| # | Migration | Description | Status |
|---|-----------|-------------|--------|
| 001 | `create families table` | Root entity for multi-tenant isolation | ✅ |
| 002 | `create users table` | User authentication with bcrypt | ✅ |
| 003 | `create accounts table` | Financial accounts (cash/checking/savings) | ✅ |
| 004 | `create categories table` | Hierarchical income/expense categories | ✅ |
| 005 | `create transactions table` | Core transactions with auto-balance trigger | ✅ |
| 006 | `create exchange rates table` | Multi-currency support with rates | ✅ |
| 007 | `create audit log table` | Comprehensive audit trail with JSONB | ✅ |
| 009 | `create default categories function` | Function creating default income/expense categories for a family | ✅ |
| 010 | `fix account balance trigger` | Balance uses native `amount` instead of RSD-converted `amount_base` | ✅ |
| 011 | `add user role` | `role` column on users (family owner vs member) | ✅ |
| 012 | `exchange rates provenance` | `fetched_at` column; currency CHECKs widened to allow USD | ✅ |
| 013 | `account name unique` | Case-insensitive unique active account names per family | ✅ |
| 014 | `fix balance trigger account change` | Recalculate both old and new account when `account_id` changes | ✅ |

### Seed Data (`database/seeds/008_demo_seed_data.sql`)

| # | Migration | Description | Status |
|---|-----------|-------------|--------|
| 008 | `demo seed data` (`database/seeds/`, not a migration) | Demo data for portfolio/testing | ✅ |

**Demo Credentials:**
- Email: `demo@example.com`
- Password: `Demo123!`

## How to Use Migrations

### Apply migrations (in order):
```bash
001_create_families_table.up.sql
002_create_users_table.up.sql
003_create_accounts_table.up.sql
004_create_categories_table.up.sql
005_create_transactions_table.up.sql
006_create_exchange_rates_table.up.sql
007_create_audit_log_table.up.sql
009_create_default_categories_function.up.sql
010_fix_account_balance_trigger.up.sql
011_add_user_role.up.sql
012_exchange_rates_provenance.up.sql
013_account_name_unique.up.sql
014_fix_balance_trigger_account_change.up.sql
```

### Load seed data:
```bash
database/seeds/008_demo_seed_data.sql
```

### Rollback (in reverse order):
```bash
007 create audit log table - drop.sql
006 create exchange rates table - drop.sql
# ... and so on
```

### Verify Schema

```sql
-- List all tables:
\dt

-- Describe specific table:
\d families
\d transactions

-- List all triggers:
SELECT trigger_name, event_object_table 
FROM information_schema.triggers 
WHERE trigger_schema = 'public';

-- List all functions:
\df

-- List all views:
\dv
```

## Migration Best Practices

1. **Never modify existing migrations** - Create new ones instead
2. **Always provide rollback (drop) migrations** - For easy reverting
3. **Test migrations on a copy of production data** before applying
4. **Keep migrations atomic** - One logical change per migration
5. **Use transactions** - Wrap migrations in BEGIN/COMMIT blocks
6. **Run migrations in order** - Dependencies must be respected

## Database Schema Overview

```
families (root entity)
  ├── users (authentication, family members)
  ├── accounts (financial accounts: cash, checking, savings)
  ├── categories (hierarchical: parent → children)
  ├── transactions (core financial data)
  │   ├── → account_id (which account)
  │   ├── → category_id (what category)
  │   └── → created_by (which user)
  └── audit_log (automatic via triggers)
      └── logs all CUD operations

exchange_rates (shared, not family-specific)
  └── historical rates for currency conversion
```

## Key Features

### 🏠 Family Isolation (Multi-tenancy)
All data is scoped to a `family_id` to ensure data privacy and isolation between households.

### 🗑️ Soft Deletes
Tables use an `is_active` flag for logical deletion, preserving historical data.

### 🔍 Comprehensive Audit Trail
All CUD operations are automatically logged in `audit_log` via triggers with before/after JSONB snapshots.

### 💰 Automatic Account Balance Calculation
Account balances are automatically calculated via `update_account_balance()` trigger:
```
current_balance = initial_balance + SUM(income) - SUM(expense)
```

### 💱 Multi-Currency Support
- Transactions store both original currency and base currency (RSD)
- Historical exchange rates with daily updates
- Helper function `get_exchange_rate(from, to, date)` with fallback

### 📊 Advanced Database Features
- **40+ indexes** for query performance
- **DECIMAL(15,2)** for money (precise, no rounding errors)
- **DECIMAL(15,6)** for exchange rates (higher precision)
- **JSONB** for flexible audit snapshots
- **GIN indexes** on JSONB for fast searching

## Tables Breakdown

| Table | Rows (seed) | Purpose |
|-------|------------|---------|
| `families` | 1 | Root entity, multi-tenant isolation |
| `users` | 2 | Authentication, family members |
| `accounts` | 4 | Financial accounts (cash, bank, savings) |
| `categories` | 19 | Hierarchical expense/income categories |
| `transactions` | ~20 | Core financial transactions |
| `exchange_rates` | 14 | Currency rates (7 days × 2 directions) |
| `audit_log` | 40+ | Automatic audit trail |

## Triggers

| Trigger | Table | Purpose |
|---------|-------|---------|
| `trigger_*_updated_at` | 5 tables | Auto-update `updated_at` timestamp |
| `trigger_transactions_update_balance` | transactions | Auto-recalculate account balance |
| `trigger_audit_*` | 4 tables | Auto-log all changes to audit_log |

## Functions

| Function | Purpose |
|----------|---------|
| `update_updated_at_column()` | Update timestamp on record change |
| `update_account_balance()` | Recalculate account balance based on transactions |
| `get_exchange_rate(from, to, date)` | Get exchange rate with fallback |
| `audit_trigger()` | Log changes to audit_log |

## Views

| View | Purpose |
|------|---------|
| `v_recent_audit_log` | Audit log with joined family/user names |

## Environment Variables

Required in `.env`:
```env
DB_HOST=your_db_host
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=expense_tracker_dev
DB_SSLMODE=disable
```

## Troubleshooting

### Migration fails with "relation already exists"
```sql
-- Check what exists:
\dt
-- Drop specific table if needed:
DROP TABLE tablename CASCADE;
```

### Trigger not working (balance not updating)
```sql
-- Check trigger exists:
SELECT trigger_name FROM information_schema.triggers 
WHERE event_object_table = 'transactions';
```

### View missing (v_recent_audit_log)
```sql
-- Recreate view manually if needed
-- See 007 create audit log table.sql for view definition
```

## Documentation

For detailed documentation on each migration, see `../docs/`:
- Step-by-step migration guides
- Seed data documentation
- Cheatsheets for quick reference

---

**Status**: Schema complete and production-ready! 🎉