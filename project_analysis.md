# RemoteScope — Project Gap Analysis

## Overview

**RemoteScope** is a full-stack remote job aggregation platform. The codebase is well-structured across four services (Go backend, Python scraper, Python AI service, Next.js frontend). The backend, scraper, and AI service are partially implemented. Several services and components are **referenced but entirely missing**.

---

## 🔴 Missing Services / Directories

These are referenced in `docker-compose.yml` or the `Makefile` but **do not exist at all**.

| Missing | Referenced In | Notes |
|---|---|---|
| `frontend/` | `docker-compose.yml`, `Makefile` | The entire Next.js frontend is absent |
| `notification_service/` | `docker-compose.yml` | A separate Go notification service is wired up but has no directory |
| `infra/nginx/nginx.conf` | `docker-compose.yml` | Nginx config file missing; container will fail to start |
| `infra/nginx/certs/` | `docker-compose.yml` | TLS certs directory missing |
| `infra/prometheus/prometheus.yml` | `docker-compose.yml` | Prometheus config missing; monitoring won't start |
| `infra/grafana/dashboards/` | `docker-compose.yml` | Grafana dashboard provisioning missing |
| `backups/` | `Makefile` (`db-backup`) | Backup target directory doesn't exist |

---

## 🔴 Missing AI Service Routers

`ai_service/main.py` imports five routers that **don't exist**:

```python
from ai_service.routers import classify, embed, deduplicate, extract, summarize
```

There is **no `routers/` package** in `ai_service/`. This means the AI service will crash on startup.

---

## 🔴 Missing Scraper: `__init__.py` / Package Files

The `scraper/` and `ai_service/` directories lack Python package `__init__.py` files. Without them, imports like `from scraper.adapters.base import BaseScraper` will fail unless running via a proper install.

---

## 🟠 Missing Backend Components

### No `go.sum` file
The `go.sum` file is required for reproducible builds. Without it, `go build` / `docker build` will fail.

### No `backend/cmd/api` executable entrypoint wiring for `migrate` and `seed`
The `Makefile` calls `./remotescope migrate` and `./remotescope seed`, but `main.go` has no CLI argument handling (no `migrate` or `seed` subcommands). These make targets will silently do nothing or crash.

### No Rate Limiting Middleware Wired In
`go.mod` includes `github.com/ulule/limiter/v3` and `.env.example` defines `RATE_LIMIT_REQUESTS_PER_MINUTE`, but there is no rate limiting middleware applied to the router in `main.go`.

### No `docs/` Package for Swagger
`main.go` uses `@title` Swaggo annotations and the `Makefile` has a `swagger` target, but there is no `docs/` directory and `swag init` has never been run. The import `_ "github.com/remotescope/backend/docs"` is missing from `main.go`, so Swagger UI won't work.

### Missing Redis Integration in Backend
`go.mod` includes `go-redis`, `.env.example` defines `REDIS_URL`, but there is **no Redis client initialization** in `main.go` and no Redis-backed caching, session storage, or rate limiting in any backend file.

### No `application_tracking` Repository
`main.go` doesn't wire a repository for `AppliedJob` / `HiddenJob` (those actions are inline on `savedRepo`), but the `SavedJobRepository` conflates saved, applied, and hidden jobs — it should have `ListApplied`, `ListSaved`, `HideJob`, etc. The repository file should be audited to confirm all methods exist.

---

## 🟠 Missing Scraper Components

### No `scraper/__init__.py` or `scraper/main.py`
The `Makefile` runs `python -m scraper.main`, but `scraper/main.py` doesn't exist — only `scraper/scheduler.py` does. The `scraper/` directory also has no `__init__.py`.

### No `scraper/requirements.txt` Dockerfile Conflict
`scraper/requirements.txt` exists but its contents haven't been verified for all dependencies (e.g., `apscheduler` used in `scheduler.py`).

### No Test Directory
`Makefile` runs `cd scraper && python -m pytest tests/ -v` and `cd ai_service && python -m pytest tests/ -v`, but **neither `scraper/tests/` nor `ai_service/tests/` exist**.

---

## 🟠 Missing AI Service Components

### No `ai_service/requirements.txt`
The AI service has no `requirements.txt` (unlike scraper). `ai_service/Dockerfile` likely won't be able to install dependencies.

### No `ai_service/Dockerfile`
While a `Dockerfile` is listed in the directory scan for `ai_service`, the build context references it. Cross-verify it handles the `requirements.txt` gap above.

> [!NOTE]
> `ai_service/` has `main.py`, `pipeline.py`, `worker.py` — but no `routers/` package, no `requirements.txt`, and no `tests/`.

---

## 🟡 Incomplete / Stub Code

| File | Issue |
|---|---|
| [companies.go](file:///c:/Users/David/OneDrive/Documentos/job_scraper/backend/internal/api/companies.go) | Only 1 KB — likely a stub |
| [users.go](file:///c:/Users/David/OneDrive/Documentos/job_scraper/backend/internal/api/users.go) | Needs `GetMe`, `UpdateMe`, `UpdatePreferences`, `CompleteOnboarding` — needs audit |
| [admin.go](file:///c:/Users/David/OneDrive/Documentos/job_scraper/backend/internal/api/admin.go) | `TriggerScraper` calls an external scraper service — HTTP client/URL not configured |
| `main.go` | Inline handlers for `/saved-jobs` and `/applied-jobs` — should be moved to proper handler structs |
| OAuth state verification | `oauth.go` generates a state but there's no CSRF state verification (state stored in session/Redis) on callback |

---

## 🟡 Missing Infrastructure

| Missing | Impact |
|---|---|
| `.env` (not `.env.example`) | App won't run without it — `make dev` auto-creates it but secrets are empty |
| `infra/` directory entirely | Nginx, Prometheus, Grafana can't start |
| `migrations/` has only 1 file | No seed data migration (referenced by `make seed`) |
| No `Dockerfile` for `backend` for production target | `docker-compose.yml` specifies `target: development` — a `production` target may be missing |

---

## 🟡 Missing Features (Planned but Not Implemented)

Based on `.env.example` feature flags and models:

| Feature | Status |
|---|---|
| Email notifications (Resend) | Env configured, `resend-go` in `go.mod`, no sender code |
| Web Push (VAPID) | Env configured, `webpush-go` in `go.mod`, no push code |
| Telegram bot notifications | Env configured, not implemented |
| Discord / Slack webhooks | Env configured, not implemented |
| Job recommendations (`/recommendations` endpoint) | Route exists but `jobsHandler.Recommendations` may be a stub |
| Semantic search (`/search` with embeddings) | Route exists, need to confirm vector search is wired |
| Application status tracking (phone screen → offer) | Model exists, no dedicated API endpoints |

---

## Priority Order to Build Out

1. **`frontend/`** — The most impactful missing piece (entire UI)
2. **`ai_service/routers/`** — Prevents AI service from starting
3. **`infra/`** — Nginx config, Prometheus, Grafana (needed for `make dev` to fully work)
4. **`notification_service/`** — Separate service wired into compose but absent
5. **`scraper/main.py` + `__init__.py` files** — Needed for `make scrape`
6. **Backend: Redis client + rate limiting** — Security gap
7. **Backend: OAuth CSRF state verification** — Security gap
8. **Backend: `migrate`/`seed` CLI commands** — Needed for DB setup
9. **Tests** — `scraper/tests/`, `ai_service/tests/`, `backend/` tests
10. **`go.sum`** — Required for reproducible Go builds
