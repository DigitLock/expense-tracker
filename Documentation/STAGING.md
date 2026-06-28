# Expense Tracker — Staging Environment

## Overview

The staging environment is an isolated, always-on deployment of the backend
used as a **stable API target for the mobile client (v0.5.0)** and as a
**rehearsal of the eventual VPS deployment**. It mirrors the production topology
(containerized backend, dedicated Postgres, migration runner, reverse-proxied
web edge, public HTTPS hostname) on local LAN hardware so that deployment
mechanics are validated before renting a VPS.

Environment ladder:

```
dev  →  test  →  staging  →  live (VPS)
```

* **dev** — developer Mac, `make run`, local Postgres.
* **test** — integration test database (`expense_tracker_test`), CI/`make test`.
* **staging** — this document. Isolated stack, demo data, public HTTPS.
* **live** — future VPS, same compose topology with real secrets and TLS certs.

The Currency Rate Service (CRS) is **deliberately excluded** from staging. The
backend tolerates its absence: currency sync is skipped and reports fall back to
stored/identity rates (`CURRENCY_SERVICE_ADDR` is left empty in `.env.staging`).

## Topology

| Layer            | Node            | Address                                   | External |
| ---------------- | --------------- | ----------------------------------------- | -------- |
| Postgres         | 192.168.13.80   | internal Docker network only (no publish) | No       |
| Backend REST     | 192.168.13.80   | `:18080` (host) → `:8080` (container)      | Via edge |
| Backend gRPC     | 192.168.13.80   | `:15051` (host) → `:50051` (container)     | LAN only |
| Frontend (static)| 192.168.13.90   | nginx `:18090`, serves Vue `dist/`         | Via tunnel |
| Edge (proxy)     | 192.168.13.90   | nginx, `/api` → `192.168.13.80:18080`      | Via tunnel |
| Public web       | Cloudflare      | `https://staging.digitlock.systems`        | Yes      |

* **Postgres** is never published to the host — only services on the compose
  `internal` network reach it (`postgres:5432`).
* **Backend gRPC** is exposed on the LAN as **plaintext** for tooling
  (`grpcurl`, mobile dev against LAN). It is **not** routed through Cloudflare.
* **Public web** is a Cloudflare Tunnel from the edge node terminating at
  `localhost:18090`; TLS is provided by Cloudflare.

## Components

* **Compose project:** `expense-tracker-staging`
  (`docker-compose.staging.yml`, invoked with `-p expense-tracker-staging`).
* **Services:**
  * `postgres` — `postgres:16-alpine`, named volume `pgdata`, healthcheck
    `pg_isready`.
  * `migrate` — `migrate/migrate:v4.17.1`, one-shot, runs `up` then exits
    (`restart: "no"`), `depends_on: postgres (service_healthy)`.
  * `backend` — built from the repo `Dockerfile` (multi-stage: buf codegen →
    static Go binary → `alpine:3.20`, runs as unprivileged `appuser`).
    `depends_on: postgres (service_healthy)` + `migrate
    (service_completed_successfully)`.
* **Configuration:**
  * `.env.staging` — **gitignored**, holds the real secret *values*; lives only
    on the staging node.
  * `.env.staging.example` — committed template (variable names, no secret
    values). Copy it to `.env.staging` and fill in the blanks.
* **Frontend build:** `npm run build` → `frontend/dist/`. Built with the
  **default relative API base `/api/v1`** (no `VITE_API_BASE_URL`). Because the
  edge serves the bundle and proxies `/api` to the backend on the same origin,
  **no CORS configuration is required**.

## Bring-up

> All secret values come from `.env.staging` on the staging node. See
> `.env.staging.example` for the full list of variables to populate
> (`DB_PASSWORD`, `JWT_SECRET`, etc.). Never copy secret values into this doc or
> into git.

Backend node — `192.168.13.80`:

1. **Sync the tree** from the Mac to the staging node:

   ```sh
   rsync -av --delete --exclude '.git' --exclude 'bin' --exclude 'frontend/node_modules' \
     ./ digitlock@192.168.13.80:~/expense-tracker-staging/
   ```

2. **Build the backend image:**

   ```sh
   cd ~/expense-tracker-staging
   docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging build backend
   ```

