# 03 — Echo

> alright, let's get into Echo. this is the framework we're using to handle HTTP in this project and honestly, once you get the hang of it, you'll wonder how you lived without it.

---

## what is Echo?

Echo is a high-performance, minimalist web framework for Go. It sits on top of Go's standard `net/http` package and gives you the things the standard library is awkward about — routing, middleware, context, binding, and response helpers.

We're using **Echo v5** in this project. v5 is a significant rewrite from v4. The biggest change you'll notice is that `echo.Context` is now a **struct** (not an interface like it was in v4). That changes how you extend it, which we'll get to.

---

## setting up Echo

Here's the minimum viable Echo server:

```go
package main

import (
    "net/http"
    "github.com/labstack/echo/v5"
)

func main() {
    e := echo.New()

    e.GET("/", func(c *echo.Context) error {
        return c.String(http.StatusOK, "hello world")
    })

    e.Start(":8080")
}
```

That's it. Run it, hit `localhost:8080`, get a response. Dead simple.

---

## routing

Echo's router is built on a radix tree — it's fast and supports named parameters, wildcards, and route groups.

### basic routes

```go
e.GET("/users", getUsers)
e.POST("/users", createUser)
e.PUT("/users/:id", updateUser)
e.DELETE("/users/:id", deleteUser)
```

### path parameters

Use `:name` in the path and `c.Param("name")` to read it:

```go
e.GET("/users/:id", func(c *echo.Context) error {
    id := c.Param("id")
    return c.String(http.StatusOK, "user id: " + id)
})
```

### query parameters

```go
e.GET("/search", func(c *echo.Context) error {
    query := c.QueryParam("q")
    return c.String(http.StatusOK, "searching for: " + query)
})
```

### route groups

Group routes to share a common prefix or middleware:

```go
api := e.Group("/api")

api.GET("/health", healthHandler)
api.POST("/create", createHandler)
api.GET("/list", listHandler)
```

In our project, all API routes live under `/api`:

```go
api := e.Group("/api")
api.Any("/health", routes.HealthCheck)
api.POST("/create", routes.CreateLog)
api.GET("/list", routes.FetchLogs)
api.GET("/fetch/i/:id", routes.FetchID)
api.GET("/fetch/t/:timestamp", routes.FetchTimestamp)
api.GET("/fetch/f/:flag", routes.FetchFlag)
api.DELETE("/delete/:id", routes.DeleteLog)
```

---

## handlers

Every handler in Echo has the same signature:

```go
func(c *echo.Context) error
```

That's it. You get a context, you return an error (or nil if everything went fine).

### responding with JSON

```go
func MyHandler(c *echo.Context) error {
    return c.JSON(http.StatusOK, map[string]string{
        "message": "Yeppers, seems good.",
    })
}
```

### responding with a string

```go
func MyHandler(c *echo.Context) error {
    return c.String(http.StatusOK, "Hellow from the other side")
}
```

### responding with an error

Return a non-nil error and Echo will handle it. Or return a JSON response manually with a non-200 status code — which is what we do throughout this project:

```go
return c.JSON(http.StatusBadRequest, map[string]interface{}{
    "error": "bad request, i had better expectations from you.",
})
```

---

## binding

Binding is how you parse incoming request data into a Go struct. Echo does this automatically from JSON, form data, or query params.

```go
var payload struct {
    Flag    string `json:"flag"`
    Message string `json:"message"`
}

if err := c.Bind(&payload); err != nil {
    return c.JSON(http.StatusBadRequest, map[string]interface{}{
        "error": "bad request, i had better expectations from you.",
    })
}
```

Pass a pointer to your struct. Echo reads the `Content-Type` header and parses accordingly. For JSON requests, it reads the body. For form requests, it reads form fields. Simple.

---

## middleware

Middleware is a function that wraps a handler. In Echo, a middleware looks like this:

```go
func MyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
        // do something before the handler
        err := next(c)
        // do something after the handler
        return err
    }
}
```

Register it globally with `e.Use()`:

```go
e.Use(MyMiddleware)
```

Or on a specific group:

```go
api := e.Group("/api")
api.Use(MyMiddleware)
```

### built-in middleware we use

**RequestLogger** — logs every incoming request:
```go
e.Use(middleware.RequestLogger())
```

**Secure** — adds security headers (XSS protection, content type sniffing, clickjacking):
```go
e.Use(middleware.Secure())
```

**RateLimiter** — throttles requests. The rate comes from our `LIMIT_RATE` environment variable:
```go
e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(limitRate)))
```

---

## the context — and how we extend it

In Echo v5, `echo.Context` is a struct. Every handler receives a `*echo.Context`. The context carries everything about the current request — path params, query params, the request/response objects, and any custom values you set.

### storing values

```go
c.Set("key", someValue)
```

### reading values

```go
value := c.Get("key")
```

### our AppContext pattern

We need to access the database in every handler. Rather than using a global variable (please don't), we inject the database into every request context using a middleware:

```go
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
        c.Set("app", &models.AppContext{DB: db})
        return next(c)
    }
})
```

`AppContext` is a simple struct that holds our database:

```go
// models/context.go
type AppContext struct {
    DB *gorm.DB
}
```

Then in any handler, we retrieve it like this:

```go
db := c.Get("app").(*models.AppContext).DB
```

The `(*models.AppContext)` part is a type assertion — we're telling Go "I know this `any` value is actually a `*models.AppContext`, treat it as one." If it's not, it panics. Since the middleware always sets it, that's fine.

---

## error handling

If a handler returns an error, Echo catches it. By default it returns a 500 response. You can customise the error handler:

```go
e.HTTPErrorHandler = func(err error, c *echo.Context) {
    // custom error handling
}
```

In our project we prefer returning explicit JSON error responses directly from handlers rather than relying on the global error handler — it's more predictable and gives us full control over the message:

```go
return c.JSON(http.StatusForbidden, map[string]string{
    "error": "EEEEYYYY! You can't do that!",
})
```

---

## starting the server

```go
if err := e.Start(":" + os.Getenv("PORT")); err != nil {
    e.Logger.Error("Failed to start echo application", "error", err)
}
```

We read the port from an environment variable. That's the right way to do it — hardcoding ports is bad practice and will bite you when you deploy.

---

## putting it all together

Here's our full `server.go` setup condensed:

```go
func main() {
    modules.LoadEnv("dev")

    db := modules.InitDB()

    e := echo.New()
    e.Use(middleware.RequestLogger())
    e.Use(middleware.Secure())
    e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(limitRate)))

    migrations.Run(db)

    // inject DB into every request
    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {
            c.Set("app", &models.AppContext{DB: db})
            return next(c)
        }
    })

    api := e.Group("/api")
    api.POST("/create", routes.CreateLog)
    api.GET("/list", routes.FetchLogs)
    // ...

    e.Start(":" + os.Getenv("PORT"))
}
```

Clean, readable, and each piece has a clear job. That's what we're going for.

---

## alright, that's Echo

You know how to set up the server, register routes, write handlers, use middleware, and extend the context. Next up — we need somewhere to store data. Let's talk databases.

→ next: [04 — Databases](./04-databases.md)

---

*MIT License — [devsimsek.mit-license.org](https://devsimsek.mit-license.org)*