# 05 — GORM

> writing raw SQL everywhere gets old fast. GORM is how we talk to the database from Go without losing our minds.

---

## what is GORM?

GORM is an ORM — Object Relational Mapper. It maps your Go structs to database tables and gives you a clean API to create, read, update and delete records without writing SQL by hand.

Does that mean you never write SQL? No. Sometimes raw SQL is the right tool. But for 90% of your day-to-day operations — GORM handles it cleanly and you'll thank it later.

We're using **GORM v2** (specifically v1.31.x) with the **Postgres driver**.

---

## installing GORM

```sh
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

---

## connecting to the database

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

db, err := gorm.Open(postgres.New(postgres.Config{
    DSN:                  "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
    PreferSimpleProtocol: true,
}), &gorm.Config{})

if err != nil {
    panic("connection failed: " + err.Error())
}
```

`PreferSimpleProtocol: true` disables implicit prepared statements. Fine for development. In production, you might want to set it to `false` for better performance with repeated queries.

In our project, this lives in `modules/db.go` and the DSN comes from the environment — never hardcoded:

```go
// modules/db.go
func InitDB() *gorm.DB {
    if os.Getenv("DSN") == "" {
        panic("DSN environment variable is not set")
    }

    db, err := gorm.Open(postgres.New(postgres.Config{
        DSN:                  os.Getenv("DSN"),
        PreferSimpleProtocol: true,
    }), &gorm.Config{})

    if err != nil {
        panic("There exist a connection error, check logs manually: " + err.Error())
    }

    return db
}
```

---

## defining models

A GORM model is a Go struct with some extra tags. GORM reads those tags and knows how to map the struct to a table.

```go
type Log struct {
    ID        uuid.UUID `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
    Flag      FlagEnum  `gorm:"type:log_flag"`
    Message   string    `gorm:"type:text;not null"`
    Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
```

Let's break down those struct tags:

| tag | what it does |
|---|---|
| `gorm:"primarykey"` | marks this field as the primary key |
| `gorm:"type:uuid"` | tells GORM the DB column type is UUID |
| `gorm:"default:uuid_generate_v4()"` | the DB will generate the value using this function |
| `gorm:"type:log_flag"` | tells GORM to use our custom Postgres enum type |
| `gorm:"type:text;not null"` | TEXT column, cannot be null |
| `gorm:"default:CURRENT_TIMESTAMP"` | the DB fills this in automatically |

Notice we're using a custom `FlagEnum` type — not a plain `string`. That's intentional. It gives us type safety in Go and maps to the Postgres `log_flag` enum in the database. More on enums in a moment.

---

## CRUD — the basics

GORM covers the four fundamental database operations: Create, Read, Update, Delete.

### create

```go
log := models.Log{
    Flag:    models.InfoFlag,
    Message: "something happened",
}

result := db.Create(&log)
if result.Error != nil {
    // handle error
}

// after create, log.ID and log.Timestamp are populated by the DB
fmt.Println(log.ID) // uuid filled in
```

Pass a pointer. GORM will fill in the DB-generated fields (ID, Timestamp) after the insert.

### read — find all

```go
var logs []models.Log
db.Find(&logs)
```

### read — find one by condition

```go
var log models.Log
db.First(&log, "id = ?", someUUID)
```

`First` returns the first matching record ordered by primary key. If nothing is found, `db.Error` will be `gorm.ErrRecordNotFound`.

### read — with conditions

```go
var logs []models.Log
db.Where("flag = ?", "info").Find(&logs)
```

Always use parameterized queries like the `?` placeholder above. Never string-concatenate user input into queries — that's a SQL injection vulnerability.

```go
// DO THIS
db.Where("flag = ?", userInput).Find(&logs)

// NEVER DO THIS
db.Where("flag = '" + userInput + "'").Find(&logs) // SQL injection waiting to happen
```

### read — order and limit

```go
var log models.Log
db.Where("timestamp <= ?", ts).Order("timestamp desc").First(&log)
```

### update

```go
db.Model(&log).Update("message", "updated message")

