# echo-basics

> A simple remote logging REST-ish API built with Go, Echo v5, GORM and Postgres. This is an in-person tutorial project — my students and I put this together to learn Go web development the right way, using the latest versions of everything.

Fair warning: I've intentionally left some bad practices in the code. If you spot them, open a PR — that's part of the exercise.

---

## What is this?

A remote logging API. You send logs to it, it stores them, you can fetch and delete them. That's it. Simple on the surface, but there's enough going on under the hood to learn a lot from it — migrations, enums, context injection, rate limiting, and more.

## Stack

- [Go 1.25+](https://go.dev/)
- [Echo v5](https://github.com/labstack/echo) — HTTP framework
- [GORM v2](https://gorm.io/) — ORM
- [gormigrate](https://github.com/go-gormigrate/gormigrate) — versioned migrations
- [PostgreSQL](https://www.postgresql.org/) — database

---

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL running locally (or remotely, doesn't matter)
- `jq` (optional, but makes test output much nicer)

### Environment

Create a `.dev.env` file in the project root (or a `.env` file — both are loaded):

```
DSN=postgres://user:pass@localhost:5432/dbname?sslmode=disable
PORT=6070
LIMIT_RATE=20
```

### Run

```sh
go run .
```

That's it. Migrations run automatically on startup — they're idempotent so running them multiple times won't blow anything up.

---

## Project Structure

```
echo-basics/
├── docs/               # documentation (you're reading it)
├── migrations/         # gormigrate migration definitions
├── models/             # GORM models, enums, app context
├── modules/            # db init, env loader, utilities
├── routes/             # route handlers
├── tests/              # curl-based smoke tests
├── server.go           # entry point
├── go.mod
└── LICENSE
```

---

## API Reference

Base path: `/api`

### Health

```
GET /api/health
```

Just checking if we're alive.

**Response 200**
```json
{
  "status": "ok",
  "message": "Yeppers, seems good."
}
```

---

### Create Log

```
POST /api/create
```

Creates a new log entry. `flag` defaults to `info` if not provided. Case-insensitive — send `INFO` or `info`, we don't mind.

**Request body**
```json
{
  "flag": "info",
  "message": "something happened"
}
```

**Allowed flag values**

| Flag    | Level |
|---------|-------|
| `log`   | 0     |
| `debug` | 1     |
| `info`  | 2     |
| `warn`  | 3     |
| `error` | 4     |
| `trace` | 5     |

**Response 201**
```json
{
  "ID": "6af05bdd-2b64-4365-a600-b7d87a169da5",
  "Flag": "info",
  "Message": "something happened",
  "Timestamp": "2026-02-23T00:00:00Z"
}
```

**Response 400** — bad payload or invalid flag  
**Response 500** — something went wrong on our end

---

### List Logs

```
GET /api/list
```

Returns all logs. Yes, all of them. Pagination is a task left for you — go ahead and open a PR.

**Response 200** — array of log objects

---

### Fetch by ID

```
GET /api/fetch/i/:id
```

Fetch a single log by its UUID.

**Response 200** — single log object  
**Response 400** — that's not a UUID  
**Response 404** — not found

---

### Fetch by Timestamp

```
GET /api/fetch/t/:timestamp
```

Returns the latest log at or before the given timestamp. Timestamp must be RFC3339 format (e.g. `2026-02-23T00:00:00Z`).

**Response 200** — single log object  
**Response 400** — bad timestamp  
**Response 404** — nothing found before that timestamp

---

### Fetch by Flag

```
GET /api/fetch/f/:flag
```

Returns all logs with the given flag, ordered by timestamp descending. Case-insensitive.

**Response 200** — array of log objects  
**Response 400** — invalid flag value  
**Response 500** — DB error

---

### Delete Log

```
DELETE /api/delete/:id
```

Deletes a log by UUID. There's a catch — you can only delete logs with a flag level **below 4** (i.e. `log`, `debug`, `info`, `warn`). `error` and `trace` are off limits.

**Response 200** — deleted  
**Response 400** — bad UUID  
**Response 403** — flag level too high, not allowed  
**Response 404** — log not found  
**Response 500** — DB error

---

## Running Tests

Smoke tests live under `tests/`. They use `curl` — no frameworks, no fuss.

```sh
# Run everything in order
./tests/run_all.sh

# Or run individually
./tests/health.sh
./tests/create.sh
./tests/list.sh
./tests/fetch_by_flag.sh INFO
./tests/fetch_by_id.sh <uuid>
./tests/delete.sh <uuid>
```

Install `jq` to get pretty-printed output and automatic UUID extraction between steps.

---

## Migrations

Migrations are versioned using `gormigrate` and run automatically at startup via `migrations.Run(db)`. They're idempotent — safe to re-run.

Current migrations:
- `1771799054_init_uuid_and_logs` — creates `uuid-ossp` extension, `log_flag` enum and the `logs` table.

To add a new migration, append a new `*gormigrate.Migration` entry in `migrations/migrations.go` with a unique incrementing ID.

---

## Contributing

If you spot something wrong or want to practice your Go and PR skills — go for it. There are intentional bad practices hidden in the codebase. Find them, fix them, open a PR.

---

## License

[MIT](https://devsimsek.mit-license.org) — Metin Şimşek (devsimsek)