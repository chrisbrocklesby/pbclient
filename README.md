# PocketBase Go Client (pbclient) v1.0.1

Lightweight, production-friendly Go client for PocketBase using only `net/http`, typed collections via generics, batch support, and SSE realtime with automatic reconnect.

## Install

Replace `github.com/chrisbrocklesby/pbclient` with your repo path after publishing:

```bash
go get github.com/chrisbrocklesby/pbclient
```

Go 1.20+ recommended.

## Quickstart

```go
package main

import (
    "log"

    pbclient "github.com/chrisbrocklesby/pbclient"
)

type Post struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}

func main() {
    if _, err := pbclient.New(pbclient.Config{BaseURL: "http://127.0.0.1:8090", SuperEmail: "user@example.com", SuperPassword: "password1234", Logger: log.Default()}); err != nil {
        log.Fatal(err)
    }

    posts := pbclient.Collection[Post]("posts")

    // Create
    _, _ = posts.Create(map[string]any{"title": "Hello", "content": "World"})

    // List
    res, _ := posts.List("perPage=10")
    log.Printf("fetched %d posts", len(res.Items))
}
```

## Explicit client (recommended for prod)

```go
c, err := pbclient.NewClient(pbclient.Config{BaseURL: "https://pb.example.com", SuperEmail: "admin@example.com", SuperPassword: "secret"})
if err != nil { /* handle */ }
posts := pbclient.Collection[Post]("posts", c)
first, err := posts.First("filter=title~'Hello'")
```

## Batch requests

```go
batch := pbclient.NewBatch() // or pbclient.NewBatch(c) to use a specific client
bc := batch.Collection("posts")
bc.Create(map[string]any{"title": "a"})
bc.Create(map[string]any{"title": "b"})
resp, err := batch.Send()
```

## Realtime (SSE)

```go
rt := pbclient.NewRealtime() // or pbclient.NewRealtime(c) to use a specific client
if err := rt.Subscribe("posts"); err != nil { /* handle */ }
if err := rt.Connect(); err != nil { /* handle */ }
for ev := range rt.Events {
    log.Printf("event=%s data=%s", ev.Event, string(ev.Data))
}
```

## Health + readiness

```go
if err := c.WaitReady(10 * time.Second); err != nil { /* handle */ }
```

## Notes

- Retries on GETs for transient network/server errors.
- `Collection[T]` helpers use the package default client; pass a client as the optional second argument to override.
- `Realtime` automatically reconnects and resubscribes (defaults to the package client if none is passed).

## Example app

`main.go` shows a minimal HTTP demo that lists and mutates `posts` and `users` collections using the package-level helpers. Update credentials/base URL as needed.
