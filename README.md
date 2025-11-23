# Expense Tracker

Personal and family finance management system with multi-currency support and automatic balance calculation.

## 🎯 Project Status

- ✅ **Business Requirements** - Complete
- ✅ **System Requirements** - Complete  
- ✅ **Database Schema** - Complete (7 tables, production-ready)
- 🔄 **Backend API** - In Progress
- 📋 **Frontend** - Planned

## ✨ Features

- 💰 **Multi-currency support** (RSD/EUR with automatic conversion)
- 🏦 **Multiple account types** (cash, checking, savings)
- 🏷️ **Hierarchical categories** (parent-child structure)
- 📊 **Automatic balance calculation** via database triggers
- 👥 **Multi-user families** with data isolation
- 🔍 **Complete audit trail** with before/after snapshots
- 📈 **Historical exchange rates** for accurate reporting

## 🏗️ Tech Stack

- **Backend**: Go 1.21+ (planned)
- **Database**: PostgreSQL 16
- **Frontend**: Vue.js 3 (planned)

## 📚 Documentation

### Business & System Requirements

Located in `Documentation/`:

- [`expense_tracker_brd.md`](Documentation/expense_tracker_brd.md) – Business Requirements Document
- [`expense_tracker_srs_mvp.md`](Documentation/expense_tracker_srs_mvp.md) – System Requirements (MVP)
- PDF exports available in `Documentation/PDF/`

## 🗄️ Database Schema

Production-ready PostgreSQL schema with:

- **7 core tables**: families, users, accounts, categories, transactions, exchange_rates, audit_log
- **5 triggers**: automatic timestamp updates, balance calculation, audit logging
- **4 functions**: balance recalculation, exchange rate lookup, audit trail
- **40+ indexes**: optimized for common query patterns
- **Complete rollback migrations**: every migration has a corresponding drop script

See [`database/migrations/README.md`](database/migrations/README.md) for details.

### Quick Start (Database)

```bash
# Apply all migrations (in order):
001 create families table.sql
002 create users table.sql
003 create accounts table.sql
004 create categories table.sql
005 create transactions table.sql
006 create exchange rates table.sql
007 create audit log table.sql

# Load demo data:
009 demo seed data.sql
```

**Demo credentials:**
- Email: `demo@example.com`
- Password: `Demo123!`

## 🎨 Demo

Live demo coming soon with pre-loaded sample data for portfolio showcase.

## 📋 Project Structure

```
expense-tracker/
├── Documentation/          # Business and system requirements
├── database/
│   ├── migrations/        # SQL migration files
│   ├── docs/             # Database documentation (planned)
│   └── schema/           # Schema exports (planned)
├── internal/             # Go backend code (planned)
└── cmd/                  # Application entrypoints (planned)
```

## 🚀 Roadmap

### Phase 1: Database Foundation ✅
- [x] Schema design
- [x] Migration scripts
- [x] Automatic balance calculation
- [x] Audit logging
- [x] Demo seed data

### Phase 2: Backend API 🔄
- [ ] Database package (Go)
- [ ] REST API endpoints
- [ ] JWT authentication
- [ ] Business logic layer
- [ ] API documentation

### Phase 3: Frontend 📋
- [ ] Vue.js 3 setup
- [ ] Authentication UI
- [ ] Dashboard
- [ ] Transaction management
- [ ] Reports and analytics

## 📄 License

This project is licensed under the **MIT License**.  
See the [`LICENSE`](LICENSE) file for details.

## 👤 Author
**Igor Kudinov**  

This project is part of my professional portfolio demonstrating:
- Requirements analysis and documentation
- Database design and implementation
- Backend development (Go)
- Frontend development (Vue.js)
- DevOps and deployment