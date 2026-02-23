# 02 — Frameworks

> you could build everything from scratch. the standard library lets you. but should you?

---

## what is a framework?

A framework is a pre-written collection of code that gives you a structure to build on top of. Instead of solving the same problems every single project (routing, middleware, request parsing, response formatting...), someone already solved them and packaged it up. You use that, and focus on your actual problem.

Think of it like building a house. You *could* make your own bricks, smelt your own metal, cut your own wood. Or you could just buy those things and focus on building the house. Frameworks are the pre-made bricks.

---

## do you always need one?

No. And in Go especially, the standard library is good enough that a lot of people skip frameworks entirely for small services.

For example, this is a perfectly valid Go HTTP handler with zero dependencies:

```go
package main

import (
    "encoding/json"
    "net/http"
)

func main() {
    http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })
    http.ListenAndServe(":8080", nil)
}
```

That works. That's production-ready for simple cases. No framework needed.

But the moment you start adding:
- URL path parameters (e.g. `/users/:id`)
- middleware chains (logging, auth, rate limiting...)
- request binding and validation
- grouped routes

...the standard library starts to get awkward. You end up writing the same boilerplate over and over. That's when a framework earns its place.

---

## what does a web framework typically give you?

Most Go web frameworks give you some or all of these:

| feature | what it solves |
|---|---|
| **router** | match URL patterns to handler functions, including path params |
| **middleware** | run code before/after every request (logging, auth, etc.) |
| **context** | carry request-scoped data through the handler chain |
| **binding** | parse JSON/form/query params into Go structs automatically |
| **response helpers** | write JSON, HTML, status codes without boilerplate |
| **group routing** | prefix a set of routes (e.g. `/api/v1/...`) cleanly |

---

## the popular Go web frameworks

Here's what's out there:

**net/http (standard library)**
- no dependencies, ships with Go
- great for simple stuff
- gets verbose quickly for complex routing

**Gin**
- the most popular Go web framework by stars
- fast, well-documented
- v1 has been around forever, stable

**Echo**
- what we're using in this project
- clean API, great middleware support, very fast
- v5 is a big improvement over v4

**Fiber**
- inspired by Express.js, built on top of fasthttp instead of net/http
- very fast, familiar if you've used Node.js
- not compatible with the standard `net/http` ecosystem (a trade-off worth knowing)

**Chi**
- lightweight, composable, 100% compatible with net/http
- great if you want something minimal

---

## why Echo for this project?

A few reasons:

1. **it's clean.** the API feels natural, handlers are easy to read, and middleware is straightforward to write.
2. **it's fast.** Echo uses a radix tree router which is efficient for route matching.
3. **middleware is first-class.** logging, rate limiting, security headers, CORS — all available out of the box.
4. **context is easy to extend.** we can store our database connection in the request context and access it from every handler without globals.
5. **v5 is modern.** Echo v5 cleaned up a lot of the rough edges from v4.

Could we have used Gin? Sure. Chi? Also fine. But Echo fits this project well and it's worth knowing.

---

## the middleware concept

This is worth spending a minute on because you'll see it everywhere.

Middleware is a function that wraps a handler. It runs code *before* and/or *after* the actual handler. Multiple middleware functions are chained together — each one calls the next.

Conceptually:

```
Request → Middleware A → Middleware B → Handler → Middleware B → Middleware A → Response
```

In Go/Echo it looks like this:

```go
func MyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
        // runs before the handler
        fmt.Println("before")

        err := next(c) // call the next handler

        // runs after the handler
        fmt.Println("after")

        return err
    }
}
```

In our project we use middleware for:
- **request logging** — every request gets logged
- **secure headers** — XSS protection, content type sniffing prevention, clickjacking protection
- **rate limiting** — throttle requests per IP
- **app context injection** — attach the database to every request context

---

## the request lifecycle in Echo

When a request comes in, here's what happens in our app:

```
Incoming Request
    ↓
RequestLogger middleware  (logs the request)
    ↓
Secure middleware         (adds security headers)
    ↓
RateLimiter middleware    (checks rate limit)
    ↓
AppContext middleware      (injects DB into context)
    ↓
Route Handler             (does the actual work)
    ↓
Response sent back
```

Each middleware can short-circuit the chain — for example, if the rate limiter decides you've sent too many requests, it returns a 429 response and the handler never runs.

---

## alright, that's frameworks

You know what they are, why they exist, and why we picked Echo. Next up — let's actually look at Echo and how we're using it.

→ next: [03 — Echo](./03-echo.md)

---

*MIT License — [devsimsek.mit-license.org](https://devsimsek.mit-license.org)*