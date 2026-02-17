package pbclient

import (
	"fmt"
	"net/http"
	"time"
)

// Health checks /api/health.
func (c *Client) Health() error {
	var out map[string]any
	return c.doJSON(http.MethodGet, "/api/health", "", nil, &out)
}

// WaitReady polls /api/health until it succeeds or the timeout elapses.
func (c *Client) WaitReady(timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	deadline := time.Now().Add(timeout)
	backoff := 100 * time.Millisecond
	for {
		if err := c.Health(); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("pbclient: server not ready within %s", timeout)
		}
		time.Sleep(backoff)
		if backoff < 750*time.Millisecond {
			backoff *= 2
		}
	}
}
