# Expense Tracker

Personal and family finance management system with multi-currency support and automatic balance calculation.

## 🎯 Project Status

- ✅ **Business Requirements** - Complete
- ✅ **System Requirements** - Complete
- ✅ **Database Schema** - Complete (7 tables, production-ready)
- ✅ **Backend API** - Complete (23 REST endpoints with JWT auth)
- ✅ **OpenAPI Documentation** - Complete (Swagger UI available)
- 🔄 **Frontend** - In Progress (Stage 4)

## ✨ Features

- 💰 **Multi-currency support** (RSD/EUR with automatic conversion)
- 🏦 **Multiple account types** (cash, checking, savings)
- 🏷️ **Hierarchical categories** (parent-child structure)
- 📊 **Automatic balance calculation** via database triggers
- 👥 **Multi-user families** with data isolation
- 📝 **Complete audit trail** with before/after snapshots
- 📈 **Historical exchange rates** for accurate reporting
- 🔐 **JWT authentication** with family-based access control
- 📊 **Financial reports** (monthly summary, spending by category)
- 📚 **Interactive API documentation** with Swagger UI

## 🏗️ Tech Stack

- **Backend**: Go 1.23+ with Chi router
- **Database**: PostgreSQL 16
- **API**: REST with JWT authentication
- **Code Generation**: sqlc for type-safe database queries
- **API Documentation**: Swagger/OpenAPI 2.0
- **Frontend**: Vue.js 3 (planned)

## 📚 Documentation

### Business & System Requirements

Located in `Documentation/`:

- [`expense_tracker_brd.md`](Documentation/expense_tracker_brd.md) – Business Requirements Document
- [`expense_tracker_srs_mvp.md`](Documentation/expense_tracker_srs_mvp.md) – System Requirements (MVP)
- PDF exports available in `Documentation/PDF/`

### API Documentation

**Interactive Swagger UI**: Available at `/swagger/index.html` when running the server

- **OpenAPI 2.0 specification** with full endpoint documentation
- **Request/Response examples** for all 23 endpoints
- **Try it out** functionality for testing endpoints directly
- **Schema definitions** for all DTOs
- **Authentication flow** documentation

Generated specification files in `Documentation/swagger/`:
- `swagger.json` - OpenAPI specification (machine-readable)
- `swagger.yaml` - OpenAPI specification (human-readable)
- `docs.go` - Embedded Go documentation

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

## 🚀 API Endpoints

The REST API includes 23 endpoints across 6 categories:

### Authentication
- `POST /api/v1/auth/login` - User login with JWT

### Health
- `GET /health` - Health check
- `GET /ready` - Readiness probe

### Accounts
- `GET /api/v1/accounts` - List all accounts
- `POST /api/v1/accounts` - Create account
- `GET /api/v1/accounts/{id}` - Get account details
- `PATCH /api/v1/accounts/{id}` - Update account
- `DELETE /api/v1/accounts/{id}` - Delete account
- `GET /api/v1/accounts/{id}/balance` - Get account balance

### Categories
- `GET /api/v1/categories` - List all categories
- `POST /api/v1/categories` - Create category
- `GET /api/v1/categories/{id}` - Get category details
- `PATCH /api/v1/categories/{id}` - Update category
- `DELETE /api/v1/categories/{id}` - Delete category

### Transactions
- `GET /api/v1/transactions` - List transactions (with filters & pagination)
- `POST /api/v1/transactions` - Create transaction
- `GET /api/v1/transactions/{id}` - Get transaction details
- `PATCH /api/v1/transactions/{id}` - Update transaction
- `DELETE /api/v1/transactions/{id}` - Delete transaction

### Reports
- `GET /api/v1/reports/spending-by-category` - Spending analysis
- `GET /api/v1/reports/monthly-summary` - Monthly financial summary

### Currencies
- `GET /api/v1/currencies/rates` - Get exchange rates
- `GET /api/v1/currencies/convert` - Convert currency

**📚 Full documentation with examples**: Visit `/swagger/index.html` after starting the server

## 🎨 Demo

Live demo coming soon with pre-loaded sample data for portfolio showcase.

**Testing environment**: `https://api.test.expensetracker.digitlock.systems`

## 📋 Project Structure

```
expense-tracker/
├── Documentation/          # Business and system requirements
│   ├── expense_tracker_brd.md
│   ├── expense_tracker_srs_mvp.md
│   ├── swagger/           # OpenAPI documentation
│   │   ├── docs.go        # Generated Swagger docs
│   │   ├── swagger.json   # OpenAPI 2.0 spec
│   │   └── swagger.yaml   # OpenAPI 2.0 spec (YAML)
│   └── *_SUMMARY.md       # Development stage summaries
├── database/
│   └── migrations/        # SQL migration files
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── api/              # HTTP handlers and routing
│   │   ├── handlers/     # Request handlers (with Swagger annotations)
│   │   └── middleware/   # Auth, logging, recovery
│   ├── auth/             # JWT service
│   ├── config/           # Configuration management
│   ├── database/         # Database layer
│   │   ├── queries/      # SQL queries for sqlc
│   │   └── sqlc/         # Generated type-safe code
│   ├── dto/              # Data transfer objects (with Swagger tags)
│   └── repository/       # Business logic layer
├── .env                  # Environment variables
├── go.mod               # Go module definition
└── sqlc.yaml            # sqlc configuration
```

## 🚀 Roadmap

### Phase 1: Database Foundation ✅
- [x] Schema design
- [x] Migration scripts
- [x] Automatic balance calculation
- [x] Audit logging
- [x] Demo seed data

### Phase 2: Backend API ✅
- [x] Database package (Go + sqlc)
- [x] REST API endpoints (23 endpoints)
- [x] JWT authentication
- [x] Business logic layer
- [x] Input validation
- [x] CORS configuration
- [x] Family-based data isolation

### Phase 3: Documentation ✅
- [x] OpenAPI/Swagger specification
- [x] Interactive API documentation (Swagger UI)
- [x] Request/Response examples
- [x] Authentication flow documentation
- [ ] Postman collection export

### Phase 4: Frontend 🔄
- [ ] Vue.js 3 setup
- [ ] API client generation from OpenAPI spec
- [ ] Authentication UI
- [ ] Dashboard
- [ ] Transaction management
- [ ] Reports and analytics
- [ ] Responsive design

### Phase 5: Testing & Deployment 📋
- [ ] Unit tests
- [ ] Integration tests
- [ ] Docker containerization
- [ ] CI/CD pipeline
- [ ] Production deployment

## 📄 License

This project is licensed under the **MIT License**.  
See the [`LICENSE`](LICENSE) file for details.

## 👤 Author
**Igor Kudinov**  

This project is part of my professional portfolio demonstrating:
- Requirements analysis and documentation
- Database design and implementation
- Backend development (Go)
- REST API design
- OpenAPI/Swagger documentation
- Type-safe code generation (sqlc)
- Frontend development (Vue.js)
- DevOps and deployment

## 🔗 Links

- [GitHub Repository](https://github.com/DigitLock/expense-tracker)
- Portfolio: [portfolio.digitlock.systems](https://portfolio.digitlock.systems)
- API Documentation: Available at `/swagger/index.html` when running