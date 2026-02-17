package pbclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

func (c *Client) doJSON(method, endpoint, rawQuery string, in any, out any) error {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, endpoint)
	if rawQuery != "" {
		u.RawQuery = rawQuery
	}

	var payload []byte
	if in != nil {
		payload, err = json.Marshal(in)
		if err != nil {
			return err
		}
	}

	attempts := 1
	if method == http.MethodGet {
		attempts = 3
	}
	backoff := 150 * time.Millisecond

	for i := 0; i < attempts; i++ {
		var body io.Reader
		if payload != nil {
			body = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(c.ctx, method, u.String(), body)
		if err != nil {
			return err
		}
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		if tok := c.Token(); tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			if method == http.MethodGet && i < attempts-1 && isTransient(err) {
				if c.logger != nil {
					c.logger.Printf("pbclient: retry GET %s (%v)", u.String(), err)
				}
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return err
		}

		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()

		if resp.StatusCode >= 500 && resp.StatusCode <= 599 && method == http.MethodGet && i < attempts-1 {
			if c.logger != nil {
				c.logger.Printf("pbclient: retry GET %s (http %d)", u.String(), resp.StatusCode)
			}
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return parseAPIError(resp.StatusCode, b)
		}

		if out == nil {
			return nil
		}
		return json.Unmarshal(b, out)
	}

	return fmt.Errorf("pbclient: request failed after retries")
}
