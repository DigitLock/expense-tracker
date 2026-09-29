# Expense Tracker

Personal and family finance management system with multi-currency support and automatic balance calculation.

## 🎯 Project Status

- ✅ **Business Requirements** - Complete
- ✅ **System Requirements** - Complete
- ✅ **Database Schema** - Complete (7 tables, production-ready)
- ✅ **Backend API** - Complete (26 REST endpoints with JWT auth)
- ✅ **OpenAPI Documentation** - Complete (Swagger UI available)
- ✅ **Frontend MVP** - Complete (Full CRUD for Accounts, Categories, Transactions)
- ✅ **gRPC API** - Complete (Stage 6: Mobile API with dual-protocol architecture)
- 🚀 **Current release:** v0.4.1 on staging; v0.5.0 (Flutter client) in development

## ✨ Features

- 💰 **Multi-currency support** (RSD/EUR/USD with automatic conversion)
- 🏦 **Multiple account types** (cash, bank, credit, savings, investment)
- 🏷️ **Hierarchical categories** (parent-child structure for income/expense)
- 📊 **Automatic balance calculation** via database triggers
- 👥 **Multi-user families** with data isolation
- 🔐 **JWT authentication** with family-based access control
- 📈 **Transaction management** with filters and pagination
- 📊 **Financial dashboard** with monthly summary
- 📱 **Responsive UI** built with Vue.js 3 and Tailwind CSS
- 🔍 **Advanced filtering** by type, account, date range
- 📄 **Pagination** for large transaction lists
- 📡 **gRPC API** for mobile clients (dual-protocol: REST + gRPC)

## 🏗️ Tech Stack

### Backend
- **Language**: Go 1.25+
- **Router**: Chi v5
- **Database**: PostgreSQL 16
- **Query Builder**: sqlc (type-safe SQL)
- **Authentication**: JWT tokens
- **API Docs**: Swagger/OpenAPI 2.0
- **gRPC**: Protocol Buffers + gRPC (mobile API)

### Frontend
- **Framework**: Vue.js 3 (Composition API)
- **Styling**: Tailwind CSS
- **Form Validation**: VeeValidate + Zod
- **State Management**: Pinia
- **HTTP Client**: Axios
- **Build Tool**: Vite
- **UI Components**: Custom component library

## 📚 Documentation

### Business & System Requirements

Located in `Documentation/`:

- [`expense_tracker_brd.md`](Documentation/expense_tracker_brd.md) – Business Requirements Document
- [`expense_tracker_srs_mvp.md`](Documentation/expense_tracker_srs_mvp.md) – System Requirements (MVP)
- [`expense_tracker_tdd_grpc_integration.md`](Documentation/expense_tracker_tdd_grpc_integration.md) – Technical Design Document: gRPC Integration (Stage 6)
- [`expense_tracker_fs_flutter_mobile_client.md`](Documentation/expense_tracker_fs_flutter_mobile_client.md) – Functional Specification: Flutter Mobile Client (v0.4.0)
- PDF exports available in `Documentation/PDF/`

### API Documentation

**Interactive Swagger UI**: Available at `/swagger/index.html` when running the server

- **OpenAPI 2.0 specification** with full endpoint documentation
- **Request/Response examples** for all 26 endpoints
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
- **10 triggers**: automatic timestamp updates, balance calculation, audit logging
- **5 functions**: timestamp updates, balance update, balance recalculation, audit trail, default categories
- **33 indexes**: optimized for common query patterns
- **Complete rollback migrations**: every migration has a corresponding drop script

See [`database/migrations/README.md`](database/migrations/README.md) for details.

### Quick Start (Database)

```bash
# Apply all migrations (in order; 008 is intentionally absent):
001 create families table
002 create users table
003 create accounts table
004 create categories table
005 create transactions table
006 create exchange rates table
007 create audit log table
009 create default categories function
010 fix account balance trigger
011 add user role
012 exchange rates provenance
013 account name unique
014 fix balance trigger account change

# Load demo data (not a migration):
database/seeds/008_demo_seed_data.sql
```

