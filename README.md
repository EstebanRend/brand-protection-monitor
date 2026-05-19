# Brand Protection Monitor PoC

A full-stack Proof-of-Concept application that monitors recent Certificate Transparency log entries and detects possible brand phishing/domain abuse based on configurable keywords.

## Tech Stack

- Backend: Go
- Database: PostgreSQL
- Frontend: React, TypeScript, Tailwind CSS
- Communication: REST APIs
- Infrastructure: Docker Compose

## Requirements (Tested / Expected)

- Go `1.24.x` (see `backend/go.mod`, toolchain `go1.24.13`)
- PostgreSQL `16` (local container uses `postgres:16-alpine`)
- Node.js `20+` and npm `10+` (recommended for frontend tooling)
- Docker Engine + Docker Compose v2 (for local Postgres)

## Architecture

```text
React Dashboard
  ↓ REST API
Go HTTP API
  ↓
Keyword Service / Certificate Service / Export Service
  ↓
PostgreSQL

Background Monitor Worker
  ↓
CT Log Client
  ↓
Certificate Parser
  ↓
Keyword Matcher
  ↓
PostgreSQL matched_certificates
```

## Repository Structure

```text
.
├─ backend/
│  ├─ cmd/api/                  # App entrypoint, wiring, HTTP server startup
│  └─ internal/
│     ├─ certificates/          # Matched certificate model + persistence
│     ├─ config/                # Environment config loading
│     ├─ db/                    # DB connection + migration runner
│     ├─ exporter/              # CSV export writer
│     ├─ httpapi/               # REST routes + handlers + CORS middleware
│     ├─ keywords/              # Keyword model + repository
│     └─ monitor/               # CT client, monitor worker, state tracking
├─ frontend/
│  ├─ src/
│  │  ├─ api/                   # API client wrapper
│  │  ├─ components/            # Dashboard UI sections
│  │  ├─ types/                 # Shared frontend data types
│  │  └─ App.tsx                # Main dashboard composition
├─ infra/docker-compose.yml     # Local PostgreSQL service
├─ migrations/                  # SQL schema migrations
├─ scripts/migrate.mjs          # Root migration helper script
└─ package.json                 # Root scripts (migration commands)
```

## Implemented Features

- Add, remove, and view monitored keywords.
- Persist keywords in PostgreSQL.
- Periodically process a recent batch of CT log entries.
- Parse certificate data from CT log `get-entries` response.
- Match Common Name and SAN domains against stored keywords.
- Store only matched certificates.
- Dashboard showing monitor status and processing metrics.
- Highlight matched certificates in the UI.
- CSV export endpoint and frontend export button.
- Configurable CT log URL, CT request timeout, CORS allowed origins, batch size, monitor interval, and database connection.

## Setup Instructions

### 1. Start PostgreSQL

```bash
cd infra
docker compose up -d
```

### 2. Run Backend

```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/api
```

The backend runs on:

```text
http://localhost:8080
```

If monitor logs show `get-sth failed with status 404`, set `CT_LOG_BASE_URL` to an RFC6962-compatible endpoint (default in this repo):

```text
https://ct.googleapis.com/logs/us1/argon2026h1/ct/v1
```

Migrations are executed automatically on backend startup using `golang-migrate`.
By default, backend reads migrations from `../migrations` (configurable via `MIGRATIONS_DIR`).

### Database Migrations

Migration files live in `migrations` (repository root) and follow this format:

- `000001_name.up.sql`
- `000001_name.down.sql`

Install migration CLI once:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Then use root-level npm scripts:

```bash
npm run mig:new -- add_keyword_description
npm run mig:up
npm run mig:down -- 1
npm run mig:version
```

This generates matching `up` and `down` SQL files for `mig:new`. Add your schema changes in:

- `*.up.sql` for apply
- `*.down.sql` for rollback

Apply latest migrations automatically through backend startup:

```bash
cd backend
go run ./cmd/api
```

Manual rollback example (via script):

```bash
npm run mig:down -- 1
```

### 3. Run Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend runs on:

```text
http://localhost:5173
```

Frontend environment variable:

- `VITE_API_BASE_URL` (optional): backend base URL (default `http://localhost:8080`).

## Useful API Endpoints

```text
GET    /health
GET    /api/keywords
POST   /api/keywords
DELETE /api/keywords/{id}
GET    /api/matches
GET    /api/status
POST   /api/monitor/run-once
GET    /api/export.csv
```

## Data Model (Current Schema)

- `keywords`: monitored keyword values (unique).
- `matched_certificates`: matched domain, issuer, validity window, matched keyword, and source log.
- `monitor_state`: singleton row tracking last processed tree size, last cycle count, status, and timestamps.

Uniqueness rule:

- `matched_certificates` uses `UNIQUE(domain, matched_keyword, source_log)` to avoid duplicate findings for the same source log.

## Testing and Verification

Backend tests:

```bash
cd backend
go test ./...
```

Frontend build verification:

```bash
cd frontend
npm run build
```

## Design Decisions

- The monitor runs as a background worker inside the Go API process to keep the PoC simple.
- The batch size is configurable via `BATCH_SIZE`.
- Only matched certificates are persisted, as required by the PoC scope.
- Duplicate matches are prevented using a unique database constraint on domain, keyword, and source log.
- A manual `run-once` endpoint was added to make demos and testing easier.
- Keyword matching is currently case-insensitive substring matching.

## Limitations and Future Improvements

- The CT parser focuses on the first certificate embedded in `extra_data`, which is enough for this PoC but could be expanded for full chain processing.
- The monitor currently checks the latest batch every cycle, so older unmatched entries are not stored.
- Future improvements could include WebSocket live updates, pagination, authentication, multiple CT log sources, regex matching, severity scoring, and queue-based processing.