// or update multiple fields
db.Model(&log).Updates(models.Log{Message: "updated"})
```

### delete

```go
db.Delete(&log)
```

If your model has a `DeletedAt` field (GORM's soft delete), GORM will set that field instead of actually deleting the row. Our model doesn't use soft delete — we do a hard delete.

---

## using context

Always pass the request context to GORM operations. This ensures that if the HTTP request is cancelled (the client disconnected, request timed out), the DB operation is cancelled too — saving resources:

```go
db.WithContext(c.Request().Context()).Create(&log)
db.WithContext(c.Request().Context()).Find(&logs)
db.WithContext(c.Request().Context()).Delete(&log)
```

We do this in every handler in this project. Don't skip it.

---

## transactions

Wrap multiple operations in a transaction when they need to succeed or fail together:

```go
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&log).Error; err != nil {
        return err // rolls back automatically
    }
    // add more operations here if needed
    return nil // commits
})
```

If the function returns an error — rollback. If it returns nil — commit. Simple as that.

We use this in `CreateLog`:

```go
if err := db.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&log).Error; err != nil {
        return err
    }
    return nil
}); err != nil {
    return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
```

---

## error handling

Every GORM operation returns a `*gorm.DB` result. Check `.Error` on it:

```go
result := db.Create(&log)
if result.Error != nil {
    // something went wrong
}
```

Or chain it directly:

```go
if err := db.Create(&log).Error; err != nil {
    // handle
}
```

For not-found errors specifically, use `errors.Is`:

```go
import "errors"

res := db.First(&log, "id = ?", id)
if res.Error != nil {
    if errors.Is(res.Error, gorm.ErrRecordNotFound) {
        return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "not found"})
    }
    return c.JSON(http.StatusInternalServerError, map[string]string{"error": res.Error.Error()})
}
```

---

## enums with GORM

GORM doesn't create Postgres ENUM types for you — that's your job, in a migration. But once the enum exists, you can map a Go type to it.

We define a typed alias:

```go
// models/enums.go
type FlagEnum string

const (
    LogFlag   FlagEnum = "log"
    DebugFlag FlagEnum = "debug"
    InfoFlag  FlagEnum = "info"
    WarnFlag  FlagEnum = "warn"
    ErrorFlag FlagEnum = "error"
    TraceFlag FlagEnum = "trace"
)
```

And use it in the model with the GORM tag pointing to the Postgres enum type:

```go
Flag FlagEnum `gorm:"type:log_flag"`
```

When inserting, we validate and normalize the flag in the handler before it ever touches the DB:

```go
flagStr := strings.ToLower(strings.TrimSpace(string(payload.Flag)))

allowed := map[string]struct{}{
    string(models.LogFlag):   {},
    string(models.DebugFlag): {},
    string(models.InfoFlag):  {},
    string(models.WarnFlag):  {},
    string(models.ErrorFlag): {},
    string(models.TraceFlag): {},
}

if _, ok := allowed[flagStr]; !ok {
    return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "invalid flag value"})
}
```

The database enforces valid enum values. The application normalizes and validates before inserting. Both layers doing their job.

---

## what about AutoMigrate?

GORM has an `AutoMigrate` function that creates or alters tables to match your models:

```go
db.AutoMigrate(&models.Log{})
```

It's useful for quick prototyping. But there are serious limitations:

- It only adds columns, never removes them.
- It can't create custom Postgres types like ENUM types.
- It can try to ALTER existing columns in ways that fail (for example, changing a `varchar` column to an enum type — Postgres won't do that automatically).
- It has no version history — you can't see what changed or roll it back.

**We removed AutoMigrate from this project.** We use explicit, versioned migrations instead. See the next section for how that works.

---

## the gorm.DB instance — keep one, share it everywhere

You open the database connection once and reuse it everywhere. The `*gorm.DB` instance manages a connection pool internally — it's safe to use concurrently.

In our project, `InitDB()` returns one `*gorm.DB`. That instance is stored in `AppContext` and injected into every request via middleware. Every handler accesses the same instance:

```go
db := c.Get("app").(*models.AppContext).DB
```

Never open a new connection per request. That's expensive and will exhaust your Postgres `max_connections` limit in no time.

---

## alright, that's GORM

You know how to connect, define models, run CRUD operations, use transactions, handle errors, and deal with custom types. Next up — migrations. Because the database schema doesn't just appear out of thin air.

→ next: [06 — Migrations](./06-migrations.md)

---

*MIT License — [devsimsek.mit-license.org](https://devsimsek.mit-license.org)*