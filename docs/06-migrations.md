# 06 — Migrations

> your database schema doesn't just appear out of thin air. migrations are how you get it there — and how you change it over time without breaking everything.

---

## what is a migration?

A migration is a versioned, repeatable script that changes your database schema. Adding a table, dropping a column, creating an enum type — all of these are schema changes that need to happen in a controlled, auditable way.

Think of migrations as a changelog for your database. Every change is a numbered entry in that log. You can apply them forward (migrate up) or reverse them (migrate down/rollback). And crucially — you always know exactly what state your database is in.

---

## why not just use AutoMigrate?

Because AutoMigrate is a shortcut that works until it doesn't.

Here's what AutoMigrate does:
- it reads your Go model structs
- it compares them to the current database schema
- it tries to bring the database in line with your models

Sounds great. The problems:

- **it never removes columns.** renamed a field? the old column stays forever.
- **it can't create custom Postgres types.** no ENUM creation, no extensions, nothing fancy.
- **it can fail silently or destructively** when column types change in ways Postgres can't cast automatically.
- **there's no history.** no log, no version, no rollback. if something goes wrong, good luck.
- **it runs on every startup.** that slow `SELECT count(*) FROM information_schema.tables` you saw in the logs? that was AutoMigrate checking if the table exists. every. single. startup.

We removed AutoMigrate from this project for exactly these reasons. You should too, once you're past prototyping.

---

## what we use instead — gormigrate

We use [gormigrate](https://github.com/go-gormigrate/gormigrate) — a small, simple migration library built on top of GORM. It gives you:

- versioned migrations with unique IDs
- `Migrate` (up) and `Rollback` (down) functions per migration
- a `migrations` table in your database that tracks which migrations have been applied
- idiomatic Go — migrations are just functions, not SQL files

Install it:

```sh
go get github.com/go-gormigrate/gormigrate/v2
```

---

## how gormigrate works

When you call `gormigrate.New(db, options, migrations).Migrate()`, it:

1. checks (or creates) a `migrations` table in your database
2. looks at which migration IDs are already recorded there
3. runs the `Migrate` function of any migration that hasn't been applied yet
4. records the migration ID in the `migrations` table on success

That's it. On the next run, it sees those IDs already recorded and skips them. Safe to call every startup.

---

## our migration setup

Our migration lives in `migrations/migrations.go`:

```go
package migrations

import (
    "github.com/go-gormigrate/gormigrate/v2"
    "go.smsk.dev/pkgs/basics/echo-basics/models"
    "gorm.io/gorm"
)

func Run(db *gorm.DB) error {
    m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
        {
            ID: "1771799054_init_uuid_and_logs",
            Migrate: func(tx *gorm.DB) error {
                // Ensure uuid extension exists (Postgres)
                if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
                    return err
                }

                // Create enum type only if it doesn't exist
                var typCount int64
                if err := tx.Raw(`SELECT count(*) FROM pg_type WHERE typname = ?`, "log_flag").Scan(&typCount).Error; err != nil {
                    return err
                }
                if typCount == 0 {
                    if err := tx.Exec(`CREATE TYPE log_flag AS ENUM ('log', 'debug', 'info', 'warn', 'error', 'trace');`).Error; err != nil {
                        return err
                    }
                }

                // Create logs table only if it doesn't exist
                if !tx.Migrator().HasTable(&models.Log{}) {
                    if err := tx.Migrator().CreateTable(&models.Log{}); err != nil {
                        return err
                    }
                }

                return nil
            },
            Rollback: func(tx *gorm.DB) error {
                // Drop logs table if exists
                if tx.Migrator().HasTable(&models.Log{}) {
                    if err := tx.Migrator().DropTable(&models.Log{}); err != nil {
                        return err
                    }
                }

                // Drop enum type if exists
                if err := tx.Exec(`
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'log_flag') THEN
    DROP TYPE log_flag;
  END IF;
END
$$;`).Error; err != nil {
                    return err
                }

                return nil
            },
        },
    })

    return m.Migrate()
}
```

And in `server.go`, we call it right after opening the DB — before any routes are registered or requests are served:

```go
if err := migrations.Run(db); err != nil {
    e.Logger.Error("migrations failed", "error", err)
}
```

---

## breaking down our first migration

Let's go through what this migration actually does, step by step.

### step 1 — uuid extension

```go
tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
```

Installs the `uuid-ossp` Postgres extension. This gives us `uuid_generate_v4()` — the function our `id` column uses to generate UUIDs automatically. `IF NOT EXISTS` makes it safe to run multiple times.

