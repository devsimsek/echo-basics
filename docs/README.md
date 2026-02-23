# docs

> welcome. if you're here you're either learning Go, debugging something, or both. either way, you're in the right place.

this directory contains the full documentation for the echo-basics project — a remote logging REST-ish API built as an in-person Go tutorial. we cover everything from the ground up: Go itself, frameworks, Echo, databases, GORM, migrations, and a full walkthrough of the codebase.

read them in order if you're new. jump around if you know what you're looking for.

---

## table of contents

| # | topic | what's in it |
|---|---|---|
| [01](./01-introduction-to-go.md) | Introduction to Go | what Go is, why it exists, the basics you need to know |
| [02](./02-frameworks.md) | Frameworks | what a framework is, why you'd use one, what's out there in Go |
| [03](./03-echo.md) | Echo | Echo v5 — routing, handlers, middleware, context injection |
| [04](./04-databases.md) | Databases | relational databases, Postgres, SQL basics, transactions |
| [05](./05-gorm.md) | GORM | the ORM we use — models, CRUD, transactions, enums, error handling |
| [06](./06-migrations.md) | Migrations | why AutoMigrate isn't enough, how gormigrate works, our migration setup |
| [07](./07-codebase.md) | Codebase Walkthrough | every file in this project explained, top to bottom |
| [08](./08-api.md) | API Reference | every endpoint, every parameter, every response — with curl examples |

---

## how to use these docs

if you're completely new to Go — start at [01](./01-introduction-to-go.md) and read straight through. each doc links to the next one at the bottom.

if you already know Go and just want to understand the project — jump to [07 — Codebase Walkthrough](./07-codebase.md).

if you just want to know what the API does — go to [08 — API Reference](./08-api.md).

---

## the project in one paragraph

echo-basics is a remote logging API. you send it log entries (a flag and a message), it stores them in Postgres, and you can fetch or delete them via a REST-ish API. it's built with Echo v5 for HTTP, GORM for database access, gormigrate for versioned migrations, and Postgres as the database. there are intentional bad practices in the codebase — find them, fix them, open a PR.

---

## contributing

spot something wrong? think something could be explained better? open a PR. the docs live alongside the code so they can be updated together.

---

*MIT License — [devsimsek.mit-license.org](https://devsimsek.mit-license.org)*