# Expense Tracker v0.3.0 — Currency Sync QA Report

Acceptance run for the Currency Rate Service integration (v0.3.0). Use cases UC-03-1…6
(spec §11) executed against a live stack with fresh evidence; acceptance criteria (spec §12)
mapped to the UC / automated test that proves each. No commits were made for this run.

# 1. Environment & Rate Anchors

Run date: 2026-06-16. All evidence below was captured fresh in this run (not reused).

| Item | Value |
|------|-------|
| Go | go1.25.3 darwin/arm64 |
| Dev DB (`expense_tracker_dev`, 192.168.13.30) | `schema_migrations` version = **12**, dirty = false |
| Currency Rate Service (CRS) | gRPC `localhost:50052`, health `:8090` |
| ET backend | HTTP `:8080`, gRPC `:50051` |
| Auth | JWT via `POST /api/v1/auth/login` (demo@example.com / Demo123!) → 200, token acquired |
| Convention | `rate` = RSD per 1 foreign unit; stored `from=<foreign>, to=RSD`; reports convert `amount_base / rate` |

Rate anchors (latest rows, `SELECT … WHERE to_currency='RSD'`):

```
 from | to  |    rate    |    date     |   source    |          fetched_at
------+-----+------------+-------------+-------------+------------------------------
 EUR  | RSD | 117.332262 | 2026-06-16  | fawazahmed0 | 2026-06-16 21:34:59.732286+00
 USD  | RSD | 101.317282 | 2026-06-16  | fawazahmed0 | 2026-06-16 21:34:59.732286+00
```

(Older `manual` rows for EUR→RSD on 2026-03..06 remain as history; `GetLatestRate` selects the newest.)

# 2. Use Case Results (UC-03-1…6)

| ID | Scenario | Steps | Evidence | Result |
|----|----------|-------|----------|--------|
| UC-03-1 | Startup sync, CRS available | Start CRS, restart backend; SELECT exchange_rates | Startup log `currency sync: upserted 2 pair(s), missing 0, fetched_at=…T21:32:59Z`; DB EUR→RSD & USD→RSD with `source=fawazahmed0`, `fetched_at` advanced `21:08:16 → 21:32:59` | **Pass** |
| UC-03-2 | Scheduled tick | `CURRENCY_SYNC_INTERVAL=5s`, run ~12s; SELECT before/after | Ticks logged at +5s/+10s (`…T21:33:04Z`, `…T21:33:09Z`); DB `fetched_at` advanced to `21:33:24` (later ticks) | **Pass** |
| UC-03-3 | Manual trigger | `POST /api/v1/exchange-rates/sync` w/ JWT | `200` `{"success":true,"data":{"synced_pairs":2,"source":"fawazahmed0","fetched_at":"2026-06-16T21:34:59.732286Z"}}`; DB `fetched_at` `21:34:48 → 21:34:59` | **Pass** |
| UC-03-4 | CRS unavailable | Stop CRS while backend running; hit health/login/reports/trigger; restart w/ CRS down for logs | `/health=200`, `login=200`, reports EUR still converts (`total_expenses=502.63`, last rates), `POST sync=503`; logs `WARN: initial currency sync failed … connection refused` + `WARN: currency sync failed …` on ticks; backend alive | **Pass** |
| UC-03-5 | Reports USD | `GET monthly-summary?month=2025-11&currency=USD` | `total_expenses=582.08` = `58975 / 101.317282` (✓); `account_balances.total=232923` (native, unchanged vs RSD); switcher exposes USD — `ReportsView.vue:23 currencyOptions = ['RSD','EUR','USD']` | **Pass** (number-proven; final visual click → Igor) |
| UC-03-6 | EUR regression | Same month `currency=EUR` and `RSD` | EUR `total_expenses=502.63` = `58975 / 117.332262` (✓); RSD `58975` (no conversion); `account_total=232923` native — unchanged from v0.2.0 behavior | **Pass** |

## 2.1 Raw evidence

UC-03-1/2 — startup + ticks (interval 5s) and resulting DB rows:
```
Currency rate sync scheduled every 5s (service localhost:50052)
Server starting on port 8080
currency sync: upserted 2 pair(s), missing 0, fetched_at=2026-06-16T21:32:59Z   ← UC-03-1 startup (t=0)
currency sync: upserted 2 pair(s), missing 0, fetched_at=2026-06-16T21:33:04Z   ← UC-03-2 tick +5s
currency sync: upserted 2 pair(s), missing 0, fetched_at=2026-06-16T21:33:09Z   ← UC-03-2 tick +10s

 from | to  |    rate    |    date    |   source    |          fetched_at
------+-----+------------+------------+-------------+------------------------------
 EUR  | RSD | 117.332262 | 2026-06-16 | fawazahmed0 | 2026-06-16 21:33:24.800973+00
 USD  | RSD | 101.317282 | 2026-06-16 | fawazahmed0 | 2026-06-16 21:33:24.800973+00
```

