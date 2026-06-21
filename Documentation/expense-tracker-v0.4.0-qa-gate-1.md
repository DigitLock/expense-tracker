# Expense Tracker v0.4.0 — QA-gate 1 (gRPC Contract Freeze)

**Project:** `github.com/DigitLock/expense-tracker`
**Version:** v0.4.0
**Document type:** QA gate / contract-freeze sign-off
**Gate purpose:** Freeze the backend/gRPC contract before mobile development (v0.5.0). After this gate, proto/service/message changes are made only through deliberate versioning.

---

## 1. Summary

v0.4.0 completes the gRPC backend: all **16 of 16 RPC methods** across 5 services are implemented over a **shared application (service) layer**, with REST and gRPC reduced to thin transport adapters. No business logic is duplicated between protocols (ADR-2). `make test` is green across all 9 packages; the full 16-method surface was exercised via `grpcurl` in a single consolidated session. **0 P0 / 0 P1.**

| Metric | Result |
|--------|--------|
| RPC methods implemented | 16 / 16 |
| Services registered on gRPC server | 5 / 5 (Auth, Account, Transaction, Category, Report) |
| Test packages green (`make test`) | 9 / 9 |
| P0 / P1 defects | 0 / 0 |
| Business-logic duplication | None (rules live in `internal/service/<domain>`) |

---

## 2. Architecture conformance

| Principle | Status | Notes |
|-----------|--------|-------|
| Shared logic, no duplication (ADR-2) | ✅ | Business rules extracted into `internal/service/<domain>`; REST and gRPC are thin adapters calling the same service methods |
| Domain models transport-agnostic | ✅ | `internal/domain/*` carry clean types; no `pgtype`/`family_id`/`deleted_by` leak upward |
| Money precision | ✅ | `decimal.Decimal` end-to-end in domain/service/REST; `double` confined to the gRPC adapter edge (`NewFromFloat(...).Round(2)` in, `.Float64()` out) |
| Family isolation | ✅ | `family_id` sourced from JWT context; foreign resource → `PERMISSION_DENIED` |
| Balance integrity | ✅ | Recalculated only by the DB trigger (migration 010, fixed by 014); the service never computes `current_balance` |
| gRPC auth (ADR-4) | ✅ | Plaintext `:50051`, Bearer token in metadata; interceptor exempts exactly the two public AuthService methods |
| Unified error taxonomy | ✅ | Domain sentinel errors → single gRPC mapper (`internal/grpc/errmap`) and single REST mapper |

---

## 3. Method verification matrix (16 / 16)

Status legend: ✅ verified via `grpcurl` (live) and/or service test, error codes per TDD §3/§9.1.

| Service | Method | Status | Key checks |
|---------|--------|--------|-----------|
| AuthService | Login | ✅ | token+user+expires_in; bad creds → `UNAUTHENTICATED(16)`; missing field → `INVALID_ARGUMENT(3)`; works without metadata token (interceptor exemption) |
| AuthService | ValidateToken | ✅ | valid token → `valid:true`+user; invalid → `valid:false` (not an error); token read from request body |
| AccountService | ListAccounts | ✅ | family-scoped list with balances |
| AccountService | CreateAccount | ✅ | type/currency/initial≥0; duplicate name → `ALREADY_EXISTS(6)` (index 013); invalid type → `INVALID_ARGUMENT(3)` |
| AccountService | UpdateAccount | ✅ | partial; not found → `NOT_FOUND(5)`; other family → `PERMISSION_DENIED(7)` |
| AccountService | DeleteAccount | ✅ | soft delete; already inactive → `FAILED_PRECONDITION(9)`; has transactions → `FAILED_PRECONDITION(9)` |
| TransactionService | ListTransactions | ✅ | filters + pagination |
| TransactionService | CreateTransaction | ✅ | currency = account currency; category type match; non-existent account → `NOT_FOUND(5)`; balance recalculated |
| TransactionService | UpdateTransaction | ✅ | amount change → balance recalculated; account change → **both** old and new accounts recalculated (migration 014) |
| TransactionService | DeleteTransaction | ✅ | soft delete; balance recalculated |
| CategoryService | ListCategories | ✅ | type filter; `include_inactive` |
| CategoryService | CreateCategory | ✅ | parent type match; ≤ 2 levels; duplicate → `ALREADY_EXISTS(6)` |
| CategoryService | UpdateCategory | ✅ | circular parent → `INVALID_ARGUMENT(3)`; 2-level guard |
| CategoryService | DeleteCategory | ✅ | active subcategories → `FAILED_PRECONDITION(9)`; transactions → `FAILED_PRECONDITION(9)` |
| ReportService | GetSpendingByCategory | ✅ | default = current month/expense/RSD; custom period; invalid date → `INVALID_ARGUMENT(3)` |
| ReportService | GetMonthlySummary | ✅ | income/expenses/net/counts; empty month → zeros |

---

## 4. Acceptance criteria (spec §6 / §7)

