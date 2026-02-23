# 04 — Databases

> before we touch GORM or write a single query, let's make sure we're on the same page about what a database actually is and why we need one.

---

## what is a database?

A database is a structured way to store, retrieve and manage data that persists beyond the lifetime of your application. When your app restarts, the data is still there. When ten users hit your API at the same time, the database handles concurrent reads and writes safely.

Without a database, your app's data lives in memory — and the moment the process dies, so does everything in it. That's fine for a cache, not fine for a logging API.

---

## types of databases

There are two main categories you'll encounter:

### relational databases (SQL)

Data is stored in **tables** with **rows** and **columns**. Tables relate to each other via foreign keys. You query them with SQL (Structured Query Language).

Examples: PostgreSQL, MySQL, SQLite, MariaDB

This is what we're using. Specifically **PostgreSQL**.

### non-relational databases (NoSQL)

Data is stored in documents, key-value pairs, graphs, or time-series formats — depending on the type. No fixed schema, no SQL.

Examples: MongoDB (documents), Redis (key-value), Cassandra (wide-column)

Great for specific use cases. Not what we need here.

---

## why PostgreSQL?

We picked Postgres specifically. Here's why:

- **it's rock solid.** Postgres has been in active development since 1996. It's battle-tested, well-documented, and trusted at scale.
- **it has great types.** UUID support, JSONB, custom ENUM types, arrays — Postgres has types that other databases don't.
- **it's standards-compliant.** It follows the SQL standard closely, which means knowledge transfers well.
- **it handles concurrency properly.** MVCC (Multi-Version Concurrency Control) means reads don't block writes and vice versa.
- **it's free.** As in free forever, not "free tier with a 500MB limit."

---

## core concepts you need to know

### tables

A table is like a spreadsheet. It has columns (fields) and rows (records). Each column has a fixed type.

In our project, we have one table: `logs`.

```sql
CREATE TABLE logs (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag      log_flag NOT NULL,
    message   TEXT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

### rows

A row is one record in a table. Every time someone calls `POST /api/create`, a new row gets inserted into `logs`.

### primary keys

Every row should have a unique identifier — the primary key. In our case, that's `id`, a UUID. No two rows will ever have the same `id`.

Why UUID and not an auto-incrementing integer? A few reasons:

- UUIDs are globally unique — you can generate them anywhere (client, server, DB) without coordination.
- They don't leak information. An integer ID of `42` tells an attacker there are at least 42 records. A UUID tells them nothing.
- They're safer to expose in URLs.

The downside is they're larger (16 bytes vs 4 bytes for an int) and slightly slower to index. For most applications, that trade-off is worth it.

### data types

Postgres has a rich type system. The ones we use in this project:

| type | what it is |
|---|---|
| `UUID` | universally unique identifier, 128-bit |
| `TEXT` | variable-length string, no size limit |
| `TIMESTAMPTZ` | timestamp with timezone info |
| `ENUM` | a fixed set of allowed string values |

### ENUM types

This one deserves its own paragraph because it's not something every database does well.

In Postgres, you can define a custom ENUM type — a column that only accepts a predefined list of values. We use this for our `flag` column:

```sql
CREATE TYPE log_flag AS ENUM ('log', 'debug', 'info', 'warn', 'error', 'trace');
```

Once created, any attempt to insert a value outside this list will be rejected by the database itself — not by your application code. That's a hard guarantee at the data layer.

```sql
INSERT INTO logs (flag, message) VALUES ('info', 'something happened'); -- works
INSERT INTO logs (flag, message) VALUES ('INFO', 'something happened'); -- fails — case sensitive
INSERT INTO logs (flag, message) VALUES ('critical', 'something happened'); -- fails — not in enum
```

This is why we normalize flag values to lowercase in the application before inserting. The database enforces the allowed values; the application normalizes the input.

---

## how your app talks to the database

Your Go application talks to Postgres over a network connection (even if Postgres is running on the same machine — it's still a TCP connection). The connection details are in the **DSN** (Data Source Name):

```
postgres://user:password@host:5432/dbname?sslmode=disable
```

Breaking that down:

| part | meaning |
|---|---|
| `postgres://` | the driver / protocol |
| `user:password` | credentials |
| `host` | where Postgres is running |
| `5432` | default Postgres port |
| `dbname` | which database to connect to |
| `sslmode=disable` | disable SSL (fine for local, not for production) |

We store the DSN in an environment variable (`DSN`) and never hardcode it. That's not optional — hardcoding credentials in code is a serious security issue and you will regret it.

---

## connection pooling

Opening a new database connection is expensive — it involves a TCP handshake, authentication, and setup. You don't want to do that for every request.

Instead, you maintain a **connection pool** — a set of pre-opened connections that are reused across requests. Your application borrows a connection from the pool when it needs it and returns it when done.

GORM (and the underlying `database/sql` package) handles this for you automatically. You can tune the pool size:

```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)      // max connections open at once
sqlDB.SetMaxIdleConns(10)      // max idle connections in the pool
sqlDB.SetConnMaxLifetime(time.Hour) // max lifetime of a connection
```

For this project we're using defaults, which is fine for a tutorial. In production, tune these based on your Postgres `max_connections` setting and expected traffic.

---

## transactions

A transaction is a group of database operations that either all succeed or all fail together. No partial states.

The classic example is a bank transfer: debit one account, credit another. If the credit fails after the debit, you've just lost money. A transaction ensures either both happen or neither does.

In our project, we wrap our create and delete operations in transactions:

```go
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&log).Error; err != nil {
        return err // transaction rolls back automatically
    }
    return nil // transaction commits
})
```

If the function returns an error, the transaction rolls back. If it returns nil, it commits. Clean and safe.

---

## extensions

Postgres supports extensions — optional modules that add extra functionality. We use one:

**uuid-ossp** — provides the `uuid_generate_v4()` function, which generates random UUIDs. Without this extension, we can't use `DEFAULT uuid_generate_v4()` on our `id` column.

We install it in our first migration:

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

The `IF NOT EXISTS` makes it safe to run multiple times — it won't error if the extension is already installed.

---

## SQL basics — just enough to read what's going on

You don't need to be a SQL expert for this project, but you should be able to read basic queries. Here's a cheat sheet:

```sql
-- select all rows
SELECT * FROM logs;

-- select specific columns
SELECT id, flag, message FROM logs;

-- filter with WHERE
SELECT * FROM logs WHERE flag = 'error';

-- filter by timestamp
SELECT * FROM logs WHERE timestamp <= '2026-02-23T00:00:00Z' ORDER BY timestamp DESC LIMIT 1;

-- insert a row
INSERT INTO logs (flag, message) VALUES ('info', 'hello') RETURNING *;

-- delete a row
DELETE FROM logs WHERE id = 'some-uuid';

-- check the table structure
\d+ logs
```

---

## alright, you know databases now

You understand what a relational database is, why we use Postgres, what the `logs` table looks like, how connections work, and why transactions matter. Next up — we need a way to talk to all of this from Go without writing raw SQL everywhere. That's GORM.

→ next: [05 — GORM](./05-gorm.md)

---

*MIT License — [devsimsek.mit-license.org](https://devsimsek.mit-license.org)*