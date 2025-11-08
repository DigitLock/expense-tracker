# Database Migrations

This directory contains SQL migration files for the Expense Tracker database schema.

## Migration Naming Convention

Migrations follow the pattern: `{version}_{description}.{direction}.sql`

- `version`: Sequential number (001, 002, etc.)
- `description`: Brief description of what the migration does
- `direction`: Either `up` (apply migration) or `down` (rollback migration)

## Migration Files

### 001_initial_schema
Creates the initial database schema with all core tables:
- `families` - Family/household grouping for shared financial data
- `users` - User accounts with authentication
- `accounts` - Financial accounts (cash, checking, savings)
- `categories` - Income and expense categories
- `transactions` - Core financial transactions
- `exchange_rates` - Currency exchange rates for multi-currency support
- `audit_log` - Audit trail for all data modifications

### 002_seed_data (coming next)
Inserts initial reference data:
- Default categories for common expenses
- Sample exchange rates
- Test family and users for development

## How to Use Migrations

### Apply migrations (up)
```bash
make db-migrate
```

### Rollback migrations (down)
```bash
make db-rollback
```

### Reset database (drops all data and re-applies migrations)
```bash
make db-reset
```

### Verify schema
```bash
make db-shell
\dt  -- List all tables
\d tablename  -- Describe specific table
```

## Migration Best Practices

1. **Never modify existing migrations** - Create new ones instead
2. **Always provide rollback (down) migrations** - For easy reverting
3. **Test migrations on a copy of production data** before applying
4. **Keep migrations atomic** - One logical change per migration
5. **Use transactions** - Wrap migrations in BEGIN/COMMIT blocks
6. **Document breaking changes** - Add comments for important changes

## Database Schema Overview

```
families
  └── users (many users per family)
  └── accounts (many accounts per family)
  └── categories (many categories per family)
  └── transactions (many transactions per family)
      └── references account
      └── references category
  └── exchange_rates (shared across families)
  └── audit_log (many logs per family)
```

## Key Features

### Family Isolation
All data is scoped to a `family_id` to ensure data privacy and isolation between different households.

### Soft Deletes
Tables use an `is_active` flag for logical deletion instead of physical deletion, preserving historical data.

### Audit Trail
All CUD (Create, Update, Delete) operations are automatically logged in the `audit_log` table via database triggers.

### Account Balance Automation
Account balances are automatically calculated and updated via database triggers when transactions are added/modified/deleted.

### Multi-Currency Support
- All transactions store both original currency amount and base currency (RSD) amount
- Exchange rates are stored in a separate table with daily updates
- Base currency conversion is automatic during transaction creation