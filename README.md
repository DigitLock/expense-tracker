# Expense Tracker

Personal and family finance management system with multi-currency support and automatic balance calculation.

## 🎯 Project Status

- ✅ **Business Requirements** - Complete
- ✅ **System Requirements** - Complete
- ✅ **Database Schema** - Complete (7 tables, production-ready)
- ✅ **Backend API** - Complete (23 REST endpoints with JWT auth)
- ✅ **OpenAPI Documentation** - Complete (Swagger UI available)
- ✅ **Frontend MVP** - Complete (Full CRUD for Accounts, Categories, Transactions)
- 🧪 **Testing & QA** - In Progress (Stage 5)

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

## 🏗️ Tech Stack

### Backend
- **Language**: Go 1.23+
- **Router**: Chi v5
- **Database**: PostgreSQL 16
- **Query Builder**: sqlc (type-safe SQL)
- **Authentication**: JWT tokens
- **API Docs**: Swagger/OpenAPI 2.0

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

⚠️ **Demo Environment Status**: QA Testing in progress

**Testing environment**: `https://api.test.expensetracker.digitlock.systems`

**Note**: The application is currently undergoing quality assurance testing. Minor bugs and UI improvements are being addressed.

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
- [ ] Unit tests (backend)
- [ ] Integration tests
- [ ] E2E tests (frontend)
- [ ] Performance optimization
- [ ] Security audit

### Phase 6: Deployment 📋
- [ ] Docker containerization
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
- Full-stack application architecture
- DevOps and deployment

## 🔗 Links

- [GitHub Repository](https://github.com/DigitLock/expense-tracker)
- Portfolio: [portfolio.digitlock.systems](https://portfolio.digitlock.systems)
- API Documentation: Available at `/swagger/index.html` when running