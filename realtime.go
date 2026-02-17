package pbclient

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

type RealtimeEvent struct {
	Event string
	Data  json.RawMessage
}

type Realtime struct {
	c *Client

	mu            sync.RWMutex
	clientId      string
	subscriptions []string

	Events chan RealtimeEvent

	cancel context.CancelFunc
	wg     sync.WaitGroup

	readyCh chan struct{}
	errOnce sync.Once
}

func NewRealtime(params ...*Client) *Realtime {
	var c *Client
	if len(params) > 0 {
		c = params[0]
	}
	if c == nil {
		c = mustDefault()
	}
	return &Realtime{c: c, Events: make(chan RealtimeEvent, 128), readyCh: make(chan struct{})}
}

func (rt *Realtime) ClientID() string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.clientId
}

// Connect starts the realtime loop. It will reconnect with backoff and resubscribe automatically.
func (rt *Realtime) Connect() error {
	ctx, cancel := context.WithCancel(rt.c.ctx)
	rt.cancel = cancel

	rt.wg.Add(1)
	go func() {
		defer rt.wg.Done()
		rt.loop(ctx)
	}()

	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()

	for {
		rt.mu.RLock()
		cid := rt.clientId
		rt.mu.RUnlock()
		if cid != "" {
			return nil
		}
		select {
		case <-rt.readyCh:
			continue
		case <-deadline.C:
			return fmt.Errorf("pbclient: realtime connect timeout")
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(25 * time.Millisecond)
		}
	}
}

func (rt *Realtime) Subscribe(subscriptions ...string) error {
	rt.mu.Lock()
	rt.subscriptions = append([]string(nil), subscriptions...)
	cid := rt.clientId
	rt.mu.Unlock()

	if cid == "" {
		return nil
	}
	return rt.applySubscriptions(cid, subscriptions)
}

func (rt *Realtime) Close() {
	if rt.cancel != nil {
		rt.cancel()
	}
	rt.wg.Wait()
	close(rt.Events)
}

func (rt *Realtime) loop(ctx context.Context) {
	backoff := 200 * time.Millisecond
	maxBackoff := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cid, err := rt.connectOnce(ctx)
		if err != nil {
			if !isTransient(err) {
				rt.Events <- RealtimeEvent{Event: "PB_ERROR", Data: json.RawMessage(strconvJSON(err.Error()))}
				return
			}
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		backoff = 200 * time.Millisecond

		rt.mu.RLock()
		subs := append([]string(nil), rt.subscriptions...)
		rt.mu.RUnlock()

		if len(subs) > 0 {
			_ = rt.applySubscriptions(cid, subs)
		}
	}
}

func (rt *Realtime) connectOnce(ctx context.Context) (string, error) {
	u, err := url.Parse(rt.c.baseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "/api/realtime")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := rt.c.http.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		return "", parseAPIError(resp.StatusCode, b)
	}

	defer resp.Body.Close()

	br := bufio.NewReader(resp.Body)
	cid := ""
	for {
		ev, data, err := readSSEEvent(br)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return cid, fmt.Errorf("pbclient: realtime disconnected")
			}
			return cid, err
		}

		if ev == "" && len(data) == 0 {
			continue
		}

		if ev == "PB_CONNECT" && cid == "" {
			var tmp struct {
				ClientID string `json:"clientId"`
			}
			_ = json.Unmarshal(data, &tmp)
			if tmp.ClientID != "" {
				cid = tmp.ClientID
				rt.mu.Lock()
				rt.clientId = cid
				rt.mu.Unlock()
				rt.errOnce.Do(func() { close(rt.readyCh) })
			}
		}

		select {
		case rt.Events <- RealtimeEvent{Event: ev, Data: append(json.RawMessage(nil), data...)}:
		case <-ctx.Done():
			return cid, ctx.Err()
		}
	}
}

func (rt *Realtime) applySubscriptions(clientId string, subscriptions []string) error {
	body := map[string]any{"clientId": clientId, "subscriptions": subscriptions}
	return rt.c.doJSON(http.MethodPost, "/api/realtime", "", body, nil)
}

func readSSEEvent(r *bufio.Reader) (event string, data []byte, err error) {
	var ev string
	var dd []byte
	for {
		line, e := r.ReadString('\n')
		if e != nil {
			return ev, dd, e
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			return ev, dd, nil
		}
		if strings.HasPrefix(line, "event:") {
			ev = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			part := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if len(dd) > 0 {
				dd = append(dd, '\n')
			}
			dd = append(dd, part...)
			continue
		}
	}
}

func strconvJSON(s string) []byte {
	b, _ := json.Marshal(map[string]string{"error": s})
	return b
}