| Criterion | Status |
|-----------|--------|
| All 14 new methods respond via `grpcurl` with valid JWT | ✅ |
| Error codes match the taxonomy (3/5/6/7/9/16) on negative cases | ✅ |
| Transactional ops recalculate balance (create / update amount / update account / delete) | ✅ |
| Categories: 2-level hierarchy, parent-type match, cycle prevention, delete-with-dependencies | ✅ |
| gRPC reports return the same values as REST | ✅ (byte-identical parity confirmed on 2025-11) |
| gRPC Login token valid in both REST and gRPC; ValidateToken distinguishes valid/invalid | ✅ |
| Family isolation: foreign resource → `PERMISSION_DENIED` | ✅ |
| No business-logic duplication (thin adapters) | ✅ |
| `make test` green; consolidated `grpcurl` matrix passed | ✅ |
| Full 16-method run, 0 P0/P1 | ✅ |

---

## 5. REST ↔ gRPC parity nuances (deliberate, recorded)

The two protocols give **identical domain results for the same operation**. The following are deliberate, documented differences, surfaced during implementation and accepted at this gate:

| # | Difference | Rationale | Decision |
|---|-----------|-----------|----------|
| P1 | REST error codes aligned to the unified taxonomy: foreign resource `404 → 403`, delete-with-dependencies `400 → 409`, `initial_balance = 0` now allowed | Unification (one taxonomy for both protocols); the TDD already chose `PERMISSION_DENIED` for foreign resources, REST was the outlier | Accepted |
| P2 | gRPC `UpdateTransaction` can change `account_id` / `type`; REST PATCH cannot (unchanged) | Matches the gRPC contract (optional fields in `UpdateTransactionRequest`). gRPC is a superset; "identical result" holds for the same operation | Accepted; REST extension is a possible future enhancement |
| P3 | gRPC `CreateCategory` has no inactive-restore-offer (REST-only UX); an inactive-name duplicate creates a new active category (active-name duplicate still → `ALREADY_EXISTS`) | Restore-offer is an interactive REST flow not in the gRPC contract | Accepted; documented edge: restoring an inactive category after a gRPC active-duplicate would hit the unique index |
| P4 | REST reports now **validate** `start_date`/`end_date`/`type`/`currency` (→ `400`) instead of silently defaulting | gRPC contract requires `INVALID_ARGUMENT`; REST aligned. Valid requests return unchanged values | Accepted |
| P5 | `USER_INACTIVE` REST response folded into `INVALID_CREDENTIALS` | The previous branch was unreachable (`GetUserByEmail` is active-only); no observable behavior change | Resolved (no change) |

**Open ratification item:** the report `currency` parameter accepts `USD` in addition to `RSD`/`EUR` (matching existing REST behavior and available CRS rates), whereas TDD §5.5 lists `{RSD, EUR}`. The proto field is an unconstrained `string`; this is an accepted-values decision, not a proto-shape change. **Recommendation: freeze with `{RSD, EUR, USD}` for the report `currency` param.** Pending Igor's ratification.

---

## 6. Database migrations (pending application)

Both migration files are authored and validated (up/down round-trip on the test DB); **not yet applied to the dev DB** (`192.168.13.30`, currently `version = 12`). To be applied to dev now and to the VPS at v0.5.0 deploy.

| Migration | Purpose | Apply order |
|-----------|---------|-------------|
| 013 `account_name_unique` | Partial unique index on active account names per family (`ALREADY_EXISTS` rule); mirrors the categories index (`LOWER(name)`, `WHERE is_active`) | 1st |
| 014 `fix_balance_trigger_account_change` | Recalculate **both** old and new account on transaction `account_id` change; one-time recalc heals accumulated drift | 2nd |

> Until 013/014 are applied to dev, the duplicate-account-name (`ALREADY_EXISTS`) and account-change-both-balances cases are proven by service tests against the test DB (where the migrations are applied via `test-setup`), not by live dev `grpcurl`.

---

## 7. Contract freeze declaration

As of QA-gate 1, the following are **frozen**. Subsequent changes require deliberate versioning, not in-place edits.

- Proto files (flat `proto/`, ADR-5): `auth.proto`, `accounts.proto`, `transactions.proto`, `categories.proto`, `reports.proto`, `common.proto` (shared `DeleteResponse`).
- Service / method / message names and field numbers in `expense_tracker.v1`.
- `optional` field-presence semantics on all `Update*` partial fields.
- The error taxonomy (domain sentinel → gRPC code / HTTP status).

Generated `internal/grpc/pb/*` are tracked in the repo (ADR-5).

---

## 8. Deferred items (recorded with rationale)

| Item | Planned | Rationale |
|------|---------|-----------|
| EUR↔USD storage currency pair | v1.0.0 | Only RSD-pairs synced pre-v1.0; report `currency` conversion is display-only |
| Period-specific exchange rates in reports | v1.0.0 | Reports currently use available rate; period-by-`endDate` rate deferred |
| Server streaming (real-time balance) | Future | Not required for parity |
| TLS for gRPC | Production | Demo is plaintext (ADR-4) |
| REST `UpdateTransaction` account/type change | Future (optional) | gRPC superset is sufficient for parity (P2) |
| gRPC `CreateCategory` restore-offer | Future (optional) | REST-only UX (P3) |

---

## 9. Sign-off

| Check | Result |
|-------|--------|
| 16/16 methods verified | ✅ |
| 0 P0 / P1 | ✅ |
| REST ↔ gRPC domain parity confirmed | ✅ |
| Contract frozen | ✅ (pending §5 ratification of report `currency` values) |
| Migrations 013/014 ready (not yet applied to dev) | ✅ |

**Gate result: PASS.** Proceed to v0.4.0 release (tag, session summary — no deploy). Mobile development (v0.5.0) builds on the frozen contract.
