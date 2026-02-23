# 08 — API Reference

> everything you need to know to talk to this API. no fluff, just the endpoints, the expected inputs, and what you'll get back.

---

## base URL

```
http://localhost:{PORT}
```

`PORT` is set in your environment. Default in development is `6070`.

All API endpoints live under `/api`:

```
http://localhost:6070/api
```

---

## authentication

there isn't any. this is a tutorial project. if you're deploying this somewhere public, add auth before you do anything else — that's a good PR by the way.

---

## request format

for endpoints that accept a body, send JSON with the `Content-Type: application/json` header:

```sh
curl -X POST http://localhost:6070/api/create \
  -H "Content-Type: application/json" \
  -d '{"flag":"info","message":"something happened"}'
```

---

## response format

all responses are JSON. errors look like this:

```json
{
  "error": "some message explaining what went wrong"
}
```

success responses vary by endpoint — either a single object or an array.

---

## rate limiting

requests are rate-limited per IP. the limit is controlled by the `LIMIT_RATE` environment variable (requests per second). if you exceed the limit, you'll get a `429 Too Many Requests` response. back off and try again.

---

## endpoints

### health check

```
GET /api/health
ANY /api/health
```

checks if the server is alive. responds to any HTTP method.

**response 200**
```json
{
  "status": "ok",
  "message": "Yeppers, seems good."
}
```

**curl example**
```sh
./tests/health.sh
# or manually:
curl http://localhost:6070/api/health
```

---

### create log

```
POST /api/create
```

creates a new log entry. `flag` is case-insensitive — we normalize it to lowercase for you. if you don't provide a `flag`, it defaults to `info`.

**request body**

| field | type | required | description |
|---|---|---|---|
| `flag` | string | no | log severity flag. defaults to `info` if omitted |
| `message` | string | yes | the log message |

**allowed flag values**

| value | level | deletable? |
|---|---|---|
| `log` | 0 | yes |
| `debug` | 1 | yes |
| `info` | 2 | yes |
| `warn` | 3 | yes |
| `error` | 4 | no |
| `trace` | 5 | no |

**request example**
```json
{
  "flag": "info",
  "message": "user logged in successfully"
}
```

**response 201** — the created log record
```json
{
  "ID": "6af05bdd-2b64-4365-a600-b7d87a169da5",
  "Flag": "info",
  "Message": "user logged in successfully",
  "Timestamp": "2026-02-23T00:30:00Z"
}
```

**response 400** — malformed body or invalid flag
```json
{
  "error": "bad request, i had better expectations from you."
}
```

**response 400** — flag not in allowed set
```json
{
  "error": "invalid flag value"
}
```

**response 500** — database error
```json
{
  "error": "..." 
}
```

**curl example**
```sh
./tests/create.sh
# or manually:
curl -X POST http://localhost:6070/api/create \
  -H "Content-Type: application/json" \
  -d '{"flag":"info","message":"hello from curl"}'

# override flag and message via env:
FLAG=warn MESSAGE="something is off" ./tests/create.sh
```

---

### list all logs

```
GET /api/list
```