**Demo credentials:**
- Email: `demo@example.com`
- Password: `Demo123!`

## 🚀 API Endpoints

### REST API (port 8080)

The REST API includes 26 endpoints across 8 groups:

#### Authentication
- `POST /api/v1/auth/login` - User login with JWT
- `POST /api/v1/auth/register` - Register a new user and family

#### Health
- `GET /health` - Health check
- `GET /ready` - Readiness probe

#### Accounts
- `GET /api/v1/accounts` - List all accounts
- `POST /api/v1/accounts` - Create account
- `GET /api/v1/accounts/{id}` - Get account details
- `PATCH /api/v1/accounts/{id}` - Update account
- `DELETE /api/v1/accounts/{id}` - Delete account
- `GET /api/v1/accounts/{id}/balance` - Get account balance

#### Categories
- `GET /api/v1/categories` - List all categories
- `POST /api/v1/categories` - Create category
- `GET /api/v1/categories/{id}` - Get category details
- `PATCH /api/v1/categories/{id}` - Update category
- `DELETE /api/v1/categories/{id}` - Delete category
- `POST /api/v1/categories/{id}/restore` - Restore a deleted category

#### Transactions
- `GET /api/v1/transactions` - List transactions (with filters & pagination)
- `POST /api/v1/transactions` - Create transaction
- `GET /api/v1/transactions/{id}` - Get transaction details
- `PATCH /api/v1/transactions/{id}` - Update transaction
- `DELETE /api/v1/transactions/{id}` - Delete transaction

#### Reports
- `GET /api/v1/reports/spending-by-category` - Spending analysis
- `GET /api/v1/reports/monthly-summary` - Monthly financial summary

#### Currencies
- `GET /api/v1/currencies/rates` - Get exchange rates
- `GET /api/v1/currencies/convert` - Convert currency

#### Exchange Rates
- `POST /api/v1/exchange-rates/sync` - Force exchange-rate sync from the Currency Rate Service

**📚 Full documentation with examples**: Visit `/swagger/index.html` after starting the server

### gRPC API (port 50051)

Mobile API using Protocol Buffers: 5 services, 16 RPCs. Proto definitions in `proto/` (7 files).

| Service | RPCs | Methods |
|---------|------|---------|
| AuthService | 2 | `Login`, `ValidateToken` |
| AccountService | 4 | `ListAccounts`, `CreateAccount`, `UpdateAccount`, `DeleteAccount` |
| TransactionService | 4 | `ListTransactions`, `CreateTransaction`, `UpdateTransaction`, `DeleteTransaction` |
| CategoryService | 4 | `ListCategories`, `CreateCategory`, `UpdateCategory`, `DeleteCategory` |
| ReportService | 2 | `GetSpendingByCategory`, `GetMonthlySummary` |

`common.proto` holds shared messages; `currency_rate.proto` is the client contract for the external Currency Rate Service (not served by this backend).

**Authentication**: JWT token via gRPC metadata header `authorization: Bearer <token>`

**Testing with grpcurl** (from project root):
```bash
# Get JWT token first
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"Demo123!"}' | jq .data.token

# List accounts
grpcurl -plaintext \
  -proto proto/accounts.proto \
  -H "authorization: Bearer <token>" \
  localhost:50051 expense_tracker.v1.AccountService/ListAccounts

# List transactions with filters
grpcurl -plaintext \
  -proto proto/transactions.proto \
  -H "authorization: Bearer <token>" \
  -d '{"month": "2025-12", "page": 1, "per_page": 20}' \
  localhost:50051 expense_tracker.v1.TransactionService/ListTransactions
```

## 🎨 Demo

**Demo environment** (Hetzner VPS):

