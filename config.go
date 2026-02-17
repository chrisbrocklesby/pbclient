package pbclient

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Config struct {
	BaseURL string
	Timeout time.Duration
	HTTP    *http.Client

	UserEmail      string
	UserPassword   string
	UserCollection string

	AdminEmail    string
	AdminPassword string

	SuperEmail    string
	SuperPassword string

	Logger *log.Logger
}

type Client struct {
	baseURL string
	http    *http.Client
	ctx     context.Context

	mu    sync.RWMutex
	token string

	logger *log.Logger
}

func normalizeBaseURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "http://127.0.0.1:8090"
	}
	return strings.TrimRight(s, "/")
}

func NewClient(cfg Config) (*Client, error) {
	base := normalizeBaseURL(cfg.BaseURL)
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	hc := cfg.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: timeout}
	}

	c := &Client{
		baseURL: base,
		http:    hc,
		ctx:     context.Background(),
		logger:  cfg.Logger,
	}

	if err := c.LoginFromConfig(cfg); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) WithContext(ctx context.Context) *Client {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx
	return c
}

func (c *Client) SetToken(token string) {
	c.mu.Lock()
	c.token = token
	c.mu.Unlock()
}

func (c *Client) Token() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}