returns all log records. no filtering, no pagination — yet. (that's a task left for you, go ahead and open a PR.)

**response 200** — array of log objects (empty array if no logs exist)
```json
[
  {
    "ID": "6af05bdd-2b64-4365-a600-b7d87a169da5",
    "Flag": "info",
    "Message": "user logged in successfully",
    "Timestamp": "2026-02-23T00:30:00Z"
  },
  {
    "ID": "9bc12def-3c75-5476-b711-c8e98b270eb6",
    "Flag": "warn",
    "Message": "disk usage above 80%",
    "Timestamp": "2026-02-23T00:35:00Z"
  }
]
```

**response 500** — database error

**curl example**
```sh
./tests/list.sh
# or manually:
curl http://localhost:6070/api/list
```

---

### fetch log by ID

```
GET /api/fetch/i/:id
```

fetches a single log record by its UUID. `:id` must be a valid UUID — if it's not, you'll hear about it.

**path parameters**

| param | type | description |
|---|---|---|
| `id` | UUID | the ID of the log to fetch |

**response 200** — single log object
```json
{
  "ID": "6af05bdd-2b64-4365-a600-b7d87a169da5",
  "Flag": "info",
  "Message": "user logged in successfully",
  "Timestamp": "2026-02-23T00:30:00Z"
}
```

**response 400** — not a valid UUID
```json
{
  "error": "bad request, i had better expectations from you."
}
```

**response 404** — no log found with that ID
```json
{
  "error": "ummm, wait a bit... Nope, its not here. Maybe you didn't send me that log?"
}
```

**response 500** — database error

**curl example**
```sh
./tests/fetch_by_id.sh 6af05bdd-2b64-4365-a600-b7d87a169da5
# or manually:
curl http://localhost:6070/api/fetch/i/6af05bdd-2b64-4365-a600-b7d87a169da5
```

---

### fetch log by timestamp

```
GET /api/fetch/t/:timestamp
```

returns the **single latest log** with a timestamp at or before the provided value. useful for finding what was happening at a specific point in time.

**path parameters**

| param | type | description |
|---|---|---|
| `timestamp` | RFC3339 string | e.g. `2026-02-23T00:30:00Z` |

the timestamp must be URL-encoded when passed in the path — most `curl` and HTTP clients handle this automatically.

**response 200** — the latest log at or before the given timestamp
```json
{
  "ID": "6af05bdd-2b64-4365-a600-b7d87a169da5",
  "Flag": "info",
  "Message": "user logged in successfully",
  "Timestamp": "2026-02-23T00:30:00Z"
}
```

**response 400** — timestamp missing or not valid RFC3339
```json
{
  "error": "bad request, i had better expectations from you."
}
```

**response 404** — no logs exist at or before that timestamp
```json
{
  "error": "ummm, wait a bit... Nope, its not here. Maybe you didn't send me that log?"
}
```

**response 500** — database error

**curl example**
```sh
curl "http://localhost:6070/api/fetch/t/2026-02-23T00:30:00Z"
```

---

### fetch logs by flag

```
GET /api/fetch/f/:flag
```

returns **all logs** matching the given flag, ordered by timestamp descending (newest first). case-insensitive — send `INFO` or `info`, doesn't matter.

**path parameters**

| param | type | description |
|---|---|---|
| `flag` | string | one of: `log`, `debug`, `info`, `warn`, `error`, `trace` |

**response 200** — array of matching log objects (empty array if none match)
```json
[
  {
    "ID": "9bc12def-3c75-5476-b711-c8e98b270eb6",
    "Flag": "warn",
    "Message": "disk usage above 80%",
    "Timestamp": "2026-02-23T00:35:00Z"
  }
]
```

**response 400** — flag not in allowed set
```json
{
  "error": "WHAT'S THAT FLAG? I HAVE NEVER SEEN THAT!!!"
}
```

**response 500** — database error

**curl example**
```sh
./tests/fetch_by_flag.sh warn
# or manually:
curl http://localhost:6070/api/fetch/f/warn
```

---

### delete log

```
DELETE /api/delete/:id
```

deletes a log by UUID. but there's a rule: **you can only delete logs with a flag level below 4**. that means `log`, `debug`, `info`, and `warn` are fair game. `error` and `trace` are protected — those are your audit trail and you're not allowed to delete them. try it and see what happens.

**path parameters**

| param | type | description |
|---|---|---|
| `id` | UUID | the ID of the log to delete |

**response 200** — deleted successfully
```json
{
  "message": "That record is long gone now. Don't worry, our secret is now safe."
}
```

**response 400** — not a valid UUID
```json
{
  "error": "bad request, i had better expectations from you."
}
```

**response 403** — flag level is 4 or above (error or trace), deletion not allowed
```json
{
  "error": "EEEEYYYY! You can't do that!"
}
```

**response 404** — no log found with that ID
```json
{
  "error": "ummm, wait a bit... Nope, its not here. Maybe you didn't send me that log?"
}
```

**response 500** — database error

**curl example**
```sh
./tests/delete.sh 6af05bdd-2b64-4365-a600-b7d87a169da5
# or manually:
curl -X DELETE http://localhost:6070/api/delete/6af05bdd-2b64-4365-a600-b7d87a169da5
```

---

## running the full test suite

the `tests/` directory has curl-based smoke tests that cover all of the above. run them against a live server:

```sh
# start the server first
go run .

# in another terminal, from project root:
./tests/run_all.sh
```

`run_all.sh` runs the tests in a sensible order — health check, create, list, fetch by flag, fetch by ID, delete, then list again to confirm deletion. it needs `jq` installed to auto-extract the created UUID between steps. without `jq` you'll need to paste UUIDs manually.

---

## HTTP status codes used

| status | meaning |
|---|---|
| `200 OK` | request succeeded |
| `201 Created` | record created successfully |
| `400 Bad Request` | your input is wrong |
| `403 Forbidden` | you can't do that (delete protected log) |
| `404 Not Found` | the record doesn't exist |
| `429 Too Many Requests` | slow down, you're being rate limited |
| `500 Internal Server Error` | something went wrong on our end |

---

→ you've reached the end of the docs. go write some code.

---

*MIT License — [devsimsek.mit-license.org](https://devsimsek.mit-license.org)*