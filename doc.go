// Package pbclient is a lightweight PocketBase client built on net/http.
//
// Usage:
//
//	// Explicit client
//	c, _ := pbclient.NewClient(cfg)
//	posts := pbclient.Collection[Post]("posts", c)
//
//	// Package default client
//	_, _ = pbclient.New(cfg)
//	posts := pbclient.Collection[Post]("posts")
//
// Features:
//   - Typed collections with generics (Create/Get/Update/Delete/List/First)
//   - Batch helper for /api/batch with shared auth context
//   - Realtime SSE client with reconnect + resubscribe and a buffered Events channel
//   - Auth helpers for users/admins/superusers with token storage
//   - GET retries on transient failures; Health and WaitReady utilities
//
// Keep credentials in Config; prefer explicit clients for services, with the default client for quick scripts.
package pbclient