UC-03-3 — manual trigger:
```
fetched_at BEFORE: EUR->RSD / USD->RSD = 2026-06-16 21:34:48.242776+00
POST /api/v1/exchange-rates/sync  → HTTP/1.1 200 OK
{"success":true,"data":{"synced_pairs":2,"source":"fawazahmed0","fetched_at":"2026-06-16T21:34:59.732286Z"}}
fetched_at AFTER:  EUR->RSD / USD->RSD = 2026-06-16 21:34:59.732286+00
```

UC-03-4 — CRS unavailable (backend stays up):
```
CRS stopped
health -> 200
login -> 200
reports currency=EUR total_expenses=502.63 note=''   (served from last known rates)
POST sync -> 503                                       (CURRENCY_SERVICE_UNAVAILABLE)
WARN: initial currency sync failed: currency sync: fetch: currency: GetRates: rpc error:
  code = Unavailable … dial tcp 127.0.0.1:50052: connect: connection refused
WARN: currency sync failed: … connection refused      (repeats each tick; loop survives)
health -> 200                                          (still alive)
```

UC-03-5 / UC-03-6 — reports conversion (month 2025-11, expenses 58975 RSD):
```
currency=RSD  total_income=25000  total_expenses=58975   account_total=232923  note=''
currency=EUR  total_income=213.07 total_expenses=502.63  account_total=232923  note=''
currency=USD  total_income=246.75 total_expenses=582.08  account_total=232923  note=''
cross-check: 58975/117.332262 = 502.63 (EUR)   58975/101.317282 = 582.08 (USD)
```
`account_total` is identical (232923) across RSD/EUR/USD → account balances are NOT converted (native), as required.

# 3. Acceptance Criteria Checklist (spec §12)

| # | Criterion | Proven by | Status |
|---|-----------|-----------|--------|
| AC-1 | Rates synced from CRS at startup when available (all available pairs) | UC-03-1 | ✅ |
| AC-2 | Rates re-synced on a schedule (configurable interval) | UC-03-2 | ✅ |
| AC-3 | Authenticated manual sync endpoint returns synced_pairs/source/fetched_at | UC-03-3; autotest `TestExchangeRateHandler_Sync_Success` | ✅ |
| AC-4 | Manual sync without JWT rejected (401) | middleware (proven live st5); autotests cover 200/503 paths | ✅ |
| AC-5 | CRS rates inverted into ET convention (RSD per foreign), rounded to 6 dp | UC-03-1/3 (DB rows); autotest `TestSync_MappingInversion` (117.332262 / 101.317282) | ✅ |
| AC-6 | Sync is idempotent per (from,to,date) — update, not duplicate; fetched_at advances | autotest `TestSync_Integration` | ✅ |
| AC-7 | Partial CRS response handled — write returned pairs, report missing | autotest `TestSync_PartialResponse` (Missing=[USD]) | ✅ |
| AC-8 | Zero / outdated rates handled (skip zero; write outdated) | autotests `TestSync_ZeroRateSkipped`, `TestSync_OutdatedStillWritten` | ✅ |
| AC-9 | CRS outage never crashes backend; logs WARN; last rates served | UC-03-4; autotest `TestSync_FetchError` (DB untouched) | ✅ |
| AC-10 | Manual sync returns 503 when CRS unavailable | UC-03-4; autotests `TestExchangeRateHandler_Sync_FetchError` / `_NilSyncer` | ✅ |
| AC-11 | Reports support USD (amount_base / USD→RSD) | UC-03-5 | ✅ |
| AC-12 | EUR/RSD reports unchanged (v0.2.0 regression) | UC-03-6 | ✅ |
| AC-13 | Account balances not currency-converted | UC-03-5/6 (account_total constant across currencies) | ✅ |
| AC-14 | Migration 012 (source, fetched_at, USD-widened CHECK) applied & reversible | dev DB version=12; roundtrip verified in st3; `make test` applies 012 to test DB | ✅ |

Automated regression (st7a, `make test`, serial `-p 1`): `internal/currency` 7 tests + `internal/api/handlers` (currency 3 + register 4) + `internal/repository` 2 — all PASS.

# 4. Defects & Conclusion

Defects found: **0**.

No Fail occurred during the run; no fix/retry cycles were needed.

Conclusion: **0 P0/P1 defects.** All six use cases (UC-03-1…6) Pass and all §12 acceptance
criteria are satisfied. v0.3.0 currency sync is acceptance-ready. Remaining manual step:
final visual confirmation of the Reports USD switcher in the browser by Igor (data already
proven numerically in UC-03-5).
