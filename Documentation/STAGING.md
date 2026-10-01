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
backend tolerates its absence: with `CURRENCY_SERVICE_ADDR` left empty in
`.env.staging`, currency sync still runs, fails harmlessly against the default
`localhost:50052` (see *Notes*), and reports fall back to stored/identity rates.

## Topology

| Layer            | Node            | Address                                   | External |
| ---------------- | --------------- | ----------------------------------------- | -------- |
| Postgres         | <staging-host>  | internal Docker network only (no publish) | No       |
| Backend REST     | <staging-host>  | `:18080` (host) → `:8080` (container)      | Via edge |
| Backend gRPC     | <staging-host>  | `:15051` (host) → `:50051` (container)     | LAN only |
| Frontend (static)| <edge-host>     | nginx `:18090`, serves Vue `dist/`         | Via tunnel |
| Edge (proxy)     | <edge-host>     | nginx, `/api` → `<staging-host>:18080`     | Via tunnel |
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
  * `postgres` — `postgres:16-alpine`, named volume `pgdata`
    (`expense-tracker-staging_pgdata`), healthcheck `pg_isready`,
    `restart: unless-stopped`.
  * `migrate` — `migrate/migrate:v4.17.1`, one-shot, runs `up` then exits
    (`restart: "no"`), `depends_on: postgres (service_healthy)`.
  * `backend` — built from the repo `Dockerfile` (multi-stage: buf codegen →
    static Go binary → `alpine:3.20`, runs as unprivileged `appuser`).
    `depends_on: postgres (service_healthy)` + `migrate
    (service_completed_successfully)`, `restart: unless-stopped`.
* **Configuration:**
  * `.env.staging` — **gitignored**, holds the real secret *values*. The master
    copy lives on the Mac and travels to the staging node with the rsync of the
    tree, so both copies stay identical. `.dockerignore` excludes `.env*`, so it
    never enters the backend image.
  * `.env.staging.example` — committed template (variable names, no secret
    values). Copy it to `.env.staging` and fill in the blanks.
* **Frontend build:** `npm run build` → `frontend/dist/`. Built with the
  **default relative API base `/api/v1`** (no `VITE_API_BASE_URL`). Because the
  edge serves the bundle and proxies `/api` to the backend on the same origin,
  **no CORS configuration is required**.

## Bring-up

> All secret values come from `.env.staging` (synced from the Mac). See
> `.env.staging.example` for the full list of variables to populate
> (`DB_PASSWORD`, `JWT_SECRET`, etc.). Never copy secret values into this doc or
> into git.

> **Docker needs root on the staging node.** The deploy user is not in the
> `docker` group, so every `docker compose` / `docker` command below runs with
> `sudo` (password prompt). Run them in your own SSH session.

Backend node — `<staging-host>`:

1. **Sync the tree** from the Mac to the staging node. Always dry-run first
   (`-n -i`) and check the `*deleting` lines, then run it without `-n -i`:

   ```sh
   rsync -avn -i --delete \
     --exclude '.git' --exclude 'bin' --exclude 'frontend/node_modules' \
     --exclude '/.env' --exclude '/server' --exclude '/expense-tracker' \
     --exclude '.claude' --exclude '.idea' --exclude '.vscode' --exclude '.DS_Store' \
     --exclude 'frontend/dist' \
     ./ "<user>@<staging-host>:~/expense-tracker-staging/"
   ```

   * `/.env` (local dev config) and the root binaries `/server`,
     `/expense-tracker` (macOS builds) must not reach the node. Keep the leading
     `/`: an unanchored `server` would also drop `cmd/server/`.
   * Excluded paths are also protected from `--delete`, so leftovers already on
     the node are kept, not removed.

2. **Build the backend image:**

   ```sh
   cd ~/expense-tracker-staging
   sudo docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging build backend
   ```

3. **Start Postgres, then run migrations** (one-shot, brings schema to
   version **14**):

   ```sh
   sudo docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging up -d postgres
   sudo docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging up migrate
   ```

4. **Seed staging data.** Postgres is not published, so the seed is piped into
   `psql` inside the container:

   ```sh
   sudo docker compose -f docker-compose.staging.yml --env-file .env.staging \
     -p expense-tracker-staging exec -T postgres \
     psql -U expense_staging -d expense_tracker_staging < database/seeds/staging_seed.sql
   ```

   The seed is **idempotent** — it `DELETE`s this family's rows by `family_id`
   and recreates them, so re-running it never duplicates data. It provisions the
   demo login `demo@example.com` / `Demo123!`.

