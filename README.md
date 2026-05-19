# Brand Protection Monitor PoC

A full-stack Proof-of-Concept application that monitors recent Certificate Transparency log entries and detects possible brand phishing/domain abuse based on configurable keywords.

## Tech Stack

- Backend: Go
- Database: PostgreSQL
- Frontend: React, TypeScript, Tailwind CSS
- Communication: REST APIs
- Infrastructure: Docker Compose

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
- Configurable CT log URL, batch size, monitor interval, and database connection.

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
