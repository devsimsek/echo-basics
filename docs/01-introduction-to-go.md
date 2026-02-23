# 01 — Introduction to Go

> before we write a single line of code, let's talk about what Go actually is and why we picked it for this project.

---

## what is Go?

Go (or Golang if you're googling it) is a statically typed, compiled programming language designed at Google in 2007 and released publicly in 2009. It was built by people who were tired of the tradeoffs in other languages — C++ was too slow to compile, Java was too verbose, Python was too slow to run. So they made Go.

The pitch is simple: **fast to write, fast to compile, fast to run.** And honestly? It delivers.

---

## why Go and not something else?

Fair question. Here's the honest answer:

- **Python** is great for scripting and data stuff. But when you need to handle thousands of concurrent HTTP requests, it starts sweating.
- **Node.js** is fine but the ecosystem is a nightmare and `node_modules` is a war crime.
- **Java** works, but you'll be writing `AbstractBeanFactoryManagerServiceImpl` before you know it.
- **Rust** is incredible but you'll spend more time arguing with the borrow checker than writing your app.

Go sits in a sweet spot. It's simple enough to read at a glance, fast enough to run in production, and the standard library is so good you barely need third-party packages for most things.

For web APIs specifically — Go is genuinely one of the best choices you can make right now.

---

## the basics — things you need to know

### types

Go is statically typed. Every variable has a type and it's known at compile time. No surprises.

```go
var name string = "devsimsek"
age := 25 // Go infers the type as int
```

The `:=` shorthand is the one you'll use most of the time inside functions. Outside functions (package level), you use `var`.

### functions

Functions are first-class citizens in Go. They look like this:

```go
func greet(name string) string {
    return "hello, " + name
}
```

Go functions can return multiple values — and this is used *everywhere*, especially for error handling:

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("cannot divide by zero")
    }
    return a / b, nil
}
```

### error handling

This is the thing that trips people up coming from other languages. Go doesn't have exceptions. Instead, functions return errors as values and **you are expected to check them**.

```go
result, err := divide(10, 0)
if err != nil {
    // handle it. don't ignore it.
    log.Fatal(err)
}
```

It feels verbose at first. You'll get used to it, and eventually you'll appreciate it — errors are explicit, they're right there in your face, you can't accidentally miss them.

### structs

Go doesn't have classes. It has structs. A struct is just a collection of fields:

```go
type User struct {
    ID    int
    Name  string
    Email string
}
```

You add behaviour to structs using methods:

```go
func (u User) Greet() string {
    return "hey, i'm " + u.Name
}
```

### interfaces

Interfaces define behaviour, not data. Any type that implements the methods of an interface satisfies it — no explicit declaration needed.

```go
type Logger interface {
    Log(message string)
}

type ConsoleLogger struct{}

func (c ConsoleLogger) Log(message string) {
    fmt.Println(message)
}

// ConsoleLogger satisfies Logger — no "implements Logger" keyword needed
```

This is called **implicit interface satisfaction** and it's one of Go's best features.

### goroutines and concurrency

Go has built-in concurrency via goroutines. A goroutine is a lightweight thread managed by the Go runtime. You start one with the `go` keyword:

```go
go func() {
    fmt.Println("this runs concurrently")
}()
```

Goroutines communicate via channels. We're not going deep on this in this tutorial, but know it's there and it's powerful.

### packages and modules

Go code is organised into packages. Every `.go` file starts with a `package` declaration:

```go
package main // this is the entry point package
package models // this is a library package
```

Modules are how Go manages dependencies. You'll see a `go.mod` file at the root of every Go project — that's your module definition. It specifies the module name and the Go version:

```go
module go.smsk.dev/pkgs/basics/echo-basics

go 1.25.0
```

To add a dependency: `go get github.com/some/package`  
To tidy up unused ones: `go mod tidy`

---

## the Go toolchain

You'll be using these commands constantly:

| command | what it does |
|---|---|
| `go run .` | compile and run the current package |
| `go build .` | compile to a binary |
| `go test ./...` | run all tests |
| `go mod tidy` | clean up go.mod and go.sum |
| `go vet ./...` | static analysis — catches common mistakes |
| `go fmt ./...` | format all code |

---

## what Go is not

- It's not object-oriented in the traditional sense. No inheritance. Get over it — composition is better anyway.
- It's not a scripting language. It compiles. That's a feature.
- It doesn't have generics in the traditional sense... well, it does now (since 1.18), but you won't need them for most things.

---

## the thing that makes Go feel different

The Go standard library is *really* good. Before reaching for a third-party package, check if `net/http`, `encoding/json`, `os`, `strings`, `time`, or `fmt` already does what you need. Often they do.

```go
// a full HTTP server in the standard library — no dependencies
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "hello world")
    })
    http.ListenAndServe(":8080", nil)
}
```

That's it. That runs. That serves HTTP requests. No npm install, no pip install, nothing.

---

## alright, enough theory

You now know enough Go to not be completely lost. In the next section we'll talk about frameworks — what they are, why they exist, and why we picked Echo for this project.

→ next: [02 — Frameworks](./02-frameworks.md)

---

*MIT License — [devsimsek.mit-license.org](https://devsimsek.mit-license.org)*