5. **Start the backend:**

   ```sh
   sudo docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging up -d backend
   ```

Edge / frontend node — `<edge-host>`:

6. **Build & publish the frontend.** nginx serves `/var/www/staging`, which is
   owned by `www-data` and not writable by the deploy user (and the edge node
   has no `sudo`). Publishing is two steps: rsync into `~/staging-web/`, then
   copy into place as root.

   ```sh
   # On the Mac (VITE_API_BASE_URL must be unset → relative /api/v1):
   cd frontend && npm ci && npm run build          # → dist/
   rsync -avn -i --delete dist/ "<user>@<edge-host>:~/staging-web/"   # dry run
   rsync -av --delete dist/ "<user>@<edge-host>:~/staging-web/"

   # On <edge-host>, as root:
   su -c 'rsync -a --delete /home/<user>/staging-web/ /var/www/staging/ && chown -R www-data:www-data /var/www/staging'
   ```

   nginx site `expense-staging` listens on `:18090`, serves
   `/var/www/staging` (SPA fallback to `index.html`), reverse-proxies
   `location /api/ → http://<staging-host>:18080`, and exposes the backend
   health check as `location = /health`.

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

## Updating staging (redeploy)

For a release **without new migrations** (for example v0.4.2). Steps 1 and the
web publish run from the Mac; the `sudo` steps run on `<staging-host>` in
`~/expense-tracker-staging`.

1. **Sync the tree** — dry run, review deletions, real run (Bring-up step 1).
2. **Build the new image** while the old backend keeps serving:

   ```sh
   sudo docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging build backend
   ```

3. **Recreate Postgres only if its service definition changed** (image,
   environment, restart policy). Do it deliberately rather than as a side effect
   of starting the backend; the data stays on the `pgdata` volume, but the
   database is briefly down:

   ```sh
   sudo docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging up -d postgres
   ```

4. **Restart the backend.** `migrate` runs again as a dependency; with no new
   migrations it reports `no change`:

   ```sh
   sudo docker compose -f docker-compose.staging.yml \
     --env-file .env.staging -p expense-tracker-staging up -d backend
   ```

5. **Verify schema and logs** — expect `version` = the number of the latest
   file in `database/migrations` and `dirty` = `f` (`14|f` as of v0.4.2), and
   no startup errors (the currency sync warning under *Notes* is expected):

   ```sh
   sudo docker compose -f docker-compose.staging.yml --env-file .env.staging \
     -p expense-tracker-staging exec -T postgres \
     sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "SELECT version, dirty FROM schema_migrations"'
   sudo docker compose -f docker-compose.staging.yml --env-file .env.staging \
     -p expense-tracker-staging logs --since 10m migrate backend
   ```

6. **Publish the frontend** if it changed (Bring-up step 6), then run the
   checks under *Verification*.

## Demo credentials

* **Email:** `demo@example.com`
* **Password:** `Demo123!`
* Data: 2 accounts (one **RSD**, one **EUR**), ~180 transactions spread over the
  last 60 days, balances kept positive by the seed's income blocks. Account
  balances are computed automatically by the transactions trigger.

## Verification (health / smoke)

| Check        | Command                                                      |
| ------------ | ----------------------------------------------------------- |
| REST health  | `curl "http://<staging-host>:18080/health"`                 |
| gRPC (LAN)   | `grpcurl -plaintext "<staging-host>:15051" list`            |
| Public web   | `curl -I https://staging.digitlock.systems/`                |
| API via edge | `curl https://staging.digitlock.systems/health`             |

The backend serves health only at `/health` (not under `/api/v1`), and the
edge proxies exactly that path, so `/api/v1/health` returns 404 by design.

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
* **`JWT_SECRET` is required.** The backend exits at startup with
  `JWT_SECRET is not set` if it is empty. Only `APP_ENV=dev` enables the
  insecure development fallback — do not set `APP_ENV` on staging.
* **CRS absence is a supported mode**, not an outage — see Overview. When a real
  CRS is added later, set `CURRENCY_SERVICE_ADDR` in `.env.staging` and restart
  the backend.
* **Known issue — currency sync warning in the logs.** With
  `CURRENCY_SERVICE_ADDR` empty, the backend falls back to `localhost:50052`,
  finds nothing there, and logs `WARN: initial currency sync failed: …` on start
  and `WARN: currency sync failed: …` every `CURRENCY_SYNC_INTERVAL` (6h). This
  is expected on staging and is **not** a deploy failure.