### step 2 — create the enum type

```go
var typCount int64
tx.Raw(`SELECT count(*) FROM pg_type WHERE typname = ?`, "log_flag").Scan(&typCount)
if typCount == 0 {
    tx.Exec(`CREATE TYPE log_flag AS ENUM ('log', 'debug', 'info', 'warn', 'error', 'trace');`)
}
```

We check whether the `log_flag` type already exists in `pg_type` (Postgres's type catalog) before creating it. Why? Because `CREATE TYPE` doesn't have an `IF NOT EXISTS` option in older Postgres versions. This check makes the migration idempotent — safe to run multiple times without failing.

### step 3 — create the table

```go
if !tx.Migrator().HasTable(&models.Log{}) {
    tx.Migrator().CreateTable(&models.Log{})
}
```

We check if the `logs` table already exists before creating it. Same reason — idempotency. If the table is already there, we skip creation. No errors, no drama.

### the rollback

The `Rollback` function reverses the migration cleanly:
1. drops the `logs` table if it exists
2. drops the `log_flag` enum type if it exists (using a `DO $$` block for conditional dropping)

Note the order — table first, then type. You can't drop an enum type that's still being used by a table column.

---

## idempotency — why it matters

An idempotent operation is one that can be run multiple times and always produces the same result. Our migration is idempotent — you can run it ten times in a row and nothing breaks after the first run.

This matters because:
- migrations run on every startup in our current setup
- development databases get reset and recreated frequently
- CI pipelines run migrations fresh on every test run
- operators sometimes accidentally run migrations twice

A non-idempotent migration that fails on second run will break your startup and your CI. Don't write those.

---

## migration IDs

Every migration has a unique ID:

```go
ID: "1771799054_init_uuid_and_logs",
```

We use a Unix timestamp prefix followed by a descriptive name. The timestamp ensures ordering — migrations are applied in ascending ID order. The descriptive name tells you what the migration does without having to read it.

Keep IDs unique and increasing. Don't edit an existing migration ID after it's been applied to any database — gormigrate uses the ID to track what's been applied.

---

## adding new migrations

When you need to change the schema — add a column, add a new table, change a type — you add a new entry to the migrations slice. You never edit existing migrations.

```go
gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
    {
        // existing migration — don't touch this
        ID: "1771799054_init_uuid_and_logs",
        Migrate: func(tx *gorm.DB) error { ... },
        Rollback: func(tx *gorm.DB) error { ... },
    },
    {
        // new migration — add new ones here
        ID: "1771900000_add_source_to_logs",
        Migrate: func(tx *gorm.DB) error {
            return tx.Exec(`ALTER TABLE logs ADD COLUMN IF NOT EXISTS source TEXT;`).Error
        },
        Rollback: func(tx *gorm.DB) error {
            return tx.Exec(`ALTER TABLE logs DROP COLUMN IF EXISTS source;`).Error
        },
    },
})
```

---

## what about production?

Running migrations at application startup works fine for development and small deployments. But in production, especially with multiple instances of your app, it can cause problems:

- two instances start simultaneously and both try to apply the same migration — race condition
- a long-running migration blocks your app from starting
- if the migration fails, your app fails to start

The better approach for production is a dedicated migration runner — a separate process that runs migrations once during the deploy pipeline, before your app instances start:

```
deploy pipeline:
  1. build app binary
  2. run migrations (one process, once)
  3. start app instances
```

We have a `task:` left in the codebase for this — a `cmd/migrate` binary would be the right solution. For now, running at startup is fine for development and learning purposes.

---

## common pitfalls

**changing an existing migration after it's been applied**
Don't. gormigrate recorded that migration ID as done. Your changes will never run. Add a new migration instead.

**forgetting to handle the rollback**
If you ever need to roll back, an empty `Rollback` function will do nothing. Write meaningful rollbacks — future you will appreciate it.

**creating Postgres objects in the wrong order**
Extensions before types. Types before tables that use them. Tables before foreign keys that reference them. Get the order wrong and your migration errors out.

**altering a column with a default value to an enum type**
Postgres can't automatically cast a default value when changing column types. Drop the default first, alter the column, then re-add the default. Or do it in a controlled migration that handles existing data explicitly.

---

## alright, that's migrations

You know what they are, why they matter, why AutoMigrate isn't enough for production, and how our gormigrate setup works. Next up — let's walk through the actual codebase and see how all of this comes together.

→ next: [07 — Codebase Walkthrough](./07-codebase.md)

---

*MIT License — [devsimsek.mit-license.org](https://devsimsek.mit-license.org)*