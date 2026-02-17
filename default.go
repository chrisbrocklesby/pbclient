package pbclient

import "sync"

var (
	defaultMu     sync.RWMutex
	defaultClient *Client
)

func SetDefault(c *Client) {
	defaultMu.Lock()
	defaultClient = c
	defaultMu.Unlock()
}

func Default() *Client {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultClient
}

func mustDefault() *Client {
	c := Default()
	if c == nil {
		panic("pbclient: default client not initialized. Call pbclient.New(...) or pbclient.SetDefault(pbclient.NewClient(...))")
	}
	return c
}

// New initializes and sets the package-level default client.
func New(cfg Config) (*Client, error) {
	c, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	SetDefault(c)
	return c, nil
}