3. **Start Postgres, then run migrations** (one-shot, brings schema to
   version **14**):

   ```sh
   docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging up -d postgres
   docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging up migrate
   ```

4. **Seed staging data.** Postgres is not published, so the seed is piped into
   `psql` inside the container:

   ```sh
   docker compose -f docker-compose.staging.yml --env-file .env.staging \
     -p expense-tracker-staging exec -T postgres \
     psql -U expense_staging -d expense_tracker_staging < database/seeds/staging_seed.sql
   ```

   The seed is **idempotent** — it `DELETE`s this family's rows by `family_id`
   and recreates them, so re-running it never duplicates data. It provisions the
   demo login `demo@example.com` / `Demo123!`.

5. **Start the backend:**

   ```sh
   docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging up -d backend
   ```

Edge / frontend node — `192.168.13.90`:

6. **Build & publish the frontend:**

   ```sh
   cd frontend && npm ci && npm run build          # → dist/, relative /api/v1
   rsync -av dist/  user@192.168.13.90:/var/www/staging/
   ```

   nginx site `expense-staging` listens on `:18090`, serves
   `/var/www/staging` (SPA fallback to `index.html`), and reverse-proxies
   `location /api/ → http://192.168.13.80:18080`.

7. **Expose publicly via Cloudflare Tunnel** (token-based, runs as a systemd
   service):

   ```sh
   # Install cloudflared from Cloudflare's apt repository (one-time), then:
   sudo cloudflared service install <TUNNEL_TOKEN>   # token from the Zero Trust dashboard
   ```

   The Public Hostname mapping
   (`staging.digitlock.systems` → `HTTP localhost:18090`) is configured in the
   **Zero Trust dashboard** under *Networks → Tunnels → Public Hostname* — not in
   a local `config.yml`. Cloudflare creates the DNS record automatically.

## Demo credentials

* **Email:** `demo@example.com`
* **Password:** `Demo123!`
* Data: 2 accounts (one **RSD**, one **EUR**), ~180 transactions spread over the
  last 60 days, balances kept positive by the seed's income blocks. Account
  balances are computed automatically by the transactions trigger.

## Verification (health / smoke)

| Check        | Command                                                      |
| ------------ | ----------------------------------------------------------- |
| REST health  | `curl http://192.168.13.80:18080/health`                    |
| gRPC (LAN)   | `grpcurl -plaintext 192.168.13.80:15051 list`               |
| Public web   | `curl -I https://staging.digitlock.systems/`                |
| API via edge | `curl -I https://staging.digitlock.systems/api/v1/health`   |

A successful bring-up: `/health` returns 200, `grpcurl list` enumerates the
exactly **5** registered services — `AuthService`, `AccountService`,
`CategoryService`, `TransactionService`, `ReportService` (plus gRPC reflection)
— and the public hostname serves the Vue app with API calls routed through the
edge. There is no currency gRPC service in the backend; CRS, when present, is a
separate upstream the backend calls as a client.

## Ports

| Service          | Host port            | Container port |
| ---------------- | -------------------- | -------------- |
| Postgres         | — (not published)    | 5432           |
| Backend REST     | 18080                | 8080           |
| Backend gRPC     | 15051                | 50051          |
| Frontend / edge  | 18090 (nginx)        | —              |

## Notes

* **gRPC is not proxied through Cloudflare.** The free plan does not carry
  plaintext gRPC, and the channel is intentionally plaintext for LAN tooling, so
  gRPC stays **LAN-only by design**. Mobile/dev clients use REST over the public
  hostname or gRPC directly on the LAN.
* **Migration gap at 008 is expected.** `008_demo_seed_data.sql` is a *seed*,
  not a migration — the migration sequence is `001–007, 009–014`.
  golang-migrate permits non-contiguous version numbers, so the missing `008`
  is harmless; staging applies migrations up to schema version **14** and seeds
  separately via `staging_seed.sql`.
* **`JWT_SECRET` is unique to staging.** It is generated per environment and not
  shared with dev or live, so tokens are not portable across environments. The
  same applies to `DB_PASSWORD`.
* **CRS absence is a supported mode**, not an outage — see Overview. When a real
  CRS is added later, set `CURRENCY_SERVICE_ADDR` in `.env.staging` and restart
  the backend.