| Service | URL |
|---------|-----|
| Web Frontend | [`https://demo-expensetracker.digitlock.systems`](https://demo-expensetracker.digitlock.systems) |
| REST API | [`https://api-demo-expensetracker.digitlock.systems`](https://api-demo-expensetracker.digitlock.systems) |
| REST API (direct) | `<demo-host>:8081` |
| gRPC API | `<demo-host>:50051` (plaintext) |

**Demo credentials**: `demo@example.com` / `Demo123!`

**Note**: gRPC is accessible via direct IP only — Cloudflare free plan does not proxy gRPC traffic.

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
├── proto/                  # Protocol Buffer definitions
│   ├── auth.proto          # AuthService (2 RPCs)
│   ├── accounts.proto      # AccountService (4 RPCs)
│   ├── transactions.proto  # TransactionService (4 RPCs)
│   ├── categories.proto    # CategoryService (4 RPCs)
│   ├── reports.proto       # ReportService (2 RPCs)
│   ├── common.proto        # Shared messages
│   └── currency_rate.proto # Currency Rate Service client contract
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
│   ├── grpc/             # gRPC server
│   │   ├── pb/           # Generated protobuf Go code
│   │   ├── handlers/     # gRPC service implementations
│   │   ├── interceptors/ # Auth and logging interceptors
│   │   └── server.go     # gRPC server setup
│   └── repository/       # Business logic layer
├── frontend/             # Vue.js 3 application
│   ├── src/
│   │   ├── api/         # API client
│   │   ├── components/  # Reusable components
│   │   │   ├── forms/   # Form components
│   │   │   ├── modals/  # Modal dialogs
│   │   │   └── ui/      # Base UI components
│   │   ├── composables/ # Vue composables
│   │   ├── router/      # Vue Router config
│   │   ├── schemas/     # Zod validation schemas
│   │   ├── stores/      # Pinia stores
│   │   ├── types/       # TypeScript types
│   │   └── views/       # Page components
│   ├── public/
│   └── package.json
├── buf.yaml              # buf lint configuration
├── buf.gen.yaml          # buf code generation configuration
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
- [x] REST API endpoints (26 endpoints)
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

### Phase 4: Frontend ✅
- [x] Vue.js 3 setup with Vite
- [x] API client with Axios
- [x] Authentication UI (Login)
- [x] Dashboard with financial summary
- [x] Accounts management (full CRUD)
- [x] Categories management (full CRUD with hierarchy)
- [x] Transactions management (full CRUD with filters)
- [x] Form validation with VeeValidate + Zod
- [x] Responsive design with Tailwind CSS
- [x] Reusable component library

### Phase 5: Testing & QA 🧪
- [x] Manual testing (in progress)
- [ ] Bug fixes and improvements
- [x] Unit tests (backend)
- [x] Integration tests
- [ ] E2E tests (frontend)
- [ ] Performance optimization
- [ ] Security audit

### Phase 6: gRPC API ✅
- [x] Protocol Buffer definitions (5 services, 16 RPCs)
- [x] Code generation with buf + protoc
- [x] gRPC server with dual-protocol architecture (REST :8080 + gRPC :50051)
- [x] JWT auth interceptor (reuses existing JWT service)
- [x] Logging interceptor
- [x] All 16 RPCs across Auth, Account, Transaction, Category and Report services (v0.4.0)
- [x] Tested with grpcurl

### Phase 7: Deployment 📋
- [x] Docker containerization
- [ ] CI/CD pipeline
- [ ] Production deployment
- [ ] Monitoring and logging
- [ ] Backup strategy

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
- Frontend development (Vue.js 3)
- **gRPC API design and implementation**
- **Dual-protocol architecture (REST + gRPC)**
- **Protocol Buffers and code generation**
- Full-stack application architecture
- DevOps and deployment

## 🔗 Links

- [GitHub Repository](https://github.com/DigitLock/expense-tracker)
- Portfolio: [portfolio.digitlock.systems](https://portfolio.digitlock.systems)
- REST API Documentation: Available at `/swagger/index.html` when running
- gRPC: port `50051` (proto files in `proto/`